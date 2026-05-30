package entities

import (
	"testing"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

func newTestFeedback(t *testing.T) (*Feedback, valueobjects.FeedbackID, valueobjects.TransactionID) {
	t.Helper()

	id, err := valueobjects.FeedbackIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("FeedbackIDFromString() error = %v", err)
	}

	userID, err := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("UserIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	feedback, err := NewFeedback(id, userID, transactionID, valueobjects.ThumbsUpFeedbackKind(), now)
	if err != nil {
		t.Fatalf("NewFeedback() error = %v", err)
	}

	return feedback, id, transactionID
}

// TestNewFeedback verifies the new feedback behavior and the expected outcome asserted below.
func TestNewFeedback(t *testing.T) {
	t.Parallel()

	feedback, id, transactionID := newTestFeedback(t)

	if !feedback.ID().Equal(id) {
		t.Fatalf("ID() = %q, want %q", feedback.ID().String(), id.String())
	}

	if feedback.TransactionID() != transactionID.String() {
		t.Fatalf("TransactionID() = %q, want %q", feedback.TransactionID(), transactionID.String())
	}

	if feedback.Kind() != "thumbs_up" {
		t.Fatalf("Kind() = %q, want %q", feedback.Kind(), "thumbs_up")
	}
}

// TestNewFeedbackRejectsZeroTimestamp verifies the new feedback rejects zero timestamp behavior and the expected outcome asserted below.
func TestNewFeedbackRejectsZeroTimestamp(t *testing.T) {
	t.Parallel()

	id, _ := valueobjects.FeedbackIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	userID, _ := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	transactionID, _ := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")

	if _, err := NewFeedback(id, userID, transactionID, valueobjects.ThumbsUpFeedbackKind(), time.Time{}); err == nil {
		t.Fatal("expected error for zero timestamp")
	}
}

// TestFeedbackChangeKind verifies the feedback change kind behavior and the expected outcome asserted below.
func TestFeedbackChangeKind(t *testing.T) {
	t.Parallel()

	feedback, _, _ := newTestFeedback(t)
	now := time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)

	if err := feedback.ChangeKind(valueobjects.ThumbsDownFeedbackKind(), now); err != nil {
		t.Fatalf("ChangeKind() error = %v", err)
	}

	if feedback.Kind() != "thumbs_down" {
		t.Fatalf("Kind() = %q, want %q", feedback.Kind(), "thumbs_down")
	}

	if !feedback.UpdatedAt().Equal(now) {
		t.Fatalf("UpdatedAt() = %v, want %v", feedback.UpdatedAt(), now)
	}
}

// TestFeedbackChangeKindRejectsZeroTimestamp verifies the feedback change kind rejects zero timestamp behavior and the expected outcome asserted below.
func TestFeedbackChangeKindRejectsZeroTimestamp(t *testing.T) {
	t.Parallel()

	feedback, _, _ := newTestFeedback(t)
	if err := feedback.ChangeKind(valueobjects.ThumbsDownFeedbackKind(), time.Time{}); err == nil {
		t.Fatal("expected error for zero timestamp")
	}
}

// TestFeedbackRecordUpserted verifies the feedback record upserted behavior and the expected outcome asserted below.
func TestFeedbackRecordUpserted(t *testing.T) {
	t.Parallel()

	feedback, _, _ := newTestFeedback(t)
	now := time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)

	if err := feedback.RecordUpserted(now, "01962b8f-aeb2-7e03-a8ff-1edce1300002", "01962b8f-aeb2-7e03-a8ff-1edce1300001"); err != nil {
		t.Fatalf("RecordUpserted() error = %v", err)
	}

	if len(feedback.PullDomainEvents()) != 1 {
		t.Fatal("expected one domain event")
	}
}

// TestFeedbackRecordDeleted verifies the feedback record deleted behavior and the expected outcome asserted below.
func TestFeedbackRecordDeleted(t *testing.T) {
	t.Parallel()

	feedback, _, _ := newTestFeedback(t)
	now := time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)

	if err := feedback.RecordDeleted(now, "01962b8f-aeb2-7e03-a8ff-1edce1300002", "01962b8f-aeb2-7e03-a8ff-1edce1300001"); err != nil {
		t.Fatalf("RecordDeleted() error = %v", err)
	}

	if len(feedback.PullDomainEvents()) != 1 {
		t.Fatal("expected one domain event")
	}
}

// TestFeedbackEqual verifies the feedback equal behavior and the expected outcome asserted below.
func TestFeedbackEqual(t *testing.T) {
	t.Parallel()

	f1, id, _ := newTestFeedback(t)

	userID, _ := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	txID, _ := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	f2 := ReconstituteFeedback(id, userID, txID, valueobjects.ThumbsDownFeedbackKind(), now, now)

	if !f1.Equal(f2) {
		t.Fatal("expected feedbacks with same id to be equal")
	}

	otherId, _ := valueobjects.FeedbackIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300003")
	f3 := ReconstituteFeedback(otherId, userID, txID, valueobjects.ThumbsUpFeedbackKind(), now, now)
	if f1.Equal(f3) {
		t.Fatal("expected feedbacks with different ids to not be equal")
	}
}
