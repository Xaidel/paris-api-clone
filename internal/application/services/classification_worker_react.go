package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	"go.uber.org/zap"
)

const reactClassificationWorkerIdlePollInterval = 250 * time.Millisecond

const defaultReactClassificationWorkerBatchSize = 10

const reactClassificationWorkerBatchScanMultiplier = 10

// ReActClassificationWorker polls the queue table and dispatches ReAct classification jobs.
type ReActClassificationWorker struct {
	queue        ports.ClassificationJobQueue
	handler      *ReActClassificationJobHandler
	txManager    ports.TransactionManager
	logger       *zap.Logger
	interval     time.Duration
	batchSize    int
	flushTimeout time.Duration
	now          func() time.Time
}

// NewReActClassificationWorker builds a ReActClassificationWorker.
func NewReActClassificationWorker(queue ports.ClassificationJobQueue, handler *ReActClassificationJobHandler, txManager ports.TransactionManager, logger *zap.Logger, opts ...ReActClassificationWorkerOption) *ReActClassificationWorker {
	worker := &ReActClassificationWorker{
		queue:        queue,
		handler:      handler,
		txManager:    txManager,
		logger:       logger,
		interval:     reactClassificationWorkerIdlePollInterval,
		batchSize:    defaultReactClassificationWorkerBatchSize,
		flushTimeout: 2 * time.Second,
		now:          time.Now,
	}

	for _, opt := range opts {
		opt(worker)
	}

	return worker
}

// ReActClassificationWorkerOption configures the ReAct worker.
type ReActClassificationWorkerOption func(*ReActClassificationWorker)

// WithReActClassificationWorkerBatchSize configures the max batch size.
func WithReActClassificationWorkerBatchSize(batchSize int) ReActClassificationWorkerOption {
	return func(worker *ReActClassificationWorker) {
		worker.batchSize = batchSize
	}
}

// WithReActClassificationWorkerFlushTimeout configures the max batch wait time.
func WithReActClassificationWorkerFlushTimeout(flushTimeout time.Duration) ReActClassificationWorkerOption {
	return func(worker *ReActClassificationWorker) {
		worker.flushTimeout = flushTimeout
	}
}

// WithReActClassificationWorkerNow configures the clock.
func WithReActClassificationWorkerNow(now func() time.Time) ReActClassificationWorkerOption {
	return func(worker *ReActClassificationWorker) {
		worker.now = now
	}
}

// Run polls for queued ReAct work until the context is cancelled.
func (w *ReActClassificationWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		processed, err := w.processNext(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return nil
			}

			return err
		}

		if processed {
			continue
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (w *ReActClassificationWorker) processNext(ctx context.Context) (bool, error) {
	if w.txManager == nil {
		return false, fmt.Errorf("react classification worker transaction manager is required")
	}

	processed := false
	requiresCompletionAfterRollback := false
	jobsToComplete := make([]ports.ClassificationJob, 0)
	err := w.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		batchJobs, err := w.selectBatch(txCtx)
		if err != nil {
			return err
		}

		if len(batchJobs) == 0 {
			return nil
		}
		processed = true

		if err := w.handler.HandleBatch(txCtx, batchJobs); err != nil {
			var handledErr *reactClassificationHandledError
			if errors.As(err, &handledErr) {
				w.warnHandledFailure(batchJobs, handledErr)
				if handledErr.requiresRollback {
					requiresCompletionAfterRollback = true
					jobsToComplete = append(jobsToComplete[:0], batchJobs...)
					return handledErr
				}
			} else {
				transactionIDs := make([]string, 0, len(batchJobs))
				for _, job := range batchJobs {
					transactionIDs = append(transactionIDs, job.Payload.TransactionID)
				}

				w.warn("react classification job failed",
					zap.String("task_name", ports.TransactionClassifyReactTaskName),
					zap.Strings("transaction_ids", transactionIDs),
					zap.Error(err),
				)

				return err
			}
		}

		for _, job := range batchJobs {
			if err := w.queue.Complete(txCtx, job); err != nil {
				return fmt.Errorf("completing react classification job for transaction %s: %w", job.Payload.TransactionID, err)
			}
		}

		return nil
	})
	if err != nil {
		if requiresCompletionAfterRollback {
			if completionErr := w.completeHandledJobs(ctx, jobsToComplete); completionErr != nil {
				return processed, completionErr
			}

			return processed, nil
		}

		return processed, err
	}

	return processed, nil
}

