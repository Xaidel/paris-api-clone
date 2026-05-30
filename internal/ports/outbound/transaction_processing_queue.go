package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionProcessingQueue enqueues transactions for asynchronous processing.
type TransactionProcessingQueue interface {
	Enqueue(ctx context.Context, taskName string, transactionID valueobjects.TransactionID) error
}
