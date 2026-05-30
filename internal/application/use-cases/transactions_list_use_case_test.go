package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	services "github.com/gyud-adb/paris-api/internal/application/services"
	"github.com/gyud-adb/paris-api/internal/domain"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// This test verifies list queries are parsed into repository filters, results
// are mapped back out, and the list operation records its audit event.
func TestListTransactionsUseCaseExecute(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	transaction, err := entities.NewUploadedTransaction(transactionID, uploadID, 2, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewUploadedTransaction() error = %v", err)
	}

	repository := &transactionRepositoryMock{listResult: []*entities.Transaction{transaction}}
	recorder := &adminEventRecorderMock{}
	auditService := services.NewTransactionAuditService(recorder)

	result, err := NewListTransactionsUseCase(repository, &transactionStep4RepositoryMock{}, &transactionStep5RepositoryMock{}, &sectorRepositoryMock{}, auditService).Execute(context.Background(), inboundports.ListTransactionsQuery{UploadID: uploadID.String(), CreatedAtFrom: "2026-04-01", CreatedAtTo: "2026-04-30", ApplicantCountry: "Philippines", BeneficiaryCountry: "Japan", SourceCountry: "Thailand", DestinationCountry: "Philippines", TransactionCountMin: "1", TransactionCountMax: "10", Classification: "unclassified", Status: "processing", SortBy: "transaction_count", SortOrder: "asc", ActorUserID: "admin-1", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result.Transactions) != 1 {
		t.Fatalf("len(result.Transactions) = %d, want %d", len(result.Transactions), 1)
	}

	if repository.listFilter.UploadID == nil || repository.listFilter.UploadID.String() != uploadID.String() {
		t.Fatalf("repository.listFilter.UploadID = %v, want %q", repository.listFilter.UploadID, uploadID.String())
	}

	if repository.listFilter.CreatedAtFrom == nil || repository.listFilter.CreatedAtFrom.Format(time.DateOnly) != "2026-04-01" {
		t.Fatalf("repository.listFilter.CreatedAtFrom = %v, want %q", repository.listFilter.CreatedAtFrom, "2026-04-01")
	}

	if repository.listFilter.CreatedAtTo == nil || repository.listFilter.CreatedAtTo.Format(time.DateOnly) != "2026-04-30" {
		t.Fatalf("repository.listFilter.CreatedAtTo = %v, want %q", repository.listFilter.CreatedAtTo, "2026-04-30")
	}

	if repository.listFilter.ApplicantCountry == nil || *repository.listFilter.ApplicantCountry != "Philippines" {
		t.Fatalf("repository.listFilter.ApplicantCountry = %v, want %q", repository.listFilter.ApplicantCountry, "Philippines")
	}

	if repository.listFilter.BeneficiaryCountry == nil || *repository.listFilter.BeneficiaryCountry != "Japan" {
		t.Fatalf("repository.listFilter.BeneficiaryCountry = %v, want %q", repository.listFilter.BeneficiaryCountry, "Japan")
	}

	if repository.listFilter.SourceCountry == nil || *repository.listFilter.SourceCountry != "Thailand" {
		t.Fatalf("repository.listFilter.SourceCountry = %v, want %q", repository.listFilter.SourceCountry, "Thailand")
	}

	if repository.listFilter.DestinationCountry == nil || *repository.listFilter.DestinationCountry != "Philippines" {
		t.Fatalf("repository.listFilter.DestinationCountry = %v, want %q", repository.listFilter.DestinationCountry, "Philippines")
	}

	if repository.listFilter.TransactionCountMin == nil || *repository.listFilter.TransactionCountMin != 1 {
		t.Fatalf("repository.listFilter.TransactionCountMin = %v, want %d", repository.listFilter.TransactionCountMin, 1)
	}

	if repository.listFilter.TransactionCountMax == nil || *repository.listFilter.TransactionCountMax != 10 {
		t.Fatalf("repository.listFilter.TransactionCountMax = %v, want %d", repository.listFilter.TransactionCountMax, 10)
	}

	if repository.listFilter.Classification == nil || *repository.listFilter.Classification != "unclassified" {
		t.Fatalf("repository.listFilter.Classification = %v, want %q", repository.listFilter.Classification, "unclassified")
	}

	if repository.listFilter.Status == nil || *repository.listFilter.Status != "processing" {
		t.Fatalf("repository.listFilter.Status = %v, want %q", repository.listFilter.Status, "processing")
	}

	if repository.listFilter.SortBy != outboundports.TransactionSortByTransactionCount {
		t.Fatalf("repository.listFilter.SortBy = %q, want %q", repository.listFilter.SortBy, outboundports.TransactionSortByTransactionCount)
	}

	if repository.listFilter.SortOrder != outboundports.TransactionSortOrderAscending {
		t.Fatalf("repository.listFilter.SortOrder = %q, want %q", repository.listFilter.SortOrder, outboundports.TransactionSortOrderAscending)
	}

	if recorder.command.EventType != listTransactionsAdminEventType {
		t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, listTransactionsAdminEventType)
	}
}

