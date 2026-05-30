package adapters

import (
	"context"
	"regexp"
	"testing"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/pashagolub/pgxmock/v4"
)

func TestPostgresTransactionUploadPreviewRepositorySaveAndFindByUploadID(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	preview := ports.TransactionUploadPreviewRecord{
		UploadID:  uploadID.String(),
		Columns:   []string{"Product", "Year"},
		Rows:      [][]string{{"CG", "2026"}},
		TotalRows: 1,
		ValidationErrors: []ports.TransactionFileValidationError{{
			Code:        "missing_required_value",
			Message:     "row 2 column \"Year\" is required",
			RowNumber:   2,
			ColumnName:  "Year",
			ColumnIndex: 2,
		}},
	}

	repository := NewPostgresTransactionUploadPreviewRepository(mock)
	mock.ExpectExec(regexp.QuoteMeta(saveTransactionUploadPreviewQuery)).
		WithArgs(preview.UploadID, toJSONB(t, preview.Columns), toJSONB(t, preview.Rows), preview.TotalRows, toJSONB(t, preview.ValidationErrors)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Save(context.Background(), preview); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	rows := pgxmock.NewRows([]string{"upload_id", "columns", "rows", "total_rows", "validation_errors"}).
		AddRow(preview.UploadID, toJSONB(t, preview.Columns), toJSONB(t, preview.Rows), preview.TotalRows, toJSONB(t, preview.ValidationErrors))
	mock.ExpectQuery(regexp.QuoteMeta(findTransactionUploadPreviewByUploadIDQuery)).WithArgs(uploadID.String()).WillReturnRows(rows)

	loaded, err := repository.FindByUploadID(context.Background(), uploadID)
	if err != nil {
		t.Fatalf("FindByUploadID() error = %v", err)
	}

	if loaded == nil {
		t.Fatal("expected preview")
	}

	if loaded.UploadID != preview.UploadID {
		t.Fatalf("loaded.UploadID = %q, want %q", loaded.UploadID, preview.UploadID)
	}

	if len(loaded.Columns) != 2 {
		t.Fatalf("len(loaded.Columns) = %d, want 2", len(loaded.Columns))
	}

	if len(loaded.Rows) != 1 {
		t.Fatalf("len(loaded.Rows) = %d, want 1", len(loaded.Rows))
	}

	if loaded.Rows[0][0] != "CG" {
		t.Fatalf("loaded.Rows[0][0] = %q, want %q", loaded.Rows[0][0], "CG")
	}

	if len(loaded.ValidationErrors) != 1 {
		t.Fatalf("len(loaded.ValidationErrors) = %d, want 1", len(loaded.ValidationErrors))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func TestPostgresTransactionUploadPreviewRepositorySavePreservesNilRows(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	repository := NewPostgresTransactionUploadPreviewRepository(mock)
	preview := ports.TransactionUploadPreviewRecord{
		UploadID:         "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f",
		Columns:          []string{"Product"},
		Rows:             nil,
		TotalRows:        0,
		ValidationErrors: nil,
	}

	mock.ExpectExec(regexp.QuoteMeta(saveTransactionUploadPreviewQuery)).
		WithArgs(preview.UploadID, toJSONB(t, preview.Columns), toJSONB(t, preview.Rows), preview.TotalRows, toJSONB(t, preview.ValidationErrors)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Save(context.Background(), preview); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

func toJSONB(t *testing.T, value any) string {
	t.Helper()

	data, err := marshalPreviewJSON(value)
	if err != nil {
		t.Fatalf("marshalPreviewJSON() error = %v", err)
	}

	return string(data)
}
