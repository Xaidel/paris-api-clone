package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// GetTransactionUploadUseCase gets an upload and its accepted transactions by identifier.
type GetTransactionUploadUseCase struct {
	uploadRepository      outboundports.TransactionUploadRepository
	transactionRepository outboundports.TransactionRepository
	transactionStep4Repo  outboundports.TransactionStep4Repository
	transactionStep5Repo  outboundports.TransactionStep5Repository
	sectorRepository      outboundports.SectorRepository
	eventRecorder         adminEventRecorder
	actorDirectory        outboundports.ActorDirectory
	now                   func() time.Time
}

// NewGetTransactionUploadUseCase builds a GetTransactionUploadUseCase.
func NewGetTransactionUploadUseCase(uploadRepository outboundports.TransactionUploadRepository, transactionRepository outboundports.TransactionRepository, transactionStep4Repo outboundports.TransactionStep4Repository, transactionStep5Repo outboundports.TransactionStep5Repository, sectorRepository outboundports.SectorRepository, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *GetTransactionUploadUseCase {
	return &GetTransactionUploadUseCase{uploadRepository: uploadRepository, transactionRepository: transactionRepository, transactionStep4Repo: transactionStep4Repo, transactionStep5Repo: transactionStep5Repo, sectorRepository: sectorRepository, eventRecorder: eventRecorder, actorDirectory: actorDirectory, now: time.Now}
}

// Execute loads an upload and its accepted transactions by identifier.
func (uc *GetTransactionUploadUseCase) Execute(ctx context.Context, query inboundports.GetTransactionUploadQuery) (outboundports.TransactionUploadDetailsResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, query.ActorUserID, query.ActorGroupID); err != nil {
		return outboundports.TransactionUploadDetailsResult{}, err
	}

	uploadID, err := valueobjects.UploadIDFromString(query.ID)
	if err != nil {
		return outboundports.TransactionUploadDetailsResult{}, fmt.Errorf("parsing upload id: %w", err)
	}

	upload, err := uc.uploadRepository.FindByID(ctx, uploadID)
	if err != nil {
		return outboundports.TransactionUploadDetailsResult{}, fmt.Errorf("finding upload by id: %w", err)
	}

	if upload == nil {
		return outboundports.TransactionUploadDetailsResult{}, &NotFoundError{Resource: "transaction upload", ID: query.ID}
	}

	if upload.GroupID().String() != query.ActorGroupID {
		return outboundports.TransactionUploadDetailsResult{}, &ForbiddenError{Resource: "transaction upload", Reason: "actor group does not have access to this upload"}
	}

	transactions, err := uc.transactionRepository.ListByUploadIDs(ctx, []valueobjects.UploadID{uploadID})
	if err != nil {
		return outboundports.TransactionUploadDetailsResult{}, fmt.Errorf("listing transactions for upload: %w", err)
	}

	if err := upload.RecordRead(uc.now(), query.ActorUserID, query.ActorGroupID); err != nil {
		return outboundports.TransactionUploadDetailsResult{}, fmt.Errorf("recording transaction upload read event: %w", err)
	}

	if err := publishDomainEvents(ctx, uc.eventRecorder, upload.PullDomainEvents()); err != nil {
		return outboundports.TransactionUploadDetailsResult{}, fmt.Errorf("publishing transaction upload events: %w", err)
	}

	results := make([]outboundports.TransactionResult, 0, len(transactions))
	for _, transaction := range transactions {
		step4, sector, loadErr := loadTransactionStep4Details(ctx, uc.transactionStep4Repo, uc.sectorRepository, transaction.ID())
		if loadErr != nil {
			return outboundports.TransactionUploadDetailsResult{}, loadErr
		}

		step5, loadErr := loadTransactionStep5Details(ctx, uc.transactionStep5Repo, transaction.ID())
		if loadErr != nil {
			return outboundports.TransactionUploadDetailsResult{}, loadErr
		}

		results = append(results, newTransactionResult(transaction, step4, sector, step5))
	}

	return outboundports.TransactionUploadDetailsResult{
		TransactionUploadResult: newTransactionUploadResult(upload),
		Transactions:            results,
	}, nil
}
