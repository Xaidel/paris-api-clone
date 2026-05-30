package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// BugReportTitle is a non-empty trimmed string value object for a bug report title.
type BugReportTitle struct {
	value string
}

// NewBugReportTitle validates and builds a BugReportTitle.
func NewBugReportTitle(value string) (BugReportTitle, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return BugReportTitle{}, domain.ErrInvalidBugReportTitle
	}

	return BugReportTitle{value: normalized}, nil
}

// String returns the title value.
func (t BugReportTitle) String() string {
	return t.value
}
