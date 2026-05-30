package valueobjects

import (
	"github.com/gyud-adb/paris-api/internal/domain"
)

type uploadIDTag struct{}

// UploadID is a value object representing a UUIDv7 upload identifier.
type UploadID = UUIDv7ID[uploadIDTag]

// NewUploadID creates a new UUIDv7 upload identifier.
func NewUploadID() (UploadID, error) {
	return newUUIDv7ID[uploadIDTag]()
}

// UploadIDFromString parses and validates a UUIDv7 upload identifier.
func UploadIDFromString(value string) (UploadID, error) {
	parsedValue, err := parseUUIDv7ID[uploadIDTag](value)
	if err != nil {
		return UploadID{}, domain.ErrInvalidUploadID
	}

	return parsedValue, nil
}
