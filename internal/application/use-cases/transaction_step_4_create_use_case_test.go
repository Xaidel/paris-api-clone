package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/gyud-adb/paris-api/internal/domain"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

type transactionStep4RepositoryMock struct {
	createdStep4           *entities.TransactionStep4
	findByTransactionID    *entities.TransactionStep4
	findByTransactionIDErr error
	createErr              error
}

func (m *transactionStep4RepositoryMock) Create(_ context.Context, step4 *entities.TransactionStep4) error {
	m.createdStep4 = step4
	return m.createErr
}

// TestCreateTransactionStep4UseCaseExecute verifies the create transaction step 4 use case execute behavior and the expected outcome asserted below.
func TestCreateTransactionStep4UseCaseExecute(t *testing.T) {
	t.Parallel()

	transactionID := "01962b8f-aeb2-7e03-a8ff-1edce1300001"
	sectorID := "01962b8f-aeb2-7e03-a8ff-1edce1300002"
	otherSectorID := "01962b8f-aeb2-7e03-a8ff-1edce1300003"

	eligibleTransaction := buildProfessionallyReviewableTransaction(t, transactionID)
	ineligibleTransaction := buildUnreviewableTransaction(t, transactionID)
	sector := buildSector(t, sectorID)

	trueValue := true
	falseValue := false
	transactionIDValue, err := valueobjects.TransactionIDFromString(transactionID)
	if err != nil {
		t.Fatalf("TransactionIDFromString(%q) error = %v", transactionID, err)
	}

	sectorIDValue, err := valueobjects.SectorIDFromString(sectorID)
	if err != nil {
		t.Fatalf("SectorIDFromString(%q) error = %v", sectorID, err)
	}

	otherSectorIDValue, err := valueobjects.SectorIDFromString(otherSectorID)
	if err != nil {
		t.Fatalf("SectorIDFromString(%q) error = %v", otherSectorID, err)
	}

	additionalContextValue, err := valueobjects.NewTransactionStep4AdditionalContext("Reviewed by analyst")
	if err != nil {
		t.Fatalf("NewTransactionStep4AdditionalContext() error = %v", err)
	}

	tests := []struct {
		name         string
		repository   *transactionStep4RepositoryMock
		sectors      *sectorRepositoryMock
		transactions *transactionRepositoryMock
		manager      *transactionManagerMock
		recorder     *adminEventRecorderMock
		command      inboundports.CreateTransactionStep4Command
		assertError  func(t *testing.T, err error)
		assert       func(t *testing.T, result outboundports.TransactionStep4Result, repository *transactionStep4RepositoryMock, transactions *transactionRepositoryMock, recorder *adminEventRecorderMock)
	}{
		{
			name:         "creates step 4 and marks transaction aligned when reviewer says not high emitting",
			repository:   &transactionStep4RepositoryMock{},
			sectors:      &sectorRepositoryMock{findByID: sector},
			transactions: &transactionRepositoryMock{findByID: eligibleTransaction},
			manager:      &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command: inboundports.CreateTransactionStep4Command{
				TransactionID:     transactionIDValue,
				SectorID:          sectorIDValue,
				AdditionalContext: additionalContextValue,
				IsHighEmitting:    &falseValue,
				ActorUserID:       "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				ActorGroupID:      "group-1",
			},
			assert: func(t *testing.T, result outboundports.TransactionStep4Result, repository *transactionStep4RepositoryMock, transactions *transactionRepositoryMock, recorder *adminEventRecorderMock) {
				t.Helper()
				if repository.createdStep4 == nil {
					t.Fatal("expected step 4 record to be created")
				}
				if repository.createdStep4.SectorID().String() != sectorID {
					t.Fatalf("createdStep4.SectorID().String() = %q, want %q", repository.createdStep4.SectorID().String(), sectorID)
				}
				if transactions.updatedTransaction == nil {
					t.Fatal("expected transaction to be updated")
				}
				if transactions.updatedTransaction.Classification() != valueobjects.AlignedTransactionClassification().String() {
					t.Fatalf("updatedTransaction.Classification() = %q, want %q", transactions.updatedTransaction.Classification(), valueobjects.AlignedTransactionClassification().String())
				}
				if transactions.updatedTransaction.Status() != valueobjects.ProfessionallyReviewedTransactionStatus().String() {
					t.Fatalf("updatedTransaction.Status() = %q, want %q", transactions.updatedTransaction.Status(), valueobjects.ProfessionallyReviewedTransactionStatus().String())
				}
				if !result.Classification.Equal(valueobjects.AlignedTransactionClassification()) {
					t.Fatalf("result.Classification.String() = %q, want %q", result.Classification.String(), valueobjects.AlignedTransactionClassification().String())
				}
				if !result.Status.Equal(valueobjects.ProfessionallyReviewedTransactionStatus()) {
					t.Fatalf("result.Status.String() = %q, want %q", result.Status.String(), valueobjects.ProfessionallyReviewedTransactionStatus().String())
				}
				if recorder.command.EventType != "CreateTransactionStep4" {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, "CreateTransactionStep4")
				}
			},
		},
		{
			name:         "creates step 4 and keeps transaction at next step when reviewer says high emitting",
			repository:   &transactionStep4RepositoryMock{},
			sectors:      &sectorRepositoryMock{findByID: sector},
			transactions: &transactionRepositoryMock{findByID: buildProfessionallyReviewableTransaction(t, transactionID)},
			manager:      &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command: inboundports.CreateTransactionStep4Command{
				TransactionID:     transactionIDValue,
				SectorID:          sectorIDValue,
				AdditionalContext: additionalContextValue,
				IsHighEmitting:    &trueValue,
				ActorUserID:       "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				ActorGroupID:      "group-1",
			},
			assert: func(t *testing.T, result outboundports.TransactionStep4Result, _ *transactionStep4RepositoryMock, transactions *transactionRepositoryMock, _ *adminEventRecorderMock) {
				t.Helper()
				if transactions.updatedTransaction == nil {
					t.Fatal("expected transaction to be updated")
				}
				if transactions.updatedTransaction.Classification() != valueobjects.NextStepTransactionClassification().String() {
					t.Fatalf("updatedTransaction.Classification() = %q, want %q", transactions.updatedTransaction.Classification(), valueobjects.NextStepTransactionClassification().String())
				}
				if transactions.updatedTransaction.Status() != valueobjects.ProfessionallyReviewedTransactionStatus().String() {
					t.Fatalf("updatedTransaction.Status() = %q, want %q", transactions.updatedTransaction.Status(), valueobjects.ProfessionallyReviewedTransactionStatus().String())
				}
				if !result.Classification.Equal(valueobjects.NextStepTransactionClassification()) {
					t.Fatalf("result.Classification.String() = %q, want %q", result.Classification.String(), valueobjects.NextStepTransactionClassification().String())
				}
				if !result.Status.Equal(valueobjects.ProfessionallyReviewedTransactionStatus()) {
					t.Fatalf("result.Status.String() = %q, want %q", result.Status.String(), valueobjects.ProfessionallyReviewedTransactionStatus().String())
				}
			},
		},
		{
			name:         "returns not found when transaction is missing",
			repository:   &transactionStep4RepositoryMock{},
			sectors:      &sectorRepositoryMock{findByID: sector},
			transactions: &transactionRepositoryMock{},
			manager:      &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command: inboundports.CreateTransactionStep4Command{
				TransactionID:     transactionIDValue,
				SectorID:          sectorIDValue,
				AdditionalContext: additionalContextValue,
				IsHighEmitting:    &falseValue,
			},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
				if notFoundErr.Resource != "transaction" {
					t.Fatalf("notFoundErr.Resource = %q, want %q", notFoundErr.Resource, "transaction")
				}
			},
		},
		{
			name:         "returns not found when sector is missing",
			repository:   &transactionStep4RepositoryMock{},
			sectors:      &sectorRepositoryMock{},
			transactions: &transactionRepositoryMock{findByID: eligibleTransaction},
			manager:      &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command: inboundports.CreateTransactionStep4Command{
				TransactionID:     transactionIDValue,
				SectorID:          otherSectorIDValue,
				AdditionalContext: additionalContextValue,
				IsHighEmitting:    &falseValue,
			},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
				if notFoundErr.Resource != "sector" {
					t.Fatalf("notFoundErr.Resource = %q, want %q", notFoundErr.Resource, "sector")
				}
			},
		},
		{
			name:         "returns conflict when transaction is not eligible for step 4",
			repository:   &transactionStep4RepositoryMock{},
			sectors:      &sectorRepositoryMock{findByID: sector},
			transactions: &transactionRepositoryMock{findByID: ineligibleTransaction},
			manager:      &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command: inboundports.CreateTransactionStep4Command{
				TransactionID:     transactionIDValue,
				SectorID:          sectorIDValue,
				AdditionalContext: additionalContextValue,
				IsHighEmitting:    &falseValue,
			},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var conflictErr *ConflictError
				if !errors.As(err, &conflictErr) {
					t.Fatalf("expected ConflictError, got %v", err)
				}
			},
		},
		{
			name:         "returns validation error when required fields are missing",
			repository:   &transactionStep4RepositoryMock{},
			sectors:      &sectorRepositoryMock{},
			transactions: &transactionRepositoryMock{},
			manager:      &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command:      inboundports.CreateTransactionStep4Command{},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var validationErr *domain.ValidationError
				if !errors.As(err, &validationErr) {
					t.Fatalf("expected ValidationError, got %v", err)
				}
				if len(validationErr.Fields()) != 4 {
					t.Fatalf("len(validationErr.Fields()) = %d, want %d", len(validationErr.Fields()), 4)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uc := NewCreateTransactionStep4UseCase(tc.repository, tc.sectors, tc.transactions, tc.manager, tc.recorder)
			uc.now = testTime

			result, err := uc.Execute(context.Background(), tc.command)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result, tc.repository, tc.transactions, tc.recorder)
		})
	}
}

