package ports

import "context"

// DeleteU1ListCommand requests the deletion of an existing U1 list entry.
type DeleteU1ListCommand struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// DeleteU1ListResult returns the deleted U1 list identifier.
type DeleteU1ListResult struct {
	ID string `json:"id"`
}

// DeleteU1ListPort deletes an existing U1 list entry.
type DeleteU1ListPort interface {
	Execute(ctx context.Context, command DeleteU1ListCommand) (DeleteU1ListResult, error)
}
