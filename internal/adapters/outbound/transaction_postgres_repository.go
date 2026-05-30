package adapters

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/jackc/pgx/v5"
)

const (
	createTransactionQuery = `
INSERT INTO transactions (
    id,
    upload_id,
    row_number,
    product,
    processed_year,
    processed_month,
    dmc_ib,
    dmc,
    partner_bank,
    ref_num,
    transaction_value,
    classification,
    status,
    pipeline_result,
    failure_reason,
    transaction_count,
    goods_description,
    goods_classification,
    applicant_country,
    beneficiary_country,
    source_country,
	    destination_country,
	    tenor_description,
	    es_category,
	    pa_alignment,
	    created_by,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14::jsonb, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
`
	updateTransactionQuery = `
UPDATE transactions
SET upload_id = $2,
    row_number = $3,
    product = $4,
    processed_year = $5,
    processed_month = $6,
    dmc_ib = $7,
    dmc = $8,
    partner_bank = $9,
    ref_num = $10,
    transaction_value = $11,
    classification = $12,
    status = $13,
    pipeline_result = $14::jsonb,
    failure_reason = $15,
    transaction_count = $16,
    goods_description = $17,
    goods_classification = $18,
    applicant_country = $19,
    beneficiary_country = $20,
    source_country = $21,
    destination_country = $22,
    tenor_description = $23,
    es_category = $24,
    pa_alignment = $25,
    created_at = $26,
    updated_at = $27
WHERE id = $1
`
	findTransactionByIDQuery = `
SELECT id, upload_id, row_number, product, processed_year, processed_month, dmc_ib, dmc, partner_bank, ref_num, transaction_value, classification, status, pipeline_result, failure_reason, transaction_count, goods_description, goods_classification, applicant_country, beneficiary_country, source_country, destination_country, tenor_description, es_category, pa_alignment, created_by, created_at, updated_at
FROM transactions
WHERE id = $1
`
	findHistoricalClassificationByExactGoodsDescriptionQuery = `
SELECT id, goods_description, classification, status, pipeline_result
FROM transactions
WHERE goods_description = $1
  AND status <> 'failed'
  AND pipeline_result->>'version' = $2
  AND pipeline_result->>'classifier_family' = $3
  AND pipeline_result->>'prompt_version' = $4
ORDER BY updated_at DESC, created_at DESC
LIMIT 1
`
	baseListTransactionsQuery = `
	SELECT id, upload_id, row_number, product, processed_year, processed_month, dmc_ib, dmc, partner_bank, ref_num, transaction_value, classification, status, pipeline_result, failure_reason, transaction_count, goods_description, goods_classification, applicant_country, beneficiary_country, source_country, destination_country, tenor_description, es_category, pa_alignment, created_by, created_at, updated_at
FROM transactions
`
	baseListTransactionsByUploadIDsQuery = `
SELECT id, upload_id, row_number, product, processed_year, processed_month, dmc_ib, dmc, partner_bank, ref_num, transaction_value, classification, status, pipeline_result, failure_reason, transaction_count, goods_description, goods_classification, applicant_country, beneficiary_country, source_country, destination_country, tenor_description, es_category, pa_alignment, created_by, created_at, updated_at
FROM transactions
`
	deleteTransactionByIDQuery = `
DELETE FROM transactions
WHERE id = $1
`
	hasProcessingTransactionsByUploadIDQuery = `
SELECT EXISTS (
    SELECT 1
    FROM transactions
    WHERE upload_id = $1
      AND status = $2
)
`
	deleteTransactionsByUploadIDQuery = `
DELETE FROM transactions
WHERE upload_id = $1
`
)

// PostgresTransactionRepository persists transaction rows in PostgreSQL.
type PostgresTransactionRepository struct {
	pool pgxQuerier
}

// NewPostgresTransactionRepository builds a PostgresTransactionRepository.
func NewPostgresTransactionRepository(pool pgxQuerier) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{pool: pool}
}

// Create inserts a single transaction row.
func (r *PostgresTransactionRepository) Create(ctx context.Context, transaction *entities.Transaction, createdByUserID string) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(ctx, createTransactionQuery, createTransactionExecArgs(transaction, createdByUserID)...); err != nil {
		return fmt.Errorf("executing create transaction query: %w", err)
	}

	return nil
}

