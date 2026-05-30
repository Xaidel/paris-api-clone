package usecases

import (
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func newU1ListResult(entry *entities.U1ListEntry) ports.U1ListResult {
	return ports.U1ListResult{
		ID:                    entry.ID().String(),
		Sector:                entry.Sector(),
		EligibleOperationType: entry.EligibleOperationType(),
		ConditionGuidance:     entry.ConditionGuidance(),
		CreatedBy:             entry.CreatedBy(),
		CreatedAt:             entry.CreatedAt(),
		UpdatedAt:             entry.UpdatedAt(),
	}
}
