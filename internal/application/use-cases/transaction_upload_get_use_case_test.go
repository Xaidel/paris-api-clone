package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

var uploadedTransactionUploadStatus = valueobjects.UploadedTransactionUploadStatus()

// TestGetTransactionUploadUseCaseExecute verifies the get transaction upload use case execute behavior and the expected outcome asserted below.
func TestGetTransactionUploadUseCaseExecute(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	otherGroupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300003")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := entities.NewUploadedTransaction(transactionID, uploadID, 2, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewUploadedTransaction() error = %v", err)
	}

	tests := []struct {
		name        string
		uploadRepo  *transactionUploadRepositoryMock
		transaction *transactionRepositoryMock
		recorder    *adminEventRecorderMock
		actors      *actorDirectoryMock
		query       inboundports.GetTransactionUploadQuery
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.TransactionUploadDetailsResult, transactionRepo *transactionRepositoryMock, recorder *adminEventRecorderMock, actors *actorDirectoryMock)
	}{
		{
			name:        "gets transaction upload",
			uploadRepo:  &transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, uploadedTransactionUploadStatus, 1)},
			transaction: &transactionRepositoryMock{listResult: []*entities.Transaction{transaction}},
			recorder:    &adminEventRecorderMock{},
			actors:      &actorDirectoryMock{},
			query:       inboundports.GetTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "admin-1", ActorGroupID: groupID.String()},
			assert: func(t *testing.T, result outboundports.TransactionUploadDetailsResult, transactionRepo *transactionRepositoryMock, recorder *adminEventRecorderMock, actors *actorDirectoryMock) {
				t.Helper()

				if actors.userID != "admin-1" {
					t.Fatalf("actors.userID = %q, want %q", actors.userID, "admin-1")
				}

				if actors.groupID != groupID.String() {
					t.Fatalf("actors.groupID = %q, want %q", actors.groupID, groupID.String())
				}

				if result.ID != uploadID.String() {
					t.Fatalf("result.ID = %q, want %q", result.ID, uploadID.String())
				}

				if result.GroupID != groupID.String() {
					t.Fatalf("result.GroupID = %q, want %q", result.GroupID, groupID.String())
				}

				if len(result.Transactions) != 1 {
					t.Fatalf("len(result.Transactions) = %d, want %d", len(result.Transactions), 1)
				}

				if result.Status != uploadedTransactionUploadStatus.String() {
					t.Fatalf("result.Status = %q, want %q", result.Status, uploadedTransactionUploadStatus.String())
				}

				if result.GroupID != testGroupID(t).String() {
					t.Fatalf("result.GroupID = %q, want %q", result.GroupID, testGroupID(t).String())
				}

				if result.Transactions[0].GoodsDescription != "Goods" {
					t.Fatalf("result.Transactions[0].GoodsDescription = %q, want %q", result.Transactions[0].GoodsDescription, "Goods")
				}

				if result.Transactions[0].ID != transactionID.String() {
					t.Fatalf("result.Transactions[0].ID = %q, want %q", result.Transactions[0].ID, transactionID.String())
				}

				if result.Transactions[0].Classification != "unclassified" {
					t.Fatalf("result.Transactions[0].Classification = %q, want %q", result.Transactions[0].Classification, "unclassified")
				}

				if result.Transactions[0].Status != "processing" {
					t.Fatalf("result.Transactions[0].Status = %q, want %q", result.Transactions[0].Status, "processing")
				}

				if len(transactionRepo.listByUploadIDs) != 1 || transactionRepo.listByUploadIDs[0].String() != uploadID.String() {
					t.Fatalf("transactionRepo.listByUploadIDs = %v, want [%q]", transactionRepo.listByUploadIDs, uploadID.String())
				}

				if recorder.command.EventType != getTransactionUploadAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, getTransactionUploadAdminEventType)
				}
			},
		},
		{
			name:        "returns forbidden on group mismatch",
			uploadRepo:  &transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, uploadedTransactionUploadStatus, 1)},
			transaction: &transactionRepositoryMock{},
			recorder:    &adminEventRecorderMock{},
			actors:      &actorDirectoryMock{},
			query:       inboundports.GetTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "admin-1", ActorGroupID: otherGroupID.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var forbiddenErr *ForbiddenError
				if !errors.As(err, &forbiddenErr) {
					t.Fatalf("expected ForbiddenError, got %v", err)
				}
			},
			assert: func(t *testing.T, _ outboundports.TransactionUploadDetailsResult, transactionRepo *transactionRepositoryMock, recorder *adminEventRecorderMock, actors *actorDirectoryMock) {
				t.Helper()

				if actors.userID != "admin-1" {
					t.Fatalf("actors.userID = %q, want %q", actors.userID, "admin-1")
				}

				if actors.groupID != otherGroupID.String() {
					t.Fatalf("actors.groupID = %q, want %q", actors.groupID, otherGroupID.String())
				}

				if len(transactionRepo.listByUploadIDs) != 0 {
					t.Fatalf("transactionRepo.listByUploadIDs = %v, want empty", transactionRepo.listByUploadIDs)
				}

				if recorder.command.EventType != "" {
					t.Fatalf("recorder.command.EventType = %q, want empty", recorder.command.EventType)
				}
			},
		},
		{
			name:        "returns not found",
			uploadRepo:  &transactionUploadRepositoryMock{},
			transaction: &transactionRepositoryMock{},
			recorder:    &adminEventRecorderMock{},
			actors:      &actorDirectoryMock{},
			query:       inboundports.GetTransactionUploadQuery{ID: uploadID.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
			},
		},
		{
			name:        "wraps repository error",
			uploadRepo:  &transactionUploadRepositoryMock{findByIDErr: errors.New("boom")},
			transaction: &transactionRepositoryMock{},
			recorder:    &adminEventRecorderMock{},
			actors:      &actorDirectoryMock{},
			query:       inboundports.GetTransactionUploadQuery{ID: uploadID.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				if err == nil || !strings.Contains(err.Error(), "finding upload by id") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewGetTransactionUploadUseCase(tc.uploadRepo, tc.transaction, &transactionStep4RepositoryMock{}, &transactionStep5RepositoryMock{}, &sectorRepositoryMock{}, tc.recorder, tc.actors)
			useCase.now = testTime
			result, err := useCase.Execute(context.Background(), tc.query)
			if tc.assertError != nil {
				tc.assertError(t, err)
				if tc.assert != nil {
					tc.assert(t, result, tc.transaction, tc.recorder, tc.actors)
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result, tc.transaction, tc.recorder, tc.actors)
		})
	}
}