// CreateMany inserts transaction rows.
func (r *PostgresTransactionRepository) CreateMany(ctx context.Context, transactions []*entities.Transaction, createdByUserID string) error {
	querier := txQuerierFromContext(ctx, r.pool)
	for _, transaction := range transactions {
		if _, err := querier.Exec(ctx, createTransactionQuery, createTransactionExecArgs(transaction, createdByUserID)...); err != nil {
			return fmt.Errorf("executing create transaction query for transaction %s: %w", transaction.ID().String(), err)
		}
	}

	return nil
}

// Update updates a single transaction row.
func (r *PostgresTransactionRepository) Update(ctx context.Context, transaction *entities.Transaction) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, updateTransactionQuery, updateTransactionExecArgs(transaction)...)
	if err != nil {
		return fmt.Errorf("executing update transaction query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// FindByID returns a transaction row by identifier.
func (r *PostgresTransactionRepository) FindByID(ctx context.Context, id valueobjects.TransactionID) (*entities.Transaction, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	transaction, err := scanTransaction(querier.QueryRow(ctx, findTransactionByIDQuery, id.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning transaction by id: %w", err)
	}

	return transaction, nil
}

// FindHistoricalClassificationByExactGoodsDescription returns one exact historical ReAct classification match.
func (r *PostgresTransactionRepository) FindHistoricalClassificationByExactGoodsDescription(ctx context.Context, query ports.HistoricalTransactionClassificationQuery) (*ports.HistoricalTransactionClassificationMatch, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	row := querier.QueryRow(ctx, findHistoricalClassificationByExactGoodsDescriptionQuery, query.GoodsDescription, query.ResultVersion, query.ClassifierFamily, query.ClassifierVersion)

	var (
		transactionID    string
		goodsDescription string
		classification   string
		status           string
		pipelineResult   []byte
	)
	if err := row.Scan(&transactionID, &goodsDescription, &classification, &status, &pipelineResult); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("scanning historical exact transaction match: %w", err)
	}

	parsedTransactionID, err := valueobjects.TransactionIDFromString(transactionID)
	if err != nil {
		return nil, fmt.Errorf("parsing historical transaction id: %w", err)
	}

	parsedClassification, err := valueobjects.TransactionClassificationFromString(classification)
	if err != nil {
		return nil, fmt.Errorf("parsing historical transaction classification: %w", err)
	}

	parsedStatus, err := valueobjects.TransactionStatusFromString(status)
	if err != nil {
		return nil, fmt.Errorf("parsing historical transaction status: %w", err)
	}

	parsedPipelineResult, err := parsePipelineResult(pipelineResult)
	if err != nil {
		return nil, fmt.Errorf("parsing historical transaction pipeline result: %w", err)
	}
	if parsedPipelineResult == nil {
		return nil, nil
	}

	return &ports.HistoricalTransactionClassificationMatch{
		TransactionID:    parsedTransactionID,
		GoodsDescription: goodsDescription,
		Classification:   parsedClassification,
		Status:           parsedStatus,
		ReviewResult:     *parsedPipelineResult,
	}, nil
}

// GetNavigation returns one transaction and its immediate neighbors in the filtered scope.
func (r *PostgresTransactionRepository) GetNavigation(ctx context.Context, lookup ports.TransactionNavigationLookup) (*ports.TransactionNavigationResult, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	query, args := buildTransactionNavigationQuery(lookup)
	navigation, err := scanTransactionNavigation(querier.QueryRow(ctx, query, args...))
	if err != nil {
		return nil, fmt.Errorf("scanning transaction navigation: %w", err)
	}

	return navigation, nil
}

// List returns transaction rows matching the supplied filter.
func (r *PostgresTransactionRepository) List(ctx context.Context, filter ports.TransactionFilter) ([]*entities.Transaction, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	query, args := buildListTransactionsQuery(filter)
	rows, err := querier.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]*entities.Transaction, 0)
	for rows.Next() {
		transaction, scanErr := scanTransaction(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scanning transaction row: %w", scanErr)
		}

		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating transaction rows: %w", err)
	}

	return transactions, nil
}

