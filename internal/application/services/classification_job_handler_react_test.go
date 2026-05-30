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

type reactClassificationGatewayHandlerStub struct {
	decisions  []ports.TransactionClassificationDecision
	err        error
	callCount  int
	candidates []ports.TransactionClassificationCandidate
}

type transactionRepositoryHandlerMock struct {
	findByIDValue      *entities.Transaction
	findByIDErr        error
	updatedTransaction *entities.Transaction
	updateErr          error
	updateCalls        int
}

func (m *transactionRepositoryHandlerMock) Create(context.Context, *entities.Transaction, string) error {
	return nil
}

func (m *transactionRepositoryHandlerMock) CreateMany(context.Context, []*entities.Transaction, string) error {
	return nil
}

func (m *transactionRepositoryHandlerMock) Update(_ context.Context, transaction *entities.Transaction) error {
	m.updateCalls++
	m.updatedTransaction = transaction
	return m.updateErr
}

func (m *transactionRepositoryHandlerMock) FindByID(_ context.Context, _ valueobjects.TransactionID) (*entities.Transaction, error) {
	return m.findByIDValue, m.findByIDErr
}

func (m *transactionRepositoryHandlerMock) FindHistoricalClassificationByExactGoodsDescription(context.Context, ports.HistoricalTransactionClassificationQuery) (*ports.HistoricalTransactionClassificationMatch, error) {
	return nil, nil
}

func (m *transactionRepositoryHandlerMock) GetNavigation(context.Context, ports.TransactionNavigationLookup) (*ports.TransactionNavigationResult, error) {
	return nil, nil
}

func (m *transactionRepositoryHandlerMock) List(context.Context, ports.TransactionFilter) ([]*entities.Transaction, error) {
	return nil, nil
}

func (m *transactionRepositoryHandlerMock) ListByUploadIDs(context.Context, []valueobjects.UploadID) ([]*entities.Transaction, error) {
	return nil, nil
}

func (m *transactionRepositoryHandlerMock) HasProcessingByUploadID(context.Context, valueobjects.UploadID) (bool, error) {
	return false, nil
}

func (m *transactionRepositoryHandlerMock) DeleteByID(context.Context, valueobjects.TransactionID) error {
	return nil
}

func (m *transactionRepositoryHandlerMock) DeleteByUploadID(context.Context, valueobjects.UploadID) error {
	return nil
}

func (s *reactClassificationGatewayHandlerStub) Classify(_ context.Context, candidates []ports.TransactionClassificationCandidate) ([]ports.TransactionClassificationDecision, error) {
	s.callCount++
	s.candidates = append([]ports.TransactionClassificationCandidate(nil), candidates...)
	if s.err != nil {
		return nil, s.err
	}

	return append([]ports.TransactionClassificationDecision(nil), s.decisions...), nil
}

