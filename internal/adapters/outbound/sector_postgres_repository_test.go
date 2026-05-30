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

// TestPostgresSectorRepositoryCreate verifies the postgres sector repository create behavior and the expected outcome asserted below.
func TestPostgresSectorRepositoryCreate(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	id, err := valueobjects.SectorIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("SectorIDFromString() error = %v", err)
	}

	sector, err := entities.NewSector(id, "High Emitting", "Steel", "Steel production")
	if err != nil {
		t.Fatalf("NewSector() error = %v", err)
	}

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	if err := sector.SetAuditTimestamps(now, now); err != nil {
		t.Fatalf("SetAuditTimestamps() error = %v", err)
	}

	repository := NewPostgresSectorRepository(mock)
	mock.ExpectExec(regexp.QuoteMeta(createSectorEntryQuery)).
		WithArgs(sector.ID().String(), sector.Type(), sector.Name(), sector.Description(), "01962b8f-aeb2-7e03-a8ff-1edce1300002", sector.CreatedAt(), sector.UpdatedAt()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Create(context.Background(), sector, "01962b8f-aeb2-7e03-a8ff-1edce1300002"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

// TestPostgresSectorRepositoryFindByID verifies the postgres sector repository find by ID behavior and the expected outcome asserted below.
func TestPostgresSectorRepositoryFindByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rows    *pgxmock.Rows
		wantNil bool
	}{
		{name: "returns sector", rows: pgxmock.NewRows([]string{"id", "type", "name", "description", "created_by", "created_at", "updated_at"}).AddRow("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", "High Emitting", "Steel", "Steel production", "01962b8f-aeb2-7e03-a8ff-1edce1300002", time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC), time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))},
		{name: "returns nil when not found", rows: pgxmock.NewRows([]string{"id", "type", "name", "description", "created_by", "created_at", "updated_at"}), wantNil: true},
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

			id, err := valueobjects.SectorIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
			if err != nil {
				t.Fatalf("SectorIDFromString() error = %v", err)
			}

			mock.ExpectQuery(regexp.QuoteMeta(findSectorEntryByIDQuery)).WithArgs(id.String()).WillReturnRows(tc.rows)
			repository := NewPostgresSectorRepository(mock)

			sector, err := repository.FindByID(context.Background(), id)
			if err != nil {
				t.Fatalf("FindByID() error = %v", err)
			}

			if tc.wantNil {
				if sector != nil {
					t.Fatal("expected nil sector")
				}
				return
			}

			if sector.Description() != "Steel production" {
				t.Fatalf("sector.Description() = %q, want %q", sector.Description(), "Steel production")
			}

			if sector.CreatedBy() != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
				t.Fatalf("sector.CreatedBy() = %q, want %q", sector.CreatedBy(), "01962b8f-aeb2-7e03-a8ff-1edce1300002")
			}
		})
	}
}

// TestPostgresSectorRepositoryListUpdateDelete verifies the postgres sector repository list update delete behavior and the expected outcome asserted below.
func TestPostgresSectorRepositoryListUpdateDelete(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	repository := NewPostgresSectorRepository(mock)
	id, err := valueobjects.SectorIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("SectorIDFromString() error = %v", err)
	}
	sectorType, err := valueobjects.SectorTypeFromString("High Emitting")
	if err != nil {
		t.Fatalf("SectorTypeFromString() error = %v", err)
	}
	createdBy, err := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("UserIDFromString() error = %v", err)
	}
	sector := entities.ReconstituteSectorWithAudit(id, sectorType, "Steel", "Steel production", createdBy, time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC), time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC))

	mock.ExpectQuery(regexp.QuoteMeta(listSectorEntriesQuery)).WillReturnRows(
		pgxmock.NewRows([]string{"id", "type", "name", "description", "created_by", "created_at", "updated_at"}).AddRow(id.String(), "High Emitting", "Steel", "Steel production", sector.CreatedBy(), sector.CreatedAt(), sector.UpdatedAt()),
	)

	sectors, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(sectors) != 1 {
		t.Fatalf("len(sectors) = %d, want 1", len(sectors))
	}

	mock.ExpectExec(regexp.QuoteMeta(updateSectorEntryQuery)).WithArgs(id.String(), sector.Type(), sector.Name(), sector.Description(), sector.UpdatedAt()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repository.Update(context.Background(), sector); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(deleteSectorEntryQuery)).WithArgs(id.String()).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := repository.DeleteByID(context.Background(), id); err != nil {
		t.Fatalf("DeleteByID() error = %v", err)
	}
}
