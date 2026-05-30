package adapters

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/jackc/pgx/v5"
)

const (
	createTransactionUploadQuery = `
	INSERT INTO transaction_upload (id, group_id, file_name, file_format, content_md5, storage_provider, storage_key, schema_version, status, row_count, uploaded_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	findTransactionUploadByIDQuery = `
	SELECT id, group_id, file_name, file_format, content_md5, storage_provider, storage_key, schema_version, status, row_count, uploaded_at
	FROM transaction_upload
	WHERE id = $1
	`
	findTransactionUploadByContentMD5Query = `
	SELECT id, group_id, file_name, file_format, content_md5, storage_provider, storage_key, schema_version, status, row_count, uploaded_at
	FROM transaction_upload
	WHERE content_md5 = $1 AND group_id = $2
	`
	baseListTransactionUploadsQuery = `
	SELECT id, group_id, file_name, file_format, content_md5, storage_provider, storage_key, schema_version, status, row_count, uploaded_at
	FROM transaction_upload
	`
	deleteTransactionUploadQuery = `
DELETE FROM transaction_upload
WHERE id = $1
`
)

// PostgresTransactionUploadRepository persists transaction upload metadata in PostgreSQL.
type PostgresTransactionUploadRepository struct {
	pool pgxQuerier
}

// NewPostgresTransactionUploadRepository builds a PostgresTransactionUploadRepository.
func NewPostgresTransactionUploadRepository(pool pgxQuerier) *PostgresTransactionUploadRepository {
	return &PostgresTransactionUploadRepository{pool: pool}
}

// Create inserts a transaction upload record.
func (r *PostgresTransactionUploadRepository) Create(ctx context.Context, upload *entities.TransactionUpload) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(ctx, createTransactionUploadQuery, upload.ID().String(), upload.GroupID().String(), upload.FileName(), upload.FileFormat(), upload.ContentMD5(), upload.StorageProvider(), upload.StorageKey(), upload.SchemaVersion(), upload.Status(), upload.RowCount(), upload.UploadedAt()); err != nil {
		return fmt.Errorf("executing create transaction upload query: %w", err)
	}

	return nil
}

// FindByID returns an upload by identifier.
func (r *PostgresTransactionUploadRepository) FindByID(ctx context.Context, id valueobjects.UploadID) (*entities.TransactionUpload, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	uploadRecord, err := scanTransactionUploadRecord(querier.QueryRow(ctx, findTransactionUploadByIDQuery, id.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning transaction upload by id: %w", err)
	}
	if uploadRecord == nil {
		return nil, nil
	}

	return r.reconstituteTransactionUpload(ctx, querier, uploadRecord)
}

// FindByContentMD5 returns an upload by content hash within one group.
func (r *PostgresTransactionUploadRepository) FindByContentMD5(ctx context.Context, contentMD5 string, groupID valueobjects.GroupID) (*entities.TransactionUpload, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	uploadRecord, err := scanTransactionUploadRecord(querier.QueryRow(ctx, findTransactionUploadByContentMD5Query, strings.TrimSpace(strings.ToLower(contentMD5)), groupID.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning transaction upload by content md5: %w", err)
	}
	if uploadRecord == nil {
		return nil, nil
	}

	return r.reconstituteTransactionUpload(ctx, querier, uploadRecord)
}

// List returns upload history.
func (r *PostgresTransactionUploadRepository) List(ctx context.Context, filter ports.TransactionUploadFilter) ([]*entities.TransactionUpload, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	query, args := buildListTransactionUploadsQuery(filter)
	rows, err := querier.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying transaction uploads: %w", err)
	}
	defer rows.Close()

	uploads := make([]*entities.TransactionUpload, 0)
	for rows.Next() {
		uploadRecord, scanErr := scanTransactionUploadRecord(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scanning listed transaction upload: %w", scanErr)
		}

		upload, reconstituteErr := r.reconstituteTransactionUpload(ctx, querier, uploadRecord)
		if reconstituteErr != nil {
			return nil, fmt.Errorf("reconstituting listed transaction upload: %w", reconstituteErr)
		}

		uploads = append(uploads, upload)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating transaction uploads: %w", err)
	}

	return uploads, nil
}

// DeleteByID deletes an upload by identifier.
func (r *PostgresTransactionUploadRepository) DeleteByID(ctx context.Context, id valueobjects.UploadID) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, deleteTransactionUploadQuery, id.String())
	if err != nil {
		return fmt.Errorf("executing delete transaction upload query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

type transactionUploadRecord struct {
	uploadID        valueobjects.UploadID
	groupID         valueobjects.GroupID
	fileName        string
	fileFormat      string
	contentMD5      string
	storageProvider string
	storageKey      string
	schemaVersion   string
	status          valueobjects.TransactionUploadStatus
	rowCount        int
	uploadedAt      time.Time
}

func scanTransactionUploadRecord(row scanner) (*transactionUploadRecord, error) {
	var (
		uploadID        string
		groupID         string
		fileName        string
		fileFormat      string
		contentMD5      string
		storageProvider string
		storageKey      string
		schemaVersion   string
		status          string
		rowCount        int
		uploadedAt      time.Time
	)

	if err := row.Scan(&uploadID, &groupID, &fileName, &fileFormat, &contentMD5, &storageProvider, &storageKey, &schemaVersion, &status, &rowCount, &uploadedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	parsedUploadID, err := valueobjects.UploadIDFromString(uploadID)
	if err != nil {
		return nil, fmt.Errorf("parsing upload id: %w", err)
	}

	parsedGroupID, err := valueobjects.GroupIDFromString(groupID)
	if err != nil {
		return nil, fmt.Errorf("parsing group id: %w", err)
	}

	parsedStatus, err := valueobjects.TransactionUploadStatusFromString(status)
	if err != nil {
		return nil, fmt.Errorf("parsing transaction upload status: %w", err)
	}

	return &transactionUploadRecord{
		uploadID:        parsedUploadID,
		groupID:         parsedGroupID,
		fileName:        fileName,
		fileFormat:      fileFormat,
		contentMD5:      contentMD5,
		storageProvider: storageProvider,
		storageKey:      storageKey,
		schemaVersion:   schemaVersion,
		status:          parsedStatus,
		rowCount:        rowCount,
		uploadedAt:      uploadedAt,
	}, nil
}

func (r *PostgresTransactionUploadRepository) reconstituteTransactionUpload(ctx context.Context, querier pgxQuerier, record *transactionUploadRecord) (*entities.TransactionUpload, error) {
	upload, err := entities.ReconstituteTransactionUpload(record.uploadID, record.groupID, record.fileName, record.fileFormat, record.contentMD5, record.storageProvider, record.storageKey, record.schemaVersion, record.status, record.rowCount, record.uploadedAt)
	if err != nil {
		return nil, fmt.Errorf("reconstituting transaction upload: %w", err)
	}

	return upload, nil
}

func buildListTransactionUploadsQuery(filter ports.TransactionUploadFilter) (string, []any) {
	conditions := make([]string, 0)
	args := make([]any, 0)

	conditions = append(conditions, fmt.Sprintf("group_id = $%d", len(args)+1))
	args = append(args, filter.GroupID.String())

	if strings.TrimSpace(filter.FileName) != "" {
		conditions = append(conditions, fmt.Sprintf("file_name ILIKE $%d", len(args)+1))
		args = append(args, "%"+strings.TrimSpace(filter.FileName)+"%")
	}

	if filter.StartedAt != nil {
		conditions = append(conditions, fmt.Sprintf("uploaded_at >= $%d", len(args)+1))
		args = append(args, *filter.StartedAt)
	}

	if filter.EndedAt != nil {
		conditions = append(conditions, fmt.Sprintf("uploaded_at <= $%d", len(args)+1))
		args = append(args, *filter.EndedAt)
	}

	query := baseListTransactionUploadsQuery
	if len(conditions) > 0 {
		query += "WHERE " + strings.Join(conditions, " AND ") + "\n"
	}

	query += "ORDER BY uploaded_at DESC"

	return query, args
}
