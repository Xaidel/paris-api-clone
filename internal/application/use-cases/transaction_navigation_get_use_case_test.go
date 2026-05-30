package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/gyud-adb/paris-api/internal/domain"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

func TestGetTransactionNavigationUseCaseExecuteReturnsNeighborsWithoutFilters(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	previousID := "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f"
	nextID := "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"

	repository := &transactionRepositoryMock{navigationResult: &outboundports.TransactionNavigationResult{TransactionID: transactionID.String(), PreviousID: &previousID, NextID: &nextID}}

	result, err := NewGetTransactionNavigationUseCase(repository).Execute(context.Background(), inboundports.GetTransactionNavigationQuery{ID: transactionID.String()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.TransactionID != transactionID.String() {
		t.Fatalf("result.TransactionID = %q, want %q", result.TransactionID, transactionID.String())
	}

	if result.PreviousID == nil || *result.PreviousID != previousID {
		t.Fatalf("result.PreviousID = %v, want %q", result.PreviousID, previousID)
	}

	if result.NextID == nil || *result.NextID != nextID {
		t.Fatalf("result.NextID = %v, want %q", result.NextID, nextID)
	}

	if repository.navigationLookup.TransactionID.String() != transactionID.String() {
		t.Fatalf("repository.navigationLookup.TransactionID = %q, want %q", repository.navigationLookup.TransactionID.String(), transactionID.String())
	}

	if repository.navigationLookup.Filter.SortBy != outboundports.TransactionSortByCreatedAt {
		t.Fatalf("repository.navigationLookup.Filter.SortBy = %q, want %q", repository.navigationLookup.Filter.SortBy, outboundports.TransactionSortByCreatedAt)
	}

	if repository.navigationLookup.Filter.SortOrder != outboundports.TransactionSortOrderDescending {
		t.Fatalf("repository.navigationLookup.Filter.SortOrder = %q, want %q", repository.navigationLookup.Filter.SortOrder, outboundports.TransactionSortOrderDescending)
	}

	if repository.navigationLookup.Classification != nil {
		t.Fatalf("repository.navigationLookup.Classification = %v, want nil", repository.navigationLookup.Classification)
	}

	if repository.navigationLookup.Step != nil {
		t.Fatalf("repository.navigationLookup.Step = %v, want nil", repository.navigationLookup.Step)
	}

	if repository.navigationLookup.Filter.UploadID != nil {
		t.Fatalf("repository.navigationLookup.Filter.UploadID = %v, want nil", repository.navigationLookup.Filter.UploadID)
	}

	if repository.navigationLookup.Filter.Status != nil {
		t.Fatalf("repository.navigationLookup.Filter.Status = %v, want nil", repository.navigationLookup.Filter.Status)
	}

	if repository.navigationLookup.Filter.TransactionCountMin != nil {
		t.Fatalf("repository.navigationLookup.Filter.TransactionCountMin = %v, want nil", repository.navigationLookup.Filter.TransactionCountMin)
	}

	if repository.navigationLookup.Filter.TransactionCountMax != nil {
		t.Fatalf("repository.navigationLookup.Filter.TransactionCountMax = %v, want nil", repository.navigationLookup.Filter.TransactionCountMax)
	}

	if repository.navigationLookup.Filter.CreatedAtFrom != nil {
		t.Fatalf("repository.navigationLookup.Filter.CreatedAtFrom = %v, want nil", repository.navigationLookup.Filter.CreatedAtFrom)
	}

	if repository.navigationLookup.Filter.CreatedAtTo != nil {
		t.Fatalf("repository.navigationLookup.Filter.CreatedAtTo = %v, want nil", repository.navigationLookup.Filter.CreatedAtTo)
	}

	if repository.navigationLookup.Filter.ApplicantCountry != nil {
		t.Fatalf("repository.navigationLookup.Filter.ApplicantCountry = %v, want nil", repository.navigationLookup.Filter.ApplicantCountry)
	}

	if repository.navigationLookup.Filter.BeneficiaryCountry != nil {
		t.Fatalf("repository.navigationLookup.Filter.BeneficiaryCountry = %v, want nil", repository.navigationLookup.Filter.BeneficiaryCountry)
	}

	if repository.navigationLookup.Filter.SourceCountry != nil {
		t.Fatalf("repository.navigationLookup.Filter.SourceCountry = %v, want nil", repository.navigationLookup.Filter.SourceCountry)
	}

	if repository.navigationLookup.Filter.DestinationCountry != nil {
		t.Fatalf("repository.navigationLookup.Filter.DestinationCountry = %v, want nil", repository.navigationLookup.Filter.DestinationCountry)
	}

	if repository.navigationLookup.Filter.Classification != nil {
		t.Fatalf("repository.navigationLookup.Filter.Classification = %v, want nil", repository.navigationLookup.Filter.Classification)
	}
}

func TestGetTransactionNavigationUseCaseExecuteNormalizesClassificationAndStep(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	repository := &transactionRepositoryMock{navigationResult: &outboundports.TransactionNavigationResult{TransactionID: transactionID.String()}}

	_, err = NewGetTransactionNavigationUseCase(repository).Execute(context.Background(), inboundports.GetTransactionNavigationQuery{ID: transactionID.String(), Classification: "needs review", Step: "4"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if repository.navigationLookup.Classification == nil || *repository.navigationLookup.Classification != "next_step" {
		t.Fatalf("repository.navigationLookup.Classification = %v, want %q", repository.navigationLookup.Classification, "next_step")
	}

	if repository.navigationLookup.Filter.Classification == nil || *repository.navigationLookup.Filter.Classification != "next_step" {
		t.Fatalf("repository.navigationLookup.Filter.Classification = %v, want %q", repository.navigationLookup.Filter.Classification, "next_step")
	}

	if repository.navigationLookup.Step == nil || *repository.navigationLookup.Step != 4 {
		t.Fatalf("repository.navigationLookup.Step = %v, want %d", repository.navigationLookup.Step, 4)
	}
}

func TestGetTransactionNavigationUseCaseExecuteMirrorsCanonicalClassificationIntoLookupAndFilter(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	repository := &transactionRepositoryMock{navigationResult: &outboundports.TransactionNavigationResult{TransactionID: transactionID.String()}}

	_, err = NewGetTransactionNavigationUseCase(repository).Execute(context.Background(), inboundports.GetTransactionNavigationQuery{ID: transactionID.String(), Classification: "next-step"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if repository.navigationLookup.Classification == nil || *repository.navigationLookup.Classification != "next_step" {
		t.Fatalf("repository.navigationLookup.Classification = %v, want %q", repository.navigationLookup.Classification, "next_step")
	}

	if repository.navigationLookup.Filter.Classification == nil || *repository.navigationLookup.Filter.Classification != "next_step" {
		t.Fatalf("repository.navigationLookup.Filter.Classification = %v, want %q", repository.navigationLookup.Filter.Classification, "next_step")
	}
}

func TestGetTransactionNavigationUseCaseExecuteReturnsNotFoundWhenRepositoryReturnsNil(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	_, err = NewGetTransactionNavigationUseCase(&transactionRepositoryMock{}).Execute(context.Background(), inboundports.GetTransactionNavigationQuery{ID: transactionID.String()})
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("Execute() error = %v, want NotFoundError", err)
	}

	if notFoundErr.Resource != "transaction" {
		t.Fatalf("notFoundErr.Resource = %q, want %q", notFoundErr.Resource, "transaction")
	}

	if notFoundErr.ID != transactionID.String() {
		t.Fatalf("notFoundErr.ID = %q, want %q", notFoundErr.ID, transactionID.String())
	}
}

func TestGetTransactionNavigationUseCaseExecuteRejectsInvalidStep(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	_, err = NewGetTransactionNavigationUseCase(&transactionRepositoryMock{}).Execute(context.Background(), inboundports.GetTransactionNavigationQuery{ID: transactionID.String(), Step: "0"})
	var validationErr *domain.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("Execute() error = %v, want ValidationError", err)
	}

	fields := validationErr.Fields()
	if len(fields) != 1 {
		t.Fatalf("len(validationErr.Fields()) = %d, want %d", len(fields), 1)
	}

	if fields[0].Field() != "step" {
		t.Fatalf("validationErr.Fields()[0].Field() = %q, want %q", fields[0].Field(), "step")
	}
}

func TestGetTransactionNavigationUseCaseExecuteTreatsEmptyClassificationAndStepAsNoFilter(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	repository := &transactionRepositoryMock{navigationResult: &outboundports.TransactionNavigationResult{TransactionID: transactionID.String()}}

	_, err = NewGetTransactionNavigationUseCase(repository).Execute(context.Background(), inboundports.GetTransactionNavigationQuery{
		ID:                  transactionID.String(),
		UploadID:            uploadID.String(),
		CreatedAtFrom:       "2026-04-01",
		CreatedAtTo:         "2026-04-30",
		ApplicantCountry:    "Philippines",
		BeneficiaryCountry:  "Japan",
		SourceCountry:       "Thailand",
		DestinationCountry:  "Philippines",
		TransactionCountMin: "1",
		TransactionCountMax: "10",
		Status:              "processing",
		Classification:      "   ",
		Step:                "   ",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if repository.navigationLookup.Classification != nil {
		t.Fatalf("repository.navigationLookup.Classification = %v, want nil", repository.navigationLookup.Classification)
	}

	if repository.navigationLookup.Step != nil {
		t.Fatalf("repository.navigationLookup.Step = %v, want nil", repository.navigationLookup.Step)
	}

	if repository.navigationLookup.Filter.UploadID == nil || repository.navigationLookup.Filter.UploadID.String() != uploadID.String() {
		t.Fatalf("repository.navigationLookup.Filter.UploadID = %v, want %q", repository.navigationLookup.Filter.UploadID, uploadID.String())
	}

	if repository.navigationLookup.Filter.CreatedAtFrom == nil || repository.navigationLookup.Filter.CreatedAtFrom.Format("2006-01-02") != "2026-04-01" {
		t.Fatalf("repository.navigationLookup.Filter.CreatedAtFrom = %v, want %q", repository.navigationLookup.Filter.CreatedAtFrom, "2026-04-01")
	}

	if repository.navigationLookup.Filter.CreatedAtTo == nil || repository.navigationLookup.Filter.CreatedAtTo.Format("2006-01-02") != "2026-04-30" {
		t.Fatalf("repository.navigationLookup.Filter.CreatedAtTo = %v, want %q", repository.navigationLookup.Filter.CreatedAtTo, "2026-04-30")
	}

	if repository.navigationLookup.Filter.ApplicantCountry == nil || *repository.navigationLookup.Filter.ApplicantCountry != "Philippines" {
		t.Fatalf("repository.navigationLookup.Filter.ApplicantCountry = %v, want %q", repository.navigationLookup.Filter.ApplicantCountry, "Philippines")
	}

	if repository.navigationLookup.Filter.BeneficiaryCountry == nil || *repository.navigationLookup.Filter.BeneficiaryCountry != "Japan" {
		t.Fatalf("repository.navigationLookup.Filter.BeneficiaryCountry = %v, want %q", repository.navigationLookup.Filter.BeneficiaryCountry, "Japan")
	}

	if repository.navigationLookup.Filter.SourceCountry == nil || *repository.navigationLookup.Filter.SourceCountry != "Thailand" {
		t.Fatalf("repository.navigationLookup.Filter.SourceCountry = %v, want %q", repository.navigationLookup.Filter.SourceCountry, "Thailand")
	}

	if repository.navigationLookup.Filter.DestinationCountry == nil || *repository.navigationLookup.Filter.DestinationCountry != "Philippines" {
		t.Fatalf("repository.navigationLookup.Filter.DestinationCountry = %v, want %q", repository.navigationLookup.Filter.DestinationCountry, "Philippines")
	}

	if repository.navigationLookup.Filter.TransactionCountMin == nil || *repository.navigationLookup.Filter.TransactionCountMin != 1 {
		t.Fatalf("repository.navigationLookup.Filter.TransactionCountMin = %v, want %d", repository.navigationLookup.Filter.TransactionCountMin, 1)
	}

	if repository.navigationLookup.Filter.TransactionCountMax == nil || *repository.navigationLookup.Filter.TransactionCountMax != 10 {
		t.Fatalf("repository.navigationLookup.Filter.TransactionCountMax = %v, want %d", repository.navigationLookup.Filter.TransactionCountMax, 10)
	}

	if repository.navigationLookup.Filter.Status == nil || *repository.navigationLookup.Filter.Status != "processing" {
		t.Fatalf("repository.navigationLookup.Filter.Status = %v, want %q", repository.navigationLookup.Filter.Status, "processing")
	}
}
