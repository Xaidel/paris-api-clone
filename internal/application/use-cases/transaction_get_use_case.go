package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// GetTransactionUseCase gets a transaction by identifier.
type GetTransactionUseCase struct {
	repository           outboundports.TransactionRepository
	transactionStep4Repo outboundports.TransactionStep4Repository
	transactionStep5Repo outboundports.TransactionStep5Repository
	sectorRepository     outboundports.SectorRepository
	eventRecorder        adminEventRecorder
	now                  func() time.Time
}

// NewGetTransactionUseCase builds a GetTransactionUseCase.
func NewGetTransactionUseCase(repository outboundports.TransactionRepository, transactionStep4Repo outboundports.TransactionStep4Repository, transactionStep5Repo outboundports.TransactionStep5Repository, sectorRepository outboundports.SectorRepository, eventRecorder adminEventRecorder) *GetTransactionUseCase {
	return &GetTransactionUseCase{repository: repository, transactionStep4Repo: transactionStep4Repo, transactionStep5Repo: transactionStep5Repo, sectorRepository: sectorRepository, eventRecorder: eventRecorder, now: time.Now}
}

// Execute loads a transaction by identifier.
func (uc *GetTransactionUseCase) Execute(ctx context.Context, query inboundports.GetTransactionQuery) (outboundports.TransactionResult, error) {
	id, err := valueobjects.TransactionIDFromString(query.ID)
	if err != nil {
		return outboundports.TransactionResult{}, fmt.Errorf("parsing transaction id: %w", err)
	}

	transaction, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return outboundports.TransactionResult{}, fmt.Errorf("finding transaction by id: %w", err)
	}

	if transaction == nil {
		return outboundports.TransactionResult{}, &NotFoundError{Resource: "transaction", ID: query.ID}
	}

	if err := transaction.RecordRead(uc.now(), query.ActorUserID, query.ActorGroupID); err != nil {
		return outboundports.TransactionResult{}, fmt.Errorf("recording transaction read event: %w", err)
	}

	if err := publishDomainEvents(ctx, uc.eventRecorder, transaction.PullDomainEvents()); err != nil {
		return outboundports.TransactionResult{}, fmt.Errorf("publishing transaction events: %w", err)
	}

	step4, sector, err := loadTransactionStep4Details(ctx, uc.transactionStep4Repo, uc.sectorRepository, transaction.ID())
	if err != nil {
		return outboundports.TransactionResult{}, err
	}

	step5, err := loadTransactionStep5Details(ctx, uc.transactionStep5Repo, transaction.ID())
	if err != nil {
		return outboundports.TransactionResult{}, err
	}

	return newTransactionResult(transaction, step4, sector, step5), nil
}
