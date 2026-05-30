package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// GetTransactionUploadPreviewQuery requests persisted preview data for one upload.
type GetTransactionUploadPreviewQuery struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// GetTransactionUploadPreviewResult exposes one persisted upload preview.
type GetTransactionUploadPreviewResult struct {
	FileID           string
	FileName         string
	Columns          []string
	Rows             [][]string
	TotalRows        int
	ValidationErrors []outboundports.TransactionFileValidationError
}

// GetTransactionUploadPreviewPort gets persisted preview data for one upload.
type GetTransactionUploadPreviewPort interface {
	Execute(ctx context.Context, query GetTransactionUploadPreviewQuery) (GetTransactionUploadPreviewResult, error)
}