// This test ensures invalid classification filters are rejected before the
// repository is queried, preserving the API validation contract.
func TestListTransactionsUseCaseExecuteRejectsInvalidClassification(t *testing.T) {
	t.Parallel()

	repository := &transactionRepositoryMock{}
	recorder := &adminEventRecorderMock{}
	auditService := services.NewTransactionAuditService(recorder)

	_, err := NewListTransactionsUseCase(repository, &transactionStep4RepositoryMock{}, &transactionStep5RepositoryMock{}, &sectorRepositoryMock{}, auditService).Execute(context.Background(), inboundports.ListTransactionsQuery{Classification: "bad-value"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListTransactionsUseCaseExecuteAcceptsCanonicalClassificationAndSortFilters(t *testing.T) {
	t.Parallel()

	repository := &transactionRepositoryMock{}
	recorder := &adminEventRecorderMock{}
	auditService := services.NewTransactionAuditService(recorder)

	_, err := NewListTransactionsUseCase(
		repository,
		&transactionStep4RepositoryMock{},
		&transactionStep5RepositoryMock{},
		&sectorRepositoryMock{},
		auditService,
	).Execute(context.Background(), inboundports.ListTransactionsQuery{
		Classification: "next_step",
		SortBy:         outboundports.TransactionSortByCreatedAt,
		SortOrder:      outboundports.TransactionSortOrderDescending,
		ActorUserID:    "admin-1",
		ActorGroupID:   "group-1",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if repository.listFilter.Classification == nil || *repository.listFilter.Classification != "next_step" {
		t.Fatalf("repository.listFilter.Classification = %v, want %q", repository.listFilter.Classification, "next_step")
	}
}

func TestListTransactionsUseCaseExecuteRejectsMissingActorAuditMetadata(t *testing.T) {
	t.Parallel()

	repository := &transactionRepositoryMock{}
	recorder := &adminEventRecorderMock{}
	auditService := services.NewTransactionAuditService(recorder)

	_, err := NewListTransactionsUseCase(
		repository,
		&transactionStep4RepositoryMock{},
		&transactionStep5RepositoryMock{},
		&sectorRepositoryMock{},
		auditService,
	).Execute(context.Background(), inboundports.ListTransactionsQuery{})
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	if err.Error() != "creating list transactions audit event: [INVALID_ACTOR_USER_ID] actor user id is required" {
		t.Fatalf("Execute() error = %v, want %q", err, "creating list transactions audit event: [INVALID_ACTOR_USER_ID] actor user id is required")
	}
}

func TestNormalizeNavigationClassification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    *string
		wantErr string
	}{
		{name: "empty returns nil", raw: "", want: nil},
		{name: "needs review alias maps to canonical next step", raw: "needs review", want: stringPtr("next_step")},
		{name: "canonical next step preserved", raw: "next_step", want: stringPtr("next_step")},
		{name: "dash next step normalized", raw: "next-step", want: stringPtr("next_step")},
		{name: "invalid classification rejected", raw: "bad-value", wantErr: domain.ErrInvalidTransactionClassification.Error()},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeNavigationClassification(tc.raw)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatal("normalizeNavigationClassification() error = nil, want error")
				}
				if err.Error() != tc.wantErr {
					t.Fatalf("normalizeNavigationClassification() error = %v, want %q", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("normalizeNavigationClassification() error = %v", err)
			}

			if tc.want == nil {
				if got != nil {
					t.Fatalf("normalizeNavigationClassification() = %v, want nil", got)
				}
				return
			}

			if got == nil || *got != *tc.want {
				t.Fatalf("normalizeNavigationClassification() = %v, want %q", got, *tc.want)
			}
		})
	}
}

func TestNormalizeNavigationStep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		raw         string
		want        *int
		wantField   string
		wantMessage string
	}{
		{name: "empty returns nil", raw: "", want: nil},
		{name: "all returns nil", raw: "all", want: nil},
		{name: "valid step one", raw: "1", want: intPtr(1)},
		{name: "valid step five", raw: "5", want: intPtr(5)},
		{name: "invalid zero", raw: "0", wantField: "step", wantMessage: "step must be one of: 1, 2, 3, 4, 5"},
		{name: "invalid six", raw: "6", wantField: "step", wantMessage: "step must be one of: 1, 2, 3, 4, 5"},
		{name: "invalid text", raw: "abc", wantField: "step", wantMessage: "step must be one of: 1, 2, 3, 4, 5"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeNavigationStep(tc.raw)
			if tc.wantMessage != "" {
				var validationErr *domain.ValidationError
				if !errors.As(err, &validationErr) {
					t.Fatalf("normalizeNavigationStep() error = %v, want ValidationError", err)
				}

				fields := validationErr.Fields()
				if len(fields) != 1 {
					t.Fatalf("len(validationErr.Fields()) = %d, want %d", len(fields), 1)
				}

				if fields[0].Field() != tc.wantField {
					t.Fatalf("validationErr.Fields()[0].Field() = %q, want %q", fields[0].Field(), tc.wantField)
				}

				if fields[0].Message() != tc.wantMessage {
					t.Fatalf("validationErr.Fields()[0].Message() = %q, want %q", fields[0].Message(), tc.wantMessage)
				}
				return
			}

			if err != nil {
				t.Fatalf("normalizeNavigationStep() error = %v", err)
			}

			if tc.want == nil {
				if got != nil {
					t.Fatalf("normalizeNavigationStep() = %v, want nil", got)
				}
				return
			}

			if got == nil || *got != *tc.want {
				t.Fatalf("normalizeNavigationStep() = %v, want %d", got, *tc.want)
			}
		})
	}
}

