package usecases

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// DownloadTransactionUploadUseCase downloads a stored transaction upload file.
type DownloadTransactionUploadUseCase struct {
	uploadRepository outboundports.TransactionUploadRepository
	rawFileStore     outboundports.RawFileStore
	eventRecorder    adminEventRecorder
	actorDirectory   outboundports.ActorDirectory
	now              func() time.Time
}

// NewDownloadTransactionUploadUseCase builds a DownloadTransactionUploadUseCase.
func NewDownloadTransactionUploadUseCase(uploadRepository outboundports.TransactionUploadRepository, rawFileStore outboundports.RawFileStore, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *DownloadTransactionUploadUseCase {
	return &DownloadTransactionUploadUseCase{
		uploadRepository: uploadRepository,
		rawFileStore:     rawFileStore,
		eventRecorder:    eventRecorder,
		actorDirectory:   actorDirectory,
		now:              time.Now,
	}
}

// Execute validates access, reads the stored raw file, and records the download event.
func (uc *DownloadTransactionUploadUseCase) Execute(ctx context.Context, query inboundports.DownloadTransactionUploadQuery) (inboundports.DownloadTransactionUploadResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, query.ActorUserID, query.ActorGroupID); err != nil {
		return inboundports.DownloadTransactionUploadResult{}, err
	}

	uploadID, err := valueobjects.UploadIDFromString(query.ID)
	if err != nil {
		return inboundports.DownloadTransactionUploadResult{}, fmt.Errorf("parsing upload id: %w", err)
	}

	upload, err := uc.uploadRepository.FindByID(ctx, uploadID)
	if err != nil {
		return inboundports.DownloadTransactionUploadResult{}, fmt.Errorf("finding upload by id: %w", err)
	}

	if upload == nil {
		return inboundports.DownloadTransactionUploadResult{}, &NotFoundError{Resource: "transaction upload", ID: query.ID}
	}

	if upload.GroupID().String() != query.ActorGroupID {
		return inboundports.DownloadTransactionUploadResult{}, &ForbiddenError{Resource: "transaction upload", Reason: "actor group does not have access to this upload"}
	}

	file, err := uc.rawFileStore.Read(ctx, outboundports.ReadRawFileCommand{Key: upload.StorageKey()})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return inboundports.DownloadTransactionUploadResult{}, &storedFileNotFoundError{
				notFound: &NotFoundError{Resource: "transaction upload file", ID: query.ID},
				cause:    os.ErrNotExist,
			}
		}

		return inboundports.DownloadTransactionUploadResult{}, fmt.Errorf("reading raw upload file: %w", err)
	}

	if err := upload.RecordDownloaded(uc.now(), query.ActorUserID, query.ActorGroupID); err != nil {
		return inboundports.DownloadTransactionUploadResult{}, fmt.Errorf("recording transaction upload download event: %w", err)
	}

	if err := publishDomainEvents(ctx, uc.eventRecorder, upload.PullDomainEvents()); err != nil {
		return inboundports.DownloadTransactionUploadResult{}, fmt.Errorf("publishing transaction upload events: %w", err)
	}

	return inboundports.DownloadTransactionUploadResult{
		FileName:      upload.FileName(),
		ContentType:   resolveDownloadContentType(upload.FileName(), upload.FileFormat(), file.ContentType),
		ContentLength: len(file.FileBytes),
		FileBytes:     file.FileBytes,
	}, nil
}

func resolveDownloadContentType(fileName string, fileFormat string, storedContentType string) string {
	if trimmed := strings.TrimSpace(storedContentType); trimmed != "" {
		return trimmed
	}

	if inferred := contentTypeForUploadFormat(filepath.Ext(fileName)); inferred != "" {
		return inferred
	}

	if inferred := contentTypeForUploadFormat(fileFormat); inferred != "" {
		return inferred
	}

	return "application/octet-stream"
}

func contentTypeForUploadFormat(fileFormat string) string {
	normalizedFormat := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(fileFormat)), ".")
	switch normalizedFormat {
	case "csv":
		return "text/csv; charset=utf-8"
	case "xls":
		return "application/vnd.ms-excel"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return ""
	}
}

type storedFileNotFoundError struct {
	notFound *NotFoundError
	cause    error
}

func (e *storedFileNotFoundError) Error() string {
	return e.notFound.Error()
}

func (e *storedFileNotFoundError) Unwrap() []error {
	return []error{e.notFound, e.cause}
}
