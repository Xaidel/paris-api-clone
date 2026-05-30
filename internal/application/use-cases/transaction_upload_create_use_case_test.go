package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	services "github.com/gyud-adb/paris-api/internal/application/services"
	"github.com/gyud-adb/paris-api/internal/domain"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	domainevents "github.com/gyud-adb/paris-api/internal/domain/events"
	domainservices "github.com/gyud-adb/paris-api/internal/domain/services"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

type transactionUploadRepositoryMock struct {
	createdUpload    *entities.TransactionUpload
	findByMD5GroupID valueobjects.GroupID
	findByMD5Content string
	findByMD5Result  *entities.TransactionUpload
	findByMD5Err     error
	findByIDResult   *entities.TransactionUpload
	findByIDErr      error
	listFilter       outboundports.TransactionUploadFilter
	listResult       []*entities.TransactionUpload
	listErr          error
	deleteByIDErr    error
	createErr        error
	callOrder        *[]string
}

func (m *transactionUploadRepositoryMock) Create(_ context.Context, upload *entities.TransactionUpload) error {
	if m.callOrder != nil {
		*m.callOrder = append(*m.callOrder, "upload.create")
	}
	m.createdUpload = upload
	return m.createErr
}

func (m *transactionUploadRepositoryMock) FindByID(context.Context, valueobjects.UploadID) (*entities.TransactionUpload, error) {
	return m.findByIDResult, m.findByIDErr
}

func (m *transactionUploadRepositoryMock) FindByContentMD5(_ context.Context, contentMD5 string, groupID valueobjects.GroupID) (*entities.TransactionUpload, error) {
	m.findByMD5GroupID = groupID
	m.findByMD5Content = contentMD5
	if m.findByMD5Result != nil && m.findByMD5Result.GroupID().Equal(groupID) {
		return m.findByMD5Result, m.findByMD5Err
	}

	return nil, m.findByMD5Err
}

func (m *transactionUploadRepositoryMock) List(_ context.Context, filter outboundports.TransactionUploadFilter) ([]*entities.TransactionUpload, error) {
	m.listFilter = filter
	return m.listResult, m.listErr
}

func (m *transactionUploadRepositoryMock) DeleteByID(context.Context, valueobjects.UploadID) error {
	return m.deleteByIDErr
}

type transactionUploadPreviewRepositoryMock struct {
	transactionManager     *transactionManagerUseCaseMock
	savedPreview           *outboundports.TransactionUploadPreviewRecord
	saveErr                error
	savedWithinTransaction bool
	findByUploadIDResult   *outboundports.TransactionUploadPreviewRecord
	findByUploadIDErr      error
	callOrder              *[]string
}

func (m *transactionUploadPreviewRepositoryMock) Save(_ context.Context, preview outboundports.TransactionUploadPreviewRecord) error {
	if m.callOrder != nil {
		*m.callOrder = append(*m.callOrder, "preview.save")
	}
	m.savedWithinTransaction = m.transactionManager != nil && m.transactionManager.withinTransaction
	previewCopy := preview
	previewCopy.Columns = append([]string(nil), preview.Columns...)
	previewCopy.Rows = clonePreviewRows(preview.Rows)
	previewCopy.ValidationErrors = append([]outboundports.TransactionFileValidationError(nil), preview.ValidationErrors...)
	m.savedPreview = &previewCopy
	return m.saveErr
}

func (m *transactionUploadPreviewRepositoryMock) FindByUploadID(context.Context, valueobjects.UploadID) (*outboundports.TransactionUploadPreviewRecord, error) {
	return m.findByUploadIDResult, m.findByUploadIDErr
}

type rawFileStoreMock struct {
	storedCommand  outboundports.StoreRawFileCommand
	storeResult    outboundports.StoreRawFileResult
	storeErr       error
	readCommand    outboundports.ReadRawFileCommand
	readResult     outboundports.ReadRawFileResult
	readErr        error
	deletedCommand outboundports.DeleteRawFileCommand
	deleteErr      error
}

func (m *rawFileStoreMock) Store(_ context.Context, command outboundports.StoreRawFileCommand) (outboundports.StoreRawFileResult, error) {
	m.storedCommand = command
	return m.storeResult, m.storeErr
}

func (m *rawFileStoreMock) Read(_ context.Context, command outboundports.ReadRawFileCommand) (outboundports.ReadRawFileResult, error) {
	m.readCommand = command
	return m.readResult, m.readErr
}

func (m *rawFileStoreMock) Delete(_ context.Context, command outboundports.DeleteRawFileCommand) error {
	m.deletedCommand = command
	return m.deleteErr
}

type fileParserMock struct {
	result outboundports.ParseTransactionFileResult
	err    error
}

func (m *fileParserMock) Parse(context.Context, outboundports.ParseTransactionFileCommand) (outboundports.ParseTransactionFileResult, error) {
	return m.result, m.err
}

type transactionManagerUseCaseMock struct {
	err               error
	invoked           bool
	withinTransaction bool
}

func (m *transactionManagerUseCaseMock) WithinTransaction(ctx context.Context, operation func(ctx context.Context) error) error {
	m.invoked = true
	if m.err != nil {
		return m.err
	}

	m.withinTransaction = true
	defer func() { m.withinTransaction = false }()

	return operation(ctx)
}

type adminEventRecorderUploadMock struct {
	command services.RecordAdminEventCommand
	err     error
	when    time.Time
}

type transactionUploadProgressReporterMock struct {
	updates []outboundports.TransactionUploadProgressUpdate
	err     error
}

func (m *transactionUploadProgressReporterMock) Report(_ context.Context, update outboundports.TransactionUploadProgressUpdate) error {
	m.updates = append(m.updates, update)
	return m.err
}

func (m *adminEventRecorderUploadMock) RecordAdminEvent(_ context.Context, command services.RecordAdminEventCommand) error {
	m.command = command
	return m.err
}

