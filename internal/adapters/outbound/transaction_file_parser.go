package adapters

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/extrame/xls"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/xuri/excelize/v2"
)

// TransactionFileParser parses CSV, XLS, and XLSX uploads.
type TransactionFileParser struct{}

// NewTransactionFileParser builds a TransactionFileParser.
func NewTransactionFileParser() *TransactionFileParser {
	return &TransactionFileParser{}
}

// Parse reads a tabular transaction file.
func (p *TransactionFileParser) Parse(_ context.Context, command ports.ParseTransactionFileCommand) (ports.ParseTransactionFileResult, error) {
	format := normalizeUploadFormat(command.FileName)
	switch format {
	case "csv":
		return parseCSVFile(command.FileBytes)
	case "xlsx":
		return parseXLSXFile(command.FileBytes)
	case "xls":
		return parseXLSFile(command.FileBytes)
	default:
		return ports.ParseTransactionFileResult{}, fmt.Errorf("unsupported file format %q", filepath.Ext(command.FileName))
	}
}

func normalizeUploadFormat(fileName string) string {
	return strings.TrimPrefix(strings.ToLower(filepath.Ext(strings.TrimSpace(fileName))), ".")
}

func parseCSVFile(fileBytes []byte) (ports.ParseTransactionFileResult, error) {
	reader := csv.NewReader(bytes.NewReader(fileBytes))
	rows, err := reader.ReadAll()
	if err != nil {
		return ports.ParseTransactionFileResult{}, fmt.Errorf("reading csv rows: %w", err)
	}

	headers, dataRows, err := splitParsedRows(rows)
	if err != nil {
		return ports.ParseTransactionFileResult{}, err
	}

	return ports.ParseTransactionFileResult{Format: "csv", Headers: headers, Rows: dataRows}, nil
}

func parseXLSXFile(fileBytes []byte) (ports.ParseTransactionFileResult, error) {
	workbook, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return ports.ParseTransactionFileResult{}, fmt.Errorf("opening xlsx workbook: %w", err)
	}
	defer func() { _ = workbook.Close() }()

	sheets := workbook.GetSheetList()
	if len(sheets) == 0 {
		return ports.ParseTransactionFileResult{}, fmt.Errorf("xlsx workbook does not contain sheets")
	}

	rows, err := workbook.GetRows(sheets[0])
	if err != nil {
		return ports.ParseTransactionFileResult{}, fmt.Errorf("reading xlsx rows: %w", err)
	}

	headers, dataRows, err := splitParsedRows(rows)
	if err != nil {
		return ports.ParseTransactionFileResult{}, err
	}

	return ports.ParseTransactionFileResult{Format: "xlsx", Headers: headers, Rows: dataRows}, nil
}

func parseXLSFile(fileBytes []byte) (ports.ParseTransactionFileResult, error) {
	reader := bytes.NewReader(fileBytes)
	workbook, err := xls.OpenReader(reader, "utf-8")
	if err != nil {
		return ports.ParseTransactionFileResult{}, fmt.Errorf("opening xls workbook: %w", err)
	}

	if workbook.NumSheets() == 0 {
		return ports.ParseTransactionFileResult{}, fmt.Errorf("xls workbook does not contain sheets")
	}

	sheet := workbook.GetSheet(0)
	if sheet == nil {
		return ports.ParseTransactionFileResult{}, fmt.Errorf("xls workbook first sheet is missing")
	}

	rows := make([][]string, 0, int(sheet.MaxRow)+1)
	for rowIndex := 0; rowIndex <= int(sheet.MaxRow); rowIndex++ {
		row := sheet.Row(rowIndex)
		if row == nil {
			rows = append(rows, nil)
			continue
		}

		values := make([]string, 0, row.LastCol())
		for columnIndex := 0; columnIndex < row.LastCol(); columnIndex++ {
			values = append(values, row.Col(columnIndex))
		}

		rows = append(rows, values)
	}

	headers, dataRows, err := splitParsedRows(rows)
	if err != nil {
		return ports.ParseTransactionFileResult{}, err
	}

	return ports.ParseTransactionFileResult{Format: "xls", Headers: headers, Rows: dataRows}, nil
}

func splitParsedRows(rows [][]string) ([]string, [][]string, error) {
	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("uploaded file does not contain rows")
	}

	headers := trimRow(rows[0])
	if len(headers) == 0 {
		return nil, nil, fmt.Errorf("uploaded file header row is empty")
	}

	dataRows := make([][]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		trimmed := trimRow(row)
		if isBlankRow(trimmed) {
			continue
		}

		dataRows = append(dataRows, trimmed)
	}

	dataRows = trimTrailingArtifactRows(headers, dataRows)

	if len(dataRows) == 0 {
		return nil, nil, fmt.Errorf("uploaded file does not contain data rows")
	}

	return headers, dataRows, nil
}

func trimRow(row []string) []string {
	if row == nil {
		return nil
	}

	trimmed := make([]string, len(row))
	for index, value := range row {
		trimmed[index] = strings.TrimSpace(value)
	}

	return trimmed
}

func isBlankRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}

	return true
}

func trimTrailingArtifactRows(headers []string, rows [][]string) [][]string {
	trimmedRows := rows
	for len(trimmedRows) > 0 {
		lastRow := trimmedRows[len(trimmedRows)-1]
		if !isHeaderArtifactRow(headers, lastRow) {
			break
		}

		trimmedRows = trimmedRows[:len(trimmedRows)-1]
	}

	return trimmedRows
}

func isHeaderArtifactRow(headers []string, row []string) bool {
	if isBlankRow(row) {
		return true
	}

	hasNonEmptyValue := false
	for index, value := range row {
		trimmedValue := strings.TrimSpace(value)
		if trimmedValue == "" {
			continue
		}

		hasNonEmptyValue = true
		if index >= len(headers) || trimmedValue != strings.TrimSpace(headers[index]) {
			return false
		}
	}

	return hasNonEmptyValue
}
