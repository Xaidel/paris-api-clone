package usecases

import (
	"context"
	"errors"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestDeleteTransactionFeedbackUseCaseExecute verifies the delete transaction feedback use case execute behavior and the expected outcome asserted below.
func TestDeleteTransactionFeedbackUseCaseExecute(t *testing.T) {
	t.Parallel()

	feedbackID, _ := valueobjects.FeedbackIDFromString("aabbccddeeff00112233445566778899")
	transactionID, _ := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	userID, _ := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	existing, _ := entities.NewFeedback(feedbackID, userID, transactionID, valueobjects.ThumbsUpFeedbackKind(), testTime())

	tests := []struct {
		name         string
		feedbackRepo *feedbackRepositoryMock
		txManager    *transactionManagerMock
		recorder     *adminEventRecorderMock
		command      inboundports.DeleteTransactionFeedbackCommand
		assertError  func(t *testing.T, err error)
		assert       func(t *testing.T, repo *feedbackRepositoryMock)
	}{
		{
			name:         "deletes feedback",
			feedbackRepo: &feedbackRepositoryMock{findByUserAndTx: existing},
			txManager:    &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command:      inboundports.DeleteTransactionFeedbackCommand{TransactionID: transactionID, ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, repo *feedbackRepositoryMock) {
				t.Helper()
				if repo.deletedTxID != transactionID.String() {
					t.Fatalf("deletedTxID = %q, want %q", repo.deletedTxID, transactionID.String())
				}
			},
		},
		{
			name:         "returns not found when feedback missing",
			feedbackRepo: &feedbackRepositoryMock{},
			txManager:    &transactionManagerMock{},
			recorder:     &adminEventRecorderMock{},
			command:      inboundports.DeleteTransactionFeedbackCommand{TransactionID: transactionID, ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
				if notFoundErr.Resource != "transaction_feedback" {
					t.Fatalf("notFoundErr.Resource = %q, want %q", notFoundErr.Resource, "transaction_feedback")
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uc := NewDeleteTransactionFeedbackUseCase(tc.feedbackRepo, tc.txManager, tc.recorder)
			uc.now = testTime

			err := uc.Execute(context.Background(), tc.command)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			tc.assert(t, tc.feedbackRepo)
		})
	}
}
