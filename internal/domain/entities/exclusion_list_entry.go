package entities

import (
	"strings"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// ExclusionListEntry is a domain entity for a U2 exclusion list entry.
type ExclusionListEntry struct {
	aggregateRoot
	id           valueobjects.ExclusionListID
	activityType string
	createdBy    valueobjects.UserID
	createdAt    time.Time
	updatedAt    time.Time
}

// RecordCreated records the exclusion list creation event.
func (e *ExclusionListEntry) RecordCreated(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.CreateExclusionListEventType, map[string]any{
		"action":        "create",
		"resource":      "exclusion_list",
		"target_id":     e.ID().String(),
		"activity_type": e.ActivityType(),
	})
	if err != nil {
		return err
	}

	e.recordDomainEvent(event)
	return nil
}

// RecordRead records the exclusion list read event.
func (e *ExclusionListEntry) RecordRead(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.GetExclusionListEventType, map[string]any{
		"action":    "read",
		"resource":  "exclusion_list",
		"target_id": e.ID().String(),
	})
	if err != nil {
		return err
	}

	e.recordDomainEvent(event)
	return nil
}

// RecordUpdated records the exclusion list update event.
func (e *ExclusionListEntry) RecordUpdated(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.UpdateExclusionListEventType, map[string]any{
		"action":        "update",
		"resource":      "exclusion_list",
		"target_id":     e.ID().String(),
		"activity_type": e.ActivityType(),
	})
	if err != nil {
		return err
	}

	e.recordDomainEvent(event)
	return nil
}

// RecordDeleted records the exclusion list deletion event.
func (e *ExclusionListEntry) RecordDeleted(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.DeleteExclusionListEventType, map[string]any{
		"action":    "delete",
		"resource":  "exclusion_list",
		"target_id": e.ID().String(),
	})
	if err != nil {
		return err
	}

	e.recordDomainEvent(event)
	return nil
}

// NewExclusionListEntry creates a new valid exclusion list entry.
func NewExclusionListEntry(id valueobjects.ExclusionListID, activityType string) (*ExclusionListEntry, error) {
	normalizedActivityType := strings.TrimSpace(activityType)
	if normalizedActivityType == "" {
		return nil, domain.ErrInvalidActivityType
	}

	if _, err := valueobjects.ExclusionListIDFromString(id.String()); err != nil {
		return nil, err
	}

	return &ExclusionListEntry{id: id, activityType: normalizedActivityType}, nil
}

// ReconstituteExclusionListEntry rebuilds an exclusion list entry from storage.
func ReconstituteExclusionListEntry(id valueobjects.ExclusionListID, activityType string) *ExclusionListEntry {
	return ReconstituteExclusionListEntryWithAudit(id, activityType, valueobjects.UserID{}, time.Time{}, time.Time{})
}

// ReconstituteExclusionListEntryWithAudit rebuilds an exclusion list entry with audit metadata.
func ReconstituteExclusionListEntryWithAudit(id valueobjects.ExclusionListID, activityType string, createdBy valueobjects.UserID, createdAt, updatedAt time.Time) *ExclusionListEntry {
	return &ExclusionListEntry{id: id, activityType: activityType, createdBy: createdBy, createdAt: createdAt, updatedAt: updatedAt}
}

// ID returns the exclusion list identifier.
func (e *ExclusionListEntry) ID() valueobjects.ExclusionListID {
	return e.id
}

// ActivityType returns the exclusion list activity type.
func (e *ExclusionListEntry) ActivityType() string {
	return e.activityType
}

// CreatedAt returns the creation timestamp.
func (e *ExclusionListEntry) CreatedAt() time.Time {
	return e.createdAt
}

// CreatedBy returns the creator user identifier.
func (e *ExclusionListEntry) CreatedBy() string {
	return e.createdBy.String()
}

// SetCreatedBy sets the creator user identifier.
func (e *ExclusionListEntry) SetCreatedBy(createdBy valueobjects.UserID) {
	e.createdBy = createdBy
}

// UpdatedAt returns the update timestamp.
func (e *ExclusionListEntry) UpdatedAt() time.Time {
	return e.updatedAt
}

// SetAuditTimestamps sets the audit timestamps for the entry.
func (e *ExclusionListEntry) SetAuditTimestamps(createdAt, updatedAt time.Time) error {
	if err := validateTimestamp(createdAt); err != nil {
		return err
	}

	if err := validateTimestamp(updatedAt); err != nil {
		return err
	}

	e.createdAt = createdAt
	e.updatedAt = updatedAt

	return nil
}

// Touch updates the mutable audit timestamp.
func (e *ExclusionListEntry) Touch(updatedAt time.Time) error {
	if err := validateTimestamp(updatedAt); err != nil {
		return err
	}

	e.updatedAt = updatedAt

	return nil
}

// UpdateActivityType updates the mutable activity type.
func (e *ExclusionListEntry) UpdateActivityType(activityType string) error {
	normalizedActivityType := strings.TrimSpace(activityType)
	if normalizedActivityType == "" {
		return domain.ErrInvalidActivityType
	}

	e.activityType = normalizedActivityType
	return nil
}

// Equal reports whether two exclusion list entries share the same identity.
func (e *ExclusionListEntry) Equal(other *ExclusionListEntry) bool {
	if e == nil || other == nil {
		return false
	}

	return e.id.Equal(other.id)
}