// This test confirms list results embed step-4 classification details when the
// supporting step-4 and sector records exist.
func TestListTransactionsUseCaseExecuteIncludesStep4Classification(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := entities.NewTransaction(transactionID, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	sectorID, err := valueobjects.SectorIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("SectorIDFromString() error = %v", err)
	}

	additionalContext, err := valueobjects.NewTransactionStep4AdditionalContext("Reviewed by analyst")
	if err != nil {
		t.Fatalf("NewTransactionStep4AdditionalContext() error = %v", err)
	}

	step4, err := entities.NewTransactionStep4(transactionID, sectorID, additionalContext, false, testTime())
	if err != nil {
		t.Fatalf("NewTransactionStep4() error = %v", err)
	}

	sector, err := entities.NewSector(sectorID, "High Emitting", "Energy", "High emitting energy sector")
	if err != nil {
		t.Fatalf("NewSector() error = %v", err)
	}

	repository := &transactionRepositoryMock{listResult: []*entities.Transaction{transaction}}
	recorder := &adminEventRecorderMock{}
	auditService := services.NewTransactionAuditService(recorder)

	result, err := NewListTransactionsUseCase(repository, &transactionStep4RepositoryMock{findByTransactionID: step4}, &transactionStep5RepositoryMock{}, &sectorRepositoryMock{findByID: sector}, auditService).Execute(context.Background(), inboundports.ListTransactionsQuery{ActorUserID: "admin-1", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Transactions[0].Step4Classification == nil {
		t.Fatal("result.Transactions[0].Step4Classification = nil, want step 4 classification")
	}

	if result.Transactions[0].Step4Classification.IdentifiedSector != "Energy" {
		t.Fatalf("result.Transactions[0].Step4Classification.IdentifiedSector = %q, want %q", result.Transactions[0].Step4Classification.IdentifiedSector, "Energy")
	}

	if result.Transactions[0].Step4Classification.Result != "aligned" {
		t.Fatalf("result.Transactions[0].Step4Classification.Result = %q, want %q", result.Transactions[0].Step4Classification.Result, "aligned")
	}
}

// This test confirms list results embed step-5 reviewer output when a step-5
// record exists for the transaction.
func TestListTransactionsUseCaseExecuteIncludesStep5Classification(t *testing.T) {
	t.Parallel()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := entities.NewTransaction(transactionID, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	question1Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion1Justification("question 1 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion1Justification() error = %v", err)
	}

	question2Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion2Justification("question 2 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion2Justification() error = %v", err)
	}

	step5, err := entities.NewTransactionStep5(transactionID, true, question1Justification, false, question2Justification, valueobjects.NewTransactionStep5ReviewerNotes(nil), true, testTime())
	if err != nil {
		t.Fatalf("NewTransactionStep5() error = %v", err)
	}

	repository := &transactionRepositoryMock{listResult: []*entities.Transaction{transaction}}
	recorder := &adminEventRecorderMock{}
	auditService := services.NewTransactionAuditService(recorder)

	result, err := NewListTransactionsUseCase(repository, &transactionStep4RepositoryMock{}, &transactionStep5RepositoryMock{findByTransactionID: step5}, &sectorRepositoryMock{}, auditService).Execute(context.Background(), inboundports.ListTransactionsQuery{ActorUserID: "admin-1", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Transactions[0].Step5Classification == nil {
		t.Fatal("result.Transactions[0].Step5Classification = nil, want step 5 classification")
	}

	if result.Transactions[0].Step5Classification.Result != "not-aligned" {
		t.Fatalf("result.Transactions[0].Step5Classification.Result = %q, want %q", result.Transactions[0].Step5Classification.Result, "not-aligned")
	}
}

func stringPtr(value string) *string {
	return &value
}

func intPtr(value int) *int {
	return &value
}
