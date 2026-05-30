package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// GetBugReportQuery requests a single bug report by identifier.
type GetBugReportQuery struct {
	ID           valueobjects.BugReportID
	ActorUserID  string
	ActorGroupID string
}

// GetBugReportPort gets a single bug report.
type GetBugReportPort interface {
	Execute(ctx context.Context, query GetBugReportQuery) (outboundports.BugReportResult, error)
}
