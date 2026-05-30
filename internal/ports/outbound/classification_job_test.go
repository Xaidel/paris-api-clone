package ports

import (
	"testing"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// This test pins the default queue contract: a new classification job must use
// the standard task name and must serialize the transaction identifier into the
// queue-safe payload shape consumed by workers.
func TestNewClassificationJob(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	job := NewClassificationJob(transactionID)
	if job.TaskName != TransactionClassifyReactTaskName {
		t.Fatalf("job.TaskName = %q, want %q", job.TaskName, TransactionClassifyReactTaskName)
	}

	if job.Payload.TransactionID != transactionID.String() {
		t.Fatalf("job.Payload.TransactionID = %q, want %q", job.Payload.TransactionID, transactionID.String())
	}
}
