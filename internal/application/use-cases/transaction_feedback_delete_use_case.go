package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// DeleteTransactionFeedbackUseCase removes a user's feedback on a transaction.
type DeleteTransactionFeedbackUseCase struct {
	feedbackRepository outboundports.FeedbackRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	now                func() time.Time
}

// NewDeleteTransactionFeedbackUseCase builds a DeleteTransactionFeedbackUseCase.
func NewDeleteTransactionFeedbackUseCase(
	feedbackRepository outboundports.FeedbackRepository,
	transactionManager outboundports.TransactionManager,
	eventRecorder adminEventRecorder,
) *DeleteTransactionFeedbackUseCase {
	return &DeleteTransactionFeedbackUseCase{
		feedbackRepository: feedbackRepository,
		transactionManager: transactionManager,
		eventRecorder:      eventRecorder,
		now:                time.Now,
	}
}

// Execute removes the acting user's feedback on the given transaction.
func (uc *DeleteTransactionFeedbackUseCase) Execute(ctx context.Context, command inboundports.DeleteTransactionFeedbackCommand) error {
	userID, err := valueobjects.UserIDFromString(command.ActorUserID)
	if err != nil {
		return fmt.Errorf("parsing user id: %w", err)
	}

	existing, err := uc.feedbackRepository.FindByUserAndTransaction(ctx, userID, command.TransactionID)
	if err != nil {
		return fmt.Errorf("finding feedback: %w", err)
	}
	if existing == nil {
		return &NotFoundError{Resource: "transaction_feedback", ID: command.TransactionID.String()}
	}

	now := uc.now()

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.feedbackRepository.DeleteByUserAndTransaction(txCtx, userID, command.TransactionID); err != nil {
			return fmt.Errorf("deleting feedback: %w", err)
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, deleteTransactionFeedbackAdminEventType, map[string]any{
			"action":         "delete",
			"resource":       "transaction_feedback",
			"target_id":      existing.ID().String(),
			"transaction_id": existing.TransactionID(),
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return fmt.Errorf("deleting transaction feedback: %w", err)
	}

	return nil
}
