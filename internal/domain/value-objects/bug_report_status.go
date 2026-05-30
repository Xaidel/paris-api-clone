package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

const (
	bugReportStatusOpen   = "Open"
	bugReportStatusClosed = "Closed"
)

// BugReportStatus is a value object describing a supported bug report status.
type BugReportStatus struct {
	value string
}

// OpenBugReportStatus returns the canonical Open bug report status..
func OpenBugReportStatus() BugReportStatus {
	return BugReportStatus{value: bugReportStatusOpen}
}

// ClosedBugReportStatus returns the canonical Closed bug report status.
func ClosedBugReportStatus() BugReportStatus {
	return BugReportStatus{value: bugReportStatusClosed}
}

// BugReportStatusFromString parses and normalizes a supported bug report status.
func BugReportStatusFromString(value string) (BugReportStatus, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case strings.ToLower(bugReportStatusOpen):
		return OpenBugReportStatus(), nil
	case strings.ToLower(bugReportStatusClosed):
		return ClosedBugReportStatus(), nil
	default:
		return BugReportStatus{}, domain.ErrInvalidBugReportStatus
	}
}

// String returns the canonical bug report status value.
func (t BugReportStatus) String() string {
	return t.value
}

// Equal reports whether two bug report statuses are equal.
func (t BugReportStatus) Equal(other BugReportStatus) bool {
	return t.value == other.value
}
