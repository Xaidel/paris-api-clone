package valueobjects

import "testing"

// TestTransactionStatusFromString verifies the transaction status from string behavior and the expected outcome asserted below.
func TestTransactionStatusFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "pending", input: "pending", want: "pending"},
		{name: "processing", input: " processing ", want: "processing"},
		{name: "ai reviewed", input: "ai-reviewed", want: "ai-reviewed"},
		{name: "from previous transactions", input: "from-previous-transactions", want: "from-previous-transactions"},
		{name: "professionally reviewed", input: "professionally-reviewed", want: "professionally-reviewed"},
		{name: "failed", input: "failed", want: "failed"},
		{name: "legacy uploaded", input: "uploaded", want: "uploaded"},
		{name: "legacy classified maps ai reviewed", input: "classified", want: "ai-reviewed"},
		{name: "invalid", input: "done", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status, err := TransactionStatusFromString(tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("TransactionStatusFromString(%q) error = %v, wantErr = %v", tc.input, err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if status.String() != tc.want {
				t.Fatalf("TransactionStatusFromString(%q) = %q, want %q", tc.input, status.String(), tc.want)
			}
		})
	}
}
