package ports

import "context"

// DeleteSectorCommand requests the deletion of an existing sector entry.
type DeleteSectorCommand struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// DeleteSectorResult returns the deleted sector identifier.
type DeleteSectorResult struct {
	ID string `json:"id"`
}

// DeleteSectorPort deletes an existing sector entry.
type DeleteSectorPort interface {
	Execute(ctx context.Context, command DeleteSectorCommand) (DeleteSectorResult, error)
}
