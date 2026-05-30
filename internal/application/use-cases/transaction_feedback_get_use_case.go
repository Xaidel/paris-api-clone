package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// GetTransactionFeedbackUseCase retrieves the feedback a user gave on a transaction.
type GetTransactionFeedbackUseCase struct {
	feedbackRepository outboundports.FeedbackRepository
	eventRecorder      adminEventRecorder
	now                func() time.Time
}

// NewGetTransactionFeedbackUseCase builds a GetTransactionFeedbackUseCase.
func NewGetTransactionFeedbackUseCase(
	feedbackRepository outboundports.FeedbackRepository,
	eventRecorder adminEventRecorder,
) *GetTransactionFeedbackUseCase {
	return &GetTransactionFeedbackUseCase{
		feedbackRepository: feedbackRepository,
		eventRecorder:      eventRecorder,
		now:                time.Now,
	}
}

// Execute returns the acting user's current feedback for the given transaction, or nil if none exists.
func (uc *GetTransactionFeedbackUseCase) Execute(ctx context.Context, query inboundports.GetTransactionFeedbackQuery) (*outboundports.FeedbackResult, error) {
	userID, err := valueobjects.UserIDFromString(query.ActorUserID)
	if err != nil {
		return nil, fmt.Errorf("parsing user id: %w", err)
	}

	feedback, err := uc.feedbackRepository.FindByUserAndTransaction(ctx, userID, query.TransactionID)
	if err != nil {
		return nil, fmt.Errorf("finding feedback: %w", err)
	}

	now := uc.now()

	if err := recordAdminEvent(ctx, uc.eventRecorder, now, query.ActorUserID, query.ActorGroupID, getTransactionFeedbackAdminEventType, map[string]any{
		"action":         "read",
		"resource":       "transaction_feedback",
		"transaction_id": query.TransactionID.String(),
	}); err != nil {
		return nil, err
	}

	if feedback == nil {
		return nil, nil
	}

	result := newFeedbackResult(feedback)
	return &result, nil
}
