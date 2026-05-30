package ports

import (
	"testing"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func TestCreateTransactionUploadResultSkippedRows(t *testing.T) {
	t.Parallel()

	result := CreateTransactionUploadResult{
		Upload: outboundports.TransactionUploadResult{ID: "upload-1"},
		SkippedRows: []outboundports.TransactionUploadSkippedRow{
			{RowNumber: 3, Reason: outboundports.TransactionUploadSkippedRowReasonMalformed},
			{RowNumber: 5, Reason: outboundports.TransactionUploadSkippedRowReasonNotValidTransaction},
		},
	}

	if len(result.SkippedRows) != 2 {
		t.Fatalf("len(result.SkippedRows) = %d, want %d", len(result.SkippedRows), 2)
	}

	if result.SkippedRows[0].RowNumber != 3 {
		t.Fatalf("result.SkippedRows[0].RowNumber = %d, want %d", result.SkippedRows[0].RowNumber, 3)
	}

	if result.SkippedRows[0].Reason != outboundports.TransactionUploadSkippedRowReasonMalformed {
		t.Fatalf("result.SkippedRows[0].Reason = %q, want %q", result.SkippedRows[0].Reason, outboundports.TransactionUploadSkippedRowReasonMalformed)
	}

	if result.SkippedRows[1].Reason != outboundports.TransactionUploadSkippedRowReasonNotValidTransaction {
		t.Fatalf("result.SkippedRows[1].Reason = %q, want %q", result.SkippedRows[1].Reason, outboundports.TransactionUploadSkippedRowReasonNotValidTransaction)
	}
}

func TestTransactionUploadProgressUpdateSkippedRows(t *testing.T) {
	t.Parallel()

	update := outboundports.TransactionUploadProgressUpdate{
		Status:   outboundports.TransactionUploadProgressStatusCompleted,
		Progress: 100,
		SkippedRows: []outboundports.TransactionUploadSkippedRow{{
			RowNumber: 4,
			Reason:    outboundports.TransactionUploadSkippedRowReasonMalformed,
		}},
	}

	if len(update.SkippedRows) != 1 {
		t.Fatalf("len(update.SkippedRows) = %d, want %d", len(update.SkippedRows), 1)
	}

	if update.SkippedRows[0].RowNumber != 4 {
		t.Fatalf("update.SkippedRows[0].RowNumber = %d, want %d", update.SkippedRows[0].RowNumber, 4)
	}
}