func (m *adminEventRecorderUploadMock) Publish(_ context.Context, recordedEvents []domain.DomainEvent) error {
	if len(recordedEvents) > 0 {
		adminEvent, ok := recordedEvents[0].(*domainevents.AdminActionOccurred)
		if !ok {
			return errors.New("unexpected domain event type")
		}

		m.command = services.RecordAdminEventCommand{
			ActorUserID:  adminEvent.ActorUserID(),
			ActorGroupID: adminEvent.ActorGroupID(),
			EventType:    adminEvent.EventType(),
			EventData:    json.RawMessage(adminEvent.EventData()),
		}
		m.when = adminEvent.OccurredAt()
	}

	return m.err
}

// This test documents the upload orchestration contract: parse and validate the
// file, store the raw asset, persist the upload and transactions, enqueue
// classification, report progress, and clean up correctly on failure.
func TestCreateTransactionUploadUseCaseExecute(t *testing.T) {
	t.Parallel()

	validator := domainservices.NewTransactionFileValidator(valueobjects.TransactionFileSchemaV1())

	parsedResult := outboundports.ParseTransactionFileResult{
		Format: "csv",
		Headers: []string{
			"Product",
			"Year",
			"Month",
			"DMC:IB",
			"DMC",
			"Partner Bank",
			"Reference Number",
			"Value of Transactions",
			"No. of Transactions",
			"Goods Description",
			"Goods Classification (Sector)",
			"Applicant (CG/RPA) or Sub-Borrower (RCF) Country",
			"Beneficiary Country",
			"Source",
			"Destination",
			"Tenor > 1 year",
			"E&S Category",
			"PA Alignment",
		},
		Rows: [][]string{{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-1", "698,436.80", "1", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"}},
	}

	fixedUploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	fixedTransactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	fixedGroupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300001")
	otherGroupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300003")

	tests := []struct {
		name        string
		uploadRepo  *transactionUploadRepositoryMock
		previewRepo *transactionUploadPreviewRepositoryMock
		callOrder   *[]string
		transaction *transactionRepositoryMock
		queue       *transactionProcessingQueueMock
		store       *rawFileStoreMock
		parser      *fileParserMock
		manager     *transactionManagerUseCaseMock
		recorder    *adminEventRecorderUploadMock
		reporter    *transactionUploadProgressReporterMock
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result inboundports.CreateTransactionUploadResult, uploadRepo *transactionUploadRepositoryMock, previewRepo *transactionUploadPreviewRepositoryMock, transaction *transactionRepositoryMock, queue *transactionProcessingQueueMock, store *rawFileStoreMock, parser *fileParserMock, manager *transactionManagerUseCaseMock, recorder *adminEventRecorderUploadMock, reporter *transactionUploadProgressReporterMock)
	}{
		{
			name:        "creates transaction upload",
			uploadRepo:  &transactionUploadRepositoryMock{},
			previewRepo: &transactionUploadPreviewRepositoryMock{},
			callOrder:   &[]string{},
			transaction: &transactionRepositoryMock{},
			queue:       &transactionProcessingQueueMock{},
			store:       &rawFileStoreMock{storeResult: outboundports.StoreRawFileResult{Provider: "local", Key: "upload/file.csv"}},
			parser:      &fileParserMock{result: parsedResult},
			manager:     &transactionManagerUseCaseMock{},
			recorder:    &adminEventRecorderUploadMock{},
			reporter:    &transactionUploadProgressReporterMock{},
			assert: func(t *testing.T, result inboundports.CreateTransactionUploadResult, uploadRepo *transactionUploadRepositoryMock, previewRepo *transactionUploadPreviewRepositoryMock, transaction *transactionRepositoryMock, queue *transactionProcessingQueueMock, store *rawFileStoreMock, parser *fileParserMock, manager *transactionManagerUseCaseMock, recorder *adminEventRecorderUploadMock, reporter *transactionUploadProgressReporterMock) {
				t.Helper()

				if uploadRepo.createdUpload == nil {
					t.Fatal("expected upload to be created")
				}

				if previewRepo.savedPreview == nil {
					t.Fatal("expected upload preview to be saved")
				}

				if len(*callOrderOrEmpty(previewRepo.callOrder)) < 2 {
					t.Fatalf("callOrder = %v, want upload.create before preview.save", *callOrderOrEmpty(previewRepo.callOrder))
				}

				if got := (*callOrderOrEmpty(previewRepo.callOrder))[0]; got != "upload.create" {
					t.Fatalf("callOrder[0] = %q, want %q", got, "upload.create")
				}

				if got := (*callOrderOrEmpty(previewRepo.callOrder))[1]; got != "preview.save" {
					t.Fatalf("callOrder[1] = %q, want %q", got, "preview.save")
				}

				if !previewRepo.savedWithinTransaction {
					t.Fatal("expected upload preview save within transaction")
				}

				if previewRepo.savedPreview.UploadID != fixedUploadID.String() {
					t.Fatalf("previewRepo.savedPreview.UploadID = %q, want %q", previewRepo.savedPreview.UploadID, fixedUploadID.String())
				}

				if len(previewRepo.savedPreview.Columns) != len(parsedResult.Headers) {
					t.Fatalf("len(previewRepo.savedPreview.Columns) = %d, want %d", len(previewRepo.savedPreview.Columns), len(parsedResult.Headers))
				}

				if len(previewRepo.savedPreview.Rows) != 1 {
					t.Fatalf("len(previewRepo.savedPreview.Rows) = %d, want %d", len(previewRepo.savedPreview.Rows), 1)
				}

				if previewRepo.savedPreview.Rows[0][6] != "REF-1" {
					t.Fatalf("previewRepo.savedPreview.Rows[0][6] = %q, want %q", previewRepo.savedPreview.Rows[0][6], "REF-1")
				}

				if previewRepo.savedPreview.TotalRows != 1 {
					t.Fatalf("previewRepo.savedPreview.TotalRows = %d, want %d", previewRepo.savedPreview.TotalRows, 1)
				}

				if len(previewRepo.savedPreview.ValidationErrors) != 0 {
					t.Fatalf("len(previewRepo.savedPreview.ValidationErrors) = %d, want 0", len(previewRepo.savedPreview.ValidationErrors))
				}

				parser.result.Rows[0][6] = "CHANGED"
				if previewRepo.savedPreview.Rows[0][6] != "REF-1" {
					t.Fatalf("previewRepo.savedPreview.Rows[0][6] after parser mutation = %q, want %q", previewRepo.savedPreview.Rows[0][6], "REF-1")
				}

				if len(transaction.createdTransactions) != 1 {
					t.Fatalf("len(transaction.createdTransactions) = %d, want %d", len(transaction.createdTransactions), 1)
				}

				created := transaction.createdTransactions[0]
				if created.Product() != "CG" {
					t.Fatalf("created.Product() = %q, want %q", created.Product(), "CG")
				}

				if created.ReferenceNumber() != "REF-1" {
					t.Fatalf("created.ReferenceNumber() = %q, want %q", created.ReferenceNumber(), "REF-1")
				}

				if created.TransactionValue() != "698,436.80" {
					t.Fatalf("created.TransactionValue() = %q, want %q", created.TransactionValue(), "698,436.80")
				}

				if len(queue.enqueuedIDs) != 1 || queue.enqueuedIDs[0] != fixedTransactionID.String() {
					t.Fatalf("queue.enqueuedIDs = %v, want [%q]", queue.enqueuedIDs, fixedTransactionID.String())
				}

				if transaction.updatedTransaction == nil {
					t.Fatal("expected transaction status update")
				}

				if !manager.invoked {
					t.Fatal("expected transaction manager invocation")
				}

				if result.Upload.FileFormat != "csv" {
					t.Fatalf("result.Upload.FileFormat = %q, want %q", result.Upload.FileFormat, "csv")
				}

				if result.Upload.GroupID != testGroupID(t).String() {
					t.Fatalf("result.Upload.GroupID = %q, want %q", result.Upload.GroupID, testGroupID(t).String())
				}

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}

				if payload["upload_id"] != result.Upload.ID {
					t.Fatalf("payload[upload_id] = %v, want %q", payload["upload_id"], result.Upload.ID)
				}

				if !uploadRepo.findByMD5GroupID.Equal(fixedGroupID) {
					t.Fatalf("uploadRepo.findByMD5GroupID = %q, want %q", uploadRepo.findByMD5GroupID.String(), fixedGroupID.String())
				}

				assertTransactionUploadProgressStatuses(t, reporter.updates, []string{
					outboundports.TransactionUploadProgressStatusParsed,
					outboundports.TransactionUploadProgressStatusStoredRawFile,
					outboundports.TransactionUploadProgressStatusValidated,
					outboundports.TransactionUploadProgressStatusSavedUpload,
					outboundports.TransactionUploadProgressStatusRecordedEvent,
					outboundports.TransactionUploadProgressStatusPersistedTransactions,
					outboundports.TransactionUploadProgressStatusCompleted,
				})

				assertTransactionUploadProgressMonotonic(t, reporter.updates)
			},
		},
		{
			name:        "persists failed upload when all rows are malformed",
			uploadRepo:  &transactionUploadRepositoryMock{},
			previewRepo: &transactionUploadPreviewRepositoryMock{},
			transaction: &transactionRepositoryMock{},
			queue:       &transactionProcessingQueueMock{},
			store:       &rawFileStoreMock{storeResult: outboundports.StoreRawFileResult{Provider: "local", Key: "upload/file.csv"}},
			parser:      &fileParserMock{result: outboundports.ParseTransactionFileResult{Format: "csv", Headers: parsedResult.Headers, Rows: [][]string{{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-1", "698,436.80", "bad-count", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"}}}},
			manager:     &transactionManagerUseCaseMock{},
			recorder:    &adminEventRecorderUploadMock{},
			reporter:    &transactionUploadProgressReporterMock{},
			assert: func(t *testing.T, result inboundports.CreateTransactionUploadResult, uploadRepo *transactionUploadRepositoryMock, previewRepo *transactionUploadPreviewRepositoryMock, transaction *transactionRepositoryMock, queue *transactionProcessingQueueMock, store *rawFileStoreMock, _ *fileParserMock, manager *transactionManagerUseCaseMock, recorder *adminEventRecorderUploadMock, reporter *transactionUploadProgressReporterMock) {
				t.Helper()

				if uploadRepo.createdUpload == nil {
					t.Fatal("expected failed upload to be created")
				}

				if previewRepo.savedPreview == nil {
					t.Fatal("expected failed upload preview to be saved")
				}

				if !previewRepo.savedWithinTransaction {
					t.Fatal("expected failed upload preview save within transaction")
				}

				if len(previewRepo.savedPreview.Rows) != 0 {
					t.Fatalf("len(previewRepo.savedPreview.Rows) = %d, want 0", len(previewRepo.savedPreview.Rows))
				}

				if previewRepo.savedPreview.TotalRows != 0 {
					t.Fatalf("previewRepo.savedPreview.TotalRows = %d, want 0", previewRepo.savedPreview.TotalRows)
				}

				if len(previewRepo.savedPreview.ValidationErrors) != 0 {
					t.Fatalf("len(previewRepo.savedPreview.ValidationErrors) = %d, want 0", len(previewRepo.savedPreview.ValidationErrors))
				}

				if uploadRepo.createdUpload.Status() != valueobjects.FailedTransactionUploadStatus().String() {
					t.Fatalf("uploadRepo.createdUpload.Status() = %q, want %q", uploadRepo.createdUpload.Status(), valueobjects.FailedTransactionUploadStatus().String())
				}

				if uploadRepo.createdUpload.RowCount() != 0 {
					t.Fatalf("uploadRepo.createdUpload.RowCount() = %d, want 0", uploadRepo.createdUpload.RowCount())
				}

				if result.Upload.Status != valueobjects.FailedTransactionUploadStatus().String() {
					t.Fatalf("result.Upload.Status = %q, want %q", result.Upload.Status, valueobjects.FailedTransactionUploadStatus().String())
				}

				if result.Upload.RowCount != 0 {
					t.Fatalf("result.Upload.RowCount = %d, want 0", result.Upload.RowCount)
				}

				if recorder.command.EventType != createTransactionUploadAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, createTransactionUploadAdminEventType)
				}

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}

				if payload["upload_id"] != result.Upload.ID {
					t.Fatalf("payload[upload_id] = %v, want %q", payload["upload_id"], result.Upload.ID)
				}

				if len(transaction.createdTransactions) != 0 {
					t.Fatalf("len(transaction.createdTransactions) = %d, want %d", len(transaction.createdTransactions), 0)
				}

				if len(queue.enqueuedIDs) != 0 {
					t.Fatalf("len(queue.enqueuedIDs) = %d, want 0", len(queue.enqueuedIDs))
				}

				if !manager.invoked {
					t.Fatal("expected transaction manager invocation")
				}

				if store.deletedCommand.Key != "" {
					t.Fatalf("store.deletedCommand.Key = %q, want empty", store.deletedCommand.Key)
				}

				assertTransactionUploadProgressStatuses(t, reporter.updates, []string{
					outboundports.TransactionUploadProgressStatusParsed,
					outboundports.TransactionUploadProgressStatusStoredRawFile,
					outboundports.TransactionUploadProgressStatusValidationFailed,
				})

				lastUpdate := reporter.updates[len(reporter.updates)-1]
				if lastUpdate.Upload == nil {
					t.Fatal("expected failed upload payload in progress update")
				}

				if lastUpdate.Upload.Status != valueobjects.FailedTransactionUploadStatus().String() {
					t.Fatalf("lastUpdate.Upload.Status = %q, want %q", lastUpdate.Upload.Status, valueobjects.FailedTransactionUploadStatus().String())
				}
			},
		},
		{
			name:        "returns skipped rows warnings for mixed transaction counts",
			uploadRepo:  &transactionUploadRepositoryMock{},
			previewRepo: &transactionUploadPreviewRepositoryMock{},
			transaction: &transactionRepositoryMock{},
			queue:       &transactionProcessingQueueMock{},
			store:       &rawFileStoreMock{storeResult: outboundports.StoreRawFileResult{Provider: "local", Key: "upload/file.csv"}},
			parser: &fileParserMock{result: outboundports.ParseTransactionFileResult{
				Format:  "csv",
				Headers: parsedResult.Headers,
				Rows: [][]string{
					{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-1", "698,436.80", "1", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"},
					{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-2", "698,436.80", "2", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"},
					{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-3", "698,436.80", "bad-count", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"},
				},
			}},
			manager:  &transactionManagerUseCaseMock{},
			recorder: &adminEventRecorderUploadMock{},
			reporter: &transactionUploadProgressReporterMock{},
			assert: func(t *testing.T, result inboundports.CreateTransactionUploadResult, _ *transactionUploadRepositoryMock, _ *transactionUploadPreviewRepositoryMock, transaction *transactionRepositoryMock, _ *transactionProcessingQueueMock, _ *rawFileStoreMock, _ *fileParserMock, _ *transactionManagerUseCaseMock, _ *adminEventRecorderUploadMock, reporter *transactionUploadProgressReporterMock) {
				t.Helper()

				if len(transaction.createdTransactions) != 1 {
					t.Fatalf("len(transaction.createdTransactions) = %d, want %d", len(transaction.createdTransactions), 1)
				}

				if transaction.createdTransactions[0].ReferenceNumber() != "REF-1" {
					t.Fatalf("transaction.createdTransactions[0].ReferenceNumber() = %q, want %q", transaction.createdTransactions[0].ReferenceNumber(), "REF-1")
				}

				if result.Upload.RowCount != 1 {
					t.Fatalf("result.Upload.RowCount = %d, want %d", result.Upload.RowCount, 1)
				}

				assertCreateUploadSkippedRows(t, result.SkippedRows, []outboundports.TransactionUploadSkippedRow{
					{RowNumber: 3, Reason: outboundports.TransactionUploadSkippedRowReasonNotValidTransaction},
					{RowNumber: 4, Reason: outboundports.TransactionUploadSkippedRowReasonMalformed},
				})

				lastUpdate := reporter.updates[len(reporter.updates)-1]
				assertCreateUploadSkippedRows(t, lastUpdate.SkippedRows, []outboundports.TransactionUploadSkippedRow{
					{RowNumber: 3, Reason: outboundports.TransactionUploadSkippedRowReasonNotValidTransaction},
					{RowNumber: 4, Reason: outboundports.TransactionUploadSkippedRowReasonMalformed},
				})

				assertTransactionUploadProgressStatuses(t, reporter.updates, []string{
					outboundports.TransactionUploadProgressStatusParsed,
					outboundports.TransactionUploadProgressStatusStoredRawFile,
					outboundports.TransactionUploadProgressStatusValidated,
					outboundports.TransactionUploadProgressStatusSavedUpload,
					outboundports.TransactionUploadProgressStatusRecordedEvent,
					outboundports.TransactionUploadProgressStatusPersistedTransactions,
					outboundports.TransactionUploadProgressStatusCompleted,
				})
			},
		},
		{
			name:        "persists failed upload when no eligible rows remain",
			uploadRepo:  &transactionUploadRepositoryMock{},
			previewRepo: &transactionUploadPreviewRepositoryMock{},
			transaction: &transactionRepositoryMock{},
			queue:       &transactionProcessingQueueMock{},
			store:       &rawFileStoreMock{storeResult: outboundports.StoreRawFileResult{Provider: "local", Key: "upload/file.csv"}},
			parser: &fileParserMock{result: outboundports.ParseTransactionFileResult{
				Format: "csv",
				Headers: []string{
					"Product",
					"Month",
					"DMC:IB",
					"DMC",
					"Partner Bank",
					"Reference Number",
					"Value of Transactions",
					"No. of Transactions",
					"Goods Description",
					"Goods Classification (Sector)",
					"Applicant (CG/RPA) or Sub-Borrower (RCF) Country",
					"Beneficiary Country",
					"Source",
					"Destination",
					"Tenor > 1 year",
					"E&S Category",
					"PA Alignment",
				},
				Rows: [][]string{
					{"CG", "4", "IB", "DMC", "Partner Bank", "REF-2", "698,436.80", "2", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"},
					{"CG", "4", "IB", "DMC", "Partner Bank", "REF-3", "698,436.80", "0", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"},
				},
			}},
			manager:  &transactionManagerUseCaseMock{},
			recorder: &adminEventRecorderUploadMock{},
			reporter: &transactionUploadProgressReporterMock{},
			assert: func(t *testing.T, result inboundports.CreateTransactionUploadResult, uploadRepo *transactionUploadRepositoryMock, previewRepo *transactionUploadPreviewRepositoryMock, transaction *transactionRepositoryMock, queue *transactionProcessingQueueMock, store *rawFileStoreMock, _ *fileParserMock, manager *transactionManagerUseCaseMock, _ *adminEventRecorderUploadMock, reporter *transactionUploadProgressReporterMock) {
				t.Helper()

				if uploadRepo.createdUpload == nil {
					t.Fatal("expected failed upload to be created")
				}

				if previewRepo.savedPreview == nil {
					t.Fatal("expected zero-row upload preview to be saved")
				}

				if previewRepo.savedPreview.Rows != nil {
					t.Fatalf("previewRepo.savedPreview.Rows = %v, want nil", previewRepo.savedPreview.Rows)
				}

				if len(previewRepo.savedPreview.Rows) != 0 {
					t.Fatalf("len(previewRepo.savedPreview.Rows) = %d, want 0", len(previewRepo.savedPreview.Rows))
				}

				if previewRepo.savedPreview.TotalRows != 0 {
					t.Fatalf("previewRepo.savedPreview.TotalRows = %d, want 0", previewRepo.savedPreview.TotalRows)
				}

				if len(previewRepo.savedPreview.ValidationErrors) == 0 {
					t.Fatal("expected zero-row preview validation errors")
				}

				if uploadRepo.createdUpload.Status() != valueobjects.FailedTransactionUploadStatus().String() {
					t.Fatalf("uploadRepo.createdUpload.Status() = %q, want %q", uploadRepo.createdUpload.Status(), valueobjects.FailedTransactionUploadStatus().String())
				}

				if uploadRepo.createdUpload.RowCount() != 0 {
					t.Fatalf("uploadRepo.createdUpload.RowCount() = %d, want 0", uploadRepo.createdUpload.RowCount())
				}

				if result.Upload.Status != valueobjects.FailedTransactionUploadStatus().String() {
					t.Fatalf("result.Upload.Status = %q, want %q", result.Upload.Status, valueobjects.FailedTransactionUploadStatus().String())
				}

				if result.Upload.RowCount != 0 {
					t.Fatalf("result.Upload.RowCount = %d, want 0", result.Upload.RowCount)
				}

				if len(result.ValidationErrors) == 0 {
					t.Fatal("expected zero-row validation errors")
				}

				if len(transaction.createdTransactions) != 0 {
					t.Fatalf("len(transaction.createdTransactions) = %d, want %d", len(transaction.createdTransactions), 0)
				}

				if len(queue.enqueuedIDs) != 0 {
					t.Fatalf("len(queue.enqueuedIDs) = %d, want 0", len(queue.enqueuedIDs))
				}

				if !manager.invoked {
					t.Fatal("expected transaction manager invocation")
				}

				if store.deletedCommand.Key != "" {
					t.Fatalf("store.deletedCommand.Key = %q, want empty", store.deletedCommand.Key)
				}

				assertTransactionUploadProgressStatuses(t, reporter.updates, []string{
					outboundports.TransactionUploadProgressStatusParsed,
					outboundports.TransactionUploadProgressStatusStoredRawFile,
					outboundports.TransactionUploadProgressStatusValidationFailed,
				})

				lastUpdate := reporter.updates[len(reporter.updates)-1]
				if len(lastUpdate.ValidationErrors) == 0 {
					t.Fatal("expected zero-row progress validation errors")
				}
			},
		},
		{
			name:        "returns validation errors and skipped rows together",
			uploadRepo:  &transactionUploadRepositoryMock{},
			previewRepo: &transactionUploadPreviewRepositoryMock{},
			transaction: &transactionRepositoryMock{},
			queue:       &transactionProcessingQueueMock{},
			store:       &rawFileStoreMock{storeResult: outboundports.StoreRawFileResult{Provider: "local", Key: "upload/file.csv"}},
			parser: &fileParserMock{result: outboundports.ParseTransactionFileResult{
				Format:  "csv",
				Headers: parsedResult.Headers,
				Rows: [][]string{
					{"CG", "", "4", "IB", "DMC", "Partner Bank", "REF-1", "698,436.80", "1", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"},
					{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-2", "698,436.80", "bad-count", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"},
				},
			}},
			manager:  &transactionManagerUseCaseMock{},
			recorder: &adminEventRecorderUploadMock{},
			reporter: &transactionUploadProgressReporterMock{},
			assert: func(t *testing.T, result inboundports.CreateTransactionUploadResult, uploadRepo *transactionUploadRepositoryMock, previewRepo *transactionUploadPreviewRepositoryMock, transaction *transactionRepositoryMock, _ *transactionProcessingQueueMock, _ *rawFileStoreMock, _ *fileParserMock, _ *transactionManagerUseCaseMock, _ *adminEventRecorderUploadMock, reporter *transactionUploadProgressReporterMock) {
				t.Helper()

				if uploadRepo.createdUpload == nil {
					t.Fatal("expected failed upload to be created")
				}

				if previewRepo.savedPreview == nil {
					t.Fatal("expected validation preview to be saved")
				}

				if len(previewRepo.savedPreview.Rows) != 1 {
					t.Fatalf("len(previewRepo.savedPreview.Rows) = %d, want %d", len(previewRepo.savedPreview.Rows), 1)
				}

				if previewRepo.savedPreview.Rows[0][6] != "REF-1" {
					t.Fatalf("previewRepo.savedPreview.Rows[0][6] = %q, want %q", previewRepo.savedPreview.Rows[0][6], "REF-1")
				}

				if previewRepo.savedPreview.TotalRows != 1 {
					t.Fatalf("previewRepo.savedPreview.TotalRows = %d, want %d", previewRepo.savedPreview.TotalRows, 1)
				}

				if len(previewRepo.savedPreview.ValidationErrors) == 0 {
					t.Fatal("expected preview validation errors")
				}

				if uploadRepo.createdUpload.Status() != valueobjects.FailedTransactionUploadStatus().String() {
					t.Fatalf("uploadRepo.createdUpload.Status() = %q, want %q", uploadRepo.createdUpload.Status(), valueobjects.FailedTransactionUploadStatus().String())
				}

				if uploadRepo.createdUpload.RowCount() != 0 {
					t.Fatalf("uploadRepo.createdUpload.RowCount() = %d, want 0", uploadRepo.createdUpload.RowCount())
				}

				if result.Upload.Status != valueobjects.FailedTransactionUploadStatus().String() {
					t.Fatalf("result.Upload.Status = %q, want %q", result.Upload.Status, valueobjects.FailedTransactionUploadStatus().String())
				}

				if len(transaction.createdTransactions) != 0 {
					t.Fatalf("len(transaction.createdTransactions) = %d, want %d", len(transaction.createdTransactions), 0)
				}

				if len(result.ValidationErrors) == 0 {
					t.Fatal("expected validation errors")
				}

				assertCreateUploadSkippedRows(t, result.SkippedRows, []outboundports.TransactionUploadSkippedRow{{RowNumber: 3, Reason: outboundports.TransactionUploadSkippedRowReasonMalformed}})

				lastUpdate := reporter.updates[len(reporter.updates)-1]
				if lastUpdate.Status != outboundports.TransactionUploadProgressStatusValidationFailed {
					t.Fatalf("lastUpdate.Status = %q, want %q", lastUpdate.Status, outboundports.TransactionUploadProgressStatusValidationFailed)
				}

				if lastUpdate.Upload == nil {
					t.Fatal("expected failed upload payload in progress update")
				}

				if lastUpdate.Upload.Status != valueobjects.FailedTransactionUploadStatus().String() {
					t.Fatalf("lastUpdate.Upload.Status = %q, want %q", lastUpdate.Upload.Status, valueobjects.FailedTransactionUploadStatus().String())
				}

				assertCreateUploadSkippedRows(t, lastUpdate.SkippedRows, []outboundports.TransactionUploadSkippedRow{{RowNumber: 3, Reason: outboundports.TransactionUploadSkippedRowReasonMalformed}})
			},
		},
		{
			name:        "filters mixed rows for xls uploads",
			uploadRepo:  &transactionUploadRepositoryMock{},
			previewRepo: &transactionUploadPreviewRepositoryMock{},
			transaction: &transactionRepositoryMock{},
			queue:       &transactionProcessingQueueMock{},
			store:       &rawFileStoreMock{storeResult: outboundports.StoreRawFileResult{Provider: "local", Key: "upload/file.xls"}},
			parser: &fileParserMock{result: outboundports.ParseTransactionFileResult{
				Format:  "xls",
				Headers: parsedResult.Headers,
				Rows: [][]string{
					{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-1", "698,436.80", "1", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"},
					{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-2", "698,436.80", "2", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"},
				},
			}},
			manager:  &transactionManagerUseCaseMock{},
			recorder: &adminEventRecorderUploadMock{},
			reporter: &transactionUploadProgressReporterMock{},
			assert: func(t *testing.T, result inboundports.CreateTransactionUploadResult, _ *transactionUploadRepositoryMock, _ *transactionUploadPreviewRepositoryMock, transaction *transactionRepositoryMock, _ *transactionProcessingQueueMock, _ *rawFileStoreMock, _ *fileParserMock, _ *transactionManagerUseCaseMock, _ *adminEventRecorderUploadMock, _ *transactionUploadProgressReporterMock) {
				t.Helper()

				if result.Upload.FileFormat != "xls" {
					t.Fatalf("result.Upload.FileFormat = %q, want %q", result.Upload.FileFormat, "xls")
				}

				if len(transaction.createdTransactions) != 1 {
					t.Fatalf("len(transaction.createdTransactions) = %d, want %d", len(transaction.createdTransactions), 1)
				}

				if transaction.createdTransactions[0].ReferenceNumber() != "REF-1" {
					t.Fatalf("transaction.createdTransactions[0].ReferenceNumber() = %q, want %q", transaction.createdTransactions[0].ReferenceNumber(), "REF-1")
				}
			},
		},
		{
			name:        "rejects duplicate file hash in same group",
			uploadRepo:  &transactionUploadRepositoryMock{findByMD5Result: testReconstitutedTransactionUpload(t, fixedUploadID, valueobjects.UploadedTransactionUploadStatus(), 1)},
			previewRepo: &transactionUploadPreviewRepositoryMock{},
			transaction: &transactionRepositoryMock{},
			queue:       &transactionProcessingQueueMock{},
			store:       &rawFileStoreMock{},
			parser:      &fileParserMock{result: parsedResult},
			manager:     &transactionManagerUseCaseMock{},
			recorder:    &adminEventRecorderUploadMock{},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "upload file has already been ingested") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
		{
			name: "allows duplicate file hash in different group",
			uploadRepo: &transactionUploadRepositoryMock{findByMD5Result: func() *entities.TransactionUpload {
				upload, err := entities.ReconstituteTransactionUpload(fixedUploadID, otherGroupID, "transactions.csv", "csv", "d41d8cd98f00b204e9800998ecf8427e", "local", "upload/file.csv", "transaction-file-v1", valueobjects.UploadedTransactionUploadStatus(), 1, testTime())
				if err != nil {
					t.Fatalf("ReconstituteTransactionUpload() error = %v", err)
				}

				return upload
			}()},
			previewRepo: &transactionUploadPreviewRepositoryMock{},
			transaction: &transactionRepositoryMock{},
			queue:       &transactionProcessingQueueMock{},
			store:       &rawFileStoreMock{storeResult: outboundports.StoreRawFileResult{Provider: "local", Key: "upload/file.csv"}},
			parser:      &fileParserMock{result: parsedResult},
			manager:     &transactionManagerUseCaseMock{},
			recorder:    &adminEventRecorderUploadMock{},
			reporter:    &transactionUploadProgressReporterMock{},
			assert: func(t *testing.T, result inboundports.CreateTransactionUploadResult, uploadRepo *transactionUploadRepositoryMock, _ *transactionUploadPreviewRepositoryMock, transaction *transactionRepositoryMock, _ *transactionProcessingQueueMock, _ *rawFileStoreMock, _ *fileParserMock, manager *transactionManagerUseCaseMock, _ *adminEventRecorderUploadMock, _ *transactionUploadProgressReporterMock) {
				t.Helper()

				if result.Upload.ID == "" {
					t.Fatal("expected upload result")
				}

				if uploadRepo.createdUpload == nil {
					t.Fatal("expected upload to be created")
				}

				if len(transaction.createdTransactions) != 1 {
					t.Fatalf("len(transaction.createdTransactions) = %d, want %d", len(transaction.createdTransactions), 1)
				}

				if !manager.invoked {
					t.Fatal("expected transaction manager invocation")
				}
			},
		},
		{
			name:        "cleans up raw file when transaction write fails",
			uploadRepo:  &transactionUploadRepositoryMock{},
			previewRepo: &transactionUploadPreviewRepositoryMock{},
			transaction: &transactionRepositoryMock{createErr: errors.New("insert failed")},
			queue:       &transactionProcessingQueueMock{},
			store:       &rawFileStoreMock{storeResult: outboundports.StoreRawFileResult{Provider: "local", Key: "upload/file.csv"}},
			parser:      &fileParserMock{result: parsedResult},
			manager:     &transactionManagerUseCaseMock{},
			recorder:    &adminEventRecorderUploadMock{},
			reporter:    &transactionUploadProgressReporterMock{},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil {
					t.Fatal("expected error")
				}
			},
			assert: func(t *testing.T, _ inboundports.CreateTransactionUploadResult, _ *transactionUploadRepositoryMock, previewRepo *transactionUploadPreviewRepositoryMock, _ *transactionRepositoryMock, _ *transactionProcessingQueueMock, store *rawFileStoreMock, _ *fileParserMock, _ *transactionManagerUseCaseMock, _ *adminEventRecorderUploadMock, reporter *transactionUploadProgressReporterMock) {
				t.Helper()
				if store.deletedCommand.Key != "upload/file.csv" {
					t.Fatalf("store.deletedCommand.Key = %q, want %q", store.deletedCommand.Key, "upload/file.csv")
				}
				if previewRepo.savedPreview == nil {
					t.Fatal("expected preview save before transaction failure")
				}
				assertTransactionUploadProgressStatuses(t, reporter.updates, []string{
					outboundports.TransactionUploadProgressStatusParsed,
					outboundports.TransactionUploadProgressStatusStoredRawFile,
					outboundports.TransactionUploadProgressStatusValidated,
				})
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.previewRepo != nil {
				tt.previewRepo.transactionManager = tt.manager
				tt.previewRepo.callOrder = tt.callOrder
			}

			if tt.uploadRepo != nil {
				tt.uploadRepo.callOrder = tt.callOrder
			}

			useCase := NewCreateTransactionUploadUseCase(tt.uploadRepo, tt.previewRepo, tt.transaction, tt.queue, tt.store, tt.parser, tt.manager, tt.recorder, &actorDirectoryMock{}, validator)
			useCase.now = func() time.Time { return testTime() }
			useCase.newUploadID = func() (valueobjects.UploadID, error) { return fixedUploadID, nil }
			useCase.newTransactionID = func() (valueobjects.TransactionID, error) { return fixedTransactionID, nil }

			result, err := useCase.Execute(context.Background(), inboundports.CreateTransactionUploadCommand{FileName: "transactions.csv", FileBytes: []byte("file-content"), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: fixedGroupID.String(), ProgressReporter: tt.reporter})
			if tt.assertError != nil {
				tt.assertError(t, err)
				if tt.assert != nil {
					tt.assert(t, result, tt.uploadRepo, tt.previewRepo, tt.transaction, tt.queue, tt.store, tt.parser, tt.manager, tt.recorder, tt.reporter)
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tt.assert(t, result, tt.uploadRepo, tt.previewRepo, tt.transaction, tt.queue, tt.store, tt.parser, tt.manager, tt.recorder, tt.reporter)
		})
	}
}

func TestNewTransactionUploadPreviewRecordPreservesNilRows(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	preview := newTransactionUploadPreviewRecord(uploadID, []string{"Product"}, nil, nil)

	if preview.Rows != nil {
		t.Fatalf("preview.Rows = %v, want nil", preview.Rows)
	}

	if preview.TotalRows != 0 {
		t.Fatalf("preview.TotalRows = %d, want 0", preview.TotalRows)
	}
}

func clonePreviewRows(rows [][]string) [][]string {
	if len(rows) == 0 {
		return nil
	}

	cloned := make([][]string, len(rows))
	for i, row := range rows {
		cloned[i] = append([]string(nil), row...)
	}

	return cloned
}

func callOrderOrEmpty(order *[]string) *[]string {
	if order == nil {
		empty := []string{}
		return &empty
	}

	return order
}

func assertTransactionUploadProgressStatuses(t *testing.T, updates []outboundports.TransactionUploadProgressUpdate, want []string) {
	t.Helper()

	got := make([]string, 0, len(updates))
	for _, update := range updates {
		got = append(got, update.Status)
	}

	if len(got) != len(want) {
		t.Fatalf("len(progress statuses) = %d, want %d (%v)", len(got), len(want), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("progress status[%d] = %q, want %q (all = %v)", i, got[i], want[i], got)
		}
	}
}

func assertTransactionUploadProgressMonotonic(t *testing.T, updates []outboundports.TransactionUploadProgressUpdate) {
	t.Helper()

	if len(updates) == 0 {
		t.Fatal("expected progress updates")
	}

	lastProgress := updates[0].Progress
	for i := 1; i < len(updates); i++ {
		if updates[i].Progress < lastProgress {
			t.Fatalf("progress[%d] = %d, previous = %d", i, updates[i].Progress, lastProgress)
		}

		lastProgress = updates[i].Progress
	}
}

func assertCreateUploadSkippedRows(t *testing.T, got, want []outboundports.TransactionUploadSkippedRow) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(skippedRows) = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("skippedRows[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// This test fixes the task-selection rule used by uploads: ReAct is the
// default, explicit supported tasks are preserved, and unknown values fall back
// to the default path.
func TestUploadTransactionClassificationTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		taskName string
		want     string
	}{
		{name: "defaults to react task", taskName: "", want: outboundports.TransactionClassifyReactTaskName},
		{name: "preserves explicit react task", taskName: outboundports.TransactionClassifyReactTaskName, want: outboundports.TransactionClassifyReactTaskName},
		{name: "remaps explicit legacy task to react", taskName: outboundports.TransactionClassifyTaskName, want: outboundports.TransactionClassifyReactTaskName},
		{name: "falls back unknown task to react", taskName: "unknown", want: outboundports.TransactionClassifyReactTaskName},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := uploadTransactionClassificationTask(tc.taskName); got != tc.want {
				t.Fatalf("uploadTransactionClassificationTask(%q) = %q, want %q", tc.taskName, got, tc.want)
			}
		})
	}
}

