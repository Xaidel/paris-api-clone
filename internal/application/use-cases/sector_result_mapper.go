package usecases

import (
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func newSectorResult(sector *entities.Sector) ports.SectorResult {
	return ports.SectorResult{
		ID:          sector.ID().String(),
		Type:        sector.Type(),
		Name:        sector.Name(),
		Description: sector.Description(),
		CreatedBy:   sector.CreatedBy(),
		CreatedAt:   sector.CreatedAt(),
		UpdatedAt:   sector.UpdatedAt(),
	}
}
