package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionUploadPreviewRecord stores persisted preview data for one upload.
type TransactionUploadPreviewRecord struct {
	UploadID         string
	Columns          []string
	Rows             [][]string
	TotalRows        int
	ValidationErrors []TransactionFileValidationError
}

// TransactionUploadPreviewRepository persists and loads upload preview data.
type TransactionUploadPreviewRepository interface {
	Save(ctx context.Context, preview TransactionUploadPreviewRecord) error
	FindByUploadID(ctx context.Context, uploadID valueobjects.UploadID) (*TransactionUploadPreviewRecord, error)
}
