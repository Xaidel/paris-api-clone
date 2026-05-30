package valueobjects

import (
	"github.com/gyud-adb/paris-api/internal/domain"
)

type exclusionListIDTag struct{}

// ExclusionListID is a value object representing a UUIDv7 exclusion list identifier.
type ExclusionListID = UUIDv7ID[exclusionListIDTag]

// NewExclusionListID creates a new UUIDv7 exclusion list identifier.
func NewExclusionListID() (ExclusionListID, error) {
	return newUUIDv7ID[exclusionListIDTag]()
}

// ExclusionListIDFromString parses and validates a UUIDv7 exclusion list identifier.
func ExclusionListIDFromString(value string) (ExclusionListID, error) {
	parsedValue, err := parseUUIDv7ID[exclusionListIDTag](value)
	if err != nil {
		return ExclusionListID{}, domain.ErrInvalidExclusionListID
	}

	return parsedValue, nil
}
