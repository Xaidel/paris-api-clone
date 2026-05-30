package usecases

import (
	"context"
	"fmt"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// ListTransactionUploadsUseCase lists accepted transaction uploads.
type ListTransactionUploadsUseCase struct {
	uploadRepository      outboundports.TransactionUploadRepository
	transactionRepository outboundports.TransactionRepository
	transactionStep4Repo  outboundports.TransactionStep4Repository
	transactionStep5Repo  outboundports.TransactionStep5Repository
	sectorRepository      outboundports.SectorRepository
}

// NewListTransactionUploadsUseCase builds a ListTransactionUploadsUseCase.
func NewListTransactionUploadsUseCase(uploadRepository outboundports.TransactionUploadRepository, transactionRepository outboundports.TransactionRepository, transactionStep4Repo outboundports.TransactionStep4Repository, transactionStep5Repo outboundports.TransactionStep5Repository, sectorRepository outboundports.SectorRepository) *ListTransactionUploadsUseCase {
	return &ListTransactionUploadsUseCase{uploadRepository: uploadRepository, transactionRepository: transactionRepository, transactionStep4Repo: transactionStep4Repo, transactionStep5Repo: transactionStep5Repo, sectorRepository: sectorRepository}
}

// Execute lists transaction uploads.
func (uc *ListTransactionUploadsUseCase) Execute(ctx context.Context, query inboundports.ListTransactionUploadsQuery) (inboundports.ListTransactionUploadsResult, error) {
	groupID, err := valueobjects.GroupIDFromString(query.ActorGroupID)
	if err != nil {
		return inboundports.ListTransactionUploadsResult{}, fmt.Errorf("parsing actor group id: %w", err)
	}

	filter := outboundports.TransactionUploadFilter{
		GroupID:   groupID,
		FileName:  query.FileName,
		StartedAt: query.StartedAt,
		EndedAt:   query.EndedAt,
	}

	uploads, err := uc.uploadRepository.List(ctx, filter)
	if err != nil {
		return inboundports.ListTransactionUploadsResult{}, fmt.Errorf("listing transaction uploads: %w", err)
	}

	if len(uploads) == 0 {
		return inboundports.ListTransactionUploadsResult{Uploads: []outboundports.TransactionUploadDetailsResult{}}, nil
	}

	uploadIDs := make([]valueobjects.UploadID, 0, len(uploads))
	for _, upload := range uploads {
		uploadIDs = append(uploadIDs, upload.ID())
	}

	transactions, err := uc.transactionRepository.ListByUploadIDs(ctx, uploadIDs)
	if err != nil {
		return inboundports.ListTransactionUploadsResult{}, fmt.Errorf("listing transactions for uploads: %w", err)
	}

	transactionsByUploadID := make(map[string][]*entities.Transaction, len(uploads))
	for _, transaction := range transactions {
		uploadIDValue := transaction.UploadID()
		if uploadIDValue == nil {
			continue
		}

		uploadID := uploadIDValue.String()
		transactionsByUploadID[uploadID] = append(transactionsByUploadID[uploadID], transaction)
	}

	results := make([]outboundports.TransactionUploadDetailsResult, 0, len(uploads))
	for _, upload := range uploads {
		transactionsForUpload := transactionsByUploadID[upload.ID().String()]
		mappedTransactions := make([]outboundports.TransactionResult, 0, len(transactionsForUpload))
		for _, transaction := range transactionsForUpload {
			step4, sector, loadErr := loadTransactionStep4Details(ctx, uc.transactionStep4Repo, uc.sectorRepository, transaction.ID())
			if loadErr != nil {
				return inboundports.ListTransactionUploadsResult{}, loadErr
			}

			step5, loadErr := loadTransactionStep5Details(ctx, uc.transactionStep5Repo, transaction.ID())
			if loadErr != nil {
				return inboundports.ListTransactionUploadsResult{}, loadErr
			}

			mappedTransactions = append(mappedTransactions, newTransactionResult(transaction, step4, sector, step5))
		}

		results = append(results, outboundports.TransactionUploadDetailsResult{
			TransactionUploadResult: newTransactionUploadResult(upload),
			Transactions:            mappedTransactions,
		})
	}

	return inboundports.ListTransactionUploadsResult{Uploads: results}, nil
}
