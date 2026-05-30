package valueobjects

import (
	"errors"
	"testing"

	"github.com/gyud-adb/paris-api/internal/domain"
)

func TestTransactionUploadStatusFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    TransactionUploadStatus
		wantErr error
	}{
		{name: "uploaded", input: "uploaded", want: UploadedTransactionUploadStatus()},
		{name: "failed", input: "failed", want: FailedTransactionUploadStatus()},
		{name: "normalizes case and whitespace", input: "  UpLoAdEd  ", want: UploadedTransactionUploadStatus()},
		{name: "unsupported value", input: "processing", wantErr: domain.ErrInvalidTransactionUploadStatus},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			status, err := TransactionUploadStatusFromString(tc.input)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("TransactionUploadStatusFromString(%q) error = %v, want %v", tc.input, err, tc.wantErr)
			}
			if tc.wantErr != nil {
				return
			}

			if !status.Equal(tc.want) {
				t.Fatalf("TransactionUploadStatusFromString(%q) = %q, want %q", tc.input, status.String(), tc.want.String())
			}
		})
	}
}