func buildProfessionallyReviewableTransaction(t *testing.T, transactionID string) *entities.Transaction {
	t.Helper()

	id, err := valueobjects.TransactionIDFromString(transactionID)
	if err != nil {
		t.Fatalf("TransactionIDFromString(%q) error = %v", transactionID, err)
	}

	transaction, err := entities.NewTransaction(id, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "Y", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	step1 := valueobjects.NewStepResult(
		1,
		valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "u1", nil),
		valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "u1", nil),
	)
	step2 := valueobjects.NewStepResult(
		2,
		valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "sector", nil),
		valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.1, "sector", nil),
	)
	step3 := valueobjects.NewBooleanStepResult(3, false)
	pipelineResult := valueobjects.NewPipelineResult(
		transaction.ID().String(),
		step1,
		&step2,
		&step3,
		3,
		valueobjects.NextStepTransactionClassification(),
	)

	if err := transaction.MarkClassified(valueobjects.NextStepTransactionClassification(), pipelineResult, testTime()); err != nil {
		t.Fatalf("transaction.MarkClassified() error = %v", err)
	}

	return transaction
}

func buildUnreviewableTransaction(t *testing.T, transactionID string) *entities.Transaction {
	t.Helper()

	id, err := valueobjects.TransactionIDFromString(transactionID)
	if err != nil {
		t.Fatalf("TransactionIDFromString(%q) error = %v", transactionID, err)
	}

	transaction, err := entities.NewTransaction(id, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	return transaction
}

func buildSector(t *testing.T, sectorID string) *entities.Sector {
	t.Helper()

	id, err := valueobjects.SectorIDFromString(sectorID)
	if err != nil {
		t.Fatalf("SectorIDFromString(%q) error = %v", sectorID, err)
	}

	sector, err := entities.NewSector(id, "High Emitting", "Energy", "Energy generation")
	if err != nil {
		t.Fatalf("NewSector() error = %v", err)
	}

	return sector
}
