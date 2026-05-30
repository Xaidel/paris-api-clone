package usecases

import (
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func newExclusionListResult(entry *entities.ExclusionListEntry) ports.ExclusionListResult {
	return ports.ExclusionListResult{
		ID:           entry.ID().String(),
		ActivityType: entry.ActivityType(),
		CreatedBy:    entry.CreatedBy(),
		CreatedAt:    entry.CreatedAt(),
		UpdatedAt:    entry.UpdatedAt(),
	}
}
