package entities

import (
	"strings"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// Sector is a domain entity for an admin-managed sector list entry.
type Sector struct {
	aggregateRoot
	id          valueobjects.SectorID
	sectorType  valueobjects.SectorType
	name        string
	description string
	createdBy   valueobjects.UserID
	createdAt   time.Time
	updatedAt   time.Time
}

// RecordCreated records the sector creation event.
func (s *Sector) RecordCreated(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.CreateSectorEventType, map[string]any{
		"action":      "create",
		"resource":    "sector",
		"target_id":   s.ID().String(),
		"type":        s.Type(),
		"name":        s.Name(),
		"description": s.Description(),
	})
	if err != nil {
		return err
	}

	s.recordDomainEvent(event)
	return nil
}

// RecordRead records the sector read event.
func (s *Sector) RecordRead(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.GetSectorEventType, map[string]any{
		"action":    "read",
		"resource":  "sector",
		"target_id": s.ID().String(),
	})
	if err != nil {
		return err
	}

	s.recordDomainEvent(event)
	return nil
}

// RecordUpdated records the sector update event.
func (s *Sector) RecordUpdated(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.UpdateSectorEventType, map[string]any{
		"action":      "update",
		"resource":    "sector",
		"target_id":   s.ID().String(),
		"type":        s.Type(),
		"name":        s.Name(),
		"description": s.Description(),
	})
	if err != nil {
		return err
	}

	s.recordDomainEvent(event)
	return nil
}

// RecordDeleted records the sector deletion event.
func (s *Sector) RecordDeleted(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.DeleteSectorEventType, map[string]any{
		"action":    "delete",
		"resource":  "sector",
		"target_id": s.ID().String(),
	})
	if err != nil {
		return err
	}

	s.recordDomainEvent(event)
	return nil
}

// NewSector creates a new valid sector.
func NewSector(id valueobjects.SectorID, sectorType, name, description string) (*Sector, error) {
	normalizedType, normalizedName, normalizedDescription, err := normalizeSectorFields(sectorType, name, description)
	if err != nil {
		return nil, err
	}

	if _, err := valueobjects.SectorIDFromString(id.String()); err != nil {
		return nil, err
	}

	return &Sector{
		id:          id,
		sectorType:  normalizedType,
		name:        normalizedName,
		description: normalizedDescription,
	}, nil
}

// ReconstituteSector rebuilds a sector from storage.
func ReconstituteSector(id valueobjects.SectorID, sectorType valueobjects.SectorType, name, description string) *Sector {
	return ReconstituteSectorWithAudit(id, sectorType, name, description, valueobjects.UserID{}, time.Time{}, time.Time{})
}

// ReconstituteSectorWithAudit rebuilds a sector from storage with audit metadata.
func ReconstituteSectorWithAudit(id valueobjects.SectorID, sectorType valueobjects.SectorType, name, description string, createdBy valueobjects.UserID, createdAt, updatedAt time.Time) *Sector {
	return &Sector{
		id:          id,
		sectorType:  sectorType,
		name:        name,
		description: description,
		createdBy:   createdBy,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

// ID returns the sector identifier.
func (s *Sector) ID() valueobjects.SectorID {
	return s.id
}

// Type returns the sector classification type.
func (s *Sector) Type() string {
	return s.sectorType.String()
}

// Name returns the sector name.
func (s *Sector) Name() string {
	return s.name
}

// Description returns the sector description.
func (s *Sector) Description() string {
	return s.description
}

// CreatedAt returns the creation timestamp.
func (s *Sector) CreatedAt() time.Time {
	return s.createdAt
}

// CreatedBy returns the creator user identifier.
func (s *Sector) CreatedBy() string {
	return s.createdBy.String()
}

// SetCreatedBy sets the creator user identifier.
func (s *Sector) SetCreatedBy(createdBy valueobjects.UserID) {
	s.createdBy = createdBy
}

// UpdatedAt returns the update timestamp.
func (s *Sector) UpdatedAt() time.Time {
	return s.updatedAt
}

// SetAuditTimestamps sets the audit timestamps for the sector.
func (s *Sector) SetAuditTimestamps(createdAt, updatedAt time.Time) error {
	if err := validateTimestamp(createdAt); err != nil {
		return err
	}

	if err := validateTimestamp(updatedAt); err != nil {
		return err
	}

	s.createdAt = createdAt
	s.updatedAt = updatedAt

	return nil
}

// Touch updates the mutable audit timestamp.
func (s *Sector) Touch(updatedAt time.Time) error {
	if err := validateTimestamp(updatedAt); err != nil {
		return err
	}

	s.updatedAt = updatedAt

	return nil
}

// Update updates the mutable sector fields.
func (s *Sector) Update(sectorType, name, description string) error {
	normalizedType, normalizedName, normalizedDescription, err := normalizeSectorFields(sectorType, name, description)
	if err != nil {
		return err
	}

	s.sectorType = normalizedType
	s.name = normalizedName
	s.description = normalizedDescription

	return nil
}

// Equal reports whether two sectors share the same identity.
func (s *Sector) Equal(other *Sector) bool {
	if s == nil || other == nil {
		return false
	}

	return s.id.Equal(other.id)
}

func normalizeSectorFields(sectorType, name, description string) (valueobjects.SectorType, string, string, error) {
	normalizedType, err := valueobjects.SectorTypeFromString(sectorType)
	if err != nil {
		return valueobjects.SectorType{}, "", "", err
	}

	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return valueobjects.SectorType{}, "", "", domain.ErrInvalidSectorName
	}

	normalizedDescription := strings.TrimSpace(description)
	if normalizedDescription == "" {
		return valueobjects.SectorType{}, "", "", domain.ErrInvalidSectorDescription
	}

	return normalizedType, normalizedName, normalizedDescription, nil
}
