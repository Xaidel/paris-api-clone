package usecases

import "testing"

// TestNotFoundError verifies the not found error behavior and the expected outcome asserted below.
func TestNotFoundError(t *testing.T) {
	t.Parallel()

	err := &NotFoundError{Resource: "user", ID: "123"}
	if got := err.Error(); got != "user \"123\" was not found" {
		t.Fatalf("Error() = %q", got)
	}
}

// TestConflictError verifies the conflict error behavior and the expected outcome asserted below.
func TestConflictError(t *testing.T) {
	t.Parallel()

	err := &ConflictError{Resource: "transaction upload", Reason: "upload file has already been ingested"}
	if got := err.Error(); got != "transaction upload conflict: upload file has already been ingested" {
		t.Fatalf("Error() = %q", got)
	}
}
