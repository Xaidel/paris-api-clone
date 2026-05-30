package valueobjects

import "slices"

// TransactionFileValidationError describes a schema validation failure.
type TransactionFileValidationError struct {
	code        string
	message     string
	rowNumber   int
	columnName  string
	columnIndex int
	value       string
}

// NewTransactionFileValidationError builds a validation error value.
func NewTransactionFileValidationError(code, message string, rowNumber int, columnName string, columnIndex int, value string) TransactionFileValidationError {
	return TransactionFileValidationError{
		code:        code,
		message:     message,
		rowNumber:   rowNumber,
		columnName:  columnName,
		columnIndex: columnIndex,
		value:       value,
	}
}

// Code returns the machine-readable validation error code.
func (e TransactionFileValidationError) Code() string {
	return e.code
}

// Message returns the human-readable validation error message.
func (e TransactionFileValidationError) Message() string {
	return e.message
}

// RowNumber returns the one-based spreadsheet row number.
func (e TransactionFileValidationError) RowNumber() int {
	return e.rowNumber
}

// ColumnName returns the column associated with the error.
func (e TransactionFileValidationError) ColumnName() string {
	return e.columnName
}

// ColumnIndex returns the one-based spreadsheet column index.
func (e TransactionFileValidationError) ColumnIndex() int {
	return e.columnIndex
}

// Value returns the offending cell value when one exists.
func (e TransactionFileValidationError) Value() string {
	return e.value
}

// TransactionFileValidationReport contains the outcome of schema validation.
type TransactionFileValidationReport struct {
	schemaVersion string
	errors        []TransactionFileValidationError
}

// NewTransactionFileValidationReport builds a validation report.
func NewTransactionFileValidationReport(schemaVersion string, errors []TransactionFileValidationError) TransactionFileValidationReport {
	return TransactionFileValidationReport{
		schemaVersion: schemaVersion,
		errors:        slices.Clone(errors),
	}
}

// SchemaVersion returns the schema version applied during validation.
func (r TransactionFileValidationReport) SchemaVersion() string {
	return r.schemaVersion
}

// Valid reports whether any schema validation errors were found.
func (r TransactionFileValidationReport) Valid() bool {
	return len(r.errors) == 0
}

// Errors returns the collected validation errors.
func (r TransactionFileValidationReport) Errors() []TransactionFileValidationError {
	return slices.Clone(r.errors)
}
