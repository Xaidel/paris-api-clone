package entities

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TestTransactionUploadStoresGroupAndRecordsPreview verifies group ownership and preview audit event behavior.
func TestTransactionUploadStoresGroupAndRecordsPreview(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	uploadedAt := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	if _, err := NewTransactionUpload(
		uploadID,
		valueobjects.GroupID{},
		"transactions.xlsx",
		"xlsx",
		"d41d8cd98f00b204e9800998ecf8427e",
		"local",
		"uploads/file.xlsx",
		"transaction-file-v1",
		valueobjects.UploadedTransactionUploadStatus(),
		7,
		uploadedAt,
	); !errors.Is(err, domain.ErrInvalidGroupID) {
		t.Fatalf("NewTransactionUpload() error = %v, want %v", err, domain.ErrInvalidGroupID)
	}

	upload, err := NewTransactionUpload(
		uploadID,
		groupID,
		"transactions.xlsx",
		"xlsx",
		"d41d8cd98f00b204e9800998ecf8427e",
		"local",
		"uploads/file.xlsx",
		"transaction-file-v1",
		valueobjects.UploadedTransactionUploadStatus(),
		7,
		uploadedAt,
	)
	if err != nil {
		t.Fatalf("NewTransactionUpload() error = %v", err)
	}

	if upload.GroupID() != groupID {
		t.Fatalf("upload.GroupID() = %q, want %q", upload.GroupID().String(), groupID.String())
	}

	now := time.Date(2026, time.April, 3, 13, 0, 0, 0, time.UTC)
	if err := upload.RecordPreviewed(now, "admin-1", "group-1"); err != nil {
		t.Fatalf("RecordPreviewed() error = %v", err)
	}

	domainEvents := upload.PullDomainEvents()
	if len(domainEvents) != 1 {
		t.Fatalf("len(upload.PullDomainEvents()) = %d, want %d", len(domainEvents), 1)
	}

	event, ok := domainEvents[0].(*events.AdminActionOccurred)
	if !ok {
		t.Fatalf("upload.PullDomainEvents()[0] type = %T, want %T", domainEvents[0], &events.AdminActionOccurred{})
	}

	if event.OccurredAt() != now {
		t.Fatalf("event.OccurredAt() = %v, want %v", event.OccurredAt(), now)
	}

	if event.ActorUserID() != "admin-1" {
		t.Fatalf("event.ActorUserID() = %q, want %q", event.ActorUserID(), "admin-1")
	}

	if event.ActorGroupID() != "group-1" {
		t.Fatalf("event.ActorGroupID() = %q, want %q", event.ActorGroupID(), "group-1")
	}

	if event.EventType() != events.PreviewTransactionUploadEventType {
		t.Fatalf("event.EventType() = %q, want %q", event.EventType(), events.PreviewTransactionUploadEventType)
	}

	var payload map[string]any
	if err := json.Unmarshal(event.EventData(), &payload); err != nil {
		t.Fatalf("json.Unmarshal(event.EventData()) error = %v", err)
	}

	if payload["action"] != "preview" {
		t.Fatalf("payload[action] = %v, want %q", payload["action"], "preview")
	}

	if payload["resource"] != "transaction_upload" {
		t.Fatalf("payload[resource] = %v, want %q", payload["resource"], "transaction_upload")
	}

	if payload["upload_id"] != uploadID.String() {
		t.Fatalf("payload[upload_id] = %v, want %q", payload["upload_id"], uploadID.String())
	}

	if payload["file_name"] != "transactions.xlsx" {
		t.Fatalf("payload[file_name] = %v, want %q", payload["file_name"], "transactions.xlsx")
	}

}