// ListByUploadIDs lists transaction rows for the supplied uploads.
func (r *PostgresTransactionRepository) ListByUploadIDs(ctx context.Context, uploadIDs []valueobjects.UploadID) ([]*entities.Transaction, error) {
	if len(uploadIDs) == 0 {
		return nil, nil
	}

	querier := txQuerierFromContext(ctx, r.pool)
	query, args := buildListTransactionsByUploadIDsQuery(uploadIDs)
	rows, err := querier.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying transactions by upload ids: %w", err)
	}
	defer rows.Close()

	transactions := make([]*entities.Transaction, 0)
	for rows.Next() {
		transaction, scanErr := scanTransaction(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scanning transaction row: %w", scanErr)
		}

		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating transaction rows: %w", err)
	}

	return transactions, nil
}

// DeleteByID deletes an existing transaction row.
func (r *PostgresTransactionRepository) DeleteByID(ctx context.Context, id valueobjects.TransactionID) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, deleteTransactionByIDQuery, id.String())
	if err != nil {
		return fmt.Errorf("executing delete transaction by id query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// HasProcessingByUploadID reports whether an upload still has processing transactions.
func (r *PostgresTransactionRepository) HasProcessingByUploadID(ctx context.Context, uploadID valueobjects.UploadID) (bool, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	row := querier.QueryRow(ctx, hasProcessingTransactionsByUploadIDQuery, uploadID.String(), valueobjects.ProcessingTransactionStatus().String())

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("querying processing transactions by upload id: %w", err)
	}

	return exists, nil
}

// DeleteByUploadID deletes all rows associated with an upload.
func (r *PostgresTransactionRepository) DeleteByUploadID(ctx context.Context, uploadID valueobjects.UploadID) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(ctx, deleteTransactionsByUploadIDQuery, uploadID.String()); err != nil {
		return fmt.Errorf("executing delete transactions by upload id query: %w", err)
	}

	return nil
}

