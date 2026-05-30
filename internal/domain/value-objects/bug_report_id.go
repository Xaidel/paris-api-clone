// Package valueobjects
package valueobjects

import (
	"github.com/gyud-adb/paris-api/internal/domain"
)

type bugReportIDTag struct{}

// BugReportID is a value object representing a bug report identifier.
type BugReportID = UUIDv7ID[bugReportIDTag]

// NewBugReportID creates a new UUIDv7 bug report identifier.
func NewBugReportID() (BugReportID, error) {
	return newUUIDv7ID[bugReportIDTag]()
}

// BugReportIDFromString parses and validates a canonical UUIDv7 bug report identifier.
func BugReportIDFromString(value string) (BugReportID, error) {
	parsedValue, err := parseUUIDv7ID[bugReportIDTag](value)
	if err != nil {
		return BugReportID{}, domain.ErrInvalidBugReportID
	}

	return parsedValue, nil
}
