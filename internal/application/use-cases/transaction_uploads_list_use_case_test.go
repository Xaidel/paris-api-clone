package usecases

import (
	"context"
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestListTransactionUploadsUseCaseExecute verifies the list transaction uploads use case execute behavior and the expected outcome asserted below.
func TestListTransactionUploadsUseCaseExecute(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	repository := &transactionUploadRepositoryMock{listResult: []*entities.TransactionUpload{testReconstitutedTransactionUpload(t, uploadID, uploadedTransactionUploadStatus, 1)}}
	transaction, err := entities.NewUploadedTransaction(transactionID, uploadID, 2, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewUploadedTransaction() error = %v", err)
	}

	transactionRepository := &transactionRepositoryMock{listResult: []*entities.Transaction{transaction}}
	useCase := NewListTransactionUploadsUseCase(repository, transactionRepository, &transactionStep4RepositoryMock{}, &transactionStep5RepositoryMock{}, &sectorRepositoryMock{})

	result, err := useCase.Execute(context.Background(), inboundports.ListTransactionUploadsQuery{ActorGroupID: groupID.String()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result.Uploads) != 1 {
		t.Fatalf("len(result.Uploads) = %d, want %d", len(result.Uploads), 1)
	}

	if len(result.Uploads[0].Transactions) != 1 {
		t.Fatalf("len(result.Uploads[0].Transactions) = %d, want %d", len(result.Uploads[0].Transactions), 1)
	}

	if result.Uploads[0].GroupID != groupID.String() {
		t.Fatalf("result.Uploads[0].GroupID = %q, want %q", result.Uploads[0].GroupID, groupID.String())
	}

	if result.Uploads[0].Status != uploadedTransactionUploadStatus.String() {
		t.Fatalf("result.Uploads[0].Status = %q, want %q", result.Uploads[0].Status, uploadedTransactionUploadStatus.String())
	}

	if result.Uploads[0].GroupID != testGroupID(t).String() {
		t.Fatalf("result.Uploads[0].GroupID = %q, want %q", result.Uploads[0].GroupID, testGroupID(t).String())
	}

	if result.Uploads[0].Transactions[0].GoodsDescription != "Goods" {
		t.Fatalf("result.Uploads[0].Transactions[0].GoodsDescription = %q, want %q", result.Uploads[0].Transactions[0].GoodsDescription, "Goods")
	}

	if result.Uploads[0].Transactions[0].ID != transactionID.String() {
		t.Fatalf("result.Uploads[0].Transactions[0].ID = %q, want %q", result.Uploads[0].Transactions[0].ID, transactionID.String())
	}

	if len(transactionRepository.listByUploadIDs) != 1 {
		t.Fatalf("len(transactionRepository.listByUploadIDs) = %d, want %d", len(transactionRepository.listByUploadIDs), 1)
	}
}

// TestListTransactionUploadsUseCaseExecuteFiltersByActorGroup verifies the list transaction uploads use case forwards actor group and keeps other filters intact.
func TestListTransactionUploadsUseCaseExecuteFiltersByActorGroup(t *testing.T) {
	t.Parallel()

	groupIDValue := "01962b8f-aeb2-7e03-a8ff-1edce1300001"
	startedAt := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	endedAt := time.Date(2026, time.April, 30, 23, 59, 59, 0, time.UTC)
	repository := &transactionUploadRepositoryMock{}
	useCase := NewListTransactionUploadsUseCase(repository, &transactionRepositoryMock{}, &transactionStep4RepositoryMock{}, &transactionStep5RepositoryMock{}, &sectorRepositoryMock{})

	result, err := useCase.Execute(context.Background(), inboundports.ListTransactionUploadsQuery{
		FileName:     "transactions",
		StartedAt:    &startedAt,
		EndedAt:      &endedAt,
		ActorGroupID: groupIDValue,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result.Uploads) != 0 {
		t.Fatalf("len(result.Uploads) = %d, want %d", len(result.Uploads), 0)
	}

	groupID, err := valueobjects.GroupIDFromString(groupIDValue)
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	if !repository.listFilter.GroupID.Equal(groupID) {
		t.Fatalf("repository.listFilter.GroupID = %q, want %q", repository.listFilter.GroupID.String(), groupID.String())
	}

	if repository.listFilter.FileName != "transactions" {
		t.Fatalf("repository.listFilter.FileName = %q, want %q", repository.listFilter.FileName, "transactions")
	}

	if repository.listFilter.StartedAt == nil || !repository.listFilter.StartedAt.Equal(startedAt) {
		t.Fatalf("repository.listFilter.StartedAt = %v, want %v", repository.listFilter.StartedAt, startedAt)
	}

	if repository.listFilter.EndedAt == nil || !repository.listFilter.EndedAt.Equal(endedAt) {
		t.Fatalf("repository.listFilter.EndedAt = %v, want %v", repository.listFilter.EndedAt, endedAt)
	}
}

// TestListTransactionUploadsUseCaseExecuteIncludesStep5Classification verifies the list transaction uploads use case execute includes step 5 classification behavior and the expected outcome asserted below.
func TestListTransactionUploadsUseCaseExecuteIncludesStep5Classification(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	repository := &transactionUploadRepositoryMock{listResult: []*entities.TransactionUpload{testReconstitutedTransactionUpload(t, uploadID, uploadedTransactionUploadStatus, 1)}}
	transaction, err := entities.NewUploadedTransaction(transactionID, uploadID, 2, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewUploadedTransaction() error = %v", err)
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

	transactionRepository := &transactionRepositoryMock{listResult: []*entities.Transaction{transaction}}
	useCase := NewListTransactionUploadsUseCase(repository, transactionRepository, &transactionStep4RepositoryMock{}, &transactionStep5RepositoryMock{findByTransactionID: step5}, &sectorRepositoryMock{})

	result, err := useCase.Execute(context.Background(), inboundports.ListTransactionUploadsQuery{ActorGroupID: groupID.String()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Uploads[0].Transactions[0].Step5Classification == nil {
		t.Fatal("result.Uploads[0].Transactions[0].Step5Classification = nil, want step 5 classification")
	}

	if result.Uploads[0].Status != uploadedTransactionUploadStatus.String() {
		t.Fatalf("result.Uploads[0].Status = %q, want %q", result.Uploads[0].Status, uploadedTransactionUploadStatus.String())
	}
}
