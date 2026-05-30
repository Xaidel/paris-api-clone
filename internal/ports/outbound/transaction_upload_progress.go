package ports

import "context"

const (
	TransactionUploadProgressStatusReceived              = "received"
	TransactionUploadProgressStatusParsed                = "parsed"
	TransactionUploadProgressStatusValidated             = "validated"
	TransactionUploadProgressStatusStoredRawFile         = "stored_raw_file"
	TransactionUploadProgressStatusSavedUpload           = "saved_upload"
	TransactionUploadProgressStatusRecordedEvent         = "recorded_event"
	TransactionUploadProgressStatusPersistedTransactions = "persisted_transactions"
	TransactionUploadProgressStatusValidationFailed      = "validation_failed"
	TransactionUploadProgressStatusCompleted             = "completed"
)

// TransactionUploadProgressUpdate reports ingestion progress to callers.
type TransactionUploadProgressUpdate struct {
	Status           string
	Message          string
	Progress         int
	Upload           *TransactionUploadResult
	ValidationErrors []TransactionFileValidationError
	SkippedRows      []TransactionUploadSkippedRow
}

// TransactionUploadProgressReporter receives upload progress updates.
type TransactionUploadProgressReporter interface {
	Report(ctx context.Context, update TransactionUploadProgressUpdate) error
}