func (w *ReActClassificationWorker) completeHandledJobs(ctx context.Context, jobs []ports.ClassificationJob) error {
	if len(jobs) == 0 {
		return nil
	}

	return w.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		for _, job := range jobs {
			if err := w.queue.Complete(txCtx, job); err != nil {
				return fmt.Errorf("completing handled react classification job for transaction %s: %w", job.Payload.TransactionID, err)
			}
		}

		return nil
	})
}

func (w *ReActClassificationWorker) dequeueLimit() int {
	if w == nil || w.batchSize <= 0 {
		return defaultReactClassificationWorkerBatchSize
	}

	return w.batchSize
}

func (w *ReActClassificationWorker) dequeueScanLimit() int {
	limit := w.dequeueLimit() * reactClassificationWorkerBatchScanMultiplier
	if limit < w.dequeueLimit() {
		return w.dequeueLimit()
	}

	return limit
}

func (w *ReActClassificationWorker) selectBatch(ctx context.Context) ([]ports.ClassificationJob, error) {
	selected := make([]ports.ClassificationJob, 0, w.dequeueLimit())
	uniqueDescriptions := make(map[string]struct{}, w.dequeueLimit())
	oldestQueuedAt := time.Time{}
	jobs, err := w.queue.DequeueBatch(ctx, ports.TransactionClassifyReactTaskName, w.dequeueScanLimit())
	if err != nil {
		return nil, fmt.Errorf("dequeueing react classification jobs: %w", err)
	}

	if len(jobs) == 0 {
		return nil, nil
	}

	for _, job := range jobs {
		descriptionKey, err := w.jobDescriptionKey(ctx, job)
		if err != nil {
			return nil, err
		}

		if oldestQueuedAt.IsZero() || job.QueuedAt.Before(oldestQueuedAt) {
			oldestQueuedAt = job.QueuedAt
		}

		if len(uniqueDescriptions) >= w.dequeueLimit() {
			if _, exists := uniqueDescriptions[descriptionKey]; !exists {
				return selected, nil
			}
		}

		selected = append(selected, job)
		uniqueDescriptions[descriptionKey] = struct{}{}
	}

	if len(selected) == 0 {
		return nil, nil
	}

	if len(uniqueDescriptions) >= w.dequeueLimit() {
		return selected, nil
	}

	if w.flushTimeout > 0 && !oldestQueuedAt.IsZero() && !w.now().Before(oldestQueuedAt.Add(w.flushTimeout)) {
		return selected, nil
	}

	return nil, nil
}

func (w *ReActClassificationWorker) jobDescriptionKey(ctx context.Context, job ports.ClassificationJob) (string, error) {
	if w == nil || w.handler == nil || w.handler.txRepo == nil {
		return job.Payload.TransactionID, nil
	}

	transactionID, err := valueobjects.TransactionIDFromString(job.Payload.TransactionID)
	if err != nil {
		return "", fmt.Errorf("parsing transaction id for react batch selection: %w", err)
	}

	transaction, err := w.handler.txRepo.FindByID(ctx, transactionID)
	if err != nil {
		return "", fmt.Errorf("loading transaction %s for react batch selection: %w", transactionID.String(), err)
	}

	if transaction == nil {
		return job.Payload.TransactionID, nil
	}

	return transaction.GoodsDescription(), nil
}

func (w *ReActClassificationWorker) warn(message string, fields ...zap.Field) {
	if w == nil || w.logger == nil {
		return
	}

	w.logger.Warn(message, fields...)
}

func (w *ReActClassificationWorker) warnHandledFailure(batchJobs []ports.ClassificationJob, err error) {
	transactionIDs := make([]string, 0, len(batchJobs))
	for _, job := range batchJobs {
		transactionIDs = append(transactionIDs, job.Payload.TransactionID)
	}

	w.warn("react classification job failed and was persisted",
		zap.String("task_name", ports.TransactionClassifyReactTaskName),
		zap.Strings("transaction_ids", transactionIDs),
		zap.Error(err),
	)
}
