package usecases

import (
	"context"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestGetTransactionFeedbackUseCaseExecute verifies the get transaction feedback use case execute behavior and the expected outcome asserted below.
func TestGetTransactionFeedbackUseCaseExecute(t *testing.T) {
	t.Parallel()

	feedbackID, _ := valueobjects.FeedbackIDFromString("aabbccddeeff00112233445566778899")
	transactionID, _ := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	userID, _ := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	existing, _ := entities.NewFeedback(feedbackID, userID, transactionID, valueobjects.ThumbsUpFeedbackKind(), testTime())

	tests := []struct {
		name         string
		feedbackRepo *feedbackRepositoryMock
		recorder     *adminEventRecorderMock
		query        inboundports.GetTransactionFeedbackQuery
		assertError  func(t *testing.T, err error)
		assert       func(t *testing.T, result *outboundports.FeedbackResult)
	}{
		{
			name:         "returns feedback when found",
			feedbackRepo: &feedbackRepositoryMock{findByUserAndTx: existing},
			recorder:     &adminEventRecorderMock{},
			query:        inboundports.GetTransactionFeedbackQuery{TransactionID: transactionID, ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result *outboundports.FeedbackResult) {
				t.Helper()
				if result == nil {
					t.Fatal("expected non-nil result")
				}
				if result.Kind != "thumbs_up" {
					t.Fatalf("result.Kind = %q, want %q", result.Kind, "thumbs_up")
				}
			},
		},
		{
			name:         "returns nil when no feedback exists",
			feedbackRepo: &feedbackRepositoryMock{},
			recorder:     &adminEventRecorderMock{},
			query:        inboundports.GetTransactionFeedbackQuery{TransactionID: transactionID, ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result *outboundports.FeedbackResult) {
				t.Helper()
				if result != nil {
					t.Fatalf("expected nil result, got %v", result)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uc := NewGetTransactionFeedbackUseCase(tc.feedbackRepo, tc.recorder)
			uc.now = testTime

			result, err := uc.Execute(context.Background(), tc.query)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			tc.assert(t, result)
		})
	}
}
