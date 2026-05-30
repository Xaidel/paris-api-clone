package adapters

import (
	"context"
	"regexp"
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/pashagolub/pgxmock/v4"
)

// TestPostgresU1ListRepositoryCreate verifies the postgres U1 list repository create behavior and the expected outcome asserted below.
func TestPostgresU1ListRepositoryCreate(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	id, err := valueobjects.U1ListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("U1ListIDFromString() error = %v", err)
	}

	entry, err := entities.NewU1ListEntry(id, "energy", "grant", "rule 1")
	if err != nil {
		t.Fatalf("NewU1ListEntry() error = %v", err)
	}

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	if err := entry.SetAuditTimestamps(now, now); err != nil {
		t.Fatalf("SetAuditTimestamps() error = %v", err)
	}

	repository := NewPostgresU1ListRepository(mock)
	mock.ExpectExec(regexp.QuoteMeta(createU1ListEntryQuery)).
		WithArgs(entry.ID().String(), entry.Sector(), entry.EligibleOperationType(), entry.ConditionGuidance(), "01962b8f-aeb2-7e03-a8ff-1edce1300002", entry.CreatedAt(), entry.UpdatedAt()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Create(context.Background(), entry, "01962b8f-aeb2-7e03-a8ff-1edce1300002"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

// TestPostgresU1ListRepositoryFindByID verifies the postgres U1 list repository find by ID behavior and the expected outcome asserted below.
func TestPostgresU1ListRepositoryFindByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rows    *pgxmock.Rows
		wantNil bool
	}{
		{name: "returns entry", rows: pgxmock.NewRows([]string{"id", "sector", "eligible_operation_type", "condition_guidance", "created_by", "created_at", "updated_at"}).AddRow("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", "energy", "grant", "rule 1", "01962b8f-aeb2-7e03-a8ff-1edce1300002", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC), time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))},
		{name: "returns nil when not found", rows: pgxmock.NewRows([]string{"id", "sector", "eligible_operation_type", "condition_guidance", "created_by", "created_at", "updated_at"}), wantNil: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			if err != nil {
				t.Fatalf("pgxmock.NewPool() error = %v", err)
			}
			defer mock.Close()

			id, err := valueobjects.U1ListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
			if err != nil {
				t.Fatalf("U1ListIDFromString() error = %v", err)
			}

			mock.ExpectQuery(regexp.QuoteMeta(findU1ListEntryByIDQuery)).WithArgs(id.String()).WillReturnRows(tc.rows)
			repository := NewPostgresU1ListRepository(mock)

			entry, err := repository.FindByID(context.Background(), id)
			if err != nil {
				t.Fatalf("FindByID() error = %v", err)
			}

			if tc.wantNil {
				if entry != nil {
					t.Fatal("expected nil entry")
				}
				return
			}

			if entry.ConditionGuidance() != "rule 1" {
				t.Fatalf("entry.ConditionGuidance() = %q, want %q", entry.ConditionGuidance(), "rule 1")
			}

			if entry.CreatedBy() != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
				t.Fatalf("entry.CreatedBy() = %q, want %q", entry.CreatedBy(), "01962b8f-aeb2-7e03-a8ff-1edce1300002")
			}
		})
	}
}

// TestPostgresU1ListRepositoryListUpdateDelete verifies the postgres U1 list repository list update delete behavior and the expected outcome asserted below.
func TestPostgresU1ListRepositoryListUpdateDelete(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	repository := NewPostgresU1ListRepository(mock)
	id, err := valueobjects.U1ListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("U1ListIDFromString() error = %v", err)
	}
	createdBy, err := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("UserIDFromString() error = %v", err)
	}
	entry := entities.ReconstituteU1ListEntryWithAudit(id, "energy", "grant", "rule 1", createdBy, time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC), time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))

	mock.ExpectQuery(regexp.QuoteMeta(listU1ListEntriesQuery)).WillReturnRows(
		pgxmock.NewRows([]string{"id", "sector", "eligible_operation_type", "condition_guidance", "created_by", "created_at", "updated_at"}).AddRow(id.String(), "energy", "grant", "rule 1", entry.CreatedBy(), entry.CreatedAt(), entry.UpdatedAt()),
	)

	entries, err := repository.List(context.Background(), ports.U1ListFilter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	mock.ExpectExec(regexp.QuoteMeta(updateU1ListEntryQuery)).WithArgs(id.String(), entry.Sector(), entry.EligibleOperationType(), entry.ConditionGuidance(), entry.UpdatedAt()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repository.Update(context.Background(), entry); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(deleteU1ListEntryQuery)).WithArgs(id.String()).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := repository.DeleteByID(context.Background(), id); err != nil {
		t.Fatalf("DeleteByID() error = %v", err)
	}
}

// TestPostgresU1ListRepositoryListBySector verifies the postgres U1 list repository list by sector behavior and the expected outcome asserted below.
func TestPostgresU1ListRepositoryListBySector(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	repository := NewPostgresU1ListRepository(mock)
	mock.ExpectQuery(regexp.QuoteMeta(listU1ListEntriesBySectorQuery)).WithArgs("energy").WillReturnRows(
		pgxmock.NewRows([]string{"id", "sector", "eligible_operation_type", "condition_guidance", "created_by", "created_at", "updated_at"}).AddRow("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", "Energy", "grant", "rule 1", "01962b8f-aeb2-7e03-a8ff-1edce1300002", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC), time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)),
	)

	entries, err := repository.List(context.Background(), ports.U1ListFilter{Sector: "energy"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	if entries[0].Sector() != "Energy" {
		t.Fatalf("entries[0].Sector() = %q, want %q", entries[0].Sector(), "Energy")
	}
}
