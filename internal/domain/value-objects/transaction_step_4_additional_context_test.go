package valueobjects

import (
	"errors"
	"testing"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// TestNewTransactionStep4AdditionalContext verifies the new transaction step 4 additional context behavior and the expected outcome asserted below.
func TestNewTransactionStep4AdditionalContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		want      string
		wantField string
	}{
		{
			name:  "normalizes surrounding whitespace",
			value: "  Reviewed by analyst  ",
			want:  "Reviewed by analyst",
		},
		{
			name:      "rejects empty value",
			value:     "   ",
			wantField: "additional_context",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewTransactionStep4AdditionalContext(tc.value)
			if tc.wantField != "" {
				var validationErr *domain.ValidationError
				if !errors.As(err, &validationErr) {
					t.Fatalf("NewTransactionStep4AdditionalContext() error = %v, want validation error", err)
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
				t.Fatalf("NewTransactionStep4AdditionalContext() error = %v", err)
			}

			if got.String() != tc.want {
				t.Fatalf("NewTransactionStep4AdditionalContext().String() = %q, want %q", got.String(), tc.want)
			}
		})
	}
}
