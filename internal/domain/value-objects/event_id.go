package valueobjects

import (
	"github.com/gyud-adb/paris-api/internal/domain"
)

type eventIDTag struct{}

// EventID is a value object representing a UUIDv7 event identifier.
type EventID = UUIDv7ID[eventIDTag]

// NewEventID creates a new UUIDv7 event identifier.
func NewEventID() (EventID, error) {
	return newUUIDv7ID[eventIDTag]()
}

// EventIDFromString parses and validates a UUIDv7 event identifier.
func EventIDFromString(value string) (EventID, error) {
	parsedValue, err := parseUUIDv7ID[eventIDTag](value)
	if err != nil {
		return EventID{}, domain.ErrInvalidEventID
	}

	return parsedValue, nil
}
