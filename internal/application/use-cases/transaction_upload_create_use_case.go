package usecases

import (
	"context"
	"crypto/md5"
	"fmt"
	"reflect"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	domainservices "github.com/gyud-adb/paris-api/internal/domain/services"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// CreateTransactionUploadUseCase uploads and ingests transaction files.
type CreateTransactionUploadUseCase struct {
	uploadRepository   outboundports.TransactionUploadRepository
	previewRepository  outboundports.TransactionUploadPreviewRepository
	transactionRepo    outboundports.TransactionRepository
	processingQueue    outboundports.TransactionProcessingQueue
	rawFileStore       outboundports.RawFileStore
	fileParser         outboundports.TransactionFileParser
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	actorDirectory     outboundports.ActorDirectory
	validator          *domainservices.TransactionFileValidator
	uploadFactory      *domainservices.TransactionUploadFactory
	now                func() time.Time
	newUploadID        func() (valueobjects.UploadID, error)
	newTransactionID   func() (valueobjects.TransactionID, error)
}

const defaultTransactionUploadClassificationTask = outboundports.TransactionClassifyReactTaskName

// NewCreateTransactionUploadUseCase builds a CreateTransactionUploadUseCase.
func NewCreateTransactionUploadUseCase(
	uploadRepository outboundports.TransactionUploadRepository,
	previewRepository outboundports.TransactionUploadPreviewRepository,
	transactionRepo outboundports.TransactionRepository,
	processingQueue outboundports.TransactionProcessingQueue,
	rawFileStore outboundports.RawFileStore,
	fileParser outboundports.TransactionFileParser,
	transactionManager outboundports.TransactionManager,
	eventRecorder adminEventRecorder,
	actorDirectory outboundports.ActorDirectory,
	validator *domainservices.TransactionFileValidator,
) *CreateTransactionUploadUseCase {
	return &CreateTransactionUploadUseCase{
		uploadRepository:   uploadRepository,
		previewRepository:  previewRepository,
		transactionRepo:    transactionRepo,
		processingQueue:    processingQueue,
		rawFileStore:       rawFileStore,
		fileParser:         fileParser,
		transactionManager: transactionManager,
		eventRecorder:      eventRecorder,
		actorDirectory:     actorDirectory,
		validator:          validator,
		uploadFactory:      domainservices.NewTransactionUploadFactory(),
		now:                time.Now,
		newUploadID:        valueobjects.NewUploadID,
		newTransactionID:   valueobjects.NewTransactionID,
	}
}

// Execute validates, stores, records, and persists a transaction upload.
func (uc *CreateTransactionUploadUseCase) Execute(ctx context.Context, command inboundports.CreateTransactionUploadCommand) (inboundports.CreateTransactionUploadResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, command.ActorUserID, command.ActorGroupID); err != nil {
		return inboundports.CreateTransactionUploadResult{}, err
	}

	createdBy, err := valueobjects.UserIDFromString(command.ActorUserID)
	if err != nil {
		return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("parsing actor user id: %w", err)
	}

	actorGroupID, err := valueobjects.GroupIDFromString(command.ActorGroupID)
	if err != nil {
		return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("parsing actor group id: %w", err)
	}

	parsedFile, err := uc.fileParser.Parse(ctx, outboundports.ParseTransactionFileCommand{FileName: command.FileName, FileBytes: command.FileBytes})
	if err != nil {
		return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("parsing transaction file: %w", err)
	}
	reportTransactionUploadProgress(ctx, command.ProgressReporter, outboundports.TransactionUploadProgressStatusParsed, "file parsed", 25, nil, nil, nil)

	filteredRows, err := filterUploadRowsByTransactionCount(parsedFile.Headers, parsedFile.Rows)
	if err != nil {
		return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("filtering transaction upload rows: %w", err)
	}

	uploadID, err := uc.newUploadID()
	if err != nil {
		return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("generating upload id: %w", err)
	}

	contentMD5 := fmt.Sprintf("%x", md5.Sum(command.FileBytes))
	existingUpload, err := uc.uploadRepository.FindByContentMD5(ctx, contentMD5, actorGroupID)
	if err != nil {
		return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("checking duplicate upload: %w", err)
	}

	if existingUpload != nil {
		return inboundports.CreateTransactionUploadResult{}, &ConflictError{Resource: "transaction upload", Reason: domain.ErrDuplicateUpload.Message}
	}

	storedFile, err := uc.rawFileStore.Store(ctx, outboundports.StoreRawFileCommand{UploadID: uploadID.String(), FileName: command.FileName, FileBytes: command.FileBytes})
	if err != nil {
		return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("storing raw upload file: %w", err)
	}
	reportTransactionUploadProgress(ctx, command.ProgressReporter, outboundports.TransactionUploadProgressStatusStoredRawFile, "raw file stored", 60, nil, nil, nil)

	validationReport := uc.validator.Validate(parsedFile.Headers, filteredRows.EligibleRows)
	persistFailedUpload := func(validationErrors []outboundports.TransactionFileValidationError) (inboundports.CreateTransactionUploadResult, error) {
		upload, err := uc.newFailedTransactionUpload(uploadID, actorGroupID, command, parsedFile.Format, contentMD5, storedFile, validationReport.SchemaVersion())
		if err != nil {
			deleteErr := uc.rawFileStore.Delete(ctx, outboundports.DeleteRawFileCommand{Key: storedFile.Key})
			if deleteErr != nil {
				return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("building failed transaction upload aggregate: %w (cleanup raw file: %v)", err, deleteErr)
			}

			return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("building failed transaction upload aggregate: %w", err)
		}

		preview := newTransactionUploadPreviewRecord(uploadID, parsedFile.Headers, filteredRows.EligibleRows, validationErrors)

		if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
			if err := uc.uploadRepository.Create(txCtx, upload); err != nil {
				return fmt.Errorf("creating failed transaction upload: %w", err)
			}

			if err := uc.previewRepository.Save(txCtx, preview); err != nil {
				return fmt.Errorf("saving failed transaction upload preview: %w", err)
			}

			if err := upload.RecordCreated(command.ActorUserID, command.ActorGroupID); err != nil {
				return fmt.Errorf("recording transaction upload creation event: %w", err)
			}

			if err := publishDomainEvents(txCtx, uc.eventRecorder, upload.PullDomainEvents()); err != nil {
				return fmt.Errorf("publishing transaction upload events: %w", err)
			}

			return nil
		}); err != nil {
			deleteErr := uc.rawFileStore.Delete(ctx, outboundports.DeleteRawFileCommand{Key: storedFile.Key})
			if deleteErr != nil {
				return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("persisting failed transaction upload: %w (cleanup raw file: %v)", err, deleteErr)
			}

			return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("persisting failed transaction upload: %w", err)
		}

		result := inboundports.CreateTransactionUploadResult{
			Upload:           newTransactionUploadResult(upload),
			ValidationErrors: validationErrors,
			SkippedRows:      filteredRows.SkippedRows,
		}
		reportTransactionUploadProgress(ctx, command.ProgressReporter, outboundports.TransactionUploadProgressStatusValidationFailed, "validation failed", 100, &result.Upload, result.ValidationErrors, result.SkippedRows)
		return result, nil
	}

	if len(filteredRows.EligibleRows) == 0 {
		return persistFailedUpload(toTransactionFileValidationErrors(validationReport.Errors()))
	}

	if !validationReport.Valid() {
		return persistFailedUpload(toTransactionFileValidationErrors(validationReport.Errors()))
	}
	reportTransactionUploadProgress(ctx, command.ProgressReporter, outboundports.TransactionUploadProgressStatusValidated, "file validated", 68, nil, nil, nil)

	upload, transactions, err := uc.uploadFactory.Build(
		uploadID,
		actorGroupID,
		command.FileName,
		parsedFile.Format,
		contentMD5,
		storedFile.Provider,
		storedFile.Key,
		validationReport.SchemaVersion(),
		parsedFile.Headers,
		filteredRows.EligibleRows,
		uc.now(),
		uc.newTransactionID,
	)
	if err != nil {
		deleteErr := uc.rawFileStore.Delete(ctx, outboundports.DeleteRawFileCommand{Key: storedFile.Key})
		if deleteErr != nil {
			return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("building transaction upload aggregate: %w (cleanup raw file: %v)", err, deleteErr)
		}

		return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("building transaction upload aggregate: %w", err)
	}

	for _, transaction := range transactions {
		transaction.SetCreatedBy(createdBy)
	}

	preview := newTransactionUploadPreviewRecord(uploadID, parsedFile.Headers, filteredRows.EligibleRows, nil)

	transactionalUpdates := make([]outboundports.TransactionUploadProgressUpdate, 0, 3)

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.uploadRepository.Create(txCtx, upload); err != nil {
			return fmt.Errorf("creating transaction upload: %w", err)
		}
		transactionalUpdates = append(transactionalUpdates, newTransactionUploadProgressUpdate(outboundports.TransactionUploadProgressStatusSavedUpload, "upload record saved", 72, nil, nil, nil))

		if err := uc.previewRepository.Save(txCtx, preview); err != nil {
			return fmt.Errorf("saving transaction upload preview: %w", err)
		}

		if err := upload.RecordCreated(command.ActorUserID, command.ActorGroupID); err != nil {
			return fmt.Errorf("recording transaction upload creation event: %w", err)
		}
		transactionalUpdates = append(transactionalUpdates, newTransactionUploadProgressUpdate(outboundports.TransactionUploadProgressStatusRecordedEvent, "upload event recorded", 84, nil, nil, nil))

		if err := uc.transactionRepo.CreateMany(txCtx, transactions, command.ActorUserID); err != nil {
			return fmt.Errorf("creating transaction rows: %w", err)
		}

		for _, transaction := range transactions {
			if err := uc.processingQueue.Enqueue(txCtx, uploadTransactionClassificationTask(command.ClassificationTask), transaction.ID()); err != nil {
				return fmt.Errorf("queueing transaction %s for processing: %w", transaction.ID().String(), err)
			}

			if err := transaction.MarkProcessing(uc.now()); err != nil {
				return fmt.Errorf("marking transaction %s processing: %w", transaction.ID().String(), err)
			}

			if err := uc.transactionRepo.Update(txCtx, transaction); err != nil {
				return fmt.Errorf("updating transaction %s status: %w", transaction.ID().String(), err)
			}
		}
		transactionalUpdates = append(transactionalUpdates, newTransactionUploadProgressUpdate(outboundports.TransactionUploadProgressStatusPersistedTransactions, "transactions persisted", 95, nil, nil, nil))

		if err := publishDomainEvents(txCtx, uc.eventRecorder, upload.PullDomainEvents()); err != nil {
			return fmt.Errorf("publishing transaction upload events: %w", err)
		}

		return nil
	}); err != nil {
		deleteErr := uc.rawFileStore.Delete(ctx, outboundports.DeleteRawFileCommand{Key: storedFile.Key})
		if deleteErr != nil {
			return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("creating transaction upload transaction: %w (cleanup raw file: %v)", err, deleteErr)
		}

		return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("creating transaction upload transaction: %w", err)
	}

	for _, update := range transactionalUpdates {
		reportTransactionUploadProgressUpdate(ctx, command.ProgressReporter, update)
	}

	result := inboundports.CreateTransactionUploadResult{Upload: newTransactionUploadResult(upload), SkippedRows: filteredRows.SkippedRows}
	reportTransactionUploadProgress(ctx, command.ProgressReporter, outboundports.TransactionUploadProgressStatusCompleted, "upload completed", 100, &result.Upload, nil, result.SkippedRows)
	return result, nil
}

