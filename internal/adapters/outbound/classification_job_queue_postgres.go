package adapters

import (
	"context"
	"fmt"
	"time"

	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/jackc/pgx/v5"
)

const (
	dequeueClassificationJobQuery = `
SELECT task_name, transaction_id, created_at
FROM transaction_processing_queue
WHERE task_name = $1
ORDER BY created_at ASC
FOR UPDATE SKIP LOCKED
LIMIT 1
`
	dequeueClassificationJobBatchQuery = `
SELECT task_name, transaction_id, created_at
FROM transaction_processing_queue
WHERE task_name = $1
ORDER BY created_at ASC
FOR UPDATE SKIP LOCKED
LIMIT $2
`
	completeClassificationJobQuery = `
DELETE FROM transaction_processing_queue
WHERE task_name = $1 AND transaction_id = $2
`
)

// PostgresClassificationJobQueue dequeues transaction classification jobs from PostgreSQL.
type PostgresClassificationJobQueue struct {
	pool pgxQuerier
}

// NewPostgresClassificationJobQueue builds a PostgresClassificationJobQueue.
func NewPostgresClassificationJobQueue(pool pgxQuerier) *PostgresClassificationJobQueue {
	return &PostgresClassificationJobQueue{pool: pool}
}

// Dequeue claims the next available transaction classification job.
func (q *PostgresClassificationJobQueue) Dequeue(ctx context.Context, taskName string) (*ports.ClassificationJob, error) {
	querier := txQuerierFromContext(ctx, q.pool)
	row := querier.QueryRow(ctx, dequeueClassificationJobQuery, taskName)

	var queuedTaskName string
	var transactionID string
	var queuedAt time.Time
	if err := row.Scan(&queuedTaskName, &transactionID, &queuedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("dequeueing classification job: %w", err)
	}

	job := ports.ClassificationJob{
		TaskName: queuedTaskName,
		Payload: ports.ClassificationJobPayload{
			TransactionID: transactionID,
		},
		QueuedAt: queuedAt,
	}

	return &job, nil
}

// DequeueBatch claims the next available transaction classification jobs up to limit.
func (q *PostgresClassificationJobQueue) DequeueBatch(ctx context.Context, taskName string, limit int) ([]ports.ClassificationJob, error) {
	if limit <= 0 {
		return nil, nil
	}

	querier := txQuerierFromContext(ctx, q.pool)
	rows, err := querier.Query(ctx, dequeueClassificationJobBatchQuery, taskName, limit)
	if err != nil {
		return nil, fmt.Errorf("dequeueing classification job batch: %w", err)
	}
	defer rows.Close()

	jobs := make([]ports.ClassificationJob, 0, limit)
	for rows.Next() {
		var queuedTaskName string
		var transactionID string
		var queuedAt time.Time
		if err := rows.Scan(&queuedTaskName, &transactionID, &queuedAt); err != nil {
			return nil, fmt.Errorf("scanning classification job batch row: %w", err)
		}

		jobs = append(jobs, ports.ClassificationJob{
			TaskName: queuedTaskName,
			Payload: ports.ClassificationJobPayload{
				TransactionID: transactionID,
			},
			QueuedAt: queuedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating classification job batch rows: %w", err)
	}

	return jobs, nil
}

// Complete removes a processed classification job from the queue table.
func (q *PostgresClassificationJobQueue) Complete(ctx context.Context, job ports.ClassificationJob) error {
	querier := txQuerierFromContext(ctx, q.pool)
	if _, err := querier.Exec(ctx, completeClassificationJobQuery, job.TaskName, job.Payload.TransactionID); err != nil {
		return fmt.Errorf("completing classification job: %w", err)
	}

	return nil
}
