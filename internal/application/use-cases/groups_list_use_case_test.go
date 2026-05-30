package usecases

import (
	"context"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestListGroupsUseCaseExecute verifies the list groups use case execute behavior and the expected outcome asserted below.
func TestListGroupsUseCaseExecute(t *testing.T) {
	t.Parallel()

	groupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300001")
	useCase := NewListGroupsUseCase(&groupRepositoryUseCaseMock{listGroups: []*entities.Group{entities.ReconstituteGroup(groupID, "superadmin")}}, &adminEventRecorderMock{}, &actorDirectoryMock{})
	useCase.now = testTime

	result, err := useCase.Execute(context.Background(), inboundports.ListGroupsQuery{ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result.Groups) != 1 {
		t.Fatalf("len(result.Groups) = %d, want 1", len(result.Groups))
	}
}