// TestReActClassificationJobHandlerHandleBatch verifies the ReAct act classification job handler handle batch behavior and the expected outcome asserted below.
func TestReActClassificationJobHandlerHandleBatch(t *testing.T) {
	t.Parallel()

	transactionID := mustTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	reactReviewResult := valueobjects.NewReactReviewResult(
		valueobjects.PipelineResultSourceLLMBatch,
		"react",
		"v1",
		"v1",
		"gpt-4o-mini",
		transactionID.String(),
		nil,
		nil,
		stringPointerServices("batch-1"),
		intPointerServices(1),
		false,
		10,
		true,
		8,
		2,
		valueobjects.AlignedTransactionClassification(),
		"aligned reason",
	)
	reactPipelineResult := valueobjects.NewReactPipelineResult(reactReviewResult)
	nextStepReactResult := valueobjects.NewReactPipelineResult(valueobjects.NewReactReviewResult(
		valueobjects.PipelineResultSourceLLMBatch,
		"react",
		"v1",
		"v1",
		"gpt-4o-mini",
		transactionID.String(),
		nil,
		nil,
		stringPointerServices("batch-2"),
		intPointerServices(1),
		false,
		10,
		false,
		4,
		2,
		valueobjects.NextStepTransactionClassification(),
		"next step",
	))

	tests := []struct {
		name        string
		repository  *transactionRepositoryHandlerMock
		gateway     *reactClassificationGatewayHandlerStub
		jobs        []ports.ClassificationJob
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, repository *transactionRepositoryHandlerMock, gateway *reactClassificationGatewayHandlerStub)
	}{
		{
			name: "happy path marks transaction ai reviewed",
			repository: &transactionRepositoryHandlerMock{
				findByIDValue: mustNewHandlerTransaction(t, transactionID, valueobjects.UnclassifiedTransactionClassification(), valueobjects.ProcessingTransactionStatus(), nil, ""),
			},
			gateway: &reactClassificationGatewayHandlerStub{decisions: []ports.TransactionClassificationDecision{{
				TransactionID:  transactionID,
				Classification: valueobjects.AlignedTransactionClassification(),
				Status:         valueobjects.AIReviewedTransactionStatus(),
				ReviewResult:   reactPipelineResult,
				Source:         valueobjects.PipelineResultSourceLLMBatch,
			}}},
			jobs: []ports.ClassificationJob{{
				TaskName: ports.TransactionClassifyReactTaskName,
				Payload:  ports.ClassificationJobPayload{TransactionID: transactionID.String()},
			}},
			assert: func(t *testing.T, repository *transactionRepositoryHandlerMock, gateway *reactClassificationGatewayHandlerStub) {
				t.Helper()

				if gateway.callCount != 1 {
					t.Fatalf("gateway.callCount = %d, want %d", gateway.callCount, 1)
				}

				if repository.updateCalls != 1 {
					t.Fatalf("repository.updateCalls = %d, want %d", repository.updateCalls, 1)
				}

				if repository.updatedTransaction == nil {
					t.Fatal("repository.updatedTransaction = nil, want transaction")
				}

				if repository.updatedTransaction.Status() != valueobjects.AIReviewedTransactionStatus().String() {
					t.Fatalf("repository.updatedTransaction.Status() = %q, want %q", repository.updatedTransaction.Status(), valueobjects.AIReviewedTransactionStatus().String())
				}

				if repository.updatedTransaction.Classification() != valueobjects.AlignedTransactionClassification().String() {
					t.Fatalf("repository.updatedTransaction.Classification() = %q, want %q", repository.updatedTransaction.Classification(), valueobjects.AlignedTransactionClassification().String())
				}

				if repository.updatedTransaction.PipelineResult() == nil || repository.updatedTransaction.PipelineResult().React() == nil {
					t.Fatal("repository.updatedTransaction.PipelineResult().React() = nil, want react result")
				}

			},
		},
		{
			name: "non terminal react decision restores step 3 next_step result",
			repository: &transactionRepositoryHandlerMock{
				findByIDValue: mustNewHandlerTransactionWithTenor(t, transactionID, "Y-256", valueobjects.UnclassifiedTransactionClassification(), valueobjects.ProcessingTransactionStatus(), nil, ""),
			},
			gateway: &reactClassificationGatewayHandlerStub{decisions: []ports.TransactionClassificationDecision{{
				TransactionID:  transactionID,
				Classification: valueobjects.NextStepTransactionClassification(),
				Status:         valueobjects.AIReviewedTransactionStatus(),
				ReviewResult:   nextStepReactResult,
				Source:         valueobjects.PipelineResultSourceLLMBatch,
			}}},
			jobs: []ports.ClassificationJob{{
				TaskName: ports.TransactionClassifyReactTaskName,
				Payload:  ports.ClassificationJobPayload{TransactionID: transactionID.String()},
			}},
			assert: func(t *testing.T, repository *transactionRepositoryHandlerMock, gateway *reactClassificationGatewayHandlerStub) {
				t.Helper()

				if repository.updatedTransaction == nil {
					t.Fatal("repository.updatedTransaction = nil, want transaction")
				}

				if repository.updatedTransaction.Classification() != valueobjects.NextStepTransactionClassification().String() {
					t.Fatalf("repository.updatedTransaction.Classification() = %q, want %q", repository.updatedTransaction.Classification(), valueobjects.NextStepTransactionClassification().String())
				}

				if repository.updatedTransaction.Status() != valueobjects.AIReviewedTransactionStatus().String() {
					t.Fatalf("repository.updatedTransaction.Status() = %q, want %q", repository.updatedTransaction.Status(), valueobjects.AIReviewedTransactionStatus().String())
				}

				react := repository.updatedTransaction.PipelineResult().React()
				if react == nil {
					t.Fatal("repository.updatedTransaction.PipelineResult().React() = nil, want react result")
				}

				if react.ExitStep() != 3 {
					t.Fatalf("react.ExitStep() = %d, want %d", react.ExitStep(), 3)
				}

				if react.Source() != valueobjects.PipelineResultSourceLegacyStep3Fallback {
					t.Fatalf("react.Source() = %q, want %q", react.Source(), valueobjects.PipelineResultSourceLegacyStep3Fallback)
				}

				if react.Reason() != "next step; legacy step 3 fallback boolean result: false" {
					t.Fatalf("react.Reason() = %q, want %q", react.Reason(), "next step; legacy step 3 fallback boolean result: false")
				}
			},
		},
		{
			name: "next step decision becomes aligned for low risk tenor",
			repository: &transactionRepositoryHandlerMock{
				findByIDValue: mustNewHandlerTransaction(t, transactionID, valueobjects.UnclassifiedTransactionClassification(), valueobjects.ProcessingTransactionStatus(), nil, ""),
			},
			gateway: &reactClassificationGatewayHandlerStub{decisions: []ports.TransactionClassificationDecision{{
				TransactionID:  transactionID,
				Classification: valueobjects.NextStepTransactionClassification(),
				Status:         valueobjects.AIReviewedTransactionStatus(),
				ReviewResult:   nextStepReactResult,
				Source:         valueobjects.PipelineResultSourceLLMBatch,
			}}},
			jobs: []ports.ClassificationJob{{
				TaskName: ports.TransactionClassifyReactTaskName,
				Payload:  ports.ClassificationJobPayload{TransactionID: transactionID.String()},
			}},
			assert: func(t *testing.T, repository *transactionRepositoryHandlerMock, gateway *reactClassificationGatewayHandlerStub) {
				t.Helper()

				if repository.updatedTransaction == nil {
					t.Fatal("repository.updatedTransaction = nil, want transaction")
				}

				if repository.updatedTransaction.Classification() != valueobjects.AlignedTransactionClassification().String() {
					t.Fatalf("repository.updatedTransaction.Classification() = %q, want %q", repository.updatedTransaction.Classification(), valueobjects.AlignedTransactionClassification().String())
				}

				react := repository.updatedTransaction.PipelineResult().React()
				if react == nil {
					t.Fatal("repository.updatedTransaction.PipelineResult().React() = nil, want react result")
				}

				if react.ExitStep() != 3 {
					t.Fatalf("react.ExitStep() = %d, want %d", react.ExitStep(), 3)
				}

				if react.Source() != valueobjects.PipelineResultSourceLegacyStep3Fallback {
					t.Fatalf("react.Source() = %q, want %q", react.Source(), valueobjects.PipelineResultSourceLegacyStep3Fallback)
				}

				if react.Reason() != "next step; legacy step 3 fallback boolean result: true" {
					t.Fatalf("react.Reason() = %q, want %q", react.Reason(), "next step; legacy step 3 fallback boolean result: true")
				}

				if repository.updatedTransaction.Status() != valueobjects.AIReviewedTransactionStatus().String() {
					t.Fatalf("repository.updatedTransaction.Status() = %q, want %q", repository.updatedTransaction.Status(), valueobjects.AIReviewedTransactionStatus().String())
				}
			},
		},
		{
			name: "next step decision stays next_step for Y tenor",
			repository: &transactionRepositoryHandlerMock{
				findByIDValue: mustNewHandlerTransactionWithTenor(t, transactionID, "Y-256", valueobjects.UnclassifiedTransactionClassification(), valueobjects.ProcessingTransactionStatus(), nil, ""),
			},
			gateway: &reactClassificationGatewayHandlerStub{decisions: []ports.TransactionClassificationDecision{{
				TransactionID:  transactionID,
				Classification: valueobjects.NextStepTransactionClassification(),
				Status:         valueobjects.AIReviewedTransactionStatus(),
				ReviewResult:   nextStepReactResult,
				Source:         valueobjects.PipelineResultSourceLLMBatch,
			}}},
			jobs: []ports.ClassificationJob{{
				TaskName: ports.TransactionClassifyReactTaskName,
				Payload:  ports.ClassificationJobPayload{TransactionID: transactionID.String()},
			}},
			assert: func(t *testing.T, repository *transactionRepositoryHandlerMock, gateway *reactClassificationGatewayHandlerStub) {
				t.Helper()

				if repository.updatedTransaction == nil {
					t.Fatal("repository.updatedTransaction = nil, want transaction")
				}

				if repository.updatedTransaction.Classification() != valueobjects.NextStepTransactionClassification().String() {
					t.Fatalf("repository.updatedTransaction.Classification() = %q, want %q", repository.updatedTransaction.Classification(), valueobjects.NextStepTransactionClassification().String())
				}

				react := repository.updatedTransaction.PipelineResult().React()
				if react == nil {
					t.Fatal("repository.updatedTransaction.PipelineResult().React() = nil, want react result")
				}

				if react.ExitStep() != 3 {
					t.Fatalf("react.ExitStep() = %d, want %d", react.ExitStep(), 3)
				}

				if react.Source() != valueobjects.PipelineResultSourceLegacyStep3Fallback {
					t.Fatalf("react.Source() = %q, want %q", react.Source(), valueobjects.PipelineResultSourceLegacyStep3Fallback)
				}

				if react.Reason() != "next step; legacy step 3 fallback boolean result: false" {
					t.Fatalf("react.Reason() = %q, want %q", react.Reason(), "next step; legacy step 3 fallback boolean result: false")
				}
			},
		},
		{
			name: "next step decision uses default false for unknown tenor",
			repository: &transactionRepositoryHandlerMock{
				findByIDValue: mustNewHandlerTransactionWithTenor(t, transactionID, "UNKNOWN", valueobjects.UnclassifiedTransactionClassification(), valueobjects.ProcessingTransactionStatus(), nil, ""),
			},
			gateway: &reactClassificationGatewayHandlerStub{decisions: []ports.TransactionClassificationDecision{{
				TransactionID:  transactionID,
				Classification: valueobjects.NextStepTransactionClassification(),
				Status:         valueobjects.AIReviewedTransactionStatus(),
				ReviewResult:   nextStepReactResult,
				Source:         valueobjects.PipelineResultSourceLLMBatch,
			}}},
			jobs: []ports.ClassificationJob{{
				TaskName: ports.TransactionClassifyReactTaskName,
				Payload:  ports.ClassificationJobPayload{TransactionID: transactionID.String()},
			}},
			assert: func(t *testing.T, repository *transactionRepositoryHandlerMock, gateway *reactClassificationGatewayHandlerStub) {
				t.Helper()

				react := repository.updatedTransaction.PipelineResult().React()
				if react == nil {
					t.Fatal("repository.updatedTransaction.PipelineResult().React() = nil, want react result")
				}

				if repository.updatedTransaction.Classification() != valueobjects.NextStepTransactionClassification().String() {
					t.Fatalf("repository.updatedTransaction.Classification() = %q, want %q", repository.updatedTransaction.Classification(), valueobjects.NextStepTransactionClassification().String())
				}

				wantReason := "next step; legacy step 3 fallback boolean result: false (tenor description missing or unrecognized; defaulting step 3 result to not aligned)"
				if react.Reason() != wantReason {
					t.Fatalf("react.Reason() = %q, want %q", react.Reason(), wantReason)
				}
			},
		},
		{
			name: "gateway error marks failed and returns handled error",
			repository: &transactionRepositoryHandlerMock{
				findByIDValue: mustNewHandlerTransaction(t, transactionID, valueobjects.UnclassifiedTransactionClassification(), valueobjects.ProcessingTransactionStatus(), nil, ""),
			},
			gateway: &reactClassificationGatewayHandlerStub{err: errors.New("react boom")},
			jobs: []ports.ClassificationJob{{
				TaskName: ports.TransactionClassifyReactTaskName,
				Payload:  ports.ClassificationJobPayload{TransactionID: transactionID.String()},
			}},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var handledErr *reactClassificationHandledError
				if !errors.As(err, &handledErr) {
					t.Fatalf("HandleBatch() error = %v, want reactClassificationHandledError", err)
				}
				if handledErr.Error() != "react boom" {
					t.Fatalf("handledErr.Error() = %q, want %q", handledErr.Error(), "react boom")
				}
			},
			assert: func(t *testing.T, repository *transactionRepositoryHandlerMock, gateway *reactClassificationGatewayHandlerStub) {
				t.Helper()

				if gateway.callCount != 1 {
					t.Fatalf("gateway.callCount = %d, want %d", gateway.callCount, 1)
				}

				if repository.updateCalls != 1 {
					t.Fatalf("repository.updateCalls = %d, want %d", repository.updateCalls, 1)
				}

				if repository.updatedTransaction == nil {
					t.Fatal("repository.updatedTransaction = nil, want transaction")
				}

				if repository.updatedTransaction.Status() != valueobjects.FailedTransactionStatus().String() {
					t.Fatalf("repository.updatedTransaction.Status() = %q, want %q", repository.updatedTransaction.Status(), valueobjects.FailedTransactionStatus().String())
				}

				if repository.updatedTransaction.FailureReason() != "react boom" {
					t.Fatalf("repository.updatedTransaction.FailureReason() = %q, want %q", repository.updatedTransaction.FailureReason(), "react boom")
				}

			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := NewReActClassificationJobHandler(tc.gateway, tc.repository, zap.NewNop())
			handler.now = func() time.Time { return time.Date(2026, time.April, 22, 12, 0, 0, 0, time.UTC) }

			err := handler.HandleBatch(context.Background(), tc.jobs)
			if tc.assertError != nil {
				tc.assertError(t, err)
			} else if err != nil {
				t.Fatalf("HandleBatch() error = %v", err)
			}

			tc.assert(t, tc.repository, tc.gateway)
		})
	}
}

