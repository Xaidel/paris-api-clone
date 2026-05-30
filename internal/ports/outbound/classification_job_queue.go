package ports

import "context"

// ClassificationJobQueue dequeues queued transaction classification jobs.
type ClassificationJobQueue interface {
	Dequeue(ctx context.Context, taskName string) (*ClassificationJob, error)
	DequeueBatch(ctx context.Context, taskName string, limit int) ([]ClassificationJob, error)
	Complete(ctx context.Context, job ClassificationJob) error
}