// This test verifies the factory builds one upload aggregate and normalized
// transaction rows from a schema-complete input file.
func TestTransactionUploadFactoryBuild(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	factory := domainservices.NewTransactionUploadFactory()
	upload, transactions, err := factory.Build(
		uploadID,
		testGroupID(t),
		"transactions.csv",
		"csv",
		"d41d8cd98f00b204e9800998ecf8427e",
		"local",
		"upload/file.csv",
		"transaction-file-v1",
		[]string{"Product", "Year", "Month", "DMC:IB", "DMC", "Partner Bank", "Reference Number", "Value of Transactions", "No. of Transactions", "Goods Description", "Goods Classification (Sector)", "Applicant (CG/RPA) or Sub-Borrower (RCF) Country", "Beneficiary Country", "Source", "Destination", "Tenor > 1 year", "E&S Category", "PA Alignment"},
		[][]string{{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-1", "698,436.80", "1", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned"}},
		testTime(),
		func() (valueobjects.TransactionID, error) { return transactionID, nil },
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if upload == nil {
		t.Fatal("expected upload")
	}

	if len(transactions) != 1 {
		t.Fatalf("len(transactions) = %d, want %d", len(transactions), 1)
	}

	transaction := transactions[0]
	if transaction.Product() != "CG" {
		t.Fatalf("transaction.Product() = %q, want %q", transaction.Product(), "CG")
	}

	if transaction.TransactionValue() != "698,436.80" {
		t.Fatalf("transaction.TransactionValue() = %q, want %q", transaction.TransactionValue(), "698,436.80")
	}

	if transaction.PAAlignment() != "PA Aligned" {
		t.Fatalf("transaction.PAAlignment() = %q, want %q", transaction.PAAlignment(), "PA Aligned")
	}
}

// This test keeps optional PA alignment handling backward compatible for files
// generated before that column existed.
func TestTransactionUploadFactoryBuildWithoutPAAlignmentColumn(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	factory := domainservices.NewTransactionUploadFactory()
	_, transactions, err := factory.Build(
		uploadID,
		testGroupID(t),
		"transactions.csv",
		"csv",
		"d41d8cd98f00b204e9800998ecf8427e",
		"local",
		"upload/file.csv",
		"transaction-file-v1",
		[]string{"Product", "Year", "Month", "DMC:IB", "DMC", "Partner Bank", "Reference Number", "Value of Transactions", "No. of Transactions", "Goods Description", "Goods Classification (Sector)", "Applicant (CG/RPA) or Sub-Borrower (RCF) Country", "Beneficiary Country", "Source", "Destination", "Tenor > 1 year", "E&S Category"},
		[][]string{{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-1", "698,436.80", "1", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", ""}},
		testTime(),
		func() (valueobjects.TransactionID, error) { return transactionID, nil },
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("len(transactions) = %d, want %d", len(transactions), 1)
	}

	if transactions[0].PAAlignment() != "" {
		t.Fatalf("transactions[0].PAAlignment() = %q, want empty", transactions[0].PAAlignment())
	}
}

// This test keeps optional E&S category handling backward compatible for older
// file layouts that omit that column.
func TestTransactionUploadFactoryBuildWithoutESCategoryColumn(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	factory := domainservices.NewTransactionUploadFactory()
	_, transactions, err := factory.Build(
		uploadID,
		testGroupID(t),
		"transactions.csv",
		"csv",
		"d41d8cd98f00b204e9800998ecf8427e",
		"local",
		"upload/file.csv",
		"transaction-file-v1",
		[]string{"Product", "Year", "Month", "DMC:IB", "DMC", "Partner Bank", "Reference Number", "Value of Transactions", "No. of Transactions", "Goods Description", "Goods Classification (Sector)", "Applicant (CG/RPA) or Sub-Borrower (RCF) Country", "Beneficiary Country", "Source", "Destination", "Tenor > 1 year", "PA Alignment"},
		[][]string{{"CG", "2026", "4", "IB", "DMC", "Partner Bank", "REF-1", "698,436.80", "1", "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "PA Aligned"}},
		testTime(),
		func() (valueobjects.TransactionID, error) { return transactionID, nil },
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("len(transactions) = %d, want %d", len(transactions), 1)
	}

	if transactions[0].ESCategory() != "" {
		t.Fatalf("transactions[0].ESCategory() = %q, want empty", transactions[0].ESCategory())
	}
}
