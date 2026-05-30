package usecases

import (
	"context"
	"fmt"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// UpsertTransactionFeedbackUseCase creates or updates a transaction feedback.
type UpsertTransactionFeedbackUseCase struct {
	feedbackRepository    outboundports.FeedbackRepository
	transactionRepository outboundports.TransactionRepository
	transactionManager    outboundports.TransactionManager
	eventRecorder         adminEventRecorder
	newID                 func() (valueobjects.FeedbackID, error)
	now                   func() time.Time
}

// NewUpsertTransactionFeedbackUseCase builds a UpsertTransactionFeedbackUseCase.
func NewUpsertTransactionFeedbackUseCase(
	feedbackRepository outboundports.FeedbackRepository,
	transactionRepository outboundports.TransactionRepository,
	transactionManager outboundports.TransactionManager,
	eventRecorder adminEventRecorder,
) *UpsertTransactionFeedbackUseCase {
	return &UpsertTransactionFeedbackUseCase{
		feedbackRepository:    feedbackRepository,
		transactionRepository: transactionRepository,
		transactionManager:    transactionManager,
		eventRecorder:         eventRecorder,
		newID:                 valueobjects.NewFeedbackID,
		now:                   time.Now,
	}
}

// Execute creates or updates a transaction feedback for the acting user.
func (uc *UpsertTransactionFeedbackUseCase) Execute(ctx context.Context, command inboundports.UpsertTransactionFeedbackCommand) (outboundports.FeedbackResult, error) {
	userID, err := valueobjects.UserIDFromString(command.ActorUserID)
	if err != nil {
		return outboundports.FeedbackResult{}, fmt.Errorf("parsing user id: %w", err)
	}

	tx, err := uc.transactionRepository.FindByID(ctx, command.TransactionID)
	if err != nil {
		return outboundports.FeedbackResult{}, fmt.Errorf("checking transaction existence: %w", err)
	}
	if tx == nil {
		return outboundports.FeedbackResult{}, &NotFoundError{Resource: "transaction", ID: command.TransactionID.String()}
	}

	now := uc.now()

	existing, err := uc.feedbackRepository.FindByUserAndTransaction(ctx, userID, command.TransactionID)
	if err != nil {
		return outboundports.FeedbackResult{}, fmt.Errorf("finding existing feedback: %w", err)
	}

	var feedback *entities.Feedback

	if existing != nil {
		if err := existing.ChangeKind(command.Kind, now); err != nil {
			return outboundports.FeedbackResult{}, fmt.Errorf("updating feedback kind: %w", err)
		}
		feedback = existing
	} else {
		id, err := uc.newID()
		if err != nil {
			return outboundports.FeedbackResult{}, fmt.Errorf("generating feedback id: %w", err)
		}

		feedback, err = entities.NewFeedback(id, userID, command.TransactionID, command.Kind, now)
		if err != nil {
			return outboundports.FeedbackResult{}, fmt.Errorf("creating feedback: %w", err)
		}
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if existing != nil {
			if err := uc.feedbackRepository.Update(txCtx, feedback); err != nil {
				return fmt.Errorf("updating feedback: %w", err)
			}
		} else {
			if err := uc.feedbackRepository.Create(txCtx, feedback); err != nil {
				return fmt.Errorf("persisting feedback: %w", err)
			}
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, upsertTransactionFeedbackAdminEventType, map[string]any{
			"action":         "upsert",
			"resource":       "transaction_feedback",
			"target_id":      feedback.ID().String(),
			"transaction_id": feedback.TransactionID(),
			"kind":           feedback.Kind(),
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return outboundports.FeedbackResult{}, fmt.Errorf("upserting transaction feedback: %w", err)
	}

	return newFeedbackResult(feedback), nil
}
