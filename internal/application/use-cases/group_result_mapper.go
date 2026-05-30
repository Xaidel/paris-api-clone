package usecases

import (
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func newGroupResult(group *entities.Group) ports.GroupResult {
	return ports.GroupResult{
		ID:   group.ID().String(),
		Name: group.Name(),
	}
}
