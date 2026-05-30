package adapters

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/pashagolub/pgxmock/v4"
)

// TestPostgresTransactionUploadRepositoryCreateAndQueries verifies the postgres transaction upload repository create and queries behavior and the expected outcome asserted below.
func TestPostgresTransactionUploadRepositoryCreateAndQueries(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	upload, err := entities.NewTransactionUpload(uploadID, groupID, "transactions.csv", "csv", "d41d8cd98f00b204e9800998ecf8427e", "local", "upload/file.csv", "transaction-file-v1", valueobjects.UploadedTransactionUploadStatus(), 1, now)
	if err != nil {
		t.Fatalf("NewTransactionUpload() error = %v", err)
	}

	repository := NewPostgresTransactionUploadRepository(mock)
	mock.ExpectExec(regexp.QuoteMeta(createTransactionUploadQuery)).
		WithArgs(upload.ID().String(), upload.GroupID().String(), upload.FileName(), upload.FileFormat(), upload.ContentMD5(), upload.StorageProvider(), upload.StorageKey(), upload.SchemaVersion(), upload.Status(), upload.RowCount(), upload.UploadedAt()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Create(context.Background(), upload); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	rows := pgxmock.NewRows([]string{"id", "group_id", "file_name", "file_format", "content_md5", "storage_provider", "storage_key", "schema_version", "status", "row_count", "uploaded_at"}).AddRow(upload.ID().String(), upload.GroupID().String(), upload.FileName(), upload.FileFormat(), upload.ContentMD5(), upload.StorageProvider(), upload.StorageKey(), upload.SchemaVersion(), upload.Status(), upload.RowCount(), upload.UploadedAt())
	mock.ExpectQuery(regexp.QuoteMeta(findTransactionUploadByContentMD5Query)).WithArgs(upload.ContentMD5(), upload.GroupID().String()).WillReturnRows(rows)

	loadedByMD5, err := repository.FindByContentMD5(context.Background(), upload.ContentMD5(), groupID)
	if err != nil {
		t.Fatalf("FindByContentMD5() error = %v", err)
	}

	if loadedByMD5 == nil {
		t.Fatal("expected upload loaded by content md5 and group id")
	}

	if loadedByMD5.GroupID().String() != upload.GroupID().String() {
		t.Fatalf("loadedByMD5.GroupID() = %q, want %q", loadedByMD5.GroupID().String(), upload.GroupID().String())
	}

	rows = pgxmock.NewRows([]string{"id", "group_id", "file_name", "file_format", "content_md5", "storage_provider", "storage_key", "schema_version", "status", "row_count", "uploaded_at"}).AddRow(upload.ID().String(), upload.GroupID().String(), upload.FileName(), upload.FileFormat(), upload.ContentMD5(), upload.StorageProvider(), upload.StorageKey(), upload.SchemaVersion(), upload.Status(), upload.RowCount(), upload.UploadedAt())
	mock.ExpectQuery(regexp.QuoteMeta(findTransactionUploadByIDQuery)).WithArgs(upload.ID().String()).WillReturnRows(rows)

	loaded, err := repository.FindByID(context.Background(), uploadID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if loaded == nil {
		t.Fatal("expected loaded upload")
	}

	if loaded.GroupID().String() != upload.GroupID().String() {
		t.Fatalf("loaded.GroupID() = %q, want %q", loaded.GroupID().String(), upload.GroupID().String())
	}

	listQuery, args := buildListTransactionUploadsQuery(ports.TransactionUploadFilter{GroupID: groupID, FileName: "transactions"})
	rows = pgxmock.NewRows([]string{"id", "group_id", "file_name", "file_format", "content_md5", "storage_provider", "storage_key", "schema_version", "status", "row_count", "uploaded_at"}).AddRow(upload.ID().String(), upload.GroupID().String(), upload.FileName(), upload.FileFormat(), upload.ContentMD5(), upload.StorageProvider(), upload.StorageKey(), upload.SchemaVersion(), upload.Status(), upload.RowCount(), upload.UploadedAt())
	mock.ExpectQuery(regexp.QuoteMeta(listQuery)).WithArgs(args...).WillReturnRows(rows)

	uploads, err := repository.List(context.Background(), ports.TransactionUploadFilter{GroupID: groupID, FileName: "transactions"})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(uploads) != 1 {
		t.Fatalf("len(uploads) = %d, want %d", len(uploads), 1)
	}

	if uploads[0].GroupID().String() != upload.GroupID().String() {
		t.Fatalf("uploads[0].GroupID() = %q, want %q", uploads[0].GroupID().String(), upload.GroupID().String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ExpectationsWereMet() error = %v", err)
	}
}

// TestPostgresTransactionRepositoryCreateListCRUDAndDeleteByUpload verifies the postgres transaction repository create list crud and delete by upload behavior and the expected outcome asserted below.
func TestPostgresTransactionRepositoryCreateListCRUDAndDeleteByUpload(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := entities.NewUploadedTransaction(transactionID, uploadID, 2, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", now)
	if err != nil {
		t.Fatalf("NewUploadedTransaction() error = %v", err)
	}

	repository := NewPostgresTransactionRepository(mock)
	mock.ExpectExec(regexp.QuoteMeta(createTransactionQuery)).
		WithArgs(transaction.ID().String(), transaction.UploadID().String(), *transaction.RowNumber(), transaction.Product(), transaction.ProcessedYear(), transaction.ProcessedMonth(), transaction.DMCIB(), transaction.DMC(), transaction.PartnerBank(), transaction.ReferenceNumber(), "698436.80", transaction.Classification(), transaction.Status(), nil, transaction.FailureReason(), transaction.TransactionCount(), transaction.GoodsDescription(), transaction.GoodsClassification(), transaction.ApplicantCountry(), transaction.BeneficiaryCountry(), transaction.SourceCountry(), transaction.DestinationCountry(), transaction.TenorDescription(), transaction.ESCategory(), transaction.PAAlignment(), "01962b8f-aeb2-7e03-a8ff-1edce1300002", transaction.CreatedAt(), transaction.UpdatedAt()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Create(context.Background(), transaction, "01962b8f-aeb2-7e03-a8ff-1edce1300002"); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(findTransactionByIDQuery)).WithArgs(transaction.ID().String()).WillReturnRows(
		pgxmock.NewRows([]string{"id", "upload_id", "row_number", "product", "processed_year", "processed_month", "dmc_ib", "dmc", "partner_bank", "ref_num", "transaction_value", "classification", "status", "pipeline_result", "failure_reason", "transaction_count", "goods_description", "goods_classification", "applicant_country", "beneficiary_country", "source_country", "destination_country", "tenor_description", "es_category", "pa_alignment", "created_by", "created_at", "updated_at"}).AddRow(transaction.ID().String(), transaction.UploadID().String(), *transaction.RowNumber(), transaction.Product(), transaction.ProcessedYear(), transaction.ProcessedMonth(), transaction.DMCIB(), transaction.DMC(), transaction.PartnerBank(), transaction.ReferenceNumber(), transaction.TransactionValue(), transaction.Classification(), transaction.Status(), nil, nil, transaction.TransactionCount(), transaction.GoodsDescription(), transaction.GoodsClassification(), transaction.ApplicantCountry(), transaction.BeneficiaryCountry(), transaction.SourceCountry(), transaction.DestinationCountry(), transaction.TenorDescription(), transaction.ESCategory(), transaction.PAAlignment(), "01962b8f-aeb2-7e03-a8ff-1edce1300002", transaction.CreatedAt(), transaction.UpdatedAt()),
	)

	loaded, err := repository.FindByID(context.Background(), transactionID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if loaded == nil || loaded.ID().String() != transactionID.String() {
		t.Fatalf("loaded.ID() = %v, want %q", loaded, transactionID.String())
	}

	if loaded.CreatedBy() != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
		t.Fatalf("loaded.CreatedBy() = %q, want %q", loaded.CreatedBy(), "01962b8f-aeb2-7e03-a8ff-1edce1300002")
	}

	listAllQuery, listAllArgs := buildListTransactionsQuery(ports.TransactionFilter{})
	rows := pgxmock.NewRows([]string{"id", "upload_id", "row_number", "product", "processed_year", "processed_month", "dmc_ib", "dmc", "partner_bank", "ref_num", "transaction_value", "classification", "status", "pipeline_result", "failure_reason", "transaction_count", "goods_description", "goods_classification", "applicant_country", "beneficiary_country", "source_country", "destination_country", "tenor_description", "es_category", "pa_alignment", "created_by", "created_at", "updated_at"}).AddRow(transaction.ID().String(), transaction.UploadID().String(), *transaction.RowNumber(), transaction.Product(), transaction.ProcessedYear(), transaction.ProcessedMonth(), transaction.DMCIB(), transaction.DMC(), transaction.PartnerBank(), transaction.ReferenceNumber(), transaction.TransactionValue(), transaction.Classification(), transaction.Status(), nil, nil, transaction.TransactionCount(), transaction.GoodsDescription(), transaction.GoodsClassification(), transaction.ApplicantCountry(), transaction.BeneficiaryCountry(), transaction.SourceCountry(), transaction.DestinationCountry(), transaction.TenorDescription(), transaction.ESCategory(), transaction.PAAlignment(), "01962b8f-aeb2-7e03-a8ff-1edce1300002", transaction.CreatedAt(), transaction.UpdatedAt())
	mock.ExpectQuery(regexp.QuoteMeta(listAllQuery)).WithArgs(listAllArgs...).WillReturnRows(rows)

	listedTransactions, err := repository.List(context.Background(), ports.TransactionFilter{})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(listedTransactions) != 1 {
		t.Fatalf("len(listedTransactions) = %d, want %d", len(listedTransactions), 1)
	}

	listQuery, listArgs := buildListTransactionsByUploadIDsQuery([]valueobjects.UploadID{uploadID})
	rows = pgxmock.NewRows([]string{"id", "upload_id", "row_number", "product", "processed_year", "processed_month", "dmc_ib", "dmc", "partner_bank", "ref_num", "transaction_value", "classification", "status", "pipeline_result", "failure_reason", "transaction_count", "goods_description", "goods_classification", "applicant_country", "beneficiary_country", "source_country", "destination_country", "tenor_description", "es_category", "pa_alignment", "created_by", "created_at", "updated_at"}).AddRow(transaction.ID().String(), transaction.UploadID().String(), *transaction.RowNumber(), transaction.Product(), transaction.ProcessedYear(), transaction.ProcessedMonth(), transaction.DMCIB(), transaction.DMC(), transaction.PartnerBank(), transaction.ReferenceNumber(), transaction.TransactionValue(), transaction.Classification(), transaction.Status(), nil, nil, transaction.TransactionCount(), transaction.GoodsDescription(), transaction.GoodsClassification(), transaction.ApplicantCountry(), transaction.BeneficiaryCountry(), transaction.SourceCountry(), transaction.DestinationCountry(), transaction.TenorDescription(), transaction.ESCategory(), transaction.PAAlignment(), "01962b8f-aeb2-7e03-a8ff-1edce1300002", transaction.CreatedAt(), transaction.UpdatedAt())
	mock.ExpectQuery(regexp.QuoteMeta(listQuery)).WithArgs(listArgs...).WillReturnRows(rows)

	transactions, err := repository.ListByUploadIDs(context.Background(), []valueobjects.UploadID{uploadID})
	if err != nil {
		t.Fatalf("ListByUploadIDs() error = %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("len(transactions) = %d, want %d", len(transactions), 1)
	}

	mock.ExpectExec(regexp.QuoteMeta(updateTransactionQuery)).
		WithArgs(transaction.ID().String(), transaction.UploadID().String(), *transaction.RowNumber(), transaction.Product(), transaction.ProcessedYear(), transaction.ProcessedMonth(), transaction.DMCIB(), transaction.DMC(), transaction.PartnerBank(), transaction.ReferenceNumber(), "698436.80", transaction.Classification(), transaction.Status(), nil, transaction.FailureReason(), transaction.TransactionCount(), transaction.GoodsDescription(), transaction.GoodsClassification(), transaction.ApplicantCountry(), transaction.BeneficiaryCountry(), transaction.SourceCountry(), transaction.DestinationCountry(), transaction.TenorDescription(), transaction.ESCategory(), transaction.PAAlignment(), transaction.CreatedAt(), transaction.UpdatedAt()).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repository.Update(context.Background(), transaction); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	queue := NewPostgresTransactionProcessingQueue(mock)
	mock.ExpectExec(regexp.QuoteMeta(createTransactionProcessingQueueEntryQuery)).
		WithArgs("transaction:classify", transaction.ID().String()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	if err := queue.Enqueue(context.Background(), "transaction:classify", transaction.ID()); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(deleteTransactionByIDQuery)).WithArgs(transactionID.String()).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := repository.DeleteByID(context.Background(), transactionID); err != nil {
		t.Fatalf("DeleteByID() error = %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(deleteTransactionsByUploadIDQuery)).WithArgs(uploadID.String()).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := repository.DeleteByUploadID(context.Background(), uploadID); err != nil {
		t.Fatalf("DeleteByUploadID() error = %v", err)
	}
}

// TestPostgresTransactionRepositoryHasProcessingByUploadID verifies the processing-upload existence query behavior and the expected outcome asserted below.
func TestPostgresTransactionRepositoryHasProcessingByUploadID(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	repository := NewPostgresTransactionRepository(mock)
	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	rows := pgxmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery(regexp.QuoteMeta(hasProcessingTransactionsByUploadIDQuery)).
		WithArgs(uploadID.String(), valueobjects.ProcessingTransactionStatus().String()).
		WillReturnRows(rows)

	hasProcessing, err := repository.HasProcessingByUploadID(context.Background(), uploadID)
	if err != nil {
		t.Fatalf("HasProcessingByUploadID() error = %v", err)
	}

	if !hasProcessing {
		t.Fatal("HasProcessingByUploadID() = false, want true")
	}
}

// TestBuildListTransactionsByUploadIDsQuery verifies the build list transactions by upload i ds query behavior and the expected outcome asserted below.
func TestBuildListTransactionsByUploadIDsQuery(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	query, args := buildListTransactionsByUploadIDsQuery([]valueobjects.UploadID{uploadID})
	if query == "" {
		t.Fatal("expected query")
	}

	if len(args) != 1 {
		t.Fatalf("len(args) = %d, want %d", len(args), 1)
	}

	createdAtFrom := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	createdAtTo := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)
	applicantCountry := "Philippines"
	beneficiaryCountry := "Japan"
	sourceCountry := "Thailand"
	destinationCountry := "Philippines"
	transactionCountMin := 1
	transactionCountMax := 10
	classification := "unclassified"
	status := "processing"

	filteredQuery, filteredArgs := buildListTransactionsQuery(ports.TransactionFilter{
		UploadID:            &uploadID,
		CreatedAtFrom:       &createdAtFrom,
		CreatedAtTo:         &createdAtTo,
		ApplicantCountry:    &applicantCountry,
		BeneficiaryCountry:  &beneficiaryCountry,
		SourceCountry:       &sourceCountry,
		DestinationCountry:  &destinationCountry,
		TransactionCountMin: &transactionCountMin,
		TransactionCountMax: &transactionCountMax,
		Classification:      &classification,
		Status:              &status,
		SortBy:              ports.TransactionSortByTransactionCount,
		SortOrder:           ports.TransactionSortOrderAscending,
	})
	if filteredQuery == "" {
		t.Fatal("expected filtered query")
	}

	if len(filteredArgs) != 11 {
		t.Fatalf("len(filteredArgs) = %d, want %d", len(filteredArgs), 11)
	}
}

// TestBuildListTransactionsQueryUsesUploadAwareTieBreakersForCreatedAt verifies the default created_at ordering keeps upload row order deterministic.
func TestBuildListTransactionsQueryUsesUploadAwareTieBreakersForCreatedAt(t *testing.T) {
	t.Parallel()

	query, args := buildListTransactionsQuery(ports.TransactionFilter{})
	if len(args) != 0 {
		t.Fatalf("len(args) = %d, want %d", len(args), 0)
	}

	wantOrderBy := "ORDER BY created_at DESC, upload_id ASC NULLS LAST, row_number ASC NULLS LAST, id DESC"
	if !strings.Contains(query, wantOrderBy) {
		t.Fatalf("buildListTransactionsQuery() query = %q, want it to contain %q", query, wantOrderBy)
	}
}

// TestBuildListTransactionsQueryUsesUploadAwareTieBreakersForCreatedAtAscending verifies explicit created_at ascending keeps the same upload-aware tie-breakers.
func TestBuildListTransactionsQueryUsesUploadAwareTieBreakersForCreatedAtAscending(t *testing.T) {
	t.Parallel()

	query, args := buildListTransactionsQuery(ports.TransactionFilter{
		SortBy:    ports.TransactionSortByCreatedAt,
		SortOrder: ports.TransactionSortOrderAscending,
	})
	if len(args) != 0 {
		t.Fatalf("len(args) = %d, want %d", len(args), 0)
	}

	wantOrderBy := "ORDER BY created_at ASC, upload_id ASC NULLS LAST, row_number ASC NULLS LAST, id ASC"
	if !strings.Contains(query, wantOrderBy) {
		t.Fatalf("buildListTransactionsQuery() query = %q, want it to contain %q", query, wantOrderBy)
	}
}

// TestBuildListTransactionsQueryKeepsNonCreatedAtSortsUnchanged verifies non-created_at sorting keeps the existing simple tiebreaker behavior.
func TestBuildListTransactionsQueryKeepsNonCreatedAtSortsUnchanged(t *testing.T) {
	t.Parallel()

	query, args := buildListTransactionsQuery(ports.TransactionFilter{
		SortBy:    ports.TransactionSortByTransactionCount,
		SortOrder: ports.TransactionSortOrderAscending,
	})
	if len(args) != 0 {
		t.Fatalf("len(args) = %d, want %d", len(args), 0)
	}

	wantOrderBy := "ORDER BY transaction_count ASC, id ASC"
	if !strings.Contains(query, wantOrderBy) {
		t.Fatalf("buildListTransactionsQuery() query = %q, want it to contain %q", query, wantOrderBy)
	}

	if strings.Contains(query, "row_number ASC NULLS LAST") {
		t.Fatalf("buildListTransactionsQuery() query = %q, want non-created_at ordering to avoid upload-aware tie-breakers", query)
	}
}

// TestNormalizeTransactionValueForDB verifies the normalize transaction value for DB behavior and the expected outcome asserted below.
func TestNormalizeTransactionValueForDB(t *testing.T) {
	t.Parallel()

	got := normalizeTransactionValueForDB(" 571,251.28 ")
	if got != "571251.28" {
		t.Fatalf("normalizeTransactionValueForDB() = %q, want %q", got, "571251.28")
	}
}

// TestCreateTransactionQueryCastsPipelineResultArgument verifies the create transaction query casts pipeline result argument behavior and the expected outcome asserted below.
func TestCreateTransactionQueryCastsPipelineResultArgument(t *testing.T) {
	t.Parallel()

	if !strings.Contains(createTransactionQuery, "$14::jsonb") {
		t.Fatalf("createTransactionQuery does not cast argument 14 as jsonb: %s", createTransactionQuery)
	}

	if strings.Contains(createTransactionQuery, "$15::jsonb") {
		t.Fatalf("createTransactionQuery casts wrong argument as jsonb: %s", createTransactionQuery)
	}
}

// TestBuildListTransactionUploadsQuery verifies the build list transaction uploads query behavior and the expected outcome asserted below.
func TestBuildListTransactionUploadsQuery(t *testing.T) {
	t.Parallel()

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	startedAt := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	endedAt := time.Date(2026, time.April, 30, 23, 59, 59, 0, time.UTC)
	query, args := buildListTransactionUploadsQuery(ports.TransactionUploadFilter{GroupID: groupID, FileName: "transactions", StartedAt: &startedAt, EndedAt: &endedAt})
	if query == "" {
		t.Fatal("expected query")
	}

	if len(args) != 4 {
		t.Fatalf("len(args) = %d, want %d", len(args), 4)
	}

	if !strings.Contains(query, "WHERE group_id = $1") {
		t.Fatalf("buildListTransactionUploadsQuery() query = %q, want it to contain %q", query, "WHERE group_id = $1")
	}

	if !strings.Contains(query, "file_name ILIKE $2") {
		t.Fatalf("buildListTransactionUploadsQuery() query = %q, want it to contain %q", query, "file_name ILIKE $2")
	}

	if !strings.Contains(query, "uploaded_at >= $3") {
		t.Fatalf("buildListTransactionUploadsQuery() query = %q, want it to contain %q", query, "uploaded_at >= $3")
	}

	if !strings.Contains(query, "uploaded_at <= $4") {
		t.Fatalf("buildListTransactionUploadsQuery() query = %q, want it to contain %q", query, "uploaded_at <= $4")
	}

	if args[0] != groupID.String() {
		t.Fatalf("args[0] = %v, want %q", args[0], groupID.String())
	}
}
