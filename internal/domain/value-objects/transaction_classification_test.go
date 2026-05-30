package valueobjects

import "testing"

// TestTransactionClassificationFromString verifies the transaction classification from string behavior and the expected outcome asserted below.
func TestTransactionClassificationFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "unclassified", input: "unclassified", want: "unclassified"},
		{name: "aligned", input: "aligned", want: "aligned"},
		{name: "not aligned", input: " not-aligned ", want: "not-aligned"},
		{name: "next step", input: "next_step", want: "next_step"},
		{name: "next step alias", input: " next-step ", want: "next_step"},
		{name: "invalid", input: "unknown", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			classification, err := TransactionClassificationFromString(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("TransactionClassificationFromString(%q) error = %v, wantErr = %v", tc.input, err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if classification.String() != tc.want {
				t.Fatalf("TransactionClassificationFromString(%q) = %q, want %q", tc.input, classification.String(), tc.want)
			}
		})
	}
}

// TestTransactionClassificationIsTerminal verifies the transaction classification is terminal behavior and the expected outcome asserted below.
func TestTransactionClassificationIsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value TransactionClassification
		want  bool
	}{
		{name: "aligned is terminal", value: AlignedTransactionClassification(), want: true},
		{name: "not aligned is terminal", value: NotAlignedTransactionClassification(), want: true},
		{name: "next step is not terminal", value: NextStepTransactionClassification(), want: false},
		{name: "unclassified is not terminal", value: UnclassifiedTransactionClassification(), want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.value.IsTerminal(); got != tc.want {
				t.Fatalf("TransactionClassification.IsTerminal() = %v, want %v", got, tc.want)
			}
		})
	}
}
