package ports

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// U1ListFilter describes supported U1 list query filters.
type U1ListFilter struct {
	Sector string
}

// U1ListRepository persists U1 list entries.
type U1ListRepository interface {
	Create(ctx context.Context, entry *entities.U1ListEntry, createdByUserID string) error
	FindByID(ctx context.Context, id valueobjects.U1ListID) (*entities.U1ListEntry, error)
	List(ctx context.Context, filter U1ListFilter) ([]*entities.U1ListEntry, error)
	Update(ctx context.Context, entry *entities.U1ListEntry) error
	DeleteByID(ctx context.Context, id valueobjects.U1ListID) error
}
