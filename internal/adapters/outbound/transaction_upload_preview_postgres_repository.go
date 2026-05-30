package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	saveTransactionUploadPreviewQuery = `
	INSERT INTO transaction_upload_preview (upload_id, columns, rows, total_rows, validation_errors)
	VALUES ($1, $2::jsonb, $3::jsonb, $4, $5::jsonb)
	ON CONFLICT (upload_id) DO UPDATE
	SET columns = EXCLUDED.columns,
		rows = EXCLUDED.rows,
		total_rows = EXCLUDED.total_rows,
		validation_errors = EXCLUDED.validation_errors
	`
	findTransactionUploadPreviewByUploadIDQuery = `
	SELECT upload_id, columns, rows, total_rows, validation_errors
	FROM transaction_upload_preview
	WHERE upload_id = $1
	`
)

// PostgresTransactionUploadPreviewRepository persists upload preview data in PostgreSQL.
type PostgresTransactionUploadPreviewRepository struct {
	pool pgxQuerier
}

var _ ports.TransactionUploadPreviewRepository = (*PostgresTransactionUploadPreviewRepository)(nil)

// NewPostgresTransactionUploadPreviewRepository builds a PostgresTransactionUploadPreviewRepository.
func NewPostgresTransactionUploadPreviewRepository(pool pgxQuerier) *PostgresTransactionUploadPreviewRepository {
	return &PostgresTransactionUploadPreviewRepository{pool: pool}
}

// Save upserts persisted preview data for one upload.
func (r *PostgresTransactionUploadPreviewRepository) Save(ctx context.Context, preview ports.TransactionUploadPreviewRecord) error {
	querier := txQuerierFromContext(ctx, r.pool)

	columnsJSON, err := marshalPreviewJSON(preview.Columns)
	if err != nil {
		return fmt.Errorf("marshalling preview columns: %w", err)
	}

	rowsJSON, err := marshalPreviewJSON(preview.Rows)
	if err != nil {
		return fmt.Errorf("marshalling preview rows: %w", err)
	}

	validationErrorsJSON, err := marshalPreviewJSON(preview.ValidationErrors)
	if err != nil {
		return fmt.Errorf("marshalling preview validation errors: %w", err)
	}

	if _, err := querier.Exec(ctx, saveTransactionUploadPreviewQuery, preview.UploadID, string(columnsJSON), string(rowsJSON), preview.TotalRows, string(validationErrorsJSON)); err != nil {
		return fmt.Errorf("executing save transaction upload preview query: %w", err)
	}

	return nil
}

// FindByUploadID loads persisted preview data for one upload.
func (r *PostgresTransactionUploadPreviewRepository) FindByUploadID(ctx context.Context, uploadID valueobjects.UploadID) (*ports.TransactionUploadPreviewRecord, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	row := querier.QueryRow(ctx, findTransactionUploadPreviewByUploadIDQuery, uploadID.String())

	var preview ports.TransactionUploadPreviewRecord
	var columnsJSON pgtype.Text
	var rowsJSON pgtype.Text
	var validationErrorsJSON pgtype.Text
	if err := row.Scan(&preview.UploadID, &columnsJSON, &rowsJSON, &preview.TotalRows, &validationErrorsJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("scanning transaction upload preview row: %w", err)
	}

	if err := unmarshalPreviewJSON(columnsJSON.String, &preview.Columns); err != nil {
		return nil, fmt.Errorf("unmarshalling preview columns: %w", err)
	}

	if err := unmarshalPreviewJSON(rowsJSON.String, &preview.Rows); err != nil {
		return nil, fmt.Errorf("unmarshalling preview rows: %w", err)
	}

	if err := unmarshalPreviewJSON(validationErrorsJSON.String, &preview.ValidationErrors); err != nil {
		return nil, fmt.Errorf("unmarshalling preview validation errors: %w", err)
	}

	return &preview, nil
}

func marshalPreviewJSON(value any) ([]byte, error) {
	if value == nil {
		return []byte("null"), nil
	}

	return json.Marshal(value)
}

func unmarshalPreviewJSON(data string, target any) error {
	if data == "" {
		data = "null"
	}

	return json.Unmarshal([]byte(data), target)
}
