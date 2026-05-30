package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// CreateBugReportCommand requests the creation of a new bug report.
type CreateBugReportCommand struct {
	TransactionID valueobjects.TransactionID
	Title         valueobjects.BugReportTitle
	Description   valueobjects.BugReportDescription
	ActorUserID   string
	ActorGroupID  string
}

// CreateBugReportPort creates a new bug report.
type CreateBugReportPort interface {
	Execute(ctx context.Context, command CreateBugReportCommand) (outboundports.BugReportResult, error)
}