func stringPointerServices(value string) *string {
	copyValue := value
	return &copyValue
}

func intPointerServices(value int) *int {
	copyValue := value
	return &copyValue
}

func mustNewHandlerTransaction(t *testing.T, transactionID valueobjects.TransactionID, classification valueobjects.TransactionClassification, status valueobjects.TransactionStatus, pipelineResult *valueobjects.PipelineResult, failureReason string) *entities.Transaction {
	t.Helper()

	return mustNewHandlerTransactionWithTenor(t, transactionID, "N", classification, status, pipelineResult, failureReason)
}

func mustNewHandlerTransactionWithTenor(t *testing.T, transactionID valueobjects.TransactionID, tenorDescription string, classification valueobjects.TransactionClassification, status valueobjects.TransactionStatus, pipelineResult *valueobjects.PipelineResult, failureReason string) *entities.Transaction {
	t.Helper()

	createdBy, err := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("UserIDFromString() error = %v", err)
	}

	transaction := entities.ReconstituteTransaction(
		transactionID,
		nil,
		nil,
		"CG",
		2026,
		4,
		"IB",
		"DMC",
		"Partner Bank",
		"REF-1",
		"698436.80",
		classification,
		status,
		pipelineResult,
		failureReason,
		1,
		"Goods",
		"Classification",
		"Philippines",
		"Japan",
		"Thailand",
		"Philippines",
		tenorDescription,
		"",
		"PA Aligned",
		createdBy,
		time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
	)

	return transaction
}
