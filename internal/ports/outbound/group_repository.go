package ports

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// GroupRepository persists groups.
type GroupRepository interface {
	Create(ctx context.Context, group *entities.Group) error
	FindByID(ctx context.Context, id valueobjects.GroupID) (*entities.Group, error)
	List(ctx context.Context) ([]*entities.Group, error)
	Update(ctx context.Context, group *entities.Group) error
	DeleteByID(ctx context.Context, id valueobjects.GroupID) error
}
