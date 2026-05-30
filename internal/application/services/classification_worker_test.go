package services

import (
	"context"
	"errors"
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	"go.uber.org/zap"
)

type classificationJobQueueStub struct {
	jobs          []*ports.ClassificationJob
	batchJobs     []ports.ClassificationJob
	batchJobPages [][]ports.ClassificationJob
	dequeueErr    error
	completeErr   error
	completedJobs []ports.ClassificationJob
}

type transactionManagerServicesStub struct{}

func (transactionManagerServicesStub) WithinTransaction(ctx context.Context, operation func(ctx context.Context) error) error {
	return operation(ctx)
}

func (s *classificationJobQueueStub) Dequeue(_ context.Context, _ string) (*ports.ClassificationJob, error) {
	if s.dequeueErr != nil {
		return nil, s.dequeueErr
	}

	if len(s.jobs) == 0 {
		return nil, nil
	}

	job := s.jobs[0]
	s.jobs = s.jobs[1:]
	return job, nil
}

func (s *classificationJobQueueStub) Complete(_ context.Context, job ports.ClassificationJob) error {
	s.completedJobs = append(s.completedJobs, job)
	return s.completeErr
}

func (s *classificationJobQueueStub) DequeueBatch(_ context.Context, _ string, limit int) ([]ports.ClassificationJob, error) {
	if s.dequeueErr != nil {
		return nil, s.dequeueErr
	}

	if len(s.batchJobPages) > 0 {
		jobs := append([]ports.ClassificationJob(nil), s.batchJobPages[0]...)
		s.batchJobPages = s.batchJobPages[1:]
		if limit > 0 && len(jobs) > limit {
			jobs = jobs[:limit]
		}
		return jobs, nil
	}

	if len(s.batchJobs) == 0 {
		return nil, nil
	}

	if limit > len(s.batchJobs) {
		limit = len(s.batchJobs)
	}

	jobs := append([]ports.ClassificationJob(nil), s.batchJobs[:limit]...)
	s.batchJobs = s.batchJobs[limit:]
	return jobs, nil
}