func scanTransaction(row scanner) (*entities.Transaction, error) {
	var (
		transactionID       string
		uploadID            sql.NullString
		rowNumber           sql.NullInt32
		product             string
		processedYear       int
		processedMonth      int
		dmcIB               string
		dmc                 string
		partnerBank         string
		referenceNumber     string
		transactionValue    string
		classification      string
		status              string
		pipelineResult      []byte
		failureReason       sql.NullString
		transactionCount    int
		goodsDescription    string
		goodsClassification string
		applicantCountry    string
		beneficiaryCountry  string
		sourceCountry       string
		destinationCountry  string
		tenorDescription    string
		esCategory          string
		paAlignment         string
		createdBy           string
		createdAt           time.Time
		updatedAt           time.Time
	)

	if err := row.Scan(
		&transactionID,
		&uploadID,
		&rowNumber,
		&product,
		&processedYear,
		&processedMonth,
		&dmcIB,
		&dmc,
		&partnerBank,
		&referenceNumber,
		&transactionValue,
		&classification,
		&status,
		&pipelineResult,
		&failureReason,
		&transactionCount,
		&goodsDescription,
		&goodsClassification,
		&applicantCountry,
		&beneficiaryCountry,
		&sourceCountry,
		&destinationCountry,
		&tenorDescription,
		&esCategory,
		&paAlignment,
		&createdBy,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	parsedTransactionID, err := valueobjects.TransactionIDFromString(transactionID)
	if err != nil {
		return nil, fmt.Errorf("parsing transaction id: %w", err)
	}

	var parsedUploadID *valueobjects.UploadID
	if uploadID.Valid {
		value, parseErr := valueobjects.UploadIDFromString(uploadID.String)
		if parseErr != nil {
			return nil, fmt.Errorf("parsing upload id: %w", parseErr)
		}
		parsedUploadID = &value
	}

	var parsedRowNumber *int
	if rowNumber.Valid {
		value := int(rowNumber.Int32)
		parsedRowNumber = &value
	}

	parsedClassification, err := valueobjects.TransactionClassificationFromString(classification)
	if err != nil {
		return nil, fmt.Errorf("parsing transaction classification: %w", err)
	}

	parsedStatus, err := valueobjects.TransactionStatusFromString(status)
	if err != nil {
		return nil, fmt.Errorf("parsing transaction status: %w", err)
	}

	parsedPipelineResult, err := parsePipelineResult(pipelineResult)
	if err != nil {
		return nil, fmt.Errorf("parsing pipeline result: %w", err)
	}

	parsedCreatedBy, err := valueobjects.UserIDFromString(createdBy)
	if err != nil {
		return nil, fmt.Errorf("parsing transaction created by: %w", err)
	}

	return entities.ReconstituteTransaction(
		parsedTransactionID,
		parsedUploadID,
		parsedRowNumber,
		product,
		processedYear,
		processedMonth,
		dmcIB,
		dmc,
		partnerBank,
		referenceNumber,
		transactionValue,
		parsedClassification,
		parsedStatus,
		parsedPipelineResult,
		failureReason.String,
		transactionCount,
		goodsDescription,
		goodsClassification,
		applicantCountry,
		beneficiaryCountry,
		sourceCountry,
		destinationCountry,
		tenorDescription,
		esCategory,
		paAlignment,
		parsedCreatedBy,
		createdAt,
		updatedAt,
	), nil
}

func createTransactionExecArgs(transaction *entities.Transaction, createdByUserID string) []any {
	var uploadID any
	if transaction.UploadID() != nil {
		uploadID = transaction.UploadID().String()
	}

	var rowNumber any
	if transaction.RowNumber() != nil {
		rowNumber = *transaction.RowNumber()
	}

	pipelineResult := mustMarshalPipelineResult(transaction.PipelineResult())
	var pipelineResultArg any
	if len(pipelineResult) > 0 {
		pipelineResultArg = pipelineResult
	}

	transactionValue := normalizeTransactionValueForDB(transaction.TransactionValue())

	return []any{
		transaction.ID().String(),
		uploadID,
		rowNumber,
		transaction.Product(),
		transaction.ProcessedYear(),
		transaction.ProcessedMonth(),
		transaction.DMCIB(),
		transaction.DMC(),
		transaction.PartnerBank(),
		transaction.ReferenceNumber(),
		transactionValue,
		transaction.Classification(),
		transaction.Status(),
		pipelineResultArg,
		transaction.FailureReason(),
		transaction.TransactionCount(),
		transaction.GoodsDescription(),
		transaction.GoodsClassification(),
		transaction.ApplicantCountry(),
		transaction.BeneficiaryCountry(),
		transaction.SourceCountry(),
		transaction.DestinationCountry(),
		transaction.TenorDescription(),
		transaction.ESCategory(),
		transaction.PAAlignment(),
		createdByUserID,
		transaction.CreatedAt(),
		transaction.UpdatedAt(),
	}
}

func updateTransactionExecArgs(transaction *entities.Transaction) []any {
	var uploadID any
	if transaction.UploadID() != nil {
		uploadID = transaction.UploadID().String()
	}

	var rowNumber any
	if transaction.RowNumber() != nil {
		rowNumber = *transaction.RowNumber()
	}

	pipelineResult := mustMarshalPipelineResult(transaction.PipelineResult())
	var pipelineResultArg any
	if len(pipelineResult) > 0 {
		pipelineResultArg = pipelineResult
	}

	transactionValue := normalizeTransactionValueForDB(transaction.TransactionValue())

	return []any{
		transaction.ID().String(),
		uploadID,
		rowNumber,
		transaction.Product(),
		transaction.ProcessedYear(),
		transaction.ProcessedMonth(),
		transaction.DMCIB(),
		transaction.DMC(),
		transaction.PartnerBank(),
		transaction.ReferenceNumber(),
		transactionValue,
		transaction.Classification(),
		transaction.Status(),
		pipelineResultArg,
		transaction.FailureReason(),
		transaction.TransactionCount(),
		transaction.GoodsDescription(),
		transaction.GoodsClassification(),
		transaction.ApplicantCountry(),
		transaction.BeneficiaryCountry(),
		transaction.SourceCountry(),
		transaction.DestinationCountry(),
		transaction.TenorDescription(),
		transaction.ESCategory(),
		transaction.PAAlignment(),
		transaction.CreatedAt(),
		transaction.UpdatedAt(),
	}
}

func normalizeTransactionValueForDB(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), ",", "")
}

