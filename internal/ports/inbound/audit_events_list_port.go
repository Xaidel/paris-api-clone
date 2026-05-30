package ports

import (
	"context"
	"time"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// ListAuditEventsQuery requests a filtered list of audit events.
type ListAuditEventsQuery struct {
	EventOwner string
	EventType  string
	SessionID  string
	UserID     string
	StartedAt  *time.Time
	EndedAt    *time.Time
}

// ListAuditEventsPort lists audit events.
type ListAuditEventsPort interface {
	Execute(ctx context.Context, query ListAuditEventsQuery) (outboundports.ListAuditEventsResult, error)
}
