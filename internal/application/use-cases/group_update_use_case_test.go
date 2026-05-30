package usecases

import (
	"context"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestUpdateGroupUseCaseExecute verifies the update group use case execute behavior and the expected outcome asserted below.
func TestUpdateGroupUseCaseExecute(t *testing.T) {
	t.Parallel()

	groupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300001")
	useCase := NewUpdateGroupUseCase(&groupRepositoryUseCaseMock{findByID: entities.ReconstituteGroup(groupID, "superadmin")}, &transactionManagerMock{}, &adminEventRecorderMock{}, &actorDirectoryMock{})
	useCase.now = testTime

	result, err := useCase.Execute(context.Background(), inboundports.UpdateGroupCommand{ID: groupID.String(), Name: "admins", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.Name != "admins" {
		t.Fatalf("result.Name = %q, want %q", result.Name, "admins")
	}
}
