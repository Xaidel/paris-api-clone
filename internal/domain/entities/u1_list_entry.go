package entities

import (
	"strings"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// U1ListEntry is a domain entity for a U1 list entry.
type U1ListEntry struct {
	aggregateRoot
	id                    valueobjects.U1ListID
	sector                string
	eligibleOperationType string
	conditionGuidance     string
	createdBy             valueobjects.UserID
	createdAt             time.Time
	updatedAt             time.Time
}

// RecordCreated records the U1 list creation event.
func (e *U1ListEntry) RecordCreated(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.CreateU1ListEventType, map[string]any{
		"action":                  "create",
		"resource":                "u1_list",
		"target_id":               e.ID().String(),
		"sector":                  e.Sector(),
		"eligible_operation_type": e.EligibleOperationType(),
		"condition_guidance":      e.ConditionGuidance(),
	})
	if err != nil {
		return err
	}

	e.recordDomainEvent(event)
	return nil
}

// RecordRead records the U1 list read event.
func (e *U1ListEntry) RecordRead(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.GetU1ListEventType, map[string]any{
		"action":    "read",
		"resource":  "u1_list",
		"target_id": e.ID().String(),
	})
	if err != nil {
		return err
	}

	e.recordDomainEvent(event)
	return nil
}

// RecordUpdated records the U1 list update event.
func (e *U1ListEntry) RecordUpdated(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.UpdateU1ListEventType, map[string]any{
		"action":                  "update",
		"resource":                "u1_list",
		"target_id":               e.ID().String(),
		"sector":                  e.Sector(),
		"eligible_operation_type": e.EligibleOperationType(),
		"condition_guidance":      e.ConditionGuidance(),
	})
	if err != nil {
		return err
	}

	e.recordDomainEvent(event)
	return nil
}

// RecordDeleted records the U1 list deletion event.
func (e *U1ListEntry) RecordDeleted(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.DeleteU1ListEventType, map[string]any{
		"action":    "delete",
		"resource":  "u1_list",
		"target_id": e.ID().String(),
	})
	if err != nil {
		return err
	}

	e.recordDomainEvent(event)
	return nil
}

// NewU1ListEntry creates a new valid U1 list entry.
func NewU1ListEntry(id valueobjects.U1ListID, sector, eligibleOperationType, conditionGuidance string) (*U1ListEntry, error) {
	normalizedSector, normalizedEligibleOperationType, normalizedConditionGuidance, err := normalizeU1ListFields(sector, eligibleOperationType, conditionGuidance)
	if err != nil {
		return nil, err
	}

	if _, err := valueobjects.U1ListIDFromString(id.String()); err != nil {
		return nil, err
	}

	return &U1ListEntry{
		id:                    id,
		sector:                normalizedSector,
		eligibleOperationType: normalizedEligibleOperationType,
		conditionGuidance:     normalizedConditionGuidance,
	}, nil
}

// ReconstituteU1ListEntry rebuilds a U1 list entry from storage.
func ReconstituteU1ListEntry(id valueobjects.U1ListID, sector, eligibleOperationType, conditionGuidance string) *U1ListEntry {
	return ReconstituteU1ListEntryWithAudit(id, sector, eligibleOperationType, conditionGuidance, valueobjects.UserID{}, time.Time{}, time.Time{})
}

// ReconstituteU1ListEntryWithAudit rebuilds a U1 list entry from storage with audit metadata.
func ReconstituteU1ListEntryWithAudit(id valueobjects.U1ListID, sector, eligibleOperationType, conditionGuidance string, createdBy valueobjects.UserID, createdAt, updatedAt time.Time) *U1ListEntry {
	return &U1ListEntry{
		id:                    id,
		sector:                sector,
		eligibleOperationType: eligibleOperationType,
		conditionGuidance:     conditionGuidance,
		createdBy:             createdBy,
		createdAt:             createdAt,
		updatedAt:             updatedAt,
	}
}

// ID returns the U1 list identifier.
func (e *U1ListEntry) ID() valueobjects.U1ListID {
	return e.id
}

// Sector returns the U1 list sector.
func (e *U1ListEntry) Sector() string {
	return e.sector
}

// EligibleOperationType returns the U1 list eligible operation type.
func (e *U1ListEntry) EligibleOperationType() string {
	return e.eligibleOperationType
}

// ConditionGuidance returns the U1 list condition guidance.
func (e *U1ListEntry) ConditionGuidance() string {
	return e.conditionGuidance
}

// CreatedAt returns the creation timestamp.
func (e *U1ListEntry) CreatedAt() time.Time {
	return e.createdAt
}

// CreatedBy returns the creator user identifier.
func (e *U1ListEntry) CreatedBy() string {
	return e.createdBy.String()
}

// SetCreatedBy sets the creator user identifier.
func (e *U1ListEntry) SetCreatedBy(createdBy valueobjects.UserID) {
	e.createdBy = createdBy
}

// UpdatedAt returns the update timestamp.
func (e *U1ListEntry) UpdatedAt() time.Time {
	return e.updatedAt
}

// SetAuditTimestamps sets the audit timestamps for the entry.
func (e *U1ListEntry) SetAuditTimestamps(createdAt, updatedAt time.Time) error {
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
func (e *U1ListEntry) Touch(updatedAt time.Time) error {
	if err := validateTimestamp(updatedAt); err != nil {
		return err
	}

	e.updatedAt = updatedAt

	return nil
}

// Update updates the mutable U1 list fields.
func (e *U1ListEntry) Update(sector, eligibleOperationType, conditionGuidance string) error {
	normalizedSector, normalizedEligibleOperationType, normalizedConditionGuidance, err := normalizeU1ListFields(sector, eligibleOperationType, conditionGuidance)
	if err != nil {
		return err
	}

	e.sector = normalizedSector
	e.eligibleOperationType = normalizedEligibleOperationType
	e.conditionGuidance = normalizedConditionGuidance

	return nil
}

// Equal reports whether two U1 list entries share the same identity.
func (e *U1ListEntry) Equal(other *U1ListEntry) bool {
	if e == nil || other == nil {
		return false
	}

	return e.id.Equal(other.id)
}

func normalizeU1ListFields(sector, eligibleOperationType, conditionGuidance string) (string, string, string, error) {
	normalizedSector := strings.TrimSpace(sector)
	if normalizedSector == "" {
		return "", "", "", domain.ErrInvalidSector
	}

	normalizedEligibleOperationType := strings.TrimSpace(eligibleOperationType)
	if normalizedEligibleOperationType == "" {
		return "", "", "", domain.ErrInvalidEligibleOperationType
	}

	normalizedConditionGuidance := strings.TrimSpace(conditionGuidance)
	if normalizedConditionGuidance == "" {
		return "", "", "", domain.ErrInvalidConditionGuidance
	}

	return normalizedSector, normalizedEligibleOperationType, normalizedConditionGuidance, nil
}
