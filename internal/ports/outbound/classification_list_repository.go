package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// ClassificationListRepository loads classification list entries by list type.
type ClassificationListRepository interface {
	GetEntries(ctx context.Context, listType valueobjects.ListType) ([]string, error)
	GetEntryDocuments(ctx context.Context, listType valueobjects.ListType) ([]valueobjects.ClassificationListEntryDocument, error)
}
