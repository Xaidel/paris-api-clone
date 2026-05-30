package ports

import "context"

// DeleteTransactionCommand requests the deletion of an existing transaction.
type DeleteTransactionCommand struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// DeleteTransactionResult returns the deleted transaction identifier.
type DeleteTransactionResult struct {
	ID string `json:"id"`
}

// DeleteTransactionPort deletes an existing transaction.
type DeleteTransactionPort interface {
	Execute(ctx context.Context, command DeleteTransactionCommand) (DeleteTransactionResult, error)
}