func (uc *CreateTransactionUploadUseCase) newFailedTransactionUpload(uploadID valueobjects.UploadID, groupID valueobjects.GroupID, command inboundports.CreateTransactionUploadCommand, fileFormat, contentMD5 string, storedFile outboundports.StoreRawFileResult, schemaVersion string) (*entities.TransactionUpload, error) {
	return entities.NewTransactionUpload(
		uploadID,
		groupID,
		command.FileName,
		fileFormat,
		contentMD5,
		storedFile.Provider,
		storedFile.Key,
		schemaVersion,
		valueobjects.FailedTransactionUploadStatus(),
		0,
		uc.now(),
	)
}

func newTransactionUploadPreviewRecord(uploadID valueobjects.UploadID, columns []string, rows [][]string, validationErrors []outboundports.TransactionFileValidationError) outboundports.TransactionUploadPreviewRecord {
	return outboundports.TransactionUploadPreviewRecord{
		UploadID:         uploadID.String(),
		Columns:          append([]string(nil), columns...),
		Rows:             cloneTransactionUploadPreviewRows(rows),
		TotalRows:        len(rows),
		ValidationErrors: append([]outboundports.TransactionFileValidationError(nil), validationErrors...),
	}
}

func cloneTransactionUploadPreviewRows(rows [][]string) [][]string {
	if len(rows) == 0 {
		return nil
	}

	cloned := make([][]string, len(rows))
	for i, row := range rows {
		cloned[i] = append([]string(nil), row...)
	}

	return cloned
}

