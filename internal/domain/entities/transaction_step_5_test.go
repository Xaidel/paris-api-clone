package entities

import (
	"errors"
	"testing"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TestNewTransactionStep5 verifies the new transaction step 5 behavior and the expected outcome asserted below.
func TestNewTransactionStep5(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300101")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	question1Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion1Justification("screened against question 1")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion1Justification() error = %v", err)
	}

	question2Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion2Justification("screened against question 2")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion2Justification() error = %v", err)
	}

	reviewerNotes := valueobjects.NewTransactionStep5ReviewerNotes(stringPointer("  analyst note  "))
	now := time.Date(2026, time.April, 3, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name               string
		question1Answer    bool
		question2Answer    bool
		isFinal            bool
		now                time.Time
		wantClassification string
		wantErr            error
	}{
		{
			name:               "both false classify aligned",
			question1Answer:    false,
			question2Answer:    false,
			isFinal:            false,
			now:                now,
			wantClassification: "aligned",
		},
		{
			name:               "any true classify not aligned",
			question1Answer:    true,
			question2Answer:    false,
			isFinal:            true,
			now:                now,
			wantClassification: "not-aligned",
		},
		{
			name:            "zero timestamp rejected",
			question1Answer: false,
			question2Answer: false,
			isFinal:         false,
			now:             time.Time{},
			wantErr:         domain.ErrInvalidTimestamp,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			step5, err := NewTransactionStep5(
				transactionID,
				tc.question1Answer,
				question1Justification,
				tc.question2Answer,
				question2Justification,
				reviewerNotes,
				tc.isFinal,
				tc.now,
			)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("NewTransactionStep5() error = %v, want %v", err, tc.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("NewTransactionStep5() error = %v", err)
			}

			if step5.Classification().String() != tc.wantClassification {
				t.Fatalf("step5.Classification().String() = %q, want %q", step5.Classification().String(), tc.wantClassification)
			}

			if step5.IsFinal() != tc.isFinal {
				t.Fatalf("step5.IsFinal() = %t, want %t", step5.IsFinal(), tc.isFinal)
			}

			gotReviewerNotes := step5.ReviewerNotes().String()
			if gotReviewerNotes == nil || *gotReviewerNotes != "analyst note" {
				t.Fatalf("step5.ReviewerNotes().String() = %v, want %q", gotReviewerNotes, "analyst note")
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}