func buildTransactionNavigationQuery(lookup ports.TransactionNavigationLookup) (string, []any) {
	args := []any{lookup.TransactionID.String()}
	joins := make([]string, 0, 2)
	conditions := make([]string, 0, 14)

	appendCondition := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}

	filter := lookup.Filter
	if filter.UploadID != nil {
		appendCondition("t.upload_id = $%d", filter.UploadID.String())
	}
	if filter.CreatedAtFrom != nil {
		appendCondition("t.created_at >= $%d", *filter.CreatedAtFrom)
	}
	if filter.CreatedAtTo != nil {
		appendCondition("t.created_at < $%d", filter.CreatedAtTo.AddDate(0, 0, 1))
	}
	if filter.ApplicantCountry != nil {
		appendCondition("t.applicant_country = $%d", *filter.ApplicantCountry)
	}
	if filter.BeneficiaryCountry != nil {
		appendCondition("t.beneficiary_country = $%d", *filter.BeneficiaryCountry)
	}
	if filter.SourceCountry != nil {
		appendCondition("t.source_country = $%d", *filter.SourceCountry)
	}
	if filter.DestinationCountry != nil {
		appendCondition("t.destination_country = $%d", *filter.DestinationCountry)
	}
	if filter.TransactionCountMin != nil {
		appendCondition("t.transaction_count >= $%d", *filter.TransactionCountMin)
	}
	if filter.TransactionCountMax != nil {
		appendCondition("t.transaction_count <= $%d", *filter.TransactionCountMax)
	}

	classification := lookup.Classification
	if classification == nil {
		classification = filter.Classification
	}
	if classification != nil {
		appendCondition("t.classification = $%d", *classification)
	}
	if filter.Status != nil {
		appendCondition("t.status = $%d", *filter.Status)
	}

	if lookup.Step != nil {
		switch *lookup.Step {
		case 1, 2, 3:
			joins = append(joins,
				"LEFT JOIN transaction_step_4 ts4 ON ts4.transaction_id = t.id",
				"LEFT JOIN transaction_step_5_data ts5 ON ts5.transaction_id = t.id",
			)
			appendCondition("t.pipeline_result->>'exit_step' = $%d", strconv.Itoa(*lookup.Step))
			conditions = append(conditions, "ts4.transaction_id IS NULL", "ts5.transaction_id IS NULL")
		case 4:
			joins = append(joins,
				"JOIN transaction_step_4 ts4 ON ts4.transaction_id = t.id",
				"LEFT JOIN transaction_step_5_data ts5 ON ts5.transaction_id = t.id",
			)
			conditions = append(conditions, "ts5.transaction_id IS NULL")
		case 5:
			joins = append(joins, "JOIN transaction_step_5_data ts5 ON ts5.transaction_id = t.id")
		}
	}

	query := "WITH filtered_transactions AS (\n" +
		"    SELECT t.id\n" +
		"    FROM transactions t\n"
	if len(joins) > 0 {
		query += "    " + strings.Join(joins, "\n    ") + "\n"
	}
	if len(conditions) > 0 {
		query += "    WHERE " + strings.Join(conditions, " AND ") + "\n"
	}
	query += "),\n" +
		"previous_transaction AS (\n" +
		"    SELECT id\n" +
		"    FROM filtered_transactions\n" +
		"    WHERE id < $1\n" +
		"    ORDER BY id DESC\n" +
		"    LIMIT 1\n" +
		"),\n" +
		"next_transaction AS (\n" +
		"    SELECT id\n" +
		"    FROM filtered_transactions\n" +
		"    WHERE id > $1\n" +
		"    ORDER BY id ASC\n" +
		"    LIMIT 1\n" +
		")\n" +
		"SELECT current_transaction.id AS transaction_id,\n" +
		"       previous_transaction.id AS previous_id,\n" +
		"       next_transaction.id AS next_id\n" +
		"FROM filtered_transactions current_transaction\n" +
		"LEFT JOIN previous_transaction ON TRUE\n" +
		"LEFT JOIN next_transaction ON TRUE\n" +
		"WHERE current_transaction.id = $1\n"

	return query, args
}

func buildListTransactionsByUploadIDsQuery(uploadIDs []valueobjects.UploadID) (string, []any) {
	placeholders := make([]string, 0, len(uploadIDs))
	args := make([]any, 0, len(uploadIDs))

	for index, uploadID := range uploadIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
		args = append(args, uploadID.String())
	}

	query := baseListTransactionsByUploadIDsQuery
	query += "WHERE upload_id IN (" + strings.Join(placeholders, ", ") + ")\n"
	query += "ORDER BY upload_id, row_number ASC, id ASC"

	return query, args
}

