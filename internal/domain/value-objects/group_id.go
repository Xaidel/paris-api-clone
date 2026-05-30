package valueobjects

import (
	"github.com/gyud-adb/paris-api/internal/domain"
)

type groupIDTag struct{}

// GroupID is a value object representing a UUIDv7 group identifier.
type GroupID = UUIDv7ID[groupIDTag]

// NewGroupID creates a new UUIDv7 group identifier.
func NewGroupID() (GroupID, error) {
	return newUUIDv7ID[groupIDTag]()
}

// GroupIDFromString parses and validates a UUIDv7 group identifier.
func GroupIDFromString(value string) (GroupID, error) {
	parsedValue, err := parseUUIDv7ID[groupIDTag](value)
	if err != nil {
		return GroupID{}, domain.ErrInvalidGroupID
	}

	return parsedValue, nil
}
