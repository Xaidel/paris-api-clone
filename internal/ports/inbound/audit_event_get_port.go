package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// GetAuditEventQuery requests a single audit event by identifier.
type GetAuditEventQuery struct {
	ID string
}

// GetAuditEventPort gets a single audit event.
type GetAuditEventPort interface {
	Execute(ctx context.Context, query GetAuditEventQuery) (outboundports.AuditEventResult, error)
}
