package events

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// AdminActionOccurred records an admin-facing action as a domain event.
type AdminActionOccurred struct {
	occurredAt   time.Time
	actorUserID  string
	actorGroupID string
	eventType    string
	eventData    []byte
}

// NewAdminActionOccurred builds an AdminActionOccurred event.
func NewAdminActionOccurred(occurredAt time.Time, actorUserID, actorGroupID, eventType string, eventData any) (*AdminActionOccurred, error) {
	if occurredAt.IsZero() {
		return nil, domain.ErrInvalidTimestamp
	}

	normalizedActorUserID := strings.TrimSpace(actorUserID)
	if normalizedActorUserID == "" {
		return nil, domain.ErrInvalidActorUserID
	}

	normalizedActorGroupID := strings.TrimSpace(actorGroupID)
	if normalizedActorGroupID == "" {
		return nil, domain.ErrInvalidActorGroupID
	}

	normalizedEventType := strings.TrimSpace(eventType)
	if normalizedEventType == "" {
		return nil, domain.ErrInvalidEventType
	}

	payload, err := json.Marshal(eventData)
	if err != nil {
		return nil, fmt.Errorf("marshal event payload: %w", err)
	}

	if !json.Valid(payload) {
		return nil, domain.ErrInvalidEventData
	}

	return &AdminActionOccurred{
		occurredAt:   occurredAt,
		actorUserID:  normalizedActorUserID,
		actorGroupID: normalizedActorGroupID,
		eventType:    normalizedEventType,
		eventData:    append([]byte(nil), payload...),
	}, nil
}

// EventName returns the event name.
func (e *AdminActionOccurred) EventName() string {
	return "AdminActionOccurred"
}

// OccurredAt returns the event timestamp.
func (e *AdminActionOccurred) OccurredAt() time.Time {
	return e.occurredAt
}

// ActorUserID returns the acting user identifier.
func (e *AdminActionOccurred) ActorUserID() string {
	return e.actorUserID
}

// ActorGroupID returns the acting group identifier.
func (e *AdminActionOccurred) ActorGroupID() string {
	return e.actorGroupID
}

// EventType returns the persisted event type.
func (e *AdminActionOccurred) EventType() string {
	return e.eventType
}

// EventData returns the persisted event payload.
func (e *AdminActionOccurred) EventData() []byte {
	return append([]byte(nil), e.eventData...)
}
