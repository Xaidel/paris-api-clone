package usecases

import (
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func newTransactionUploadResult(upload *entities.TransactionUpload) ports.TransactionUploadResult {
	return ports.TransactionUploadResult{
		ID:              upload.ID().String(),
		GroupID:         upload.GroupID().String(),
		FileName:        upload.FileName(),
		FileFormat:      upload.FileFormat(),
		ContentMD5:      upload.ContentMD5(),
		StorageProvider: upload.StorageProvider(),
		StorageKey:      upload.StorageKey(),
		SchemaVersion:   upload.SchemaVersion(),
		Status:          upload.Status(),
		RowCount:        upload.RowCount(),
		UploadedAt:      upload.UploadedAt().UTC().Format(time.RFC3339),
	}
}
