package usecases

import (
	"context"
	"errors"
	"testing"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestDeleteTransactionUploadUseCaseExecute verifies the delete transaction upload use case execute behavior and the expected outcome asserted below.
func TestDeleteTransactionUploadUseCaseExecute(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	uploadRepository := &transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, valueobjects.UploadedTransactionUploadStatus(), 1)}
	transactionRepository := &transactionRepositoryMock{}
	store := &rawFileStoreMock{}
	manager := &transactionManagerUseCaseMock{}
	recorder := &adminEventRecorderUploadMock{}
	actors := &actorDirectoryMock{}
	useCase := NewDeleteTransactionUploadUseCase(uploadRepository, transactionRepository, store, manager, recorder, actors)
	useCase.now = testTime

	result, err := useCase.Execute(context.Background(), inboundports.DeleteTransactionUploadCommand{ID: uploadID.String(), ActorUserID: "admin-1", ActorGroupID: groupID.String()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ID != uploadID.String() {
		t.Fatalf("result.ID = %q, want %q", result.ID, uploadID.String())
	}

	if store.deletedCommand.Key != "upload/file.csv" {
		t.Fatalf("store.deletedCommand.Key = %q, want %q", store.deletedCommand.Key, "upload/file.csv")
	}

	if recorder.command.EventType != deleteTransactionUploadAdminEventType {
		t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, deleteTransactionUploadAdminEventType)
	}

	if actors.userID != "admin-1" {
		t.Fatalf("actors.userID = %q, want %q", actors.userID, "admin-1")
	}

	if actors.groupID != groupID.String() {
		t.Fatalf("actors.groupID = %q, want %q", actors.groupID, groupID.String())
	}
}

// TestDeleteTransactionUploadUseCaseExecuteRejectsProcessingUpload verifies uploads cannot be deleted while classification is still processing.
func TestDeleteTransactionUploadUseCaseExecuteRejectsProcessingUpload(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	uploadRepository := &transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, valueobjects.UploadedTransactionUploadStatus(), 1)}
	transactionRepository := &transactionRepositoryMock{hasProcessingResult: true}
	store := &rawFileStoreMock{}
	manager := &transactionManagerUseCaseMock{}
	recorder := &adminEventRecorderUploadMock{}
	actors := &actorDirectoryMock{}
	useCase := NewDeleteTransactionUploadUseCase(uploadRepository, transactionRepository, store, manager, recorder, actors)

	_, err = useCase.Execute(context.Background(), inboundports.DeleteTransactionUploadCommand{ID: uploadID.String(), ActorUserID: "admin-1", ActorGroupID: groupID.String()})
	if err == nil {
		t.Fatal("Execute() error = nil, want conflict")
	}

	var conflictErr *ConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("Execute() error = %T, want *ConflictError", err)
	}

	wantMessage := "transaction upload conflict: cannot delete transaction upload while transactions are still processing"
	if conflictErr.Error() != wantMessage {
		t.Fatalf("conflictErr.Error() = %q, want %q", conflictErr.Error(), wantMessage)
	}

	if !transactionRepository.hasProcessingUpload.Equal(uploadID) {
		t.Fatalf("transactionRepository.hasProcessingUpload = %q, want %q", transactionRepository.hasProcessingUpload.String(), uploadID.String())
	}

	if manager.invoked {
		t.Fatal("transaction manager should not be invoked when upload is still processing")
	}

	if transactionRepository.deletedByUploadID != "" {
		t.Fatalf("transactionRepository.deletedByUploadID = %q, want empty", transactionRepository.deletedByUploadID)
	}

	if store.deletedCommand.Key != "" {
		t.Fatalf("store.deletedCommand.Key = %q, want empty", store.deletedCommand.Key)
	}

	if recorder.command.EventType != "" {
		t.Fatalf("recorder.command.EventType = %q, want empty", recorder.command.EventType)
	}

	if actors.userID != "admin-1" {
		t.Fatalf("actors.userID = %q, want %q", actors.userID, "admin-1")
	}

	if actors.groupID != groupID.String() {
		t.Fatalf("actors.groupID = %q, want %q", actors.groupID, groupID.String())
	}
}

// TestDeleteTransactionUploadUseCaseExecuteRejectsGroupMismatch verifies uploads cannot be deleted across groups.
func TestDeleteTransactionUploadUseCaseExecuteRejectsGroupMismatch(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	otherGroupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300003")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	uploadRepository := &transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, valueobjects.UploadedTransactionUploadStatus(), 1)}
	transactionRepository := &transactionRepositoryMock{}
	store := &rawFileStoreMock{}
	manager := &transactionManagerUseCaseMock{}
	recorder := &adminEventRecorderUploadMock{}
	actors := &actorDirectoryMock{}
	useCase := NewDeleteTransactionUploadUseCase(uploadRepository, transactionRepository, store, manager, recorder, actors)

	_, err = useCase.Execute(context.Background(), inboundports.DeleteTransactionUploadCommand{ID: uploadID.String(), ActorUserID: "admin-1", ActorGroupID: otherGroupID.String()})
	if err == nil {
		t.Fatal("Execute() error = nil, want forbidden")
	}

	var forbiddenErr *ForbiddenError
	if !errors.As(err, &forbiddenErr) {
		t.Fatalf("Execute() error = %T, want *ForbiddenError", err)
	}

	if manager.invoked {
		t.Fatal("transaction manager should not be invoked when group access is forbidden")
	}

	if transactionRepository.deletedByUploadID != "" {
		t.Fatalf("transactionRepository.deletedByUploadID = %q, want empty", transactionRepository.deletedByUploadID)
	}

	if store.deletedCommand.Key != "" {
		t.Fatalf("store.deletedCommand.Key = %q, want empty", store.deletedCommand.Key)
	}

	if recorder.command.EventType != "" {
		t.Fatalf("recorder.command.EventType = %q, want empty", recorder.command.EventType)
	}

	if actors.userID != "admin-1" {
		t.Fatalf("actors.userID = %q, want %q", actors.userID, "admin-1")
	}

	if actors.groupID != otherGroupID.String() {
		t.Fatalf("actors.groupID = %q, want %q", actors.groupID, otherGroupID.String())
	}
}
