package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestUpsertTransactionFeedbackUseCaseExecute verifies the upsert transaction feedback use case execute behavior and the expected outcome asserted below.
func TestUpsertTransactionFeedbackUseCaseExecute(t *testing.T) {
	t.Parallel()

	feedbackID, _ := valueobjects.FeedbackIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	transactionID, _ := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	transaction, _ := entities.NewTransaction(transactionID, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())

	thumbsUp := valueobjects.ThumbsUpFeedbackKind()
	thumbsDown := valueobjects.ThumbsDownFeedbackKind()

	userID, _ := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	existingFeedback, _ := entities.NewFeedback(feedbackID, userID, transactionID, thumbsUp, testTime())

	tests := []struct {
		name         string
		feedbackRepo *feedbackRepositoryMock
		txRepo       *transactionRepositoryMock
		txManager    *transactionManagerMock
		recorder     *adminEventRecorderMock
		command      inboundports.UpsertTransactionFeedbackCommand
		assertError  func(t *testing.T, err error)
		assert       func(t *testing.T, result outboundports.FeedbackResult, repo *feedbackRepositoryMock)
	}{
		{
			name:         "creates new feedback",
			feedbackRepo: &feedbackRepositoryMock{},
			txRepo:       &transactionRepositoryMock{findByID: transaction},
			txManager:    &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command:      inboundports.UpsertTransactionFeedbackCommand{TransactionID: transactionID, Kind: thumbsUp, ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.FeedbackResult, repo *feedbackRepositoryMock) {
				t.Helper()
				if repo.createdFeedback == nil {
					t.Fatal("expected feedback to be created")
				}
				if result.Kind != "thumbs_up" {
					t.Fatalf("result.Kind = %q, want %q", result.Kind, "thumbs_up")
				}
			},
		},
		{
			name:         "updates existing feedback",
			feedbackRepo: &feedbackRepositoryMock{findByUserAndTx: existingFeedback},
			txRepo:       &transactionRepositoryMock{findByID: transaction},
			txManager:    &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command:      inboundports.UpsertTransactionFeedbackCommand{TransactionID: transactionID, Kind: thumbsDown, ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.FeedbackResult, repo *feedbackRepositoryMock) {
				t.Helper()
				if repo.updatedFeedback == nil {
					t.Fatal("expected feedback to be updated")
				}
				if result.Kind != "thumbs_down" {
					t.Fatalf("result.Kind = %q, want %q", result.Kind, "thumbs_down")
				}
			},
		},
		{
			name:         "returns not found when transaction missing",
			feedbackRepo: &feedbackRepositoryMock{},
			txRepo:       &transactionRepositoryMock{},
			txManager:    &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command:      inboundports.UpsertTransactionFeedbackCommand{TransactionID: transactionID, Kind: thumbsUp, ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
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
			name:         "wraps repo create error",
			feedbackRepo: &feedbackRepositoryMock{createErr: errors.New("boom")},
			txRepo:       &transactionRepositoryMock{findByID: transaction},
			txManager:    &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command:      inboundports.UpsertTransactionFeedbackCommand{TransactionID: transactionID, Kind: thumbsUp, ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "upserting transaction feedback") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uc := NewUpsertTransactionFeedbackUseCase(tc.feedbackRepo, tc.txRepo, tc.txManager, tc.recorder)
			uc.newID = func() (valueobjects.FeedbackID, error) { return feedbackID, nil }
			uc.now = testTime

			result, err := uc.Execute(context.Background(), tc.command)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			tc.assert(t, result, tc.feedbackRepo)
		})
	}
}
