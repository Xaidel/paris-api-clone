package valueobjects

import (
	"errors"
	"testing"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// TestNewTransactionStep5ScreeningQuestion2Justification verifies the new transaction step 5 screening question 2 justification behavior and the expected outcome asserted below.
func TestNewTransactionStep5ScreeningQuestion2Justification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		want      string
		wantField string
	}{
		{
			name:  "normalizes surrounding whitespace",
			value: "  reviewed counterparty disclosures  ",
			want:  "reviewed counterparty disclosures",
		},
		{
			name:      "rejects empty value",
			value:     "   ",
			wantField: "screening_question_2_justification",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewTransactionStep5ScreeningQuestion2Justification(tc.value)
			if tc.wantField != "" {
				var validationErr *domain.ValidationError
				if !errors.As(err, &validationErr) {
					t.Fatalf("NewTransactionStep5ScreeningQuestion2Justification() error = %v, want validation error", err)
				}

				fields := validationErr.Fields()
				if len(fields) != 1 {
					t.Fatalf("len(validationErr.Fields()) = %d, want %d", len(fields), 1)
				}
				if fields[0].Field() != tc.wantField {
					t.Fatalf("validationErr.Fields()[0].Field() = %q, want %q", fields[0].Field(), tc.wantField)
				}

				return
			}

			if err != nil {
				t.Fatalf("NewTransactionStep5ScreeningQuestion2Justification() error = %v", err)
			}

			if got.String() != tc.want {
				t.Fatalf("NewTransactionStep5ScreeningQuestion2Justification().String() = %q, want %q", got.String(), tc.want)
			}
		})
	}
}
