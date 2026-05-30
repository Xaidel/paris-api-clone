package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// CreateU1ListCommand requests the creation of a new U1 list entry.
type CreateU1ListCommand struct {
	Sector                string
	EligibleOperationType string
	ConditionGuidance     string
	ActorUserID           string
	ActorGroupID          string
}

// CreateU1ListPort creates a new U1 list entry.
type CreateU1ListPort interface {
	Execute(ctx context.Context, command CreateU1ListCommand) (outboundports.U1ListResult, error)
}
