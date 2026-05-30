package valueobjects

import (
	"github.com/gyud-adb/paris-api/internal/domain"
)

type u1ListIDTag struct{}

// U1ListID is a value object representing a UUIDv7 U1 list identifier.
type U1ListID = UUIDv7ID[u1ListIDTag]

// NewU1ListID creates a new UUIDv7 U1 list identifier.
func NewU1ListID() (U1ListID, error) {
	return newUUIDv7ID[u1ListIDTag]()
}

// U1ListIDFromString parses and validates a UUIDv7 U1 list identifier.
func U1ListIDFromString(value string) (U1ListID, error) {
	parsedValue, err := parseUUIDv7ID[u1ListIDTag](value)
	if err != nil {
		return U1ListID{}, domain.ErrInvalidU1ListID
	}

	return parsedValue, nil
}
