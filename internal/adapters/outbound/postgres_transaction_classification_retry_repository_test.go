package adapters

import (
	"context"
	"regexp"
	"testing"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/pashagolub/pgxmock/v4"
)

func TestPostgresTransactionClassificationRetryRepositoryListFailedByUploadID(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	repository := NewPostgresTransactionClassificationRetryRepository(mock)
	mock.ExpectQuery(regexp.QuoteMeta(listFailedTransactionsByUploadIDQuery)).WithArgs(uploadID.String()).WillReturnRows(
		pgxmock.NewRows([]string{"id"}).AddRow("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60").AddRow("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"),
	)

	transactionIDs, err := repository.ListFailedByUploadID(context.Background(), uploadID)
	if err != nil {
		t.Fatalf("ListFailedByUploadID() error = %v", err)
	}

	if len(transactionIDs) != 2 {
		t.Fatalf("len(transactionIDs) = %d, want %d", len(transactionIDs), 2)
	}
}

func TestPostgresTransactionClassificationRetryRepositoryRetryFailedTransaction(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	retriedAt := time.Date(2026, time.April, 4, 12, 0, 0, 0, time.UTC)
	repository := NewPostgresTransactionClassificationRetryRepository(mock)

	mock.ExpectQuery(regexp.QuoteMeta(findLatestTransactionClassificationAttemptQuery)).WithArgs(transactionID.String()).WillReturnRows(
		pgxmock.NewRows([]string{"id"}).AddRow("job-parent"),
	)
	mock.ExpectQuery(regexp.QuoteMeta(markTransactionForClassificationRetryQuery)).WithArgs(transactionID.String(), uploadID.String(), retriedAt, ports.TransactionClassifyReactTaskName).WillReturnRows(
		pgxmock.NewRows([]string{"retry_count", "last_retried_at"}).AddRow(2, retriedAt),
	)
	mock.ExpectQuery(regexp.QuoteMeta(createTransactionClassificationAttemptQuery)).WithArgs(transactionID.String(), uploadID.String(), ports.TransactionClassifyReactTaskName, "job-parent", 2, retriedAt).WillReturnRows(
		pgxmock.NewRows([]string{"id"}).AddRow("job-2"),
	)
	mock.ExpectExec(regexp.QuoteMeta(createTransactionProcessingQueueEntryQuery)).WithArgs(ports.TransactionClassifyReactTaskName, transactionID.String()).WillReturnResult(
		pgxmock.NewResult("INSERT", 1),
	)

	attempt, err := repository.RetryFailedTransaction(context.Background(), ports.RetryFailedTransactionCommand{
		UploadID:      uploadID,
		TransactionID: transactionID,
		TaskName:      ports.TransactionClassifyReactTaskName,
		LastRetriedAt: retriedAt,
	})
	if err != nil {
		t.Fatalf("RetryFailedTransaction() error = %v", err)
	}

	if attempt.Attempt == nil {
		t.Fatal("RetryFailedTransaction().Attempt = nil, want attempt")
	}
	if attempt.Attempt.JobID != "job-2" {
		t.Fatalf("attempt.Attempt.JobID = %q, want %q", attempt.Attempt.JobID, "job-2")
	}
	if attempt.Attempt.ParentJobID != "job-parent" {
		t.Fatalf("attempt.Attempt.ParentJobID = %q, want %q", attempt.Attempt.ParentJobID, "job-parent")
	}
	if attempt.Attempt.RetryCount != 2 {
		t.Fatalf("attempt.Attempt.RetryCount = %d, want %d", attempt.Attempt.RetryCount, 2)
	}
}

func TestPostgresTransactionClassificationRetryRepositoryRetryFailedTransactionReturnsNilWhenSkipped(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	repository := NewPostgresTransactionClassificationRetryRepository(mock)
	retriedAt := time.Date(2026, time.April, 4, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(findLatestTransactionClassificationAttemptQuery)).WithArgs(transactionID.String()).WillReturnRows(
		pgxmock.NewRows([]string{"id"}),
	)
	mock.ExpectQuery(regexp.QuoteMeta(markTransactionForClassificationRetryQuery)).WithArgs(transactionID.String(), uploadID.String(), retriedAt, ports.TransactionClassifyReactTaskName).WillReturnRows(
		pgxmock.NewRows([]string{"retry_count", "last_retried_at"}),
	)

	attempt, err := repository.RetryFailedTransaction(context.Background(), ports.RetryFailedTransactionCommand{
		UploadID:      uploadID,
		TransactionID: transactionID,
		TaskName:      ports.TransactionClassifyReactTaskName,
		LastRetriedAt: retriedAt,
	})
	if err != nil {
		t.Fatalf("RetryFailedTransaction() error = %v", err)
	}
	if !attempt.Skipped {
		t.Fatal("RetryFailedTransaction().Skipped = false, want true")
	}
	if attempt.SkipReason != "already_queued_or_not_failed" {
		t.Fatalf("attempt.SkipReason = %q, want %q", attempt.SkipReason, "already_queued_or_not_failed")
	}
	if attempt.Attempt != nil {
		t.Fatalf("RetryFailedTransaction().Attempt = %#v, want nil", attempt.Attempt)
	}
}
