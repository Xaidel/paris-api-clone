package entities

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// AdminEvent is an immutable audit log record for an admin action.
type AdminEvent struct {
	id         valueobjects.EventID
	occurredAt time.Time
	userID     string
	groupID    string
	eventType  string
	eventData  []byte
}

// NewAdminEvent creates a valid immutable admin event.
func NewAdminEvent(id valueobjects.EventID, occurredAt time.Time, userID, groupID, eventType string, eventData []byte) (*AdminEvent, error) {
	if _, err := valueobjects.EventIDFromString(id.String()); err != nil {
		return nil, err
	}

	if occurredAt.IsZero() {
		return nil, domain.ErrInvalidTimestamp
	}

	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return nil, domain.ErrInvalidActorUserID
	}

	normalizedGroupID := strings.TrimSpace(groupID)
	if normalizedGroupID == "" {
		return nil, domain.ErrInvalidActorGroupID
	}

	normalizedEventType := strings.TrimSpace(eventType)
	if normalizedEventType == "" {
		return nil, domain.ErrInvalidEventType
	}

	if !json.Valid(eventData) {
		return nil, domain.ErrInvalidEventData
	}

	dataCopy := append([]byte(nil), eventData...)

	return &AdminEvent{
		id:         id,
		occurredAt: occurredAt,
		userID:     normalizedUserID,
		groupID:    normalizedGroupID,
		eventType:  normalizedEventType,
		eventData:  dataCopy,
	}, nil
}

// ID returns the immutable event identifier.
func (e *AdminEvent) ID() valueobjects.EventID {
	return e.id
}

// OccurredAt returns the event timestamp.
func (e *AdminEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// UserID returns the acting admin user identifier.
func (e *AdminEvent) UserID() string {
	return e.userID
}

// GroupID returns the acting admin group identifier.
func (e *AdminEvent) GroupID() string {
	return e.groupID
}

// EventType returns the admin event type.
func (e *AdminEvent) EventType() string {
	return e.eventType
}

// EventData returns a defensive copy of the event payload.
func (e *AdminEvent) EventData() []byte {
	return append([]byte(nil), e.eventData...)
}
