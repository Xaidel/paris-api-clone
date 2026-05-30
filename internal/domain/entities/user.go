package entities

import (
	"strings"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// User is a domain entity identified by UserID.
type User struct {
	aggregateRoot
	id           valueobjects.UserID
	username     string
	passwordHash string
	profile      valueobjects.UserProfile
	createdAt    time.Time
	updatedAt    time.Time
}

// NewUser creates a new valid user entity.
func NewUser(id valueobjects.UserID, username, passwordHash string, profile valueobjects.UserProfile, now time.Time) (*User, error) {
	normalizedUsername := strings.TrimSpace(username)
	if normalizedUsername == "" {
		return nil, domain.ErrInvalidUsername
	}

	normalizedPasswordHash := strings.TrimSpace(passwordHash)
	if normalizedPasswordHash == "" {
		return nil, domain.ErrInvalidPasswordHash
	}

	if now.IsZero() {
		return nil, domain.ErrInvalidTimestamp
	}

	return &User{
		id:           id,
		username:     normalizedUsername,
		passwordHash: normalizedPasswordHash,
		profile:      profile,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// ReconstituteUser rebuilds a user from storage.
func ReconstituteUser(id valueobjects.UserID, username, passwordHash string, profile valueobjects.UserProfile, createdAt, updatedAt time.Time) *User {
	return &User{
		id:           id,
		username:     username,
		passwordHash: passwordHash,
		profile:      profile,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// RecordCreated records the user creation event.
func (u *User) RecordCreated(actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(u.createdAt, actorUserID, actorGroupID, events.CreateUserEventType, map[string]any{
		"action":          "create",
		"resource":        "user",
		"target_user_id":  u.ID().String(),
		"target_username": u.Username(),
	})
	if err != nil {
		return err
	}

	u.recordDomainEvent(event)
	return nil
}

// RecordRead records the user read event.
func (u *User) RecordRead(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.GetUserEventType, map[string]any{
		"action":          "read",
		"resource":        "user",
		"target_user_id":  u.ID().String(),
		"target_username": u.Username(),
	})
	if err != nil {
		return err
	}

	u.recordDomainEvent(event)
	return nil
}

// RecordUpdated records the user update event.
func (u *User) RecordUpdated(actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(u.updatedAt, actorUserID, actorGroupID, events.UpdateUserEventType, map[string]any{
		"action":          "update",
		"resource":        "user",
		"target_user_id":  u.ID().String(),
		"target_username": u.Username(),
	})
	if err != nil {
		return err
	}

	u.recordDomainEvent(event)
	return nil
}

// RecordDeleted records the user deletion event.
func (u *User) RecordDeleted(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.DeleteUserEventType, map[string]any{
		"action":          "delete",
		"resource":        "user",
		"target_user_id":  u.ID().String(),
		"target_username": u.Username(),
	})
	if err != nil {
		return err
	}

	u.recordDomainEvent(event)
	return nil
}

// ID returns the user identifier.
func (u *User) ID() valueobjects.UserID {
	return u.id
}

// Username returns the username.
func (u *User) Username() string {
	return u.username
}

// PasswordHash returns the password hash.
func (u *User) PasswordHash() string {
	return u.passwordHash
}

// CreatedAt returns the creation timestamp.
func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

// Profile returns the user profile.
func (u *User) Profile() valueobjects.UserProfile {
	return u.profile
}

// UpdatedAt returns the update timestamp.
func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// Touch updates the entity timestamp.
func (u *User) Touch(now time.Time) error {
	if now.IsZero() {
		return domain.ErrInvalidTimestamp
	}

	u.updatedAt = now

	return nil
}

// Update updates the mutable user credentials and profile.
func (u *User) Update(username, passwordHash string, profile valueobjects.UserProfile, now time.Time) error {
	normalizedUsername := strings.TrimSpace(username)
	if normalizedUsername == "" {
		return domain.ErrInvalidUsername
	}

	normalizedPasswordHash := strings.TrimSpace(passwordHash)
	if normalizedPasswordHash == "" {
		return domain.ErrInvalidPasswordHash
	}

	if err := u.Touch(now); err != nil {
		return err
	}

	u.username = normalizedUsername
	u.passwordHash = normalizedPasswordHash
	u.profile = profile

	return nil
}

// Equal reports whether two users share the same identity.
func (u *User) Equal(other *User) bool {
	if u == nil || other == nil {
		return false
	}

	return u.id.Equal(other.id)
}
