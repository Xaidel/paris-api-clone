package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// ListGroupsQuery requests all groups.
type ListGroupsQuery struct {
	ActorUserID  string
	ActorGroupID string
}

// ListGroupsPort lists all groups.
type ListGroupsPort interface {
	Execute(ctx context.Context, query ListGroupsQuery) (outboundports.ListGroupsResult, error)
}
