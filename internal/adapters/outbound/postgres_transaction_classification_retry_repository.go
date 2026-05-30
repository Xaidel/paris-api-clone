package adapters

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/jackc/pgx/v5"
)

const (
	listFailedTransactionsByUploadIDQuery = `
SELECT id
FROM transactions
WHERE upload_id = $1
  AND status = 'failed'
ORDER BY row_number ASC, id ASC
`
	findLatestTransactionClassificationAttemptQuery = `
SELECT id
FROM transaction_classification_attempt
WHERE transaction_id = $1
ORDER BY retry_count DESC, created_at DESC
LIMIT 1
`
	markTransactionForClassificationRetryQuery = `
UPDATE transactions
SET status = 'processing',
    retry_count = retry_count + 1,
    last_retried_at = $3,
    updated_at = $3
WHERE id = $1
  AND upload_id = $2
  AND status = 'failed'
  AND NOT EXISTS (
      SELECT 1
      FROM transaction_processing_queue
      WHERE task_name = $4
        AND transaction_id = $1
  )
RETURNING retry_count, last_retried_at
`
	createTransactionClassificationAttemptQuery = `
INSERT INTO transaction_classification_attempt (
    transaction_id,
    upload_id,
    task_name,
    parent_job_id,
    retry_count,
    last_retried_at
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id
`
)

// PostgresTransactionClassificationRetryRepository persists retry attempts in PostgreSQL.
type PostgresTransactionClassificationRetryRepository struct {
	pool pgxQuerier
}

// NewPostgresTransactionClassificationRetryRepository builds a PostgresTransactionClassificationRetryRepository.
func NewPostgresTransactionClassificationRetryRepository(pool pgxQuerier) *PostgresTransactionClassificationRetryRepository {
	return &PostgresTransactionClassificationRetryRepository{pool: pool}
}

// ListFailedByUploadID returns failed transactions that are eligible for manual retry.
func (r *PostgresTransactionClassificationRetryRepository) ListFailedByUploadID(ctx context.Context, uploadID valueobjects.UploadID) ([]valueobjects.TransactionID, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	rows, err := querier.Query(ctx, listFailedTransactionsByUploadIDQuery, uploadID.String())
	if err != nil {
		return nil, fmt.Errorf("querying failed transactions by upload id: %w", err)
	}
	defer rows.Close()

	transactionIDs := make([]valueobjects.TransactionID, 0)
	for rows.Next() {
		var rawTransactionID string
		if err := rows.Scan(&rawTransactionID); err != nil {
			return nil, fmt.Errorf("scanning failed transaction id: %w", err)
		}

		transactionID, err := valueobjects.TransactionIDFromString(rawTransactionID)
		if err != nil {
			return nil, fmt.Errorf("parsing failed transaction id: %w", err)
		}

		transactionIDs = append(transactionIDs, transactionID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating failed transactions by upload id: %w", err)
	}

	return transactionIDs, nil
}

// RetryFailedTransaction atomically claims a failed transaction, persists retry metadata, and enqueues it.
func (r *PostgresTransactionClassificationRetryRepository) RetryFailedTransaction(ctx context.Context, command ports.RetryFailedTransactionCommand) (ports.RetryFailedTransactionResult, error) {
	querier := txQuerierFromContext(ctx, r.pool)

	parentJobID, err := latestTransactionClassificationAttemptID(ctx, querier, command.TransactionID)
	if err != nil {
		return ports.RetryFailedTransactionResult{}, err
	}

	var (
		retryCount    int
		lastRetriedAt time.Time
	)
	if err := querier.QueryRow(ctx, markTransactionForClassificationRetryQuery, command.TransactionID.String(), command.UploadID.String(), command.LastRetriedAt, command.TaskName).Scan(&retryCount, &lastRetriedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.RetryFailedTransactionResult{Skipped: true, SkipReason: "already_queued_or_not_failed"}, nil
		}

		return ports.RetryFailedTransactionResult{}, fmt.Errorf("marking transaction for classification retry: %w", err)
	}

	var parentArg any
	if parentJobID.Valid {
		parentArg = parentJobID.String
	}

	var jobID string
	if err := querier.QueryRow(ctx, createTransactionClassificationAttemptQuery, command.TransactionID.String(), command.UploadID.String(), command.TaskName, parentArg, retryCount, lastRetriedAt).Scan(&jobID); err != nil {
		return ports.RetryFailedTransactionResult{}, fmt.Errorf("creating transaction classification retry attempt: %w", err)
	}

	if _, err := querier.Exec(ctx, createTransactionProcessingQueueEntryQuery, command.TaskName, command.TransactionID.String()); err != nil {
		return ports.RetryFailedTransactionResult{}, fmt.Errorf("queueing transaction classification retry attempt: %w", err)
	}

	attempt := &ports.TransactionClassificationRetryAttempt{
		JobID:         jobID,
		RetryCount:    retryCount,
		LastRetriedAt: lastRetriedAt,
	}
	if parentJobID.Valid {
		attempt.ParentJobID = parentJobID.String
	}

	return ports.RetryFailedTransactionResult{Attempt: attempt}, nil
}

func latestTransactionClassificationAttemptID(ctx context.Context, querier pgxQuerier, transactionID valueobjects.TransactionID) (sql.NullString, error) {
	var parentJobID sql.NullString
	if err := querier.QueryRow(ctx, findLatestTransactionClassificationAttemptQuery, transactionID.String()).Scan(&parentJobID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sql.NullString{}, nil
		}

		return sql.NullString{}, fmt.Errorf("finding latest transaction classification attempt: %w", err)
	}

	return parentJobID, nil
}
