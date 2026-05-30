package adapters

import (
	"context"
	"regexp"
	"testing"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/pashagolub/pgxmock/v4"
)

// TestPostgresTransactionProcessingQueueEnqueue verifies the postgres transaction processing queue enqueue behavior and the expected outcome asserted below.
func TestPostgresTransactionProcessingQueueEnqueue(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	queue := NewPostgresTransactionProcessingQueue(mock)
	mock.ExpectExec(regexp.QuoteMeta(createTransactionProcessingQueueEntryQuery)).
		WithArgs("transaction:classify", transactionID.String()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := queue.Enqueue(context.Background(), "transaction:classify", transactionID); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
}
