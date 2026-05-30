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

// TestPostgresExclusionListRepositoryCreate verifies the postgres exclusion list repository create behavior and the expected outcome asserted below.
func TestPostgresExclusionListRepositoryCreate(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	id, err := valueobjects.ExclusionListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("ExclusionListIDFromString() error = %v", err)
	}

	entry, err := entities.NewExclusionListEntry(id, "agriculture")
	if err != nil {
		t.Fatalf("NewExclusionListEntry() error = %v", err)
	}

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	if err := entry.SetAuditTimestamps(now, now); err != nil {
		t.Fatalf("SetAuditTimestamps() error = %v", err)
	}

	repository := NewPostgresExclusionListRepository(mock)
	mock.ExpectExec(regexp.QuoteMeta(createExclusionListEntryQuery)).
		WithArgs(entry.ID().String(), entry.ActivityType(), "01962b8f-aeb2-7e03-a8ff-1edce1300002", entry.CreatedAt(), entry.UpdatedAt()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Create(context.Background(), entry, "01962b8f-aeb2-7e03-a8ff-1edce1300002"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

// TestPostgresExclusionListRepositoryFindByID verifies the postgres exclusion list repository find by ID behavior and the expected outcome asserted below.
func TestPostgresExclusionListRepositoryFindByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rows    *pgxmock.Rows
		wantNil bool
	}{
		{name: "returns entry", rows: pgxmock.NewRows([]string{"id", "activity_type", "created_by", "created_at", "updated_at"}).AddRow("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", "agriculture", "01962b8f-aeb2-7e03-a8ff-1edce1300002", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC), time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))},
		{name: "returns nil when not found", rows: pgxmock.NewRows([]string{"id", "activity_type", "created_by", "created_at", "updated_at"}), wantNil: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			if err != nil {
				t.Fatalf("pgxmock.NewPool() error = %v", err)
			}
			defer mock.Close()

			id, err := valueobjects.ExclusionListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
			if err != nil {
				t.Fatalf("ExclusionListIDFromString() error = %v", err)
			}

			mock.ExpectQuery(regexp.QuoteMeta(findExclusionListEntryByIDQuery)).WithArgs(id.String()).WillReturnRows(tt.rows)
			repository := NewPostgresExclusionListRepository(mock)

			entry, err := repository.FindByID(context.Background(), id)
			if err != nil {
				t.Fatalf("FindByID() error = %v", err)
			}

			if tt.wantNil {
				if entry != nil {
					t.Fatal("expected nil entry")
				}
				return
			}

			if entry.ActivityType() != "agriculture" {
				t.Fatalf("entry.ActivityType() = %q, want %q", entry.ActivityType(), "agriculture")
			}

			if entry.CreatedBy() != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
				t.Fatalf("entry.CreatedBy() = %q, want %q", entry.CreatedBy(), "01962b8f-aeb2-7e03-a8ff-1edce1300002")
			}
		})
	}
}

// TestPostgresExclusionListRepositoryListUpdateDelete verifies the postgres exclusion list repository list update delete behavior and the expected outcome asserted below.
func TestPostgresExclusionListRepositoryListUpdateDelete(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	repository := NewPostgresExclusionListRepository(mock)
	id, err := valueobjects.ExclusionListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("ExclusionListIDFromString() error = %v", err)
	}
	createdBy, err := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("UserIDFromString() error = %v", err)
	}
	entry := entities.ReconstituteExclusionListEntryWithAudit(id, "agriculture", createdBy, time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC), time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))

	mock.ExpectQuery(regexp.QuoteMeta(listExclusionListEntriesQuery)).WillReturnRows(
		pgxmock.NewRows([]string{"id", "activity_type", "created_by", "created_at", "updated_at"}).AddRow(id.String(), "agriculture", entry.CreatedBy(), entry.CreatedAt(), entry.UpdatedAt()),
	)

	entries, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	mock.ExpectExec(regexp.QuoteMeta(updateExclusionListEntryQuery)).WithArgs(id.String(), entry.ActivityType(), entry.UpdatedAt()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repository.Update(context.Background(), entry); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(deleteExclusionListEntryQuery)).WithArgs(id.String()).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := repository.DeleteByID(context.Background(), id); err != nil {
		t.Fatalf("DeleteByID() error = %v", err)
	}
}