func uploadTransactionClassificationTask(taskName string) string {
	// Legacy embedding and keyword classification is temporarily disabled.
	// Uploads still accept the old task name so callers do not break, but all
	// work is normalized onto the ReAct worker.
	if taskName == outboundports.TransactionClassifyTaskName || taskName == outboundports.TransactionClassifyReactTaskName {
		return outboundports.TransactionClassifyReactTaskName
	}

	return defaultTransactionUploadClassificationTask
}

func newTransactionUploadProgressUpdate(status, message string, progress int, upload *outboundports.TransactionUploadResult, validationErrors []outboundports.TransactionFileValidationError, skippedRows []outboundports.TransactionUploadSkippedRow) outboundports.TransactionUploadProgressUpdate {
	return outboundports.TransactionUploadProgressUpdate{
		Status:           status,
		Message:          message,
		Progress:         progress,
		Upload:           upload,
		ValidationErrors: validationErrors,
		SkippedRows:      skippedRows,
	}
}

func reportTransactionUploadProgress(ctx context.Context, reporter outboundports.TransactionUploadProgressReporter, status, message string, progress int, upload *outboundports.TransactionUploadResult, validationErrors []outboundports.TransactionFileValidationError, skippedRows []outboundports.TransactionUploadSkippedRow) {
	reportTransactionUploadProgressUpdate(ctx, reporter, newTransactionUploadProgressUpdate(status, message, progress, upload, validationErrors, skippedRows))
}

func reportTransactionUploadProgressUpdate(ctx context.Context, reporter outboundports.TransactionUploadProgressReporter, update outboundports.TransactionUploadProgressUpdate) {
	if isNilTransactionUploadProgressReporter(reporter) {
		return
	}

	_ = reporter.Report(ctx, update)
}

func isNilTransactionUploadProgressReporter(reporter outboundports.TransactionUploadProgressReporter) bool {
	if reporter == nil {
		return true
	}

	value := reflect.ValueOf(reporter)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
