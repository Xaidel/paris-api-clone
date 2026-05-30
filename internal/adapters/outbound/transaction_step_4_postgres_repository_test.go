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

// TestPostgresTransactionStep4RepositoryCreate verifies the postgres transaction step 4 repository create behavior and the expected outcome asserted below.
func TestPostgresTransactionStep4RepositoryCreate(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	transactionID, err := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	sectorID, err := valueobjects.SectorIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("SectorIDFromString() error = %v", err)
	}

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	additionalContext, err := valueobjects.NewTransactionStep4AdditionalContext("Reviewed by analyst")
	if err != nil {
		t.Fatalf("NewTransactionStep4AdditionalContext() error = %v", err)
	}

	step4, err := entities.NewTransactionStep4(transactionID, sectorID, additionalContext, true, now)
	if err != nil {
		t.Fatalf("NewTransactionStep4() error = %v", err)
	}

	repository := NewPostgresTransactionStep4Repository(mock)
	mock.ExpectExec(regexp.QuoteMeta(createTransactionStep4Query)).
		WithArgs(step4.TransactionID().String(), step4.SectorID().String(), step4.AdditionalContext().String(), step4.IsHighEmitting(), step4.CreatedAt(), step4.UpdatedAt()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Create(context.Background(), step4); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations were not met: %v", err)
	}
}

// TestPostgresTransactionStep4RepositoryFindByTransactionID verifies the postgres transaction step 4 repository find by transaction ID behavior and the expected outcome asserted below.
func TestPostgresTransactionStep4RepositoryFindByTransactionID(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	transactionID, err := valueobjects.TransactionIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	sectorID, err := valueobjects.SectorIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("SectorIDFromString() error = %v", err)
	}

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	repository := NewPostgresTransactionStep4Repository(mock)
	mock.ExpectQuery(regexp.QuoteMeta(findTransactionStep4ByTransactionIDQuery)).
		WithArgs(transactionID.String()).
		WillReturnRows(pgxmock.NewRows([]string{"transaction_id", "sector_id", "additional_context", "is_high_emitting", "created_at", "updated_at"}).
			AddRow(transactionID.String(), sectorID.String(), "Reviewed by analyst", true, now, now))

	step4, err := repository.FindByTransactionID(context.Background(), transactionID)
	if err != nil {
		t.Fatalf("FindByTransactionID() error = %v", err)
	}

	if step4 == nil {
		t.Fatal("FindByTransactionID() = nil, want step 4")
	}

	if step4.TransactionID().String() != transactionID.String() {
		t.Fatalf("step4.TransactionID() = %q, want %q", step4.TransactionID().String(), transactionID.String())
	}

	if step4.SectorID().String() != sectorID.String() {
		t.Fatalf("step4.SectorID() = %q, want %q", step4.SectorID().String(), sectorID.String())
	}

	if step4.AdditionalContext().String() != "Reviewed by analyst" {
		t.Fatalf("step4.AdditionalContext() = %q, want %q", step4.AdditionalContext().String(), "Reviewed by analyst")
	}

	if !step4.IsHighEmitting() {
		t.Fatal("step4.IsHighEmitting() = false, want true")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations were not met: %v", err)
	}
}