// TestGetTransactionUploadUseCaseExecuteIncludesStep5Classification verifies the get transaction upload use case execute includes step 5 classification behavior and the expected outcome asserted below.
func TestGetTransactionUploadUseCaseExecuteIncludesStep5Classification(t *testing.T) {
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

	step5, err := entities.NewTransactionStep5(transactionID, false, question1Justification, false, question2Justification, valueobjects.NewTransactionStep5ReviewerNotes(nil), true, testTime())
	if err != nil {
		t.Fatalf("NewTransactionStep5() error = %v", err)
	}

	useCase := NewGetTransactionUploadUseCase(
		&transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, uploadedTransactionUploadStatus, 1)},
		&transactionRepositoryMock{listResult: []*entities.Transaction{transaction}},
		&transactionStep4RepositoryMock{},
		&transactionStep5RepositoryMock{findByTransactionID: step5},
		&sectorRepositoryMock{},
		&adminEventRecorderMock{},
		&actorDirectoryMock{},
	)
	useCase.now = testTime

	result, err := useCase.Execute(context.Background(), inboundports.GetTransactionUploadQuery{ID: uploadID.String(), ActorUserID: "admin-1", ActorGroupID: groupID.String()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Transactions[0].Step5Classification == nil {
		t.Fatal("result.Transactions[0].Step5Classification = nil, want step 5 classification")
	}

	if result.Status != uploadedTransactionUploadStatus.String() {
		t.Fatalf("result.Status = %q, want %q", result.Status, uploadedTransactionUploadStatus.String())
	}
}
