package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

const step5AutoDeterminedDetail = "Auto-determined based on your screening answer"

// CreateTransactionStep5UseCase computes or persists the final step 5 screening result.
type CreateTransactionStep5UseCase struct {
	repository            outboundports.TransactionStep5Repository
	transactionStep4Repo  outboundports.TransactionStep4Repository
	transactionRepository outboundports.TransactionRepository
	transactionManager    outboundports.TransactionManager
	eventRecorder         adminEventRecorder
	now                   func() time.Time
}

// NewCreateTransactionStep5UseCase builds a CreateTransactionStep5UseCase.
func NewCreateTransactionStep5UseCase(
	repository outboundports.TransactionStep5Repository,
	transactionStep4Repo outboundports.TransactionStep4Repository,
	transactionRepository outboundports.TransactionRepository,
	transactionManager outboundports.TransactionManager,
	eventRecorder adminEventRecorder,
) *CreateTransactionStep5UseCase {
	return &CreateTransactionStep5UseCase{
		repository:            repository,
		transactionStep4Repo:  transactionStep4Repo,
		transactionRepository: transactionRepository,
		transactionManager:    transactionManager,
		eventRecorder:         eventRecorder,
		now:                   time.Now,
	}
}

// Execute computes step 5 classification and optionally persists it.
func (uc *CreateTransactionStep5UseCase) Execute(ctx context.Context, command inboundports.CreateTransactionStep5Command) (outboundports.TransactionStep5Result, error) {
	if validationErr := validateCreateTransactionStep5Command(command); validationErr != nil {
		return outboundports.TransactionStep5Result{}, validationErr
	}

	transaction, err := uc.transactionRepository.FindByID(ctx, command.TransactionID)
	if err != nil {
		return outboundports.TransactionStep5Result{}, fmt.Errorf("finding transaction by id: %w", err)
	}
	if transaction == nil {
		return outboundports.TransactionStep5Result{}, &NotFoundError{Resource: "transaction", ID: command.TransactionID.String()}
	}

	step4, err := uc.transactionStep4Repo.FindByTransactionID(ctx, command.TransactionID)
	if err != nil {
		return outboundports.TransactionStep5Result{}, fmt.Errorf("finding transaction step 4 by transaction id: %w", err)
	}

	if step4 == nil || !step4.IsHighEmitting() || transaction.ClassificationValue().String() != "next_step" {
		return outboundports.TransactionStep5Result{}, &ConflictError{Resource: "transaction_step_5", Reason: "transaction step 4 result must be next_step"}
	}

	now := uc.now()
	step5, err := entities.NewTransactionStep5(
		command.TransactionID,
		*command.ScreeningQuestion1Answer,
		command.ScreeningQuestion1Justification,
		*command.ScreeningQuestion2Answer,
		command.ScreeningQuestion2Justification,
		command.ReviewerNotes,
		*command.IsFinal,
		now,
	)
	if err != nil {
		return outboundports.TransactionStep5Result{}, fmt.Errorf("creating transaction step 5: %w", err)
	}

	if !step5.IsFinal() {
		return newTransactionStep5Result(step5, false), nil
	}

	existingStep5, err := uc.repository.FindByTransactionID(ctx, command.TransactionID)
	if err != nil {
		return outboundports.TransactionStep5Result{}, fmt.Errorf("finding existing transaction step 5 by transaction id: %w", err)
	}
	if existingStep5 != nil {
		return outboundports.TransactionStep5Result{}, &ConflictError{Resource: "transaction_step_5", Reason: "transaction step 5 already exists"}
	}

	if err := transaction.MarkProfessionallyReviewed(step5.Classification(), now); err != nil {
		return outboundports.TransactionStep5Result{}, fmt.Errorf("marking transaction professionally reviewed: %w", err)
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.Create(txCtx, step5); err != nil {
			return fmt.Errorf("creating transaction step 5: %w", err)
		}

		if err := uc.transactionRepository.Update(txCtx, transaction); err != nil {
			return fmt.Errorf("updating transaction review state: %w", err)
		}

		payload := map[string]any{
			"action":                             "create",
			"resource":                           "transaction_step_5_data",
			"transaction_id":                     step5.TransactionID().String(),
			"screening_question_1_answer":        step5.ScreeningQuestion1Answer(),
			"screening_question_1_justification": step5.ScreeningQuestion1Justification().String(),
			"screening_question_2_answer":        step5.ScreeningQuestion2Answer(),
			"screening_question_2_justification": step5.ScreeningQuestion2Justification().String(),
			"is_final":                           step5.IsFinal(),
			"classification":                     step5.Classification().String(),
			"detail":                             step5AutoDeterminedDetail,
			"status":                             transaction.Status(),
		}
		if reviewerNotes := step5.ReviewerNotes().String(); reviewerNotes != nil {
			payload["reviewer_notes"] = *reviewerNotes
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, createTransactionStep5AdminEventType, payload); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return outboundports.TransactionStep5Result{}, fmt.Errorf("creating transaction step 5 transaction: %w", err)
	}

	return newTransactionStep5Result(step5, true), nil
}

func validateCreateTransactionStep5Command(command inboundports.CreateTransactionStep5Command) *domain.ValidationError {
	var validationErrors []domain.FieldValidationError

	if command.TransactionID.IsZero() {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("transaction_id", "required", "transaction_id is required"))
	}

	if command.ScreeningQuestion1Answer == nil {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("screening_question_1.answer", "required", "screening_question_1.answer is required"))
	}

	if command.ScreeningQuestion1Justification.String() == "" {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("screening_question_1.justification", "required", "screening_question_1.justification is required"))
	}

	if command.ScreeningQuestion2Answer == nil {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("screening_question_2.answer", "required", "screening_question_2.answer is required"))
	}

	if command.ScreeningQuestion2Justification.String() == "" {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("screening_question_2.justification", "required", "screening_question_2.justification is required"))
	}

	if command.IsFinal == nil {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("is_final", "required", "is_final is required"))
	}

	return domain.NewValidationError(validationErrors)
}

func newTransactionStep5Result(step5 *entities.TransactionStep5, persisted bool) outboundports.TransactionStep5Result {
	result := outboundports.TransactionStep5Result{
		TransactionID:                   step5.TransactionID(),
		ScreeningQuestion1Answer:        step5.ScreeningQuestion1Answer(),
		ScreeningQuestion1Justification: step5.ScreeningQuestion1Justification(),
		ScreeningQuestion2Answer:        step5.ScreeningQuestion2Answer(),
		ScreeningQuestion2Justification: step5.ScreeningQuestion2Justification(),
		ReviewerNotes:                   step5.ReviewerNotes(),
		IsFinal:                         step5.IsFinal(),
		Classification:                  step5.Classification(),
		Detail:                          step5AutoDeterminedDetail,
	}

	if persisted {
		createdAt := step5.CreatedAt()
		updatedAt := step5.UpdatedAt()
		result.CreatedAt = &createdAt
		result.UpdatedAt = &updatedAt
	}

	return result
}
