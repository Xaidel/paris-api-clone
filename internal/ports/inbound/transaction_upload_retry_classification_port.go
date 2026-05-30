package ports

import "context"

// RetryTransactionUploadClassificationCommand requests manual retry of failed upload classifications.
type RetryTransactionUploadClassificationCommand struct {
	UploadID     string
	ActorUserID  string
	ActorGroupID string
}

// RetryTransactionUploadClassificationFailure describes one retry creation failure.
type RetryTransactionUploadClassificationFailure struct {
	TransactionID string `json:"transaction_id"`
	Error         string `json:"error"`
}

// RetryTransactionUploadClassificationSkippedTransaction describes one skipped retry.
type RetryTransactionUploadClassificationSkippedTransaction struct {
	TransactionID string `json:"transaction_id"`
	Reason        string `json:"reason"`
}

// RetryTransactionUploadClassificationResult reports the retry outcome for one upload.
type RetryTransactionUploadClassificationResult struct {
	UploadID                   string
	EligibleFailedTransactions int
	RetriedTransactions        int
	SkippedTransactions        int
	Skipped                    []RetryTransactionUploadClassificationSkippedTransaction
	FailedRetryCreations       int
	Failures                   []RetryTransactionUploadClassificationFailure
}

// RetryTransactionUploadClassificationPort retries failed transaction classifications for one upload.
type RetryTransactionUploadClassificationPort interface {
	Execute(ctx context.Context, command RetryTransactionUploadClassificationCommand) (RetryTransactionUploadClassificationResult, error)
}
