package usecases

import (
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

func testTime() time.Time {
	return time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
}

func testGroupID(t *testing.T) valueobjects.GroupID {
	t.Helper()

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	return groupID
}

func testReconstitutedTransactionUpload(t *testing.T, uploadID valueobjects.UploadID, status valueobjects.TransactionUploadStatus, rowCount int) *entities.TransactionUpload {
	t.Helper()

	upload, err := entities.ReconstituteTransactionUpload(
		uploadID,
		testGroupID(t),
		"transactions.csv",
		"csv",
		"d41d8cd98f00b204e9800998ecf8427e",
		"local",
		"upload/file.csv",
		"transaction-file-v1",
		status,
		rowCount,
		testTime(),
	)
	if err != nil {
		t.Fatalf("ReconstituteTransactionUpload() error = %v", err)
	}

	return upload
}
