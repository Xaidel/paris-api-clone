package adapters

import (
	"context"
	"regexp"
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
)

// TestPostgresTransactionStep5RepositoryCreate verifies the postgres transaction step 5 repository create behavior and the expected outcome asserted below.
func TestPostgresTransactionStep5RepositoryCreate(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	transactionID, err := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300201")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	question1Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion1Justification("question 1 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion1Justification() error = %v", err)
	}

	question2Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion2Justification("question 2 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion2Justification() error = %v", err)
	}

	now := time.Date(2026, time.April, 4, 11, 0, 0, 0, time.UTC)
	step5, err := entities.NewTransactionStep5(
		transactionID,
		false,
		question1Justification,
		true,
		question2Justification,
		valueobjects.NewTransactionStep5ReviewerNotes(step5StringPointer("  follow up required  ")),
		true,
		now,
	)
	if err != nil {
		t.Fatalf("NewTransactionStep5() error = %v", err)
	}

	repository := NewPostgresTransactionStep5Repository(mock)
	mock.ExpectExec(regexp.QuoteMeta(createTransactionStep5Query)).
		WithArgs(
			step5.TransactionID().String(),
			step5.ScreeningQuestion1Answer(),
			step5.ScreeningQuestion1Justification().String(),
			step5.ScreeningQuestion2Answer(),
			step5.ScreeningQuestion2Justification().String(),
			"follow up required",
			step5.IsFinal(),
			step5.CreatedAt(),
			step5.UpdatedAt(),
		).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Create(context.Background(), step5); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations were not met: %v", err)
	}
}

// TestPostgresTransactionStep5RepositoryFindByTransactionID verifies the postgres transaction step 5 repository find by transaction ID behavior and the expected outcome asserted below.
func TestPostgresTransactionStep5RepositoryFindByTransactionID(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	transactionID, err := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300201")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	now := time.Date(2026, time.April, 4, 11, 0, 0, 0, time.UTC)
	repository := NewPostgresTransactionStep5Repository(mock)
	mock.ExpectQuery(regexp.QuoteMeta(findTransactionStep5ByTransactionIDQuery)).
		WithArgs(transactionID.String()).
		WillReturnRows(pgxmock.NewRows([]string{
			"transaction_id",
			"screening_question_1_answer",
			"screening_question_1_justification",
			"screening_question_2_answer",
			"screening_question_2_justification",
			"reviewer_notes",
			"is_final",
			"created_at",
			"updated_at",
		}).AddRow(transactionID.String(), false, "question 1 justification", true, "question 2 justification", nil, true, now, now))

	step5, err := repository.FindByTransactionID(context.Background(), transactionID)
	if err != nil {
		t.Fatalf("FindByTransactionID() error = %v", err)
	}

	if step5 == nil {
		t.Fatal("FindByTransactionID() = nil, want step 5")
	}

	if step5.TransactionID().String() != transactionID.String() {
		t.Fatalf("step5.TransactionID() = %q, want %q", step5.TransactionID().String(), transactionID.String())
	}

	if step5.ScreeningQuestion1Answer() {
		t.Fatal("step5.ScreeningQuestion1Answer() = true, want false")
	}

	if step5.ScreeningQuestion1Justification().String() != "question 1 justification" {
		t.Fatalf("step5.ScreeningQuestion1Justification().String() = %q, want %q", step5.ScreeningQuestion1Justification().String(), "question 1 justification")
	}

	if !step5.ScreeningQuestion2Answer() {
		t.Fatal("step5.ScreeningQuestion2Answer() = false, want true")
	}

	if step5.ScreeningQuestion2Justification().String() != "question 2 justification" {
		t.Fatalf("step5.ScreeningQuestion2Justification().String() = %q, want %q", step5.ScreeningQuestion2Justification().String(), "question 2 justification")
	}

	if step5.ReviewerNotes().String() != nil {
		t.Fatalf("step5.ReviewerNotes().String() = %v, want nil", step5.ReviewerNotes().String())
	}

	if !step5.IsFinal() {
		t.Fatal("step5.IsFinal() = false, want true")
	}

	if step5.Classification().String() != "not-aligned" {
		t.Fatalf("step5.Classification().String() = %q, want %q", step5.Classification().String(), "not-aligned")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations were not met: %v", err)
	}
}

// TestPostgresTransactionStep5RepositoryFindByTransactionIDReturnsNilWhenMissing verifies the postgres transaction step 5 repository find by transaction ID returns nil when missing behavior and the expected outcome asserted below.
func TestPostgresTransactionStep5RepositoryFindByTransactionIDReturnsNilWhenMissing(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	transactionID, err := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300201")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	repository := NewPostgresTransactionStep5Repository(mock)
	mock.ExpectQuery(regexp.QuoteMeta(findTransactionStep5ByTransactionIDQuery)).
		WithArgs(transactionID.String()).
		WillReturnError(pgx.ErrNoRows)

	step5, err := repository.FindByTransactionID(context.Background(), transactionID)
	if err != nil {
		t.Fatalf("FindByTransactionID() error = %v", err)
	}

	if step5 != nil {
		t.Fatalf("FindByTransactionID() = %v, want nil", step5)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations were not met: %v", err)
	}
}

func step5StringPointer(value string) *string {
	return &value
}
