package usecases

import (
	"context"
	"fmt"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// GetTransactionNavigationUseCase gets the previous and next transaction IDs.
type GetTransactionNavigationUseCase struct {
	repository outboundports.TransactionRepository
}

// NewGetTransactionNavigationUseCase builds a GetTransactionNavigationUseCase.
func NewGetTransactionNavigationUseCase(repository outboundports.TransactionRepository) *GetTransactionNavigationUseCase {
	return &GetTransactionNavigationUseCase{repository: repository}
}

// Execute loads the previous and next transaction IDs in the filtered scope.
func (uc *GetTransactionNavigationUseCase) Execute(ctx context.Context, query inboundports.GetTransactionNavigationQuery) (inboundports.GetTransactionNavigationResult, error) {
	transactionID, err := valueobjects.TransactionIDFromString(query.ID)
	if err != nil {
		return inboundports.GetTransactionNavigationResult{}, fmt.Errorf("parsing transaction id: %w", err)
	}

	filter, err := buildTransactionFilter(inboundports.ListTransactionsQuery{
		UploadID:            query.UploadID,
		CreatedAtFrom:       query.CreatedAtFrom,
		CreatedAtTo:         query.CreatedAtTo,
		ApplicantCountry:    query.ApplicantCountry,
		BeneficiaryCountry:  query.BeneficiaryCountry,
		SourceCountry:       query.SourceCountry,
		DestinationCountry:  query.DestinationCountry,
		TransactionCountMin: query.TransactionCountMin,
		TransactionCountMax: query.TransactionCountMax,
		Status:              query.Status,
		ActorUserID:         query.ActorUserID,
		ActorGroupID:        query.ActorGroupID,
	})
	if err != nil {
		return inboundports.GetTransactionNavigationResult{}, err
	}

	classification, err := normalizeNavigationClassification(query.Classification)
	if err != nil {
		return inboundports.GetTransactionNavigationResult{}, err
	}
	filter.Classification = classification

	step, err := normalizeNavigationStep(query.Step)
	if err != nil {
		return inboundports.GetTransactionNavigationResult{}, err
	}

	navigation, err := uc.repository.GetNavigation(ctx, outboundports.TransactionNavigationLookup{
		TransactionID:  transactionID,
		Filter:         filter,
		Classification: classification,
		Step:           step,
	})
	if err != nil {
		return inboundports.GetTransactionNavigationResult{}, fmt.Errorf("getting transaction navigation: %w", err)
	}

	if navigation == nil {
		return inboundports.GetTransactionNavigationResult{}, &NotFoundError{Resource: "transaction", ID: query.ID}
	}

	return inboundports.GetTransactionNavigationResult{
		TransactionID: navigation.TransactionID,
		PreviousID:    navigation.PreviousID,
		NextID:        navigation.NextID,
	}, nil
}
