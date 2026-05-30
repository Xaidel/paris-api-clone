package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// ValidateTransactionFileCommand requests validation for a parsed transaction table.
type ValidateTransactionFileCommand struct {
	Headers []string
	Rows    [][]string
}

// TransactionFileSchemaColumn exposes one schema column to inbound adapters.
type TransactionFileSchemaColumn struct {
	Name          string
	Description   string
	DataType      string
	Required      bool
	AllowedValues []string
}

// ValidateTransactionFileResult exposes the validation outcome to inbound adapters.
type ValidateTransactionFileResult struct {
	SchemaVersion string
	Columns       []TransactionFileSchemaColumn
	Valid         bool
	Errors        []outboundports.TransactionFileValidationError
}

// ValidateTransactionFilePort validates a parsed transaction upload table.
type ValidateTransactionFilePort interface {
	Execute(ctx context.Context, command ValidateTransactionFileCommand) (ValidateTransactionFileResult, error)
}
