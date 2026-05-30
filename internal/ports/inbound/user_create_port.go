package ports

import (
	"context"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// CreateUserCommand requests the creation of a new user.
type CreateUserCommand struct {
	Username     string
	Password     string
	FirstName    string
	MiddleName   *string
	LastName     string
	GroupID      string
	ActorUserID  string
	ActorGroupID string
}

// CreateUserPort creates a new user.
type CreateUserPort interface {
	Execute(ctx context.Context, command CreateUserCommand) (outboundports.UserResult, error)
}
