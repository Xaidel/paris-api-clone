package ports

import (
	"context"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionUploadFilter describes supported upload history query filters.
type TransactionUploadFilter struct {
	GroupID   valueobjects.GroupID
	FileName  string
	StartedAt *time.Time
	EndedAt   *time.Time
}

// TransactionUploadRepository persists and queries transaction uploads.
type TransactionUploadRepository interface {
	Create(ctx context.Context, upload *entities.TransactionUpload) error
	FindByID(ctx context.Context, id valueobjects.UploadID) (*entities.TransactionUpload, error)
	FindByContentMD5(ctx context.Context, contentMD5 string, groupID valueobjects.GroupID) (*entities.TransactionUpload, error)
	List(ctx context.Context, filter TransactionUploadFilter) ([]*entities.TransactionUpload, error)
	DeleteByID(ctx context.Context, id valueobjects.UploadID) error
}
