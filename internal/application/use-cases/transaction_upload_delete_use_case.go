package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// DeleteTransactionUploadUseCase deletes an upload and its persisted rows.
type DeleteTransactionUploadUseCase struct {
	uploadRepository   outboundports.TransactionUploadRepository
	transactionRepo    outboundports.TransactionRepository
	rawFileStore       outboundports.RawFileStore
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	actorDirectory     outboundports.ActorDirectory
	now                func() time.Time
}

// NewDeleteTransactionUploadUseCase builds a DeleteTransactionUploadUseCase.
func NewDeleteTransactionUploadUseCase(
	uploadRepository outboundports.TransactionUploadRepository,
	transactionRepo outboundports.TransactionRepository,
	rawFileStore outboundports.RawFileStore,
	transactionManager outboundports.TransactionManager,
	eventRecorder adminEventRecorder,
	actorDirectory outboundports.ActorDirectory,
) *DeleteTransactionUploadUseCase {
	return &DeleteTransactionUploadUseCase{
		uploadRepository:   uploadRepository,
		transactionRepo:    transactionRepo,
		rawFileStore:       rawFileStore,
		transactionManager: transactionManager,
		eventRecorder:      eventRecorder,
		actorDirectory:     actorDirectory,
		now:                time.Now,
	}
}

// Execute deletes an upload and all associated transactions.
func (uc *DeleteTransactionUploadUseCase) Execute(ctx context.Context, command inboundports.DeleteTransactionUploadCommand) (inboundports.DeleteTransactionUploadResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, command.ActorUserID, command.ActorGroupID); err != nil {
		return inboundports.DeleteTransactionUploadResult{}, err
	}

	uploadID, err := valueobjects.UploadIDFromString(command.ID)
	if err != nil {
		return inboundports.DeleteTransactionUploadResult{}, fmt.Errorf("parsing upload id: %w", err)
	}

	upload, err := uc.uploadRepository.FindByID(ctx, uploadID)
	if err != nil {
		return inboundports.DeleteTransactionUploadResult{}, fmt.Errorf("finding upload by id: %w", err)
	}

	if upload == nil {
		return inboundports.DeleteTransactionUploadResult{}, &NotFoundError{Resource: "transaction upload", ID: command.ID}
	}

	if upload.GroupID().String() != command.ActorGroupID {
		return inboundports.DeleteTransactionUploadResult{}, &ForbiddenError{Resource: "transaction upload", Reason: "actor group does not have access to this upload"}
	}

	hasProcessing, err := uc.transactionRepo.HasProcessingByUploadID(ctx, uploadID)
	if err != nil {
		return inboundports.DeleteTransactionUploadResult{}, fmt.Errorf("checking processing transactions for upload: %w", err)
	}

	if hasProcessing {
		return inboundports.DeleteTransactionUploadResult{}, &ConflictError{
			Resource: "transaction upload",
			Reason:   "cannot delete transaction upload while transactions are still processing",
		}
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.transactionRepo.DeleteByUploadID(txCtx, uploadID); err != nil {
			return fmt.Errorf("deleting transaction rows: %w", err)
		}

		if err := uc.uploadRepository.DeleteByID(txCtx, uploadID); err != nil {
			return fmt.Errorf("deleting transaction upload: %w", err)
		}

		if err := upload.RecordDeleted(uc.now(), command.ActorUserID, command.ActorGroupID); err != nil {
			return fmt.Errorf("recording transaction upload deletion event: %w", err)
		}

		if err := publishDomainEvents(txCtx, uc.eventRecorder, upload.PullDomainEvents()); err != nil {
			return fmt.Errorf("publishing transaction upload events: %w", err)
		}

		return nil
	}); err != nil {
		return inboundports.DeleteTransactionUploadResult{}, fmt.Errorf("deleting transaction upload transaction: %w", err)
	}

	if err := uc.rawFileStore.Delete(ctx, outboundports.DeleteRawFileCommand{Key: upload.StorageKey()}); err != nil {
		return inboundports.DeleteTransactionUploadResult{}, fmt.Errorf("deleting raw upload file: %w", err)
	}

	return inboundports.DeleteTransactionUploadResult{ID: command.ID}, nil
}
