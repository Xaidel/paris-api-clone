package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// CreateTransactionStep4UseCase creates a persisted aggregated step 4 review.
type CreateTransactionStep4UseCase struct {
	repository            outboundports.TransactionStep4Repository
	sectorRepository      outboundports.SectorRepository
	transactionRepository outboundports.TransactionRepository
	transactionManager    outboundports.TransactionManager
	eventRecorder         adminEventRecorder
	now                   func() time.Time
}

// NewCreateTransactionStep4UseCase builds a CreateTransactionStep4UseCase.
func NewCreateTransactionStep4UseCase(
	repository outboundports.TransactionStep4Repository,
	sectorRepository outboundports.SectorRepository,
	transactionRepository outboundports.TransactionRepository,
	transactionManager outboundports.TransactionManager,
	eventRecorder adminEventRecorder,
) *CreateTransactionStep4UseCase {
	return &CreateTransactionStep4UseCase{
		repository:            repository,
		sectorRepository:      sectorRepository,
		transactionRepository: transactionRepository,
		transactionManager:    transactionManager,
		eventRecorder:         eventRecorder,
		now:                   time.Now,
	}
}

// Execute creates an aggregated step 4 review and updates the transaction state.
func (uc *CreateTransactionStep4UseCase) Execute(ctx context.Context, command inboundports.CreateTransactionStep4Command) (outboundports.TransactionStep4Result, error) {
	if validationErr := validateCreateTransactionStep4Command(command); validationErr != nil {
		return outboundports.TransactionStep4Result{}, validationErr
	}

	transaction, err := uc.transactionRepository.FindByID(ctx, command.TransactionID)
	if err != nil {
		return outboundports.TransactionStep4Result{}, fmt.Errorf("finding transaction by id: %w", err)
	}
	if transaction == nil {
		return outboundports.TransactionStep4Result{}, &NotFoundError{Resource: "transaction", ID: command.TransactionID.String()}
	}

	sector, err := uc.sectorRepository.FindByID(ctx, command.SectorID)
	if err != nil {
		return outboundports.TransactionStep4Result{}, fmt.Errorf("finding sector by id: %w", err)
	}
	if sector == nil {
		return outboundports.TransactionStep4Result{}, &NotFoundError{Resource: "sector", ID: command.SectorID.String()}
	}

	if !transaction.HasStep3NextStepResult() {
		return outboundports.TransactionStep4Result{}, &ConflictError{Resource: "transaction_step_4", Reason: "transaction step 3 result must be next_step"}
	}

	now := uc.now()
	step4, err := entities.NewTransactionStep4(command.TransactionID, command.SectorID, command.AdditionalContext, *command.IsHighEmitting, now)
	if err != nil {
		return outboundports.TransactionStep4Result{}, fmt.Errorf("creating transaction step 4: %w", err)
	}

	classification := valueobjects.AlignedTransactionClassification()
	if *command.IsHighEmitting {
		classification = valueobjects.NextStepTransactionClassification()
	}

	if err := transaction.MarkProfessionallyReviewed(classification, now); err != nil {
		return outboundports.TransactionStep4Result{}, fmt.Errorf("marking transaction professionally reviewed: %w", err)
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.Create(txCtx, step4); err != nil {
			return fmt.Errorf("creating transaction step 4: %w", err)
		}

		if err := uc.transactionRepository.Update(txCtx, transaction); err != nil {
			return fmt.Errorf("updating transaction review state: %w", err)
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, createTransactionStep4AdminEventType, map[string]any{
			"action":             "create",
			"resource":           "transaction_step_4",
			"transaction_id":     step4.TransactionID().String(),
			"sector_id":          step4.SectorID().String(),
			"additional_context": step4.AdditionalContext().String(),
			"is_high_emitting":   step4.IsHighEmitting(),
			"classification":     transaction.Classification(),
			"status":             transaction.Status(),
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return outboundports.TransactionStep4Result{}, fmt.Errorf("creating transaction step 4 transaction: %w", err)
	}

	return newTransactionStep4Result(step4, transaction), nil
}

func validateCreateTransactionStep4Command(command inboundports.CreateTransactionStep4Command) *domain.ValidationError {
	var validationErrors []domain.FieldValidationError

	if command.TransactionID.IsZero() {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("transaction_id", "required", "transaction_id is required"))
	}

	if command.SectorID.IsZero() {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("sector_id", "required", "sector_id is required"))
	}

	if command.AdditionalContext.String() == "" {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("additional_context", "required", "additional_context is required"))
	}

	if command.IsHighEmitting == nil {
		validationErrors = append(validationErrors, domain.NewFieldValidationError("is_high_emitting", "required", "is_high_emitting is required"))
	}

	return domain.NewValidationError(validationErrors)
}

func newTransactionStep4Result(step4 *entities.TransactionStep4, transaction *entities.Transaction) outboundports.TransactionStep4Result {
	return outboundports.TransactionStep4Result{
		TransactionID:     step4.TransactionID(),
		SectorID:          step4.SectorID(),
		AdditionalContext: step4.AdditionalContext(),
		IsHighEmitting:    step4.IsHighEmitting(),
		Classification:    transaction.ClassificationValue(),
		Status:            transaction.StatusValue(),
		CreatedAt:         step4.CreatedAt(),
		UpdatedAt:         step4.UpdatedAt(),
	}
}
