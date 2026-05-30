package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// BugReportDescription is a non-empty trimmed string value object for a bug report description.
type BugReportDescription struct {
	value string
}

// NewBugReportDescription validates and builds a BugReportDescription.
func NewBugReportDescription(value string) (BugReportDescription, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return BugReportDescription{}, domain.ErrInvalidBugReportDescription
	}

	return BugReportDescription{value: normalized}, nil
}

// String returns the description value.
func (d BugReportDescription) String() string {
	return d.value
}
