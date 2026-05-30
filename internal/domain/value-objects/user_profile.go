package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// UserProfile is a value object describing personal user information.
type UserProfile struct {
	firstName     string
	middleName    string
	hasMiddleName bool
	lastName      string
	groupID       GroupID
}

// NewUserProfile creates a new valid user profile.
func NewUserProfile(firstName string, middleName *string, lastName string, groupID GroupID) (UserProfile, error) {
	normalizedFirstName := strings.TrimSpace(firstName)
	if normalizedFirstName == "" {
		return UserProfile{}, domain.ErrInvalidFirstName
	}

	normalizedLastName := strings.TrimSpace(lastName)
	if normalizedLastName == "" {
		return UserProfile{}, domain.ErrInvalidLastName
	}

	if _, err := GroupIDFromString(groupID.String()); err != nil {
		return UserProfile{}, err
	}

	normalizedMiddleName, hasMiddleName := normalizeMiddleName(middleName)

	return UserProfile{
		firstName:     normalizedFirstName,
		middleName:    normalizedMiddleName,
		hasMiddleName: hasMiddleName,
		lastName:      normalizedLastName,
		groupID:       groupID,
	}, nil
}

// FirstName returns the profile first name.
func (p UserProfile) FirstName() string {
	return p.firstName
}

// MiddleName returns the optional profile middle name.
func (p UserProfile) MiddleName() *string {
	if !p.hasMiddleName {
		return nil
	}

	middleName := p.middleName
	return &middleName
}

// LastName returns the profile last name.
func (p UserProfile) LastName() string {
	return p.lastName
}

// GroupID returns the profile group identifier.
func (p UserProfile) GroupID() GroupID {
	return p.groupID
}

func normalizeMiddleName(middleName *string) (string, bool) {
	if middleName == nil {
		return "", false
	}

	normalizedMiddleName := strings.TrimSpace(*middleName)
	if normalizedMiddleName == "" {
		return "", false
	}

	return normalizedMiddleName, true
}
