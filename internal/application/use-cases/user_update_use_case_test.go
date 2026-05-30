package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestUpdateUserUseCaseExecute verifies the update user use case execute behavior and the expected outcome asserted below.
func TestUpdateUserUseCaseExecute(t *testing.T) {
	t.Parallel()

	userID, err := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("UserIDFromString() error = %v", err)
	}

	groupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300001")
	updatedGroupID := mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
	profile, err := valueobjects.NewUserProfile("Alice", nil, "Admin", groupID)
	if err != nil {
		t.Fatalf("NewUserProfile() error = %v", err)
	}

	tests := []struct {
		name        string
		repository  *userRepositoryMock
		groups      *groupRepositoryMock
		transaction *transactionManagerMock
		recorder    *adminEventRecorderMock
		actors      *actorDirectoryMock
		command     inboundports.UpdateUserCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.UserResult, repository *userRepositoryMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock)
	}{
		{
			name:        "updates user",
			repository:  &userRepositoryMock{findByID: entities.ReconstituteUser(userID, "alice", "old-hash", profile, testTime(), testTime())},
			groups:      &groupRepositoryMock{findByID: entities.ReconstituteGroup(updatedGroupID, "admins")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			actors:      &actorDirectoryMock{},
			command:     inboundports.UpdateUserCommand{ID: userID.String(), Username: "bob", Password: "supersecret", FirstName: "Bob", LastName: "Builder", GroupID: updatedGroupID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assert: func(t *testing.T, result outboundports.UserResult, repository *userRepositoryMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock) {
				t.Helper()

				if repository.updatedUser == nil {
					t.Fatal("expected updated user")
				}

				if !transaction.invoked {
					t.Fatal("expected transaction manager invocation")
				}

				if result.ID != userID.String() {
					t.Fatalf("result.ID = %q, want %q", result.ID, userID.String())
				}

				if result.Username != "bob" {
					t.Fatalf("result.Username = %q, want %q", result.Username, "bob")
				}

				if result.GroupID != updatedGroupID.String() {
					t.Fatalf("result.GroupID = %q, want %q", result.GroupID, updatedGroupID.String())
				}

				if recorder.command.EventType != updateUserAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, updateUserAdminEventType)
				}
			},
		},
		{
			name:        "returns not found",
			repository:  &userRepositoryMock{},
			groups:      &groupRepositoryMock{findByID: entities.ReconstituteGroup(updatedGroupID, "admins")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			actors:      &actorDirectoryMock{},
			command:     inboundports.UpdateUserCommand{ID: userID.String(), Username: "bob", Password: "supersecret", FirstName: "Bob", LastName: "Builder", GroupID: updatedGroupID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected not found error, got %v", err)
				}
			},
		},
		{
			name:        "rejects short password",
			repository:  &userRepositoryMock{},
			groups:      &groupRepositoryMock{},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			actors:      &actorDirectoryMock{},
			command:     inboundports.UpdateUserCommand{ID: userID.String(), Username: "bob", Password: "short", FirstName: "Bob", LastName: "Builder", GroupID: updatedGroupID.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				if err == nil {
					t.Fatal("expected error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewUpdateUserUseCase(tt.repository, tt.groups, &passwordHasherMock{hashResult: "new-hash"}, tt.transaction, tt.recorder, tt.actors)
			useCase.now = func() time.Time { return testTime().Add(time.Minute) }

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
			tt.assert(t, result, tt.repository, tt.transaction, tt.recorder)
		})
	}
}
