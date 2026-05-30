package adapters

import (
	"context"
	"fmt"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

const createTransactionProcessingQueueEntryQuery = `
INSERT INTO transaction_processing_queue (task_name, transaction_id)
VALUES ($1, $2)
ON CONFLICT (task_name, transaction_id) DO NOTHING
`

// PostgresTransactionProcessingQueue persists transaction ids for background processing.
type PostgresTransactionProcessingQueue struct {
	pool pgxQuerier
}

// NewPostgresTransactionProcessingQueue builds a PostgresTransactionProcessingQueue.
func NewPostgresTransactionProcessingQueue(pool pgxQuerier) *PostgresTransactionProcessingQueue {
	return &PostgresTransactionProcessingQueue{pool: pool}
}

// Enqueue stores a transaction id in the processing queue table.
func (q *PostgresTransactionProcessingQueue) Enqueue(ctx context.Context, taskName string, transactionID valueobjects.TransactionID) error {
	querier := txQuerierFromContext(ctx, q.pool)
	if _, err := querier.Exec(ctx, createTransactionProcessingQueueEntryQuery, taskName, transactionID.String()); err != nil {
		return fmt.Errorf("executing create transaction processing queue entry query: %w", err)
	}

	return nil
}
