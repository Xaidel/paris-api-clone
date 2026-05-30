package ports

import "context"

// DeleteExclusionListCommand requests the deletion of an existing exclusion list entry.
type DeleteExclusionListCommand struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// DeleteExclusionListResult returns the deleted exclusion list identifier.
type DeleteExclusionListResult struct {
	ID string `json:"id"`
}

// DeleteExclusionListPort deletes an existing exclusion list entry.
type DeleteExclusionListPort interface {
	Execute(ctx context.Context, command DeleteExclusionListCommand) (DeleteExclusionListResult, error)
}
