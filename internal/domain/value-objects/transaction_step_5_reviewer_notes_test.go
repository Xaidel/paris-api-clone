package valueobjects

import "testing"

// TestNewTransactionStep5ReviewerNotes verifies the new transaction step 5 reviewer notes behavior and the expected outcome asserted below.
func TestNewTransactionStep5ReviewerNotes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value *string
		want  *string
	}{
		{
			name:  "returns empty when nil",
			value: nil,
			want:  nil,
		},
		{
			name:  "returns empty when whitespace only",
			value: stringPointer("   "),
			want:  nil,
		},
		{
			name:  "normalizes surrounding whitespace",
			value: stringPointer("  follow up with sector lead  "),
			want:  stringPointer("follow up with sector lead"),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := NewTransactionStep5ReviewerNotes(tc.value)
			gotValue := got.String()
			if !equalStringPointers(gotValue, tc.want) {
				t.Fatalf("NewTransactionStep5ReviewerNotes().String() = %v, want %v", gotValue, tc.want)
			}

			if got.HasValue() != (tc.want != nil) {
				t.Fatalf("NewTransactionStep5ReviewerNotes().HasValue() = %t, want %t", got.HasValue(), tc.want != nil)
			}
		})
	}
}

func equalStringPointers(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}

	return *left == *right
}

func stringPointer(value string) *string {
	return &value
}
