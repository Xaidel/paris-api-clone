package valueobjects

import (
	"errors"
	"testing"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// TestNewColumnIndex verifies the new column index behavior and the expected outcome asserted below.
func TestNewColumnIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value int
		want  int
	}{
		{name: "zero index", value: 0, want: 0},
		{name: "positive index", value: 7, want: 7},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewColumnIndex(tc.value)
			if err != nil {
				t.Fatalf("NewColumnIndex(%d) error = %v", tc.value, err)
			}

			if got.Int() != tc.want {
				t.Fatalf("NewColumnIndex(%d).Int() = %d, want %d", tc.value, got.Int(), tc.want)
			}
		})
	}
}

// TestNewColumnIndexRejectsNegativeValue verifies the new column index rejects negative value behavior and the expected outcome asserted below.
func TestNewColumnIndexRejectsNegativeValue(t *testing.T) {
	t.Parallel()

	_, err := NewColumnIndex(-1)
	if !errors.Is(err, domain.ErrInvalidColumnIndex) {
		t.Fatalf("NewColumnIndex(-1) error = %v, want %v", err, domain.ErrInvalidColumnIndex)
	}
}
