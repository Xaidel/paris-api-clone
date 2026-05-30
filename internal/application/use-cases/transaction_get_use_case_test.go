package usecases

import (
	"context"
	"errors"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// This test verifies the get use case returns the mapped transaction, records
// the read side effect, and converts a missing record into NotFoundError.
func TestGetTransactionUseCaseExecute(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := entities.NewTransaction(id, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	result, err := NewGetTransactionUseCase(&transactionRepositoryMock{findByID: transaction}, &transactionStep4RepositoryMock{}, &transactionStep5RepositoryMock{}, &sectorRepositoryMock{}, &adminEventRecorderMock{}).Execute(context.Background(), inboundports.GetTransactionQuery{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ID != id.String() {
		t.Fatalf("result.ID = %q, want %q", result.ID, id.String())
	}

	if result.Classification != "unclassified" {
		t.Fatalf("result.Classification = %q, want %q", result.Classification, "unclassified")
	}

	if result.Status != "processing" {
		t.Fatalf("result.Status = %q, want %q", result.Status, "processing")
	}

	useCase := NewGetTransactionUseCase(&transactionRepositoryMock{findByID: transaction}, &transactionStep4RepositoryMock{}, &transactionStep5RepositoryMock{}, &sectorRepositoryMock{}, &adminEventRecorderMock{})
	useCase.now = testTime
	_, err = useCase.Execute(context.Background(), inboundports.GetTransactionQuery{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	_, err = NewGetTransactionUseCase(&transactionRepositoryMock{}, &transactionStep4RepositoryMock{}, &transactionStep5RepositoryMock{}, &sectorRepositoryMock{}, &adminEventRecorderMock{}).Execute(context.Background(), inboundports.GetTransactionQuery{ID: id.String()})
	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
}

// This test ensures get responses include the enriched step-4 classification
// view when step-4 and sector data are available.
func TestGetTransactionUseCaseExecuteIncludesStep4Classification(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := entities.NewTransaction(id, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	sectorID, err := valueobjects.SectorIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("SectorIDFromString() error = %v", err)
	}

	additionalContext, err := valueobjects.NewTransactionStep4AdditionalContext("Reviewed by analyst")
	if err != nil {
		t.Fatalf("NewTransactionStep4AdditionalContext() error = %v", err)
	}

	step4, err := entities.NewTransactionStep4(id, sectorID, additionalContext, true, testTime())
	if err != nil {
		t.Fatalf("NewTransactionStep4() error = %v", err)
	}

	sector, err := entities.NewSector(sectorID, "High Emitting", "Energy", "High emitting energy sector")
	if err != nil {
		t.Fatalf("NewSector() error = %v", err)
	}

	result, err := NewGetTransactionUseCase(
		&transactionRepositoryMock{findByID: transaction},
		&transactionStep4RepositoryMock{findByTransactionID: step4},
		&transactionStep5RepositoryMock{},
		&sectorRepositoryMock{findByID: sector},
		&adminEventRecorderMock{},
	).Execute(context.Background(), inboundports.GetTransactionQuery{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Step4Classification == nil {
		t.Fatal("result.Step4Classification = nil, want step 4 classification")
	}

	if result.Step4Classification.IdentifiedSector != "Energy" {
		t.Fatalf("result.Step4Classification.IdentifiedSector = %q, want %q", result.Step4Classification.IdentifiedSector, "Energy")
	}

	if result.Step4Classification.AdditionalInformation != "Reviewed by analyst" {
		t.Fatalf("result.Step4Classification.AdditionalInformation = %q, want %q", result.Step4Classification.AdditionalInformation, "Reviewed by analyst")
	}

	if result.Step4Classification.Result != "next-step" {
		t.Fatalf("result.Step4Classification.Result = %q, want %q", result.Step4Classification.Result, "next-step")
	}
}

// This test ensures get responses include step-5 screening details and trimmed
// reviewer notes when a step-5 record exists.
func TestGetTransactionUseCaseExecuteIncludesStep5Classification(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := entities.NewTransaction(id, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	question1Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion1Justification("question 1 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion1Justification() error = %v", err)
	}

	question2Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion2Justification("question 2 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion2Justification() error = %v", err)
	}

	step5, err := entities.NewTransactionStep5(id, false, question1Justification, false, question2Justification, valueobjects.NewTransactionStep5ReviewerNotes(getTransactionStringPointer("  optional note  ")), true, testTime())
	if err != nil {
		t.Fatalf("NewTransactionStep5() error = %v", err)
	}

	result, err := NewGetTransactionUseCase(
		&transactionRepositoryMock{findByID: transaction},
		&transactionStep4RepositoryMock{},
		&transactionStep5RepositoryMock{findByTransactionID: step5},
		&sectorRepositoryMock{},
		&adminEventRecorderMock{},
	).Execute(context.Background(), inboundports.GetTransactionQuery{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Step5Classification == nil {
		t.Fatal("result.Step5Classification = nil, want step 5 classification")
	}

	if result.Step5Classification.Result != "aligned" {
		t.Fatalf("result.Step5Classification.Result = %q, want %q", result.Step5Classification.Result, "aligned")
	}

	if result.Step5Classification.ReviewerNotes == nil || *result.Step5Classification.ReviewerNotes != "optional note" {
		t.Fatalf("result.Step5Classification.ReviewerNotes = %v, want %q", result.Step5Classification.ReviewerNotes, "optional note")
	}
}

func getTransactionStringPointer(value string) *string {
	return &value
}
