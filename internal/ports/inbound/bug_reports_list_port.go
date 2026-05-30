package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// ListBugReportsQuery requests all bug reports.
type ListBugReportsQuery struct {
	ActorUserID  string
	ActorGroupID string
}

// ListBugReportsPort lists all bug reports.
type ListBugReportsPort interface {
	Execute(ctx context.Context, query ListBugReportsQuery) (outboundports.ListBugReportsResult, error)
}
