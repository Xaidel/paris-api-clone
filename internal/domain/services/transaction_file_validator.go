package services

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionFileValidator validates parsed transaction tables against a schema.
type TransactionFileValidator struct {
	schema valueobjects.TransactionFileSchema
}

// NewTransactionFileValidator builds a TransactionFileValidator.
func NewTransactionFileValidator(schema valueobjects.TransactionFileSchema) *TransactionFileValidator {
	return &TransactionFileValidator{schema: schema}
}

// Schema returns the schema used by the validator.
func (v *TransactionFileValidator) Schema() valueobjects.TransactionFileSchema {
	return v.schema
}

// Validate checks headers and rows against the configured schema.
func (v *TransactionFileValidator) Validate(headers []string, rows [][]string) valueobjects.TransactionFileValidationReport {
	columnIndexes := indexHeaders(headers)
	referenceColumnIndex, hasReferenceColumn := findReferenceColumnIndex(columnIndexes)
	columns := v.schema.Columns()
	presentColumns := make(map[string]int, len(columns))
	var validationErrors []valueobjects.TransactionFileValidationError

	for expectedIndex, column := range columns {
		actualIndex, ok := matchColumnIndex(columnIndexes, column)
		if !ok {
			if column.Required() {
				validationErrors = append(validationErrors, valueobjects.NewTransactionFileValidationError(
					"missing_required_column",
					fmt.Sprintf("required column %q is missing", column.Name()),
					1,
					column.Name(),
					expectedIndex+1,
					"",
				))
			}

			continue
		}

		presentColumns[column.Name()] = actualIndex
	}

	for rowIndex, row := range rows {
		if shouldSkipRow(row, referenceColumnIndex, hasReferenceColumn) {
			continue
		}

		rowNumber := rowIndex + 2
		for _, column := range columns {
			columnIndex, ok := presentColumns[column.Name()]
			if !ok {
				continue
			}

			value := cellValue(row, columnIndex)
			if value == "" {
				if column.Required() {
					validationErrors = append(validationErrors, valueobjects.NewTransactionFileValidationError(
						"missing_required_value",
						fmt.Sprintf("row %d column %q is required", rowNumber, column.Name()),
						rowNumber,
						column.Name(),
						columnIndex+1,
						"",
					))
				}

				continue
			}

			if err := validateColumnValue(column, value); err != nil {
				validationErrors = append(validationErrors, valueobjects.NewTransactionFileValidationError(
					validationErrorCode(column),
					err.Error(),
					rowNumber,
					column.Name(),
					columnIndex+1,
					value,
				))
			}
		}
	}

	return valueobjects.NewTransactionFileValidationReport(v.schema.Version(), validationErrors)
}

func indexHeaders(headers []string) map[string]int {
	indexes := make(map[string]int, len(headers))
	for index, header := range headers {
		normalizedHeader := normalizeHeader(header)
		if normalizedHeader == "" {
			continue
		}

		if _, exists := indexes[normalizedHeader]; !exists {
			indexes[normalizedHeader] = index
		}
	}

	return indexes
}

func matchColumnIndex(indexes map[string]int, column valueobjects.TransactionFileColumn) (int, bool) {
	if index, ok := indexes[normalizeHeader(column.Name())]; ok {
		return index, true
	}

	for _, alias := range column.Aliases() {
		if index, ok := indexes[normalizeHeader(alias)]; ok {
			return index, true
		}
	}

	return 0, false
}

func normalizeHeader(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.ReplaceAll(normalized, "-", " ")
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}

func cellValue(row []string, index int) string {
	if index >= len(row) {
		return ""
	}

	return strings.TrimSpace(row[index])
}

func isBlankRow(row []string) bool {
	for _, value := range row {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}

	return true
}

func shouldSkipRow(row []string, referenceColumnIndex int, hasReferenceColumn bool) bool {
	if isBlankRow(row) {
		return true
	}

	if hasReferenceColumn && cellValue(row, referenceColumnIndex) == "" {
		return true
	}

	return false
}

func findReferenceColumnIndex(indexes map[string]int) (int, bool) {
	for _, candidate := range []string{"Reference Number", "Reference No.", "Reference No", "Ref"} {
		if index, ok := indexes[normalizeHeader(candidate)]; ok {
			return index, true
		}
	}

	return 0, false
}

func validateColumnValue(column valueobjects.TransactionFileColumn, value string) error {
	if err := validateDataType(column, value); err != nil {
		return err
	}

	if len(column.AllowedValues()) == 0 {
		return nil
	}

	for _, allowedValue := range column.AllowedValues() {
		if strings.EqualFold(value, allowedValue) {
			return nil
		}
	}

	return fmt.Errorf("row value for column %q must be one of %s", column.Name(), strings.Join(column.AllowedValues(), ", "))
}

func validateDataType(column valueobjects.TransactionFileColumn, value string) error {
	switch column.DataType() {
	case "text":
		return nil
	case "integer":
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("row value for column %q must be an integer", column.Name())
		}
	case "decimal":
		normalizedValue := strings.ReplaceAll(strings.TrimSpace(value), ",", "")
		if _, ok := new(big.Rat).SetString(normalizedValue); !ok {
			return fmt.Errorf("row value for column %q must be a decimal", column.Name())
		}
	default:
		return fmt.Errorf("row value for column %q uses unsupported data type %q", column.Name(), column.DataType())
	}

	return nil
}

func validationErrorCode(column valueobjects.TransactionFileColumn) string {
	if len(column.AllowedValues()) > 0 {
		return "invalid_value"
	}

	return "type_mismatch"
}
