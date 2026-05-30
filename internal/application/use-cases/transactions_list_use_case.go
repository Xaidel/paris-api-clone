package usecases

import (
	"context"
	"fmt"

	services "github.com/gyud-adb/paris-api/internal/application/services"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// ListTransactionsUseCase lists transaction records.
type ListTransactionsUseCase struct {
	repository           outboundports.TransactionRepository
	transactionStep4Repo outboundports.TransactionStep4Repository
	transactionStep5Repo outboundports.TransactionStep5Repository
	sectorRepository     outboundports.SectorRepository
	auditService         *services.TransactionAuditService
}

// NewListTransactionsUseCase builds a ListTransactionsUseCase.
func NewListTransactionsUseCase(repository outboundports.TransactionRepository, transactionStep4Repo outboundports.TransactionStep4Repository, transactionStep5Repo outboundports.TransactionStep5Repository, sectorRepository outboundports.SectorRepository, auditService *services.TransactionAuditService) *ListTransactionsUseCase {
	return &ListTransactionsUseCase{repository: repository, transactionStep4Repo: transactionStep4Repo, transactionStep5Repo: transactionStep5Repo, sectorRepository: sectorRepository, auditService: auditService}
}

// Execute lists transaction records.
func (uc *ListTransactionsUseCase) Execute(ctx context.Context, query inboundports.ListTransactionsQuery) (inboundports.ListTransactionsResult, error) {
	filter, err := buildTransactionFilter(query)
	if err != nil {
		return inboundports.ListTransactionsResult{}, err
	}

	transactions, err := uc.repository.List(ctx, filter)
	if err != nil {
		return inboundports.ListTransactionsResult{}, fmt.Errorf("listing transactions: %w", err)
	}

	results := make([]outboundports.TransactionResult, 0, len(transactions))
	for _, transaction := range transactions {
		step4, sector, loadErr := loadTransactionStep4Details(ctx, uc.transactionStep4Repo, uc.sectorRepository, transaction.ID())
		if loadErr != nil {
			return inboundports.ListTransactionsResult{}, loadErr
		}

		step5, loadErr := loadTransactionStep5Details(ctx, uc.transactionStep5Repo, transaction.ID())
		if loadErr != nil {
			return inboundports.ListTransactionsResult{}, loadErr
		}

		results = append(results, newTransactionResult(transaction, step4, sector, step5))
	}

	if err := uc.auditService.RecordListTransactions(ctx, query.ActorUserID, query.ActorGroupID, query.UploadID, len(results)); err != nil {
		return inboundports.ListTransactionsResult{}, err
	}

	return inboundports.ListTransactionsResult{Transactions: results}, nil
}
