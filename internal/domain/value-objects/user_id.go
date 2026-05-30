package valueobjects

import (
	"github.com/gyud-adb/paris-api/internal/domain"
)

type userIDTag struct{}

// UserID is a value object representing a UUIDv7 user identifier.
type UserID = UUIDv7ID[userIDTag]

// NewUserID creates a new UUIDv7 user identifier.
func NewUserID() (UserID, error) {
	return newUUIDv7ID[userIDTag]()
}

// UserIDFromString parses and validates a canonical UUIDv7 user identifier.
func UserIDFromString(value string) (UserID, error) {
	parsedValue, err := parseUUIDv7ID[userIDTag](value)
	if err != nil {
		return UserID{}, domain.ErrInvalidUserID
	}

	return parsedValue, nil
}