func scanTransactionNavigation(row scanner) (*ports.TransactionNavigationResult, error) {
	var (
		transactionIDValue string
		previousIDValue    sql.NullString
		nextIDValue        sql.NullString
	)

	if err := row.Scan(&transactionIDValue, &previousIDValue, &nextIDValue); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	var previousID *string
	if previousIDValue.Valid {
		previousID = &previousIDValue.String
	}

	var nextID *string
	if nextIDValue.Valid {
		nextID = &nextIDValue.String
	}

	return &ports.TransactionNavigationResult{
		TransactionID: transactionIDValue,
		PreviousID:    previousID,
		NextID:        nextID,
	}, nil
}

func buildListTransactionsQuery(filter ports.TransactionFilter) (string, []any) {
	query := baseListTransactionsQuery
	args := make([]any, 0, 11)
	conditions := make([]string, 0, 11)

	appendCondition := func(condition string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(condition, len(args)))
	}

	if filter.UploadID != nil {
		appendCondition("upload_id = $%d", filter.UploadID.String())
	}

	if filter.CreatedAtFrom != nil {
		appendCondition("created_at >= $%d", *filter.CreatedAtFrom)
	}

	if filter.CreatedAtTo != nil {
		appendCondition("created_at < $%d", filter.CreatedAtTo.AddDate(0, 0, 1))
	}

	if filter.ApplicantCountry != nil {
		appendCondition("applicant_country = $%d", *filter.ApplicantCountry)
	}

	if filter.BeneficiaryCountry != nil {
		appendCondition("beneficiary_country = $%d", *filter.BeneficiaryCountry)
	}

	if filter.SourceCountry != nil {
		appendCondition("source_country = $%d", *filter.SourceCountry)
	}

	if filter.DestinationCountry != nil {
		appendCondition("destination_country = $%d", *filter.DestinationCountry)
	}

	if filter.TransactionCountMin != nil {
		appendCondition("transaction_count >= $%d", *filter.TransactionCountMin)
	}

	if filter.TransactionCountMax != nil {
		appendCondition("transaction_count <= $%d", *filter.TransactionCountMax)
	}

	if filter.Classification != nil {
		appendCondition("classification = $%d", *filter.Classification)
	}

	if filter.Status != nil {
		appendCondition("status = $%d", *filter.Status)
	}

	if len(conditions) > 0 {
		query += "WHERE " + strings.Join(conditions, " AND ") + "\n"
	}

	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = ports.TransactionSortByCreatedAt
	}

	sortOrder := strings.ToUpper(filter.SortOrder)
	if sortOrder == "" {
		sortOrder = strings.ToUpper(ports.TransactionSortOrderDescending)
	}

	allowedSortBy := []string{
		ports.TransactionSortByCreatedAt,
		ports.TransactionSortByApplicantCountry,
		ports.TransactionSortByBeneficiaryCountry,
		ports.TransactionSortBySourceCountry,
		ports.TransactionSortByDestinationCountry,
		ports.TransactionSortByTransactionCount,
		ports.TransactionSortByClassification,
		ports.TransactionSortByStatus,
	}
	if !slices.Contains(allowedSortBy, sortBy) {
		sortBy = ports.TransactionSortByCreatedAt
	}

	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	if sortBy == ports.TransactionSortByCreatedAt {
		query += fmt.Sprintf("ORDER BY created_at %s, upload_id ASC NULLS LAST, row_number ASC NULLS LAST, id %s", sortOrder, sortOrder)
		return query, args
	}

	query += fmt.Sprintf("ORDER BY %s %s, id %s", sortBy, sortOrder, sortOrder)
	return query, args
}

