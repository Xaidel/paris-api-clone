package usecases

import (
	"testing"

	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func TestFilterUploadRowsByTransactionCount(t *testing.T) {
	t.Parallel()

	headers := []string{
		"Product",
		"Reference Number",
		"No. of Transactions",
	}
	rows := [][]string{
		{"CG", "REF-1", "1"},
		{"CG", "REF-2", "2"},
		{"CG", "REF-3", " bad "},
		{"CG", "REF-4", " 1 "},
	}

	result, err := filterUploadRowsByTransactionCount(headers, rows)
	if err != nil {
		t.Fatalf("filterUploadRowsByTransactionCount() error = %v", err)
	}

	if len(result.EligibleRows) != 2 {
		t.Fatalf("len(result.EligibleRows) = %d, want %d", len(result.EligibleRows), 2)
	}

	if result.EligibleRows[0][1] != "REF-1" {
		t.Fatalf("result.EligibleRows[0][1] = %q, want %q", result.EligibleRows[0][1], "REF-1")
	}

	if result.EligibleRows[1][1] != "REF-4" {
		t.Fatalf("result.EligibleRows[1][1] = %q, want %q", result.EligibleRows[1][1], "REF-4")
	}

	assertSkippedRows(t, result.SkippedRows, []ports.TransactionUploadSkippedRow{
		{RowNumber: 3, Reason: ports.TransactionUploadSkippedRowReasonNotValidTransaction},
		{RowNumber: 4, Reason: ports.TransactionUploadSkippedRowReasonMalformed},
	})
}

func TestFilterUploadRowsByTransactionCountReturnsEmptyWhenNoEligibleRows(t *testing.T) {
	t.Parallel()

	headers := []string{"Reference Number", "No. of Transactions"}
	rows := [][]string{{"REF-1", "0"}, {"REF-2", "2"}, {"REF-3", "bad"}}

	result, err := filterUploadRowsByTransactionCount(headers, rows)
	if err != nil {
		t.Fatalf("filterUploadRowsByTransactionCount() error = %v", err)
	}

	if len(result.EligibleRows) != 0 {
		t.Fatalf("len(result.EligibleRows) = %d, want %d", len(result.EligibleRows), 0)
	}

	assertSkippedRows(t, result.SkippedRows, []ports.TransactionUploadSkippedRow{
		{RowNumber: 2, Reason: ports.TransactionUploadSkippedRowReasonNotValidTransaction},
		{RowNumber: 3, Reason: ports.TransactionUploadSkippedRowReasonNotValidTransaction},
		{RowNumber: 4, Reason: ports.TransactionUploadSkippedRowReasonMalformed},
	})
}

func assertSkippedRows(t *testing.T, got, want []ports.TransactionUploadSkippedRow) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(skippedRows) = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("skippedRows[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}
