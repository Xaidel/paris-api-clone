package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// CreateTransactionUploadCommand requests bulk transaction ingestion from a file.
type CreateTransactionUploadCommand struct {
	ClassificationTask string
	FileName           string
	FileBytes          []byte
	ActorUserID        string
	ActorGroupID       string
	ProgressReporter   outboundports.TransactionUploadProgressReporter
}

// CreateTransactionUploadResult exposes the ingestion outcome.
type CreateTransactionUploadResult struct {
	Upload           outboundports.TransactionUploadResult
	ValidationErrors []outboundports.TransactionFileValidationError
	SkippedRows      []outboundports.TransactionUploadSkippedRow
}

// CreateTransactionUploadPort uploads and ingests a transaction file.
type CreateTransactionUploadPort interface {
	Execute(ctx context.Context, command CreateTransactionUploadCommand) (CreateTransactionUploadResult, error)
}
