package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// DeleteTransactionUseCase deletes an existing transaction.
type DeleteTransactionUseCase struct {
	repository         outboundports.TransactionRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	now                func() time.Time
}

// NewDeleteTransactionUseCase builds a DeleteTransactionUseCase.
func NewDeleteTransactionUseCase(repository outboundports.TransactionRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder) *DeleteTransactionUseCase {
	return &DeleteTransactionUseCase{repository: repository, transactionManager: transactionManager, eventRecorder: eventRecorder, now: time.Now}
}

// Execute deletes an existing transaction.
func (uc *DeleteTransactionUseCase) Execute(ctx context.Context, command inboundports.DeleteTransactionCommand) (inboundports.DeleteTransactionResult, error) {
	id, err := valueobjects.TransactionIDFromString(command.ID)
	if err != nil {
		return inboundports.DeleteTransactionResult{}, fmt.Errorf("parsing transaction id: %w", err)
	}

	transaction, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return inboundports.DeleteTransactionResult{}, fmt.Errorf("finding transaction by id: %w", err)
	}

	if transaction == nil {
		return inboundports.DeleteTransactionResult{}, &NotFoundError{Resource: "transaction", ID: command.ID}
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.DeleteByID(txCtx, id); err != nil {
			return fmt.Errorf("deleting transaction record: %w", err)
		}

		if err := transaction.RecordDeleted(uc.now(), command.ActorUserID, command.ActorGroupID); err != nil {
			return fmt.Errorf("recording transaction deletion event: %w", err)
		}

		if err := publishDomainEvents(txCtx, uc.eventRecorder, transaction.PullDomainEvents()); err != nil {
			return fmt.Errorf("publishing transaction events: %w", err)
		}

		return nil
	}); err != nil {
		return inboundports.DeleteTransactionResult{}, fmt.Errorf("deleting transaction transaction: %w", err)
	}

	return inboundports.DeleteTransactionResult{ID: command.ID}, nil
}
