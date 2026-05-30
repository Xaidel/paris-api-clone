package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/gyud-adb/paris-api/internal/domain"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

type transactionStep5RepositoryMock struct {
	createdStep5           *entities.TransactionStep5
	findByTransactionID    *entities.TransactionStep5
	findByTransactionIDErr error
	createErr              error
}

func (m *transactionStep5RepositoryMock) Create(_ context.Context, step5 *entities.TransactionStep5) error {
	m.createdStep5 = step5
	return m.createErr
}

func (m *transactionStep5RepositoryMock) FindByTransactionID(_ context.Context, transactionID valueobjects.TransactionID) (*entities.TransactionStep5, error) {
	_ = transactionID
	return m.findByTransactionID, m.findByTransactionIDErr
}

// TestCreateTransactionStep5UseCaseExecute verifies the create transaction step 5 use case execute behavior and the expected outcome asserted below.
func TestCreateTransactionStep5UseCaseExecute(t *testing.T) {
	t.Parallel()

	transactionID := "01962b8f-aeb2-7e03-a8ff-1edce1300301"
	sectorID := "01962b8f-aeb2-7e03-a8ff-1edce1300302"

	transactionIDValue, err := valueobjects.TransactionIDFromString(transactionID)
	if err != nil {
		t.Fatalf("TransactionIDFromString(%q) error = %v", transactionID, err)
	}

	sectorIDValue, err := valueobjects.SectorIDFromString(sectorID)
	if err != nil {
		t.Fatalf("SectorIDFromString(%q) error = %v", sectorID, err)
	}

	question1Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion1Justification("question 1 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion1Justification() error = %v", err)
	}

	question2Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion2Justification("question 2 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion2Justification() error = %v", err)
	}

	highEmittingStep4, err := entities.NewTransactionStep4(
		transactionIDValue,
		sectorIDValue,
		mustTransactionStep4AdditionalContext(t, "Reviewed by analyst"),
		true,
		testTime(),
	)
	if err != nil {
		t.Fatalf("NewTransactionStep4() error = %v", err)
	}

	nonHighEmittingStep4, err := entities.NewTransactionStep4(
		transactionIDValue,
		sectorIDValue,
		mustTransactionStep4AdditionalContext(t, "Reviewed by analyst"),
		false,
		testTime(),
	)
	if err != nil {
		t.Fatalf("NewTransactionStep4() error = %v", err)
	}

	trueValue := true
	falseValue := false

	tests := []struct {
		name         string
		repository   *transactionStep5RepositoryMock
		step4Repo    *transactionStep4RepositoryMock
		transactions *transactionRepositoryMock
		manager      *transactionManagerMock
		recorder     *adminEventRecorderMock
		command      inboundports.CreateTransactionStep5Command
		assertError  func(t *testing.T, err error)
		assert       func(t *testing.T, result outboundports.TransactionStep5Result, repository *transactionStep5RepositoryMock, transactions *transactionRepositoryMock, manager *transactionManagerMock, recorder *adminEventRecorderMock)
	}{
		{
			name:         "previews aligned result without persistence when both answers are false",
			repository:   &transactionStep5RepositoryMock{},
			step4Repo:    &transactionStep4RepositoryMock{findByTransactionID: highEmittingStep4},
			transactions: &transactionRepositoryMock{findByID: buildProfessionallyReviewableTransaction(t, transactionID)},
			manager:      &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command: inboundports.CreateTransactionStep5Command{
				TransactionID:                   transactionIDValue,
				ScreeningQuestion1Answer:        &falseValue,
				ScreeningQuestion1Justification: question1Justification,
				ScreeningQuestion2Answer:        &falseValue,
				ScreeningQuestion2Justification: question2Justification,
				ReviewerNotes:                   valueobjects.NewTransactionStep5ReviewerNotes(step5TestStringPointer("  optional note  ")),
				IsFinal:                         &falseValue,
			},
			assert: func(t *testing.T, result outboundports.TransactionStep5Result, repository *transactionStep5RepositoryMock, transactions *transactionRepositoryMock, manager *transactionManagerMock, recorder *adminEventRecorderMock) {
				t.Helper()
				if repository.createdStep5 != nil {
					t.Fatal("repository.createdStep5 = non-nil, want nil")
				}
				if transactions.updatedTransaction != nil {
					t.Fatal("transactions.updatedTransaction = non-nil, want nil")
				}
				if manager.invoked {
					t.Fatal("transaction manager invoked = true, want false")
				}
				if result.Classification.String() != valueobjects.AlignedTransactionClassification().String() {
					t.Fatalf("result.Classification.String() = %q, want %q", result.Classification.String(), valueobjects.AlignedTransactionClassification().String())
				}
				if result.CreatedAt != nil {
					t.Fatalf("result.CreatedAt = %v, want nil", *result.CreatedAt)
				}
				if recorder.command.EventType != "" {
					t.Fatalf("recorder.command.EventType = %q, want empty", recorder.command.EventType)
				}
			},
		},
		{
			name:         "persists final result and updates transaction when any answer is true",
			repository:   &transactionStep5RepositoryMock{},
			step4Repo:    &transactionStep4RepositoryMock{findByTransactionID: highEmittingStep4},
			transactions: &transactionRepositoryMock{findByID: buildProfessionallyReviewableTransaction(t, transactionID)},
			manager:      &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command: inboundports.CreateTransactionStep5Command{
				TransactionID:                   transactionIDValue,
				ScreeningQuestion1Answer:        &trueValue,
				ScreeningQuestion1Justification: question1Justification,
				ScreeningQuestion2Answer:        &falseValue,
				ScreeningQuestion2Justification: question2Justification,
				ReviewerNotes:                   valueobjects.NewTransactionStep5ReviewerNotes(step5TestStringPointer("  optional note  ")),
				IsFinal:                         &trueValue,
				ActorUserID:                     "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				ActorGroupID:                    "group-1",
			},
			assert: func(t *testing.T, result outboundports.TransactionStep5Result, repository *transactionStep5RepositoryMock, transactions *transactionRepositoryMock, manager *transactionManagerMock, recorder *adminEventRecorderMock) {
				t.Helper()
				if repository.createdStep5 == nil {
					t.Fatal("repository.createdStep5 = nil, want created step 5")
				}
				if transactions.updatedTransaction == nil {
					t.Fatal("transactions.updatedTransaction = nil, want updated transaction")
				}
				if transactions.updatedTransaction.Classification() != valueobjects.NotAlignedTransactionClassification().String() {
					t.Fatalf("updatedTransaction.Classification() = %q, want %q", transactions.updatedTransaction.Classification(), valueobjects.NotAlignedTransactionClassification().String())
				}
				if !manager.invoked {
					t.Fatal("transaction manager invoked = false, want true")
				}
				if result.Classification.String() != valueobjects.NotAlignedTransactionClassification().String() {
					t.Fatalf("result.Classification.String() = %q, want %q", result.Classification.String(), valueobjects.NotAlignedTransactionClassification().String())
				}
				if result.CreatedAt == nil {
					t.Fatal("result.CreatedAt = nil, want timestamp")
				}
				if recorder.command.EventType != createTransactionStep5AdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, createTransactionStep5AdminEventType)
				}

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}
				if payload["classification"] != valueobjects.NotAlignedTransactionClassification().String() {
					t.Fatalf("payload[classification] = %v, want %q", payload["classification"], valueobjects.NotAlignedTransactionClassification().String())
				}
			},
		},
		{
			name:         "returns conflict when transaction step 4 is terminal",
			repository:   &transactionStep5RepositoryMock{},
			step4Repo:    &transactionStep4RepositoryMock{findByTransactionID: nonHighEmittingStep4},
			transactions: &transactionRepositoryMock{findByID: buildProfessionallyReviewableTransaction(t, transactionID)},
			manager:      &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command: inboundports.CreateTransactionStep5Command{
				TransactionID:                   transactionIDValue,
				ScreeningQuestion1Answer:        &falseValue,
				ScreeningQuestion1Justification: question1Justification,
				ScreeningQuestion2Answer:        &falseValue,
				ScreeningQuestion2Justification: question2Justification,
				IsFinal:                         &trueValue,
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
			name:         "returns conflict when step 5 already exists",
			repository:   &transactionStep5RepositoryMock{findByTransactionID: mustExistingTransactionStep5(t, transactionIDValue)},
			step4Repo:    &transactionStep4RepositoryMock{findByTransactionID: highEmittingStep4},
			transactions: &transactionRepositoryMock{findByID: buildProfessionallyReviewableTransaction(t, transactionID)},
			manager:      &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command: inboundports.CreateTransactionStep5Command{
				TransactionID:                   transactionIDValue,
				ScreeningQuestion1Answer:        &falseValue,
				ScreeningQuestion1Justification: question1Justification,
				ScreeningQuestion2Answer:        &falseValue,
				ScreeningQuestion2Justification: question2Justification,
				IsFinal:                         &trueValue,
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
			repository:   &transactionStep5RepositoryMock{},
			step4Repo:    &transactionStep4RepositoryMock{},
			transactions: &transactionRepositoryMock{},
			manager:      &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command:      inboundports.CreateTransactionStep5Command{},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var validationErr *domain.ValidationError
				if !errors.As(err, &validationErr) {
					t.Fatalf("expected ValidationError, got %v", err)
				}
				if len(validationErr.Fields()) != 6 {
					t.Fatalf("len(validationErr.Fields()) = %d, want %d", len(validationErr.Fields()), 6)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uc := NewCreateTransactionStep5UseCase(tc.repository, tc.step4Repo, tc.transactions, tc.manager, tc.recorder)
			uc.now = testTime

			result, err := uc.Execute(context.Background(), tc.command)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result, tc.repository, tc.transactions, tc.manager, tc.recorder)
		})
	}
}

func mustTransactionStep4AdditionalContext(t *testing.T, value string) valueobjects.TransactionStep4AdditionalContext {
	t.Helper()

	contextValue, err := valueobjects.NewTransactionStep4AdditionalContext(value)
	if err != nil {
		t.Fatalf("NewTransactionStep4AdditionalContext() error = %v", err)
	}

	return contextValue
}

func mustExistingTransactionStep5(t *testing.T, transactionID valueobjects.TransactionID) *entities.TransactionStep5 {
	t.Helper()

	question1Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion1Justification("question 1 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion1Justification() error = %v", err)
	}

	question2Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion2Justification("question 2 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion2Justification() error = %v", err)
	}

	step5, err := entities.NewTransactionStep5(
		transactionID,
		false,
		question1Justification,
		false,
		question2Justification,
		valueobjects.NewTransactionStep5ReviewerNotes(nil),
		true,
		testTime(),
	)
	if err != nil {
		t.Fatalf("NewTransactionStep5() error = %v", err)
	}

	return step5
}

func step5TestStringPointer(value string) *string {
	return &value
}
