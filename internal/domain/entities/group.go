package entities

import (
	"strings"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// Group is a domain entity for an admin-managed user group.
type Group struct {
	aggregateRoot
	id   valueobjects.GroupID
	name string
}

// NewGroup creates a new valid group.
func NewGroup(id valueobjects.GroupID, name string) (*Group, error) {
	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return nil, domain.ErrInvalidGroupName
	}

	if _, err := valueobjects.GroupIDFromString(id.String()); err != nil {
		return nil, err
	}

	return &Group{id: id, name: normalizedName}, nil
}

// ReconstituteGroup rebuilds a group from storage.
func ReconstituteGroup(id valueobjects.GroupID, name string) *Group {
	return &Group{id: id, name: name}
}

// RecordCreated records the group creation event.
func (g *Group) RecordCreated(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.CreateGroupEventType, map[string]any{
		"action":     "create",
		"resource":   "group",
		"target_id":  g.ID().String(),
		"group_name": g.Name(),
	})
	if err != nil {
		return err
	}

	g.recordDomainEvent(event)
	return nil
}

// RecordRead records the group read event.
func (g *Group) RecordRead(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.GetGroupEventType, map[string]any{
		"action":     "read",
		"resource":   "group",
		"target_id":  g.ID().String(),
		"group_name": g.Name(),
	})
	if err != nil {
		return err
	}

	g.recordDomainEvent(event)
	return nil
}

// RecordUpdated records the group update event.
func (g *Group) RecordUpdated(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.UpdateGroupEventType, map[string]any{
		"action":     "update",
		"resource":   "group",
		"target_id":  g.ID().String(),
		"group_name": g.Name(),
	})
	if err != nil {
		return err
	}

	g.recordDomainEvent(event)
	return nil
}

// RecordDeleted records the group deletion event.
func (g *Group) RecordDeleted(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.DeleteGroupEventType, map[string]any{
		"action":     "delete",
		"resource":   "group",
		"target_id":  g.ID().String(),
		"group_name": g.Name(),
	})
	if err != nil {
		return err
	}

	g.recordDomainEvent(event)
	return nil
}

// ID returns the group identifier.
func (g *Group) ID() valueobjects.GroupID {
	return g.id
}

// Name returns the group name.
func (g *Group) Name() string {
	return g.name
}

// Update updates the mutable group fields.
func (g *Group) Update(name string) error {
	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return domain.ErrInvalidGroupName
	}

	g.name = normalizedName
	return nil
}

// Equal reports whether two groups share the same identity.
func (g *Group) Equal(other *Group) bool {
	if g == nil || other == nil {
		return false
	}

	return g.id.Equal(other.id)
}