// TestReconstituteTransactionUploadRejectsInvalidGroupID verifies reconstitution rejects an invalid group id.
func TestReconstituteTransactionUploadRejectsInvalidGroupID(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	_, err = ReconstituteTransactionUpload(
		uploadID,
		valueobjects.GroupID{},
		"transactions.xlsx",
		"xlsx",
		"d41d8cd98f00b204e9800998ecf8427e",
		"local",
		"uploads/file.xlsx",
		"transaction-file-v1",
		valueobjects.UploadedTransactionUploadStatus(),
		7,
		time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
	)
	if !errors.Is(err, domain.ErrInvalidGroupID) {
		t.Fatalf("ReconstituteTransactionUpload() error = %v, want %v", err, domain.ErrInvalidGroupID)
	}
}

// TestNewTransactionUploadStoresStatusAndAllowsFailedZeroRows verifies the new transaction upload behavior and the expected outcome asserted below.
func TestNewTransactionUploadStoresStatusAndAllowsFailedZeroRows(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	tests := []struct {
		name       string
		status     valueobjects.TransactionUploadStatus
		rowCount   int
		assertErr  func(t *testing.T, err error)
		assertData func(t *testing.T, upload *TransactionUpload)
	}{
		{
			name:     "uploaded status accepts positive row count",
			status:   valueobjects.UploadedTransactionUploadStatus(),
			rowCount: 7,
			assertData: func(t *testing.T, upload *TransactionUpload) {
				t.Helper()

				if upload.RowCount() != 7 {
					t.Fatalf("upload.RowCount() = %d, want %d", upload.RowCount(), 7)
				}

				if upload.Status() != valueobjects.UploadedTransactionUploadStatus().String() {
					t.Fatalf("upload.Status() = %q, want %q", upload.Status(), valueobjects.UploadedTransactionUploadStatus().String())
				}
			},
		},
		{
			name:     "failed status accepts zero row count",
			status:   valueobjects.FailedTransactionUploadStatus(),
			rowCount: 0,
			assertData: func(t *testing.T, upload *TransactionUpload) {
				t.Helper()

				if upload.RowCount() != 0 {
					t.Fatalf("upload.RowCount() = %d, want %d", upload.RowCount(), 0)
				}

				if upload.Status() != valueobjects.FailedTransactionUploadStatus().String() {
					t.Fatalf("upload.Status() = %q, want %q", upload.Status(), valueobjects.FailedTransactionUploadStatus().String())
				}
			},
		},
		{
			name:     "uploaded status rejects zero row count",
			status:   valueobjects.UploadedTransactionUploadStatus(),
			rowCount: 0,
			assertErr: func(t *testing.T, err error) {
				t.Helper()

				if !errors.Is(err, domain.ErrInvalidRowCount) {
					t.Fatalf("NewTransactionUpload() error = %v, want %v", err, domain.ErrInvalidRowCount)
				}

				if err.Error() != "[INVALID_ROW_COUNT] row count must be zero for failed uploads or greater than zero for uploaded uploads" {
					t.Fatalf("NewTransactionUpload() error = %q, want %q", err.Error(), "[INVALID_ROW_COUNT] row count must be zero for failed uploads or greater than zero for uploaded uploads")
				}
			},
		},
		{
			name:     "invalid status is rejected",
			status:   valueobjects.TransactionUploadStatus{},
			rowCount: 7,
			assertErr: func(t *testing.T, err error) {
				t.Helper()

				if !errors.Is(err, domain.ErrInvalidTransactionUploadStatus) {
					t.Fatalf("NewTransactionUpload() error = %v, want %v", err, domain.ErrInvalidTransactionUploadStatus)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

				upload, err := NewTransactionUpload(
					uploadID,
					groupID,
					"transactions.xlsx",
					"xlsx",
					"d41d8cd98f00b204e9800998ecf8427e",
				"local",
				"uploads/file.xlsx",
				"transaction-file-v1",
				tc.status,
				tc.rowCount,
				time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
			)

			if tc.assertErr != nil {
				tc.assertErr(t, err)
				return
			}

				if err != nil {
					t.Fatalf("NewTransactionUpload() error = %v", err)
				}

				if upload.GroupID().String() != groupID.String() {
					t.Fatalf("upload.GroupID() = %q, want %q", upload.GroupID().String(), groupID.String())
				}

				tc.assertData(t, upload)
			})
		}
}

func TestNewTransactionUploadRejectsInvalidGroupID(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	upload, err := NewTransactionUpload(
		uploadID,
		valueobjects.GroupID{},
		"transactions.xlsx",
		"xlsx",
		"d41d8cd98f00b204e9800998ecf8427e",
		"local",
		"uploads/file.xlsx",
		"transaction-file-v1",
		valueobjects.UploadedTransactionUploadStatus(),
		7,
		time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
	)

	if !errors.Is(err, domain.ErrInvalidGroupID) {
		t.Fatalf("NewTransactionUpload() error = %v, want %v", err, domain.ErrInvalidGroupID)
	}

	if upload != nil {
		t.Fatalf("NewTransactionUpload() upload = %#v, want nil", upload)
	}
}

func TestReconstituteTransactionUploadStoresGroupID(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	upload, err := ReconstituteTransactionUpload(
		uploadID,
		groupID,
		"transactions.xlsx",
		"xlsx",
		"d41d8cd98f00b204e9800998ecf8427e",
		"local",
		"uploads/file.xlsx",
		"transaction-file-v1",
		valueobjects.UploadedTransactionUploadStatus(),
		7,
		time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("ReconstituteTransactionUpload() error = %v", err)
	}

	if upload.GroupID().String() != groupID.String() {
		t.Fatalf("upload.GroupID() = %q, want %q", upload.GroupID().String(), groupID.String())
	}
}

func TestTransactionUploadRecordDownloaded(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	upload, err := ReconstituteTransactionUpload(
		uploadID,
		groupID,
		"transactions.xlsx",
		"xlsx",
		"d41d8cd98f00b204e9800998ecf8427e",
		"local",
		"uploads/file.xlsx",
		"transaction-file-v1",
		valueobjects.UploadedTransactionUploadStatus(),
		7,
		time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("ReconstituteTransactionUpload() error = %v", err)
	}

	now := time.Date(2026, time.April, 3, 8, 30, 0, 0, time.UTC)
	if err := upload.RecordDownloaded(now, "01962b8f-aeb2-7e03-a8ff-1edce1300002", "01962b8f-aeb2-7e03-a8ff-1edce1300003"); err != nil {
		t.Fatalf("RecordDownloaded() error = %v", err)
	}

	domainEvents := upload.PullDomainEvents()
	if len(domainEvents) != 1 {
		t.Fatalf("len(upload.PullDomainEvents()) = %d, want %d", len(domainEvents), 1)
	}

	event, ok := domainEvents[0].(*events.AdminActionOccurred)
	if !ok {
		t.Fatalf("upload.PullDomainEvents()[0] type = %T, want %T", domainEvents[0], &events.AdminActionOccurred{})
	}

	if event.EventType() != events.DownloadTransactionUploadEventType {
		t.Fatalf("event.EventType() = %q, want %q", event.EventType(), events.DownloadTransactionUploadEventType)
	}

	var payload map[string]any
	if err := json.Unmarshal(event.EventData(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload["action"] != "download" {
		t.Fatalf("payload[action] = %v, want %q", payload["action"], "download")
	}

	if payload["resource"] != "transaction_upload" {
		t.Fatalf("payload[resource] = %v, want %q", payload["resource"], "transaction_upload")
	}

	if payload["upload_id"] != uploadID.String() {
		t.Fatalf("payload[upload_id] = %v, want %q", payload["upload_id"], uploadID.String())
	}

	if payload["file_name"] != "transactions.xlsx" {
		t.Fatalf("payload[file_name] = %v, want %q", payload["file_name"], "transactions.xlsx")
	}
}
