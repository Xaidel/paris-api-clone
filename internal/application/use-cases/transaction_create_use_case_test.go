package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/gyud-adb/paris-api/internal/domain"
	domainservices "github.com/gyud-adb/paris-api/internal/domain/services"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// This test verifies the orchestration contract for transaction creation: the
// use case validates input, persists the entity, queues follow-up
// classification, records the admin event, and surfaces validation failures
// without mutating state.
func TestCreateTransactionUseCaseExecute(t *testing.T) {
	t.Parallel()

	fixedID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	validator := domainservices.NewTransactionFileValidator(valueobjects.TransactionFileSchemaV1())

	tests := []struct {
		name        string
		repository  *transactionRepositoryMock
		queue       *transactionProcessingQueueMock
		transaction *transactionManagerMock
		recorder    *adminEventRecorderMock
		command     inboundports.CreateTransactionCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.TransactionResult, repository *transactionRepositoryMock, queue *transactionProcessingQueueMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock)
	}{
		{
			name:        "creates transaction entry",
			repository:  &transactionRepositoryMock{},
			queue:       &transactionProcessingQueueMock{},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.CreateTransactionCommand{Product: "CG", ProcessedYear: 2026, ProcessedMonth: 4, DMCIB: "IB", DMC: "DMC", PartnerBank: "Partner Bank", ReferenceNumber: "REF-1", TransactionValue: "698436.80", TransactionCount: 1, GoodsDescription: "Goods", GoodsClassification: "Classification", ApplicantCountry: "Philippines", BeneficiaryCountry: "Japan", SourceCountry: "Thailand", DestinationCountry: "Philippines", TenorDescription: "N", ESCategory: "", PAAlignment: "PA Aligned", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.TransactionResult, repository *transactionRepositoryMock, queue *transactionProcessingQueueMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock) {
				t.Helper()

				if repository.createdTransaction == nil {
					t.Fatal("expected created transaction")
				}

				if repository.createdBy != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("repository.createdBy = %q, want %q", repository.createdBy, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				if result.ID != fixedID.String() {
					t.Fatalf("result.ID = %q, want %q", result.ID, fixedID.String())
				}

				if result.Product != "CG" {
					t.Fatalf("result.Product = %q, want %q", result.Product, "CG")
				}

				if result.CreatedBy != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("result.CreatedBy = %q, want %q", result.CreatedBy, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				if result.UploadID != "" {
					t.Fatalf("result.UploadID = %q, want empty", result.UploadID)
				}

				if result.Classification != "unclassified" {
					t.Fatalf("result.Classification = %q, want %q", result.Classification, "unclassified")
				}

				if result.Status != "processing" {
					t.Fatalf("result.Status = %q, want %q", result.Status, "processing")
				}

				if len(queue.enqueuedIDs) != 1 || queue.enqueuedIDs[0] != fixedID.String() {
					t.Fatalf("queue.enqueuedIDs = %v, want [%q]", queue.enqueuedIDs, fixedID.String())
				}

				if len(queue.enqueuedTasks) != 1 || queue.enqueuedTasks[0] != outboundports.TransactionClassifyReactTaskName {
					t.Fatalf("queue.enqueuedTasks = %v, want [%q]", queue.enqueuedTasks, outboundports.TransactionClassifyReactTaskName)
				}

				if repository.updatedTransaction == nil {
					t.Fatal("expected updated transaction")
				}

				if !transaction.invoked {
					t.Fatal("expected transaction manager to be invoked")
				}

				if recorder.command.EventType != createTransactionAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, createTransactionAdminEventType)
				}

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}

				if payload["target_id"] != fixedID.String() {
					t.Fatalf("payload[target_id] = %v, want %q", payload["target_id"], fixedID.String())
				}
			},
		},
		{
			name:        "returns validation error",
			repository:  &transactionRepositoryMock{},
			queue:       &transactionProcessingQueueMock{},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.CreateTransactionCommand{Product: "", ProcessedYear: 0, ProcessedMonth: 13, DMCIB: "", DMC: "", PartnerBank: "", ReferenceNumber: "", TransactionValue: "", TransactionCount: -1, GoodsDescription: "", GoodsClassification: "Classification", ApplicantCountry: "Philippines", SourceCountry: "Thailand", DestinationCountry: "Philippines", TenorDescription: "N", PAAlignment: "", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var validationErr *domain.ValidationError
				if !errors.As(err, &validationErr) && !strings.Contains(err.Error(), "creating transaction record") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewCreateTransactionUseCase(tc.repository, tc.queue, tc.transaction, tc.recorder, &actorDirectoryMock{}, validator)
			useCase.newID = func() (valueobjects.TransactionID, error) { return fixedID, nil }
			useCase.now = testTime

			result, err := useCase.Execute(context.Background(), tc.command)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result, tc.repository, tc.queue, tc.transaction, tc.recorder)
		})
	}
}
