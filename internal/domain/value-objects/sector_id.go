package valueobjects

import (
	"github.com/gyud-adb/paris-api/internal/domain"
)

type sectorIDTag struct{}

// SectorID is a value object representing a UUIDv7 sector identifier.
type SectorID = UUIDv7ID[sectorIDTag]

// NewSectorID creates a new UUIDv7 sector identifier.
func NewSectorID() (SectorID, error) {
	return newUUIDv7ID[sectorIDTag]()
}

// SectorIDFromString parses and validates a UUIDv7 sector identifier.
func SectorIDFromString(value string) (SectorID, error) {
	parsedValue, err := parseUUIDv7ID[sectorIDTag](value)
	if err != nil {
		return SectorID{}, domain.ErrInvalidSectorID
	}

	return parsedValue, nil
}
