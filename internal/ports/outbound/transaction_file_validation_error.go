package ports

// TransactionFileValidationError exposes one structured validation error.
type TransactionFileValidationError struct {
	Code        string
	Message     string
	RowNumber   int
	ColumnName  string
	ColumnIndex int
	Value       string
}