// This test ensures the ReAct worker waits for more work when the batch has
// not reached its unique-description threshold and the flush timeout has not
// yet elapsed.
func TestReActClassificationWorkerDefersUntilFlushTimeoutWhenUniqueDescriptionsBelowBatchSize(t *testing.T) {
	t.Parallel()

	queuedAt := time.Date(2026, time.April, 22, 10, 0, 0, 0, time.UTC)
	queue := &classificationJobQueueStub{batchJobs: []ports.ClassificationJob{{
		TaskName: ports.TransactionClassifyReactTaskName,
		Payload:  ports.ClassificationJobPayload{TransactionID: "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60"},
		QueuedAt: queuedAt,
	}}}
	handler := &ReActClassificationJobHandler{
		txRepo: &reactSelectBatchRepositoryStub{transactionsByID: map[string]*entities.Transaction{
			"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60": mustNewHandlerTransaction(t, mustTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60"), mustClassification(t, "unclassified"), mustStatus(t, "processing"), nil, ""),
			"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61": mustNewHandlerTransaction(t, mustTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"), mustClassification(t, "unclassified"), mustStatus(t, "processing"), nil, ""),
		}},
		gateway: &reactClassificationGatewayStub{},
		logger:  zap.NewNop(),
		now:     testTime,
	}
	worker := NewReActClassificationWorker(queue, handler, transactionManagerServicesStub{}, zap.NewNop(), WithReActClassificationWorkerNow(func() time.Time {
		return queuedAt.Add(1500 * time.Millisecond)
	}))

	processed, err := worker.processNext(context.Background())
	if err != nil {
		t.Fatalf("processNext() error = %v", err)
	}

	if processed {
		t.Fatal("processNext() processed = true, want false")
	}
}

// This test verifies duplicate descriptions are still batched together before
// the flush timeout because they can share one classifier request.
func TestReActClassificationWorkerSelectBatchIncludesDuplicateDescriptionsBeforeFlushTimeout(t *testing.T) {
	t.Parallel()

	queuedAt := time.Date(2026, time.April, 22, 10, 0, 0, 0, time.UTC)
	queue := &classificationJobQueueStub{batchJobs: []ports.ClassificationJob{{
		TaskName: ports.TransactionClassifyReactTaskName,
		Payload:  ports.ClassificationJobPayload{TransactionID: "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60"},
		QueuedAt: queuedAt,
	}, {
		TaskName: ports.TransactionClassifyReactTaskName,
		Payload:  ports.ClassificationJobPayload{TransactionID: "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"},
		QueuedAt: queuedAt.Add(250 * time.Millisecond),
	}}}
	handler := &ReActClassificationJobHandler{
		txRepo: &reactSelectBatchRepositoryStub{transactionsByID: map[string]*entities.Transaction{
			"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60": mustNewHandlerTransaction(t, mustTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60"), mustClassification(t, "unclassified"), mustStatus(t, "processing"), nil, ""),
			"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61": mustNewHandlerTransaction(t, mustTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"), mustClassification(t, "unclassified"), mustStatus(t, "processing"), nil, ""),
		}},
		logger: zap.NewNop(),
		now:    testTime,
	}
	worker := NewReActClassificationWorker(queue, handler, transactionManagerServicesStub{}, zap.NewNop(), WithReActClassificationWorkerNow(func() time.Time {
		return queuedAt.Add(2500 * time.Millisecond)
	}))

	batch, err := worker.selectBatch(context.Background())
	if err != nil {
		t.Fatalf("selectBatch() error = %v", err)
	}

	if len(batch) != 2 {
		t.Fatalf("len(batch) = %d, want 2", len(batch))
	}
}

// This test ensures the batch is flushed immediately once the configured unique
// description limit is reached.
func TestReActClassificationWorkerSelectBatchFlushesAtUniqueDescriptionLimit(t *testing.T) {
	t.Parallel()

	queuedAt := time.Date(2026, time.April, 22, 10, 0, 0, 0, time.UTC)
	queue := &classificationJobQueueStub{batchJobs: []ports.ClassificationJob{{
		TaskName: ports.TransactionClassifyReactTaskName,
		Payload:  ports.ClassificationJobPayload{TransactionID: "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60"},
		QueuedAt: queuedAt,
	}, {
		TaskName: ports.TransactionClassifyReactTaskName,
		Payload:  ports.ClassificationJobPayload{TransactionID: "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"},
		QueuedAt: queuedAt.Add(250 * time.Millisecond),
	}}}
	handler := &ReActClassificationJobHandler{
		txRepo: &reactSelectBatchRepositoryStub{transactionsByID: map[string]*entities.Transaction{
			"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60": mustNewHandlerTransaction(t, mustTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60"), mustClassification(t, "unclassified"), mustStatus(t, "processing"), nil, ""),
			"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61": mustNewHandlerTransaction(t, mustTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"), mustClassification(t, "unclassified"), mustStatus(t, "processing"), nil, ""),
		}},
		logger: zap.NewNop(),
		now:    testTime,
	}
	worker := NewReActClassificationWorker(
		queue,
		handler,
		transactionManagerServicesStub{},
		zap.NewNop(),
		WithReActClassificationWorkerBatchSize(2),
		WithReActClassificationWorkerNow(func() time.Time { return queuedAt.Add(2500 * time.Millisecond) }),
	)

	batch, err := worker.selectBatch(context.Background())
	if err != nil {
		t.Fatalf("selectBatch() error = %v", err)
	}

	if len(batch) != 2 {
		t.Fatalf("len(batch) = %d, want 2", len(batch))
	}
}

// This test documents handled ReAct failures: the worker marks the transaction
// failed, completes the queue item, and returns no retry-triggering error.
func TestReActClassificationWorkerProcessNextCompletesHandledFailures(t *testing.T) {
	t.Parallel()

	transactionID := "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60"
	queue := &classificationJobQueueStub{batchJobs: []ports.ClassificationJob{{
		TaskName: ports.TransactionClassifyReactTaskName,
		Payload:  ports.ClassificationJobPayload{TransactionID: transactionID},
		QueuedAt: time.Date(2026, time.April, 22, 10, 0, 0, 0, time.UTC),
	}}}
	repository := &transactionRepositoryHandlerMock{
		findByIDValue: mustNewHandlerTransaction(t, mustTransactionID(t, transactionID), mustClassification(t, "unclassified"), mustStatus(t, "processing"), nil, ""),
	}
	handler := NewReActClassificationJobHandler(
		&reactClassificationGatewayHandlerStub{err: errors.New("react boom")},
		repository,
		zap.NewNop(),
	)
	handler.now = testTime
	worker := NewReActClassificationWorker(
		queue,
		handler,
		transactionManagerServicesStub{},
		zap.NewNop(),
		WithReActClassificationWorkerNow(func() time.Time { return time.Date(2026, time.April, 22, 10, 0, 3, 0, time.UTC) }),
	)

	processed, err := worker.processNext(context.Background())
	if err != nil {
		t.Fatalf("processNext() error = %v, want nil", err)
	}

	if !processed {
		t.Fatal("processNext() processed = false, want true")
	}

	if len(queue.completedJobs) != 1 {
		t.Fatalf("len(queue.completedJobs) = %d, want %d", len(queue.completedJobs), 1)
	}

	if repository.updatedTransaction == nil {
		t.Fatal("repository.updatedTransaction = nil, want transaction")
	}

	if repository.updatedTransaction.Status() != valueobjects.FailedTransactionStatus().String() {
		t.Fatalf("repository.updatedTransaction.Status() = %q, want %q", repository.updatedTransaction.Status(), valueobjects.FailedTransactionStatus().String())
	}
}

// This test documents that a failure-persistence race is treated as handled so the worker does not crash the server.
func TestReActClassificationWorkerProcessNextTreatsFailurePersistenceRaceAsHandled(t *testing.T) {
	t.Parallel()

	transactionID := "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60"
	queue := &classificationJobQueueStub{batchJobs: []ports.ClassificationJob{{
		TaskName: ports.TransactionClassifyReactTaskName,
		Payload:  ports.ClassificationJobPayload{TransactionID: transactionID},
		QueuedAt: time.Date(2026, time.April, 22, 10, 0, 0, 0, time.UTC),
	}}}
	repository := &transactionRepositoryHandlerMock{
		findByIDValue: mustNewHandlerTransaction(t, mustTransactionID(t, transactionID), mustClassification(t, "unclassified"), mustStatus(t, "processing"), nil, ""),
		updateErr:     errors.New("persisting failed transaction: executing update transaction query: ERROR: current transaction is aborted, commands ignored until end of transaction block (SQLSTATE 25P02)"),
	}
	handler := NewReActClassificationJobHandler(
		&reactClassificationGatewayHandlerStub{err: errors.New("react boom")},
		repository,
		zap.NewNop(),
	)
	handler.now = testTime
	worker := NewReActClassificationWorker(
		queue,
		handler,
		transactionManagerServicesStub{},
		zap.NewNop(),
		WithReActClassificationWorkerNow(func() time.Time { return time.Date(2026, time.April, 22, 10, 0, 3, 0, time.UTC) }),
	)

	processed, err := worker.processNext(context.Background())
	if err != nil {
		t.Fatalf("processNext() error = %v, want nil", err)
	}

	if !processed {
		t.Fatal("processNext() processed = false, want true")
	}

	if len(queue.completedJobs) != 1 {
		t.Fatalf("len(queue.completedJobs) = %d, want %d", len(queue.completedJobs), 1)
	}
}

// This test ensures batch selection can assemble one logical batch across
// paged queue reads without rescanning already fetched jobs.
func TestReActClassificationWorkerSelectBatchUsesSingleScanAcrossPages(t *testing.T) {
	t.Parallel()

	queuedAt := time.Date(2026, time.April, 22, 10, 0, 0, 0, time.UTC)
	queue := &classificationJobQueueStub{batchJobPages: [][]ports.ClassificationJob{
		{
			{
				TaskName: ports.TransactionClassifyReactTaskName,
				Payload:  ports.ClassificationJobPayload{TransactionID: "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60"},
				QueuedAt: queuedAt,
			},
			{
				TaskName: ports.TransactionClassifyReactTaskName,
				Payload:  ports.ClassificationJobPayload{TransactionID: "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"},
				QueuedAt: queuedAt.Add(250 * time.Millisecond),
			},
		},
	}}
	handler := &ReActClassificationJobHandler{
		txRepo: &reactSelectBatchRepositoryStub{transactionsByID: map[string]*entities.Transaction{
			"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60": mustNewHandlerTransaction(t, mustTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60"), mustClassification(t, "unclassified"), mustStatus(t, "processing"), nil, ""),
			"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61": mustNewHandlerTransaction(t, mustTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"), mustClassification(t, "unclassified"), mustStatus(t, "processing"), nil, ""),
		}},
		logger: zap.NewNop(),
		now:    testTime,
	}
	worker := NewReActClassificationWorker(
		queue,
		handler,
		transactionManagerServicesStub{},
		zap.NewNop(),
		WithReActClassificationWorkerBatchSize(10),
		WithReActClassificationWorkerNow(func() time.Time { return queuedAt.Add(2500 * time.Millisecond) }),
	)

	batch, err := worker.selectBatch(context.Background())
	if err != nil {
		t.Fatalf("selectBatch() error = %v", err)
	}

	if len(batch) != 2 {
		t.Fatalf("len(batch) = %d, want 2", len(batch))
	}
}

type reactSelectBatchRepositoryStub struct {
	transactionsByID map[string]*entities.Transaction
}

type reactClassificationGatewayStub struct{}

func (reactClassificationGatewayStub) Classify(context.Context, []ports.TransactionClassificationCandidate) ([]ports.TransactionClassificationDecision, error) {
	return nil, nil
}

func (s *reactSelectBatchRepositoryStub) Create(context.Context, *entities.Transaction, string) error {
	return nil
}

func (s *reactSelectBatchRepositoryStub) CreateMany(context.Context, []*entities.Transaction, string) error {
	return nil
}

func (s *reactSelectBatchRepositoryStub) Update(context.Context, *entities.Transaction) error {
	return nil
}

func (s *reactSelectBatchRepositoryStub) FindByID(_ context.Context, id valueobjects.TransactionID) (*entities.Transaction, error) {
	if s.transactionsByID == nil {
		return nil, nil
	}

	return s.transactionsByID[id.String()], nil
}

func (s *reactSelectBatchRepositoryStub) FindHistoricalClassificationByExactGoodsDescription(context.Context, ports.HistoricalTransactionClassificationQuery) (*ports.HistoricalTransactionClassificationMatch, error) {
	return nil, nil
}

func (s *reactSelectBatchRepositoryStub) GetNavigation(context.Context, ports.TransactionNavigationLookup) (*ports.TransactionNavigationResult, error) {
	return nil, nil
}

func (s *reactSelectBatchRepositoryStub) List(context.Context, ports.TransactionFilter) ([]*entities.Transaction, error) {
	return nil, nil
}

func (s *reactSelectBatchRepositoryStub) ListByUploadIDs(context.Context, []valueobjects.UploadID) ([]*entities.Transaction, error) {
	return nil, nil
}

func (s *reactSelectBatchRepositoryStub) HasProcessingByUploadID(context.Context, valueobjects.UploadID) (bool, error) {
	return false, nil
}

func (s *reactSelectBatchRepositoryStub) DeleteByID(context.Context, valueobjects.TransactionID) error {
	return nil
}

func (s *reactSelectBatchRepositoryStub) DeleteByUploadID(context.Context, valueobjects.UploadID) error {
	return nil
}

func mustTransactionID(t *testing.T, raw string) valueobjects.TransactionID {
	t.Helper()

	id, err := valueobjects.TransactionIDFromString(raw)
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	return id
}

func mustClassification(t *testing.T, raw string) valueobjects.TransactionClassification {
	t.Helper()

	classification, err := valueobjects.TransactionClassificationFromString(raw)
	if err != nil {
		t.Fatalf("TransactionClassificationFromString() error = %v", err)
	}

	return classification
}

func mustStatus(t *testing.T, raw string) valueobjects.TransactionStatus {
	t.Helper()

	status, err := valueobjects.TransactionStatusFromString(raw)
	if err != nil {
		t.Fatalf("TransactionStatusFromString() error = %v", err)
	}

	return status
}

func testTime() time.Time {
	return time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
}
