package usecases

import (
	"context"
	"errors"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestDeleteUserUseCaseExecute verifies the delete user use case execute behavior and the expected outcome asserted below.
func TestDeleteUserUseCaseExecute(t *testing.T) {
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

	tests := []struct {
		name        string
		repository  *userRepositoryMock
		recorder    *adminEventRecorderMock
		actors      *actorDirectoryMock
		command     inboundports.DeleteUserCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result inboundports.DeleteUserResult, repository *userRepositoryMock, recorder *adminEventRecorderMock)
	}{
		{
			name:       "deletes user",
			repository: &userRepositoryMock{findByID: entities.ReconstituteUser(userID, "alice", "hash", profile, testTime(), testTime())},
			recorder:   &adminEventRecorderMock{},
			actors:     &actorDirectoryMock{},
			command:    inboundports.DeleteUserCommand{ID: userID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assert: func(t *testing.T, result inboundports.DeleteUserResult, repository *userRepositoryMock, recorder *adminEventRecorderMock) {
				t.Helper()

				if repository.deletedID != userID.String() {
					t.Fatalf("deletedID = %q, want %q", repository.deletedID, userID.String())
				}

				if result.ID != userID.String() {
					t.Fatalf("result.ID = %q, want %q", result.ID, userID.String())
				}

				if recorder.command.EventType != deleteUserAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, deleteUserAdminEventType)
				}
			},
		},
		{
			name:       "returns not found",
			repository: &userRepositoryMock{},
			recorder:   &adminEventRecorderMock{},
			actors:     &actorDirectoryMock{},
			command:    inboundports.DeleteUserCommand{ID: userID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected not found error, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewDeleteUserUseCase(tt.repository, tt.recorder, tt.actors)
			result, err := useCase.Execute(context.Background(), tt.command)
			if tt.assertError != nil {
				tt := tt
				tt.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tt := tt
			tt.assert(t, result, tt.repository, tt.recorder)
		})
	}
}
