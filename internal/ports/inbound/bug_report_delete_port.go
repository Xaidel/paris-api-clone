package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// DeleteBugReportCommand requests the deletion of an existing bug report.
type DeleteBugReportCommand struct {
	ID           valueobjects.BugReportID
	ActorUserID  string
	ActorGroupID string
}

// DeleteBugReportResult returns the deleted bug report identifier.
type DeleteBugReportResult struct {
	ID string `json:"id"`
}

// DeleteBugReportPort deletes an existing bug report.
type DeleteBugReportPort interface {
	Execute(ctx context.Context, command DeleteBugReportCommand) (DeleteBugReportResult, error)
}
