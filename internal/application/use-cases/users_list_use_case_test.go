package usecases

import (
	"context"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestListUsersUseCaseExecute verifies the list users use case execute behavior and the expected outcome asserted below.
func TestListUsersUseCaseExecute(t *testing.T) {
	t.Parallel()

	userID, err := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("UserIDFromString() error = %v", err)
	}

	groupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300001")
	profile, err := valueobjects.NewUserProfile("Alice", nil, "Admin", groupID)
	if err != nil {
		t.Fatalf("NewUserProfile() error = %v", err)
	}

	repository := &userRepositoryMock{listUsers: []*entities.User{entities.ReconstituteUser(userID, "alice", "hash", profile, testTime(), testTime())}}
	recorder := &adminEventRecorderMock{}
	actors := &actorDirectoryMock{}

	useCase := NewListUsersUseCase(repository, recorder, actors)
	useCase.now = testTime
	result, err := useCase.Execute(context.Background(), inboundports.ListUsersQuery{ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result.Users) != 1 {
		t.Fatalf("len(result.Users) = %d, want 1", len(result.Users))
	}

	if result.Users[0].Username != "alice" {
		t.Fatalf("result.Users[0].Username = %q, want %q", result.Users[0].Username, "alice")
	}

	if result.Users[0].FirstName != "Alice" {
		t.Fatalf("result.Users[0].FirstName = %q, want %q", result.Users[0].FirstName, "Alice")
	}

	if recorder.command.EventType != listUsersAdminEventType {
		t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, listUsersAdminEventType)
	}

	if !recorder.when.Equal(testTime()) {
		t.Fatalf("recorder.when = %v, want %v", recorder.when, testTime())
	}
}
