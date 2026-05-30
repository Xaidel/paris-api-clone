package adapters

import (
	"context"
	"regexp"
	"testing"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/pashagolub/pgxmock/v4"
)

// TestPostgresClassificationListRepositoryGetEntries verifies the postgres classification list repository get entries behavior and the expected outcome asserted below.
func TestPostgresClassificationListRepositoryGetEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		listType valueobjects.ListType
		query    string
		want     string
	}{
		{name: "u1 entries", listType: valueobjects.U1ListType(), query: listClassificationEntriesU1Query, want: "sector: energy; eligible_operation_type: loan; condition_guidance: rule 1"},
		{name: "u2 entries", listType: valueobjects.U2ListType(), query: listClassificationEntriesU2Query, want: "activity_type: coal"},
		{name: "sector entries", listType: valueobjects.SectorListType(), query: listClassificationEntriesSectorQuery, want: "type: PA Aligned; name: steel; description: green steel"},
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

			repository := NewPostgresClassificationListRepository(mock)
			rows := pgxmock.NewRows([]string{"source_row_id", "entry_text"}).AddRow("source-1", tt.want)
			mock.ExpectQuery(regexp.QuoteMeta(tt.query)).WillReturnRows(rows)

			entries, err := repository.GetEntries(context.Background(), tt.listType)
			if err != nil {
				t.Fatalf("GetEntries() error = %v", err)
			}

			if len(entries) != 1 {
				t.Fatalf("len(entries) = %d, want %d", len(entries), 1)
			}

			if entries[0] != tt.want {
				t.Fatalf("entries[0] = %q, want %q", entries[0], tt.want)
			}
		})
	}
}

// TestPostgresClassificationListRepositoryGetEntryDocuments verifies the postgres classification list repository get entry documents behavior and the expected outcome asserted below.
func TestPostgresClassificationListRepositoryGetEntryDocuments(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	repository := NewPostgresClassificationListRepository(mock)
	rows := pgxmock.NewRows([]string{"source_row_id", "entry_text"}).AddRow("source-1", "activity_type: coal")
	mock.ExpectQuery(regexp.QuoteMeta(listClassificationEntriesU2Query)).WillReturnRows(rows)

	entryDocuments, err := repository.GetEntryDocuments(context.Background(), valueobjects.U2ListType())
	if err != nil {
		t.Fatalf("GetEntryDocuments() error = %v", err)
	}

	if len(entryDocuments) != 1 {
		t.Fatalf("len(entryDocuments) = %d, want %d", len(entryDocuments), 1)
	}

	if entryDocuments[0].SourceRowID() != "source-1" {
		t.Fatalf("entryDocuments[0].SourceRowID() = %q, want %q", entryDocuments[0].SourceRowID(), "source-1")
	}

	if entryDocuments[0].EntryText().String() != "activity_type: coal" {
		t.Fatalf("entryDocuments[0].EntryText().String() = %q, want %q", entryDocuments[0].EntryText().String(), "activity_type: coal")
	}
}

// TestPostgresClassificationListRepositoryGetEntryDocumentsUnsupportedListType verifies the postgres classification list repository get entry documents unsupported list type behavior and the expected outcome asserted below.
func TestPostgresClassificationListRepositoryGetEntryDocumentsUnsupportedListType(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	repository := NewPostgresClassificationListRepository(mock)
	unsupportedListType := valueobjects.ListType{}

	_, err = repository.GetEntryDocuments(context.Background(), unsupportedListType)
	if err == nil {
		t.Fatal("GetEntryDocuments() error = nil, want unsupported list type error")
	}

	want := "selecting \"\" classification list query: unsupported classification list type \"\""
	if err.Error() != want {
		t.Fatalf("GetEntryDocuments() error = %q, want %q", err.Error(), want)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}
