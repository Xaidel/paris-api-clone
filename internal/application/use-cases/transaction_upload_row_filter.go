package usecases

import (
	"fmt"
	"strconv"
	"strings"

	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

type transactionUploadRowFilterResult struct {
	EligibleRows [][]string
	SkippedRows  []ports.TransactionUploadSkippedRow
}

func filterUploadRowsByTransactionCount(headers []string, rows [][]string) (transactionUploadRowFilterResult, error) {
	transactionCountIndex, err := findTransactionCountColumnIndex(headers)
	if err != nil {
		return transactionUploadRowFilterResult{}, err
	}

	result := transactionUploadRowFilterResult{
		EligibleRows: make([][]string, 0, len(rows)),
		SkippedRows:  make([]ports.TransactionUploadSkippedRow, 0),
	}

	for rowIndex, row := range rows {
		rowNumber := rowIndex + 2
		value := ""
		if transactionCountIndex < len(row) {
			value = strings.TrimSpace(row[transactionCountIndex])
		}

		parsedValue, err := strconv.Atoi(value)
		if err != nil {
			result.SkippedRows = append(result.SkippedRows, ports.TransactionUploadSkippedRow{RowNumber: rowNumber, Reason: ports.TransactionUploadSkippedRowReasonMalformed})
			continue
		}

		if parsedValue != 1 {
			result.SkippedRows = append(result.SkippedRows, ports.TransactionUploadSkippedRow{RowNumber: rowNumber, Reason: ports.TransactionUploadSkippedRowReasonNotValidTransaction})
			continue
		}

		result.EligibleRows = append(result.EligibleRows, row)
	}

	return result, nil
}

func findTransactionCountColumnIndex(headers []string) (int, error) {
	for index, header := range headers {
		if strings.EqualFold(strings.TrimSpace(header), "No. of Transactions") {
			return index, nil
		}
	}

	return 0, fmt.Errorf("required transaction column is missing: No. of Transactions")
}
