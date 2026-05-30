package usecases

import (
	"context"
	"testing"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestCreateGroupUseCaseExecute verifies the create group use case execute behavior and the expected outcome asserted below.
func TestCreateGroupUseCaseExecute(t *testing.T) {
	t.Parallel()

	fixedID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300001")
	repository := &groupRepositoryUseCaseMock{}
	transaction := &transactionManagerMock{}
	recorder := &adminEventRecorderMock{}
	actors := &actorDirectoryMock{}

	useCase := NewCreateGroupUseCase(repository, transaction, recorder, actors)
	useCase.newID = func() (valueobjects.GroupID, error) { return fixedID, nil }
	useCase.now = testTime

	result, err := useCase.Execute(context.Background(), inboundports.CreateGroupCommand{Name: "superadmin", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: fixedID.String()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Name != "superadmin" {
		t.Fatalf("result.Name = %q, want %q", result.Name, "superadmin")
	}
}
