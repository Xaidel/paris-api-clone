package ports

import (
	"context"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// RetryFailedTransactionCommand describes one atomic failed-transaction retry request.
type RetryFailedTransactionCommand struct {
	UploadID      valueobjects.UploadID
	TransactionID valueobjects.TransactionID
	TaskName      string
	LastRetriedAt time.Time
}

// TransactionClassificationRetryAttempt describes one persisted retry attempt.
type TransactionClassificationRetryAttempt struct {
	JobID         string
	ParentJobID   string
	RetryCount    int
	LastRetriedAt time.Time
}

// RetryFailedTransactionResult describes one retry creation outcome.
type RetryFailedTransactionResult struct {
	Attempt    *TransactionClassificationRetryAttempt
	Skipped    bool
	SkipReason string
}

// TransactionClassificationRetryRepository persists and creates classification retry attempts.
type TransactionClassificationRetryRepository interface {
	ListFailedByUploadID(ctx context.Context, uploadID valueobjects.UploadID) ([]valueobjects.TransactionID, error)
	RetryFailedTransaction(ctx context.Context, command RetryFailedTransactionCommand) (RetryFailedTransactionResult, error)
}
