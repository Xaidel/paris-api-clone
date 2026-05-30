package ports_test

import (
	"testing"

	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func TestDownloadTransactionUploadQueryActorFields(t *testing.T) {
	t.Parallel()

	query := inboundports.DownloadTransactionUploadQuery{
		ID:           "upload-1",
		ActorUserID:  "user-1",
		ActorGroupID: "group-1",
	}

	if query.ID != "upload-1" {
		t.Fatalf("query.ID = %q, want %q", query.ID, "upload-1")
	}

	if query.ActorUserID != "user-1" {
		t.Fatalf("query.ActorUserID = %q, want %q", query.ActorUserID, "user-1")
	}

	if query.ActorGroupID != "group-1" {
		t.Fatalf("query.ActorGroupID = %q, want %q", query.ActorGroupID, "group-1")
	}
}

func TestDownloadTransactionUploadResultIncludesFileBytes(t *testing.T) {
	t.Parallel()

	result := inboundports.DownloadTransactionUploadResult{
		FileName:      "transactions.xlsx",
		ContentType:   "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		ContentLength: 7,
		FileBytes:     []byte("content"),
	}

	if result.FileName != "transactions.xlsx" {
		t.Fatalf("result.FileName = %q, want %q", result.FileName, "transactions.xlsx")
	}

	if result.ContentType != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		t.Fatalf("result.ContentType = %q, want %q", result.ContentType, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	}

	if result.ContentLength != 7 {
		t.Fatalf("result.ContentLength = %d, want %d", result.ContentLength, 7)
	}

	if string(result.FileBytes) != "content" {
		t.Fatalf("string(result.FileBytes) = %q, want %q", string(result.FileBytes), "content")
	}
}

func TestReadRawFileResultIncludesFileBytes(t *testing.T) {
	t.Parallel()

	result := outboundports.ReadRawFileResult{FileBytes: []byte("content"), ContentType: "text/csv"}

	if string(result.FileBytes) != "content" {
		t.Fatalf("string(result.FileBytes) = %q, want %q", string(result.FileBytes), "content")
	}

	if result.ContentType != "text/csv" {
		t.Fatalf("result.ContentType = %q, want %q", result.ContentType, "text/csv")
	}
}

func TestTransactionUploadResultIncludesGroupID(t *testing.T) {
	t.Parallel()

	result := outboundports.TransactionUploadResult{ID: "upload-1", GroupID: "group-1"}

	if result.GroupID != "group-1" {
		t.Fatalf("result.GroupID = %q, want %q", result.GroupID, "group-1")
	}
}