type pipelineResultRecord struct {
	Version                       string            `json:"version,omitempty"`
	Source                        string            `json:"source,omitempty"`
	ClassifierFamily              string            `json:"classifier_family,omitempty"`
	ClassifierVersion             string            `json:"classifier_version,omitempty"`
	PromptVersion                 string            `json:"prompt_version,omitempty"`
	Model                         string            `json:"model,omitempty"`
	MatchedTransactionID          *string           `json:"matched_transaction_id,omitempty"`
	MatchedGoodsDescription       *string           `json:"matched_goods_description,omitempty"`
	BatchID                       *string           `json:"batch_id,omitempty"`
	BatchSize                     *int              `json:"batch_size,omitempty"`
	NotAlignedListMatch           bool              `json:"not_aligned_list_match,omitempty"`
	NotAlignedListMatchConfidence int               `json:"not_aligned_list_match_confidence,omitempty"`
	AlignedListMatch              bool              `json:"aligned_list_match,omitempty"`
	AlignedListMatchConfidence    int               `json:"aligned_list_match_confidence,omitempty"`
	ExitStep                      int               `json:"exit_step,omitempty"`
	OverallClassification         string            `json:"overall_classification,omitempty"`
	Reason                        string            `json:"reason,omitempty"`
	TransactionID                 string            `json:"transaction_id"`
	Step1Result                   *stepResultRecord `json:"step1_result,omitempty"`
	Step2Result                   *stepResultRecord `json:"step2_result,omitempty"`
	Step3Result                   *stepResultRecord `json:"step3_result,omitempty"`
	FinalClassification           string            `json:"final_classification,omitempty"`
}

type stepResultRecord struct {
	StepNumber    int    `json:"step_number"`
	StepAlignment string `json:"step_alignment"`
	BooleanResult *bool  `json:"boolean_result,omitempty"`
	Reason        string `json:"reason,omitempty"`
}

func marshalPipelineResult(result *valueobjects.PipelineResult) ([]byte, error) {
	if result == nil {
		return nil, nil
	}

	if react := result.React(); react != nil {
		record := pipelineResultRecord{
			Version:                       react.Version(),
			Source:                        react.Source(),
			ClassifierFamily:              react.ClassifierFamily(),
			ClassifierVersion:             react.ClassifierVersion(),
			PromptVersion:                 react.PromptVersion(),
			Model:                         react.Model(),
			TransactionID:                 react.TransactionID(),
			MatchedTransactionID:          react.MatchedTransactionID(),
			MatchedGoodsDescription:       react.MatchedGoodsDescription(),
			BatchID:                       react.BatchID(),
			BatchSize:                     react.BatchSize(),
			NotAlignedListMatch:           react.NotAlignedListMatch(),
			NotAlignedListMatchConfidence: react.NotAlignedListMatchConfidence(),
			AlignedListMatch:              react.AlignedListMatch(),
			AlignedListMatchConfidence:    react.AlignedListMatchConfidence(),
			ExitStep:                      react.ExitStep(),
			OverallClassification:         react.OverallClassification().String(),
			Reason:                        react.Reason(),
		}
		encoded, err := json.Marshal(record)
		if err != nil {
			return nil, fmt.Errorf("marshalling react pipeline result: %w", err)
		}

		return encoded, nil
	}

	record := pipelineResultRecord{
		Version:             valueobjects.PipelineResultVersionLegacy,
		TransactionID:       result.TransactionID(),
		Step1Result:         stepResultRecordPointer(newLegacyStepResultRecord(result.Step1Result())),
		ExitStep:            result.ExitStep(),
		FinalClassification: result.FinalClassification().String(),
	}

	if step2Result := result.Step2Result(); step2Result != nil {
		step2Record := newLegacyStepResultRecord(*step2Result)
		record.Step2Result = &step2Record
	}

	if step3Result := result.Step3Result(); step3Result != nil {
		step3Record := newLegacyStepResultRecord(*step3Result)
		record.Step3Result = &step3Record
	}

	encoded, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshalling pipeline result: %w", err)
	}

	return encoded, nil
}

func mustMarshalPipelineResult(result *valueobjects.PipelineResult) []byte {
	encoded, err := marshalPipelineResult(result)
	if err != nil {
		panic(fmt.Errorf("marshalling pipeline result: %w", err))
	}

	return encoded
}

