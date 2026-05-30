package ports

import "context"

// DeleteUserCommand requests the deletion of an existing user.
type DeleteUserCommand struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// DeleteUserResult returns the deleted user identifier.
type DeleteUserResult struct {
	ID string
}

// DeleteUserPort deletes an existing user.
type DeleteUserPort interface {
	Execute(ctx context.Context, command DeleteUserCommand) (DeleteUserResult, error)
}
