package usecases

import (
	"context"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestDeleteTransactionUseCaseExecute verifies the delete transaction use case execute behavior and the expected outcome asserted below.
func TestDeleteTransactionUseCaseExecute(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := entities.NewTransaction(id, "CG", 2026, 4, "IB", "DMC", "Partner Bank", "REF-1", "698436.80", 1, "Goods", "Classification", "Philippines", "Japan", "Thailand", "Philippines", "N", "", "PA Aligned", testTime())
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	repository := &transactionRepositoryMock{findByID: transaction}
	manager := &transactionManagerMock{}
	recorder := &adminEventRecorderMock{}
	useCase := NewDeleteTransactionUseCase(repository, manager, recorder)
	useCase.now = testTime
	result, err := useCase.Execute(context.Background(), inboundports.DeleteTransactionCommand{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ID != id.String() {
		t.Fatalf("result.ID = %q, want %q", result.ID, id.String())
	}

	if repository.deletedID != id.String() {
		t.Fatalf("repository.deletedID = %q, want %q", repository.deletedID, id.String())
	}

	if recorder.command.EventType != deleteTransactionAdminEventType {
		t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, deleteTransactionAdminEventType)
	}
}
