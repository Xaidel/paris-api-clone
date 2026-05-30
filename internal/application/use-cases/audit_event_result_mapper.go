package usecases

import (
	"encoding/json"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func newAuditEventResult(event *entities.AdminEvent) ports.AuditEventResult {
	return ports.AuditEventResult{
		ID:         event.ID().String(),
		Timestamp:  event.OccurredAt(),
		EventOwner: "user",
		EventType:  event.EventType(),
		UserID:     event.UserID(),
		GroupID:    event.GroupID(),
		EventData:  json.RawMessage(event.EventData()),
	}
}
