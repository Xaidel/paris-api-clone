package ports

import (
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

const (
	// TransactionClassifyTaskName identifies queued transaction classification work.
	TransactionClassifyTaskName = "transaction:classify"
	// TransactionClassifyReactTaskName identifies queued ReAct transaction classification work.
	TransactionClassifyReactTaskName = "transaction:classify-react"
)

// ClassificationJobPayload identifies a transaction queued for classification.
type ClassificationJobPayload struct {
	// TransactionID is serialized as a string so queue backends can persist it
	// without depending on domain value-object types.
	TransactionID string
}

// ClassificationJob describes one transaction classification queue task.
type ClassificationJob struct {
	// TaskName selects the worker flow that should process the job payload.
	TaskName string
	// Payload holds the transaction identifier to classify.
	Payload ClassificationJobPayload
	// QueuedAt records when the queue backend accepted the job when that metadata
	// is available.
	QueuedAt time.Time
}

// NewClassificationJob builds a transaction classification job.
func NewClassificationJob(transactionID valueobjects.TransactionID) ClassificationJob {
	return NewClassificationJobWithTaskName(TransactionClassifyReactTaskName, transactionID)
}

// NewClassificationJobWithTaskName builds a transaction classification job for one task type.
func NewClassificationJobWithTaskName(taskName string, transactionID valueobjects.TransactionID) ClassificationJob {
	return ClassificationJob{
		TaskName: taskName,
		Payload:  ClassificationJobPayload{TransactionID: transactionID.String()},
	}
}
