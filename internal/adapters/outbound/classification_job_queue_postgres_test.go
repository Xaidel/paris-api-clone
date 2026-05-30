package adapters

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/pashagolub/pgxmock/v4"
)

// TestPostgresClassificationJobQueueDequeueAndComplete verifies dequeue returns
// the oldest queued job and complete removes that job successfully.
func TestPostgresClassificationJobQueueDequeueAndComplete(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	queue := NewPostgresClassificationJobQueue(mock)
	queuedAt := time.Date(2026, time.April, 22, 10, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(dequeueClassificationJobQuery)).WithArgs(ports.TransactionClassifyTaskName).WillReturnRows(
		pgxmock.NewRows([]string{"task_name", "transaction_id", "created_at"}).AddRow(ports.TransactionClassifyTaskName, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60", queuedAt),
	)

	job, err := queue.Dequeue(context.Background(), ports.TransactionClassifyTaskName)
	if err != nil {
		t.Fatalf("Dequeue() error = %v", err)
	}

	if job == nil {
		t.Fatal("Dequeue() = nil, want job")
	}

	if job.TaskName != ports.TransactionClassifyTaskName {
		t.Fatalf("job.TaskName = %q, want %q", job.TaskName, ports.TransactionClassifyTaskName)
	}

	if job.Payload.TransactionID != "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60" {
		t.Fatalf("job.Payload.TransactionID = %q, want %q", job.Payload.TransactionID, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	}

	if !job.QueuedAt.Equal(queuedAt) {
		t.Fatalf("job.QueuedAt = %v, want %v", job.QueuedAt, queuedAt)
	}

	mock.ExpectExec(regexp.QuoteMeta(completeClassificationJobQuery)).
		WithArgs(job.TaskName, job.Payload.TransactionID).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	if err := queue.Complete(context.Background(), *job); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
}

// TestPostgresClassificationJobQueueDequeueEmpty verifies dequeue returns nil
// when no matching classification jobs are available.
func TestPostgresClassificationJobQueueDequeueEmpty(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	queue := NewPostgresClassificationJobQueue(mock)
	mock.ExpectQuery(regexp.QuoteMeta(dequeueClassificationJobQuery)).WithArgs(ports.TransactionClassifyTaskName).WillReturnRows(
		pgxmock.NewRows([]string{"task_name", "transaction_id", "created_at"}),
	)

	job, err := queue.Dequeue(context.Background(), ports.TransactionClassifyTaskName)
	if err != nil {
		t.Fatalf("Dequeue() error = %v", err)
	}

	if job != nil {
		t.Fatalf("Dequeue() = %#v, want nil", job)
	}
}

// TestPostgresClassificationJobQueueDequeueBatch verifies batch dequeue returns
// multiple queued ReAct jobs in order with their queued timestamps preserved.
func TestPostgresClassificationJobQueueDequeueBatch(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	queue := NewPostgresClassificationJobQueue(mock)
	firstQueuedAt := time.Date(2026, time.April, 22, 10, 0, 0, 0, time.UTC)
	secondQueuedAt := firstQueuedAt.Add(500 * time.Millisecond)
	mock.ExpectQuery(regexp.QuoteMeta(dequeueClassificationJobBatchQuery)).WithArgs(ports.TransactionClassifyReactTaskName, 2).WillReturnRows(
		pgxmock.NewRows([]string{"task_name", "transaction_id", "created_at"}).
			AddRow(ports.TransactionClassifyReactTaskName, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60", firstQueuedAt).
			AddRow(ports.TransactionClassifyReactTaskName, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61", secondQueuedAt),
	)

	jobs, err := queue.DequeueBatch(context.Background(), ports.TransactionClassifyReactTaskName, 2)
	if err != nil {
		t.Fatalf("DequeueBatch() error = %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("len(jobs) = %d, want %d", len(jobs), 2)
	}

	if jobs[0].Payload.TransactionID != "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60" {
		t.Fatalf("jobs[0].Payload.TransactionID = %q", jobs[0].Payload.TransactionID)
	}

	if !jobs[0].QueuedAt.Equal(firstQueuedAt) {
		t.Fatalf("jobs[0].QueuedAt = %v, want %v", jobs[0].QueuedAt, firstQueuedAt)
	}
}

// TestPostgresClassificationJobQueueDequeueBatchZeroLimit verifies batch
// dequeue returns no work when the caller requests a zero-sized batch.
func TestPostgresClassificationJobQueueDequeueBatchZeroLimit(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	queue := NewPostgresClassificationJobQueue(mock)
	jobs, err := queue.DequeueBatch(context.Background(), ports.TransactionClassifyReactTaskName, 0)
	if err != nil {
		t.Fatalf("DequeueBatch() error = %v", err)
	}

	if jobs != nil {
		t.Fatalf("DequeueBatch() = %#v, want nil", jobs)
	}
}
