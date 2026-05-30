package ports

import "context"

// DeleteTransactionUploadCommand requests deletion of an upload and its transactions.
type DeleteTransactionUploadCommand struct {
	ID           string
	ActorUserID  string
	ActorGroupID string
}

// DeleteTransactionUploadResult returns the deleted upload identifier.
type DeleteTransactionUploadResult struct {
	ID string
}

// DeleteTransactionUploadPort deletes an upload and its associated transactions.
type DeleteTransactionUploadPort interface {
	Execute(ctx context.Context, command DeleteTransactionUploadCommand) (DeleteTransactionUploadResult, error)
}
