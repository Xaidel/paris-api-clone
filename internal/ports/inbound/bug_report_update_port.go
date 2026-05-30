// Package ports
package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// UpdateBugReportCommand requests an update for an existing bug report.
type UpdateBugReportCommand struct {
	ID           valueobjects.BugReportID
	Title        valueobjects.BugReportTitle
	Description  valueobjects.BugReportDescription
	Status       valueobjects.BugReportStatus
	ActorUserID  string
	ActorGroupID string
}

// UpdateBugReportPort updates an existing bug report.
type UpdateBugReportPort interface {
	Execute(ctx context.Context, command UpdateBugReportCommand) (outboundports.BugReportResult, error)
}
