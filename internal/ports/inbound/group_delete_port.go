package ports

import "context"

// DeleteGroupCommand requests the deletion of an existing group.
type DeleteGroupCommand struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// DeleteGroupResult returns the deleted group identifier.
type DeleteGroupResult struct {
	ID string
}

// DeleteGroupPort deletes an existing group.
type DeleteGroupPort interface {
	Execute(ctx context.Context, command DeleteGroupCommand) (DeleteGroupResult, error)
}
