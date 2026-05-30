package ports

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// SectorRepository persists sector entries.
type SectorRepository interface {
	Create(ctx context.Context, sector *entities.Sector, createdByUserID string) error
	FindByID(ctx context.Context, id valueobjects.SectorID) (*entities.Sector, error)
	List(ctx context.Context) ([]*entities.Sector, error)
	Update(ctx context.Context, sector *entities.Sector) error
	DeleteByID(ctx context.Context, id valueobjects.SectorID) error
}
