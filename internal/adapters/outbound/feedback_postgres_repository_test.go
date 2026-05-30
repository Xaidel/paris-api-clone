package adapters

import (
	"context"
	"regexp"
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/pashagolub/pgxmock/v4"
)

// TestPostgresFeedbackRepository verifies the postgres feedback repository behavior and the expected outcome asserted below.
func TestPostgresFeedbackRepository(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	feedbackID, _ := valueobjects.FeedbackIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	userID, _ := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	transactionID, _ := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)

	feedback, err := entities.NewFeedback(feedbackID, userID, transactionID, valueobjects.ThumbsUpFeedbackKind(), now)
	if err != nil {
		t.Fatalf("NewFeedback() error = %v", err)
	}

	repository := NewPostgresFeedbackRepository(mock)

	t.Run("Create", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(createFeedbackQuery)).
			WithArgs(feedback.ID().String(), feedback.UserID(), feedback.TransactionID(), feedback.Kind(), feedback.CreatedAt(), feedback.UpdatedAt()).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		if err := repository.Create(context.Background(), feedback); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	})

	t.Run("FindByUserAndTransaction returns row", func(t *testing.T) {
		rows := pgxmock.NewRows([]string{"id", "user_id", "transaction_id", "kind", "created_at", "updated_at"}).
			AddRow(feedbackID.String(), userID.String(), transactionID.String(), "thumbs_up", now, now)

		mock.ExpectQuery(regexp.QuoteMeta(findFeedbackByUserAndTransactionQuery)).
			WithArgs(userID.String(), transactionID.String()).
			WillReturnRows(rows)

		result, err := repository.FindByUserAndTransaction(context.Background(), userID, transactionID)
		if err != nil {
			t.Fatalf("FindByUserAndTransaction() error = %v", err)
		}
		if result == nil {
			t.Fatal("expected non-nil result")
		}
		if result.Kind() != "thumbs_up" {
			t.Fatalf("Kind() = %q, want %q", result.Kind(), "thumbs_up")
		}
	})

	t.Run("Update", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(updateFeedbackQuery)).
			WithArgs(feedback.UserID(), feedback.TransactionID(), feedback.Kind(), feedback.UpdatedAt()).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		if err := repository.Update(context.Background(), feedback); err != nil {
			t.Fatalf("Update() error = %v", err)
		}
	})

	t.Run("DeleteByUserAndTransaction", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(deleteFeedbackByUserAndTransactionQuery)).
			WithArgs(userID.String(), transactionID.String()).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		if err := repository.DeleteByUserAndTransaction(context.Background(), userID, transactionID); err != nil {
			t.Fatalf("DeleteByUserAndTransaction() error = %v", err)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations were not met: %v", err)
	}
}
