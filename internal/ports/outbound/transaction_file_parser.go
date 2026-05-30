package ports

import "context"

// ParseTransactionFileCommand requests parsing of an uploaded file.
type ParseTransactionFileCommand struct {
	FileName  string
	FileBytes []byte
}

// ParseTransactionFileResult returns the parsed header and row table.
type ParseTransactionFileResult struct {
	Format  string
	Headers []string
	Rows    [][]string
}

// TransactionFileParser parses uploaded CSV/XLS/XLSX files.
type TransactionFileParser interface {
	Parse(ctx context.Context, command ParseTransactionFileCommand) (ParseTransactionFileResult, error)
}