func parsePipelineResult(encoded []byte) (*valueobjects.PipelineResult, error) {
	if len(encoded) == 0 {
		return nil, nil
	}

	var record pipelineResultRecord
	if err := json.Unmarshal(encoded, &record); err != nil {
		return nil, fmt.Errorf("unmarshalling pipeline result: %w", err)
	}

	if record.Version == valueobjects.PipelineResultVersionReactV1 {
		classification, err := valueobjects.TransactionClassificationFromString(record.OverallClassification)
		if err != nil {
			return nil, fmt.Errorf("parsing react overall classification: %w", err)
		}

		classifierVersion := record.ClassifierVersion
		if strings.TrimSpace(classifierVersion) == "" {
			classifierVersion = record.PromptVersion
		}

		reviewResult := valueobjects.NewReactReviewResult(
			record.Source,
			record.ClassifierFamily,
			classifierVersion,
			record.PromptVersion,
			record.Model,
			record.TransactionID,
			record.MatchedTransactionID,
			record.MatchedGoodsDescription,
			record.BatchID,
			record.BatchSize,
			record.NotAlignedListMatch,
			record.NotAlignedListMatchConfidence,
			record.AlignedListMatch,
			record.AlignedListMatchConfidence,
			record.ExitStep,
			classification,
			record.Reason,
		)
		result := valueobjects.NewReactPipelineResult(reviewResult)
		return &result, nil
	}

	if record.Step1Result == nil {
		return nil, fmt.Errorf("step 1 result is required")
	}

	step1Result, err := record.Step1Result.toValue()
	if err != nil {
		return nil, fmt.Errorf("parsing step 1 result: %w", err)
	}

	var step2Result *valueobjects.StepResult
	if record.Step2Result != nil {
		parsedStep2Result, err := record.Step2Result.toValue()
		if err != nil {
			return nil, fmt.Errorf("parsing step 2 result: %w", err)
		}

		step2Result = &parsedStep2Result
	}

	var step3Result *valueobjects.StepResult
	if record.Step3Result != nil {
		parsedStep3Result, err := record.Step3Result.toValue()
		if err != nil {
			return nil, fmt.Errorf("parsing step 3 result: %w", err)
		}

		step3Result = &parsedStep3Result
	}

	finalClassification, err := valueobjects.TransactionClassificationFromString(record.FinalClassification)
	if err != nil {
		return nil, fmt.Errorf("parsing final classification: %w", err)
	}

	result := valueobjects.NewPipelineResult(
		record.TransactionID,
		step1Result,
		step2Result,
		step3Result,
		record.ExitStep,
		finalClassification,
	)

	return &result, nil
}

func stepResultRecordPointer(record stepResultRecord) *stepResultRecord {
	recordCopy := record
	return &recordCopy
}

func (r stepResultRecord) toValue() (valueobjects.StepResult, error) {
	if r.BooleanResult != nil {
		var reason error
		if strings.TrimSpace(r.Reason) != "" {
			reason = errors.New(strings.TrimSpace(r.Reason))
		}

		return valueobjects.NewBooleanStepResultWithReason(r.StepNumber, *r.BooleanResult, reason), nil
	}

	alignment, err := valueobjects.AlignmentFromString(r.StepAlignment)
	if err != nil {
		return valueobjects.StepResult{}, fmt.Errorf("parsing step alignment: %w", err)
	}

	return newLegacySummaryStepResult(r.StepNumber, alignment), nil
}

func newLegacyStepResultRecord(result valueobjects.StepResult) stepResultRecord {
	record := stepResultRecord{
		StepNumber:    result.StepNumber(),
		StepAlignment: result.StepAlignment().String(),
	}

	if booleanResult := result.BooleanResult(); booleanResult != nil {
		booleanResultCopy := *booleanResult
		record.BooleanResult = &booleanResultCopy
	}

	if reason := result.Reason(); reason != nil {
		record.Reason = reason.Error()
	}

	return record
}

func newLegacySummaryStepResult(stepNumber int, alignment valueobjects.Alignment) valueobjects.StepResult {
	alignedDecision := valueobjects.NewMatchDecision(alignment, 0, "", nil)
	semanticAlignment := valueobjects.UnalignedAlignment()
	if alignment.Equal(valueobjects.AlignedAlignment()) {
		semanticAlignment = valueobjects.AlignedAlignment()
	}
	semanticDecision := valueobjects.NewMatchDecision(semanticAlignment, 0, "", nil)

	return valueobjects.NewStepResult(stepNumber, alignedDecision, semanticDecision)
}
