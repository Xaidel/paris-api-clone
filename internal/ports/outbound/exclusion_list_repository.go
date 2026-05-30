package ports

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// ExclusionListRepository persists U2 exclusion list entries.
type ExclusionListRepository interface {
	Create(ctx context.Context, entry *entities.ExclusionListEntry, createdByUserID string) error
	FindByID(ctx context.Context, id valueobjects.ExclusionListID) (*entities.ExclusionListEntry, error)
	List(ctx context.Context) ([]*entities.ExclusionListEntry, error)
	Update(ctx context.Context, entry *entities.ExclusionListEntry) error
	DeleteByID(ctx context.Context, id valueobjects.ExclusionListID) error
}
