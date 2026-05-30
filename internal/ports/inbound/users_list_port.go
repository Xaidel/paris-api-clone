package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// ListUsersQuery requests all users.
type ListUsersQuery struct {
	ActorUserID  string
	ActorGroupID string
}

// ListUsersResult returns the listed users.
type ListUsersResult struct {
	Users []outboundports.UserResult
}

// ListUsersPort lists all users.
type ListUsersPort interface {
	Execute(ctx context.Context, query ListUsersQuery) (ListUsersResult, error)
}
