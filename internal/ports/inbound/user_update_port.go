package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// UpdateUserCommand requests an update for an existing user.
type UpdateUserCommand struct {
	ID           string
	Username     string
	Password     string
	FirstName    string
	MiddleName   *string
	LastName     string
	GroupID      string
	ActorUserID  string
	ActorGroupID string
}

// UpdateUserPort updates an existing user.
type UpdateUserPort interface {
	Execute(ctx context.Context, command UpdateUserCommand) (outboundports.UserResult, error)
}
