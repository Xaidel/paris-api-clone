package usecases

import (
	"context"
	"errors"
	"testing"

	rootdomain "github.com/gyud-adb/paris-api/internal/domain"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestGetUserUseCaseExecute verifies the get user use case execute behavior and the expected outcome asserted below.
func TestGetUserUseCaseExecute(t *testing.T) {
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

	user := entities.ReconstituteUser(userID, "alice", "hash", profile, testTime(), testTime())

	tests := []struct {
		name        string
		query       inboundports.GetUserQuery
		repository  *userRepositoryMock
		recorder    *adminEventRecorderMock
		actors      *actorDirectoryMock
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.UserResult, recorder *adminEventRecorderMock)
	}{
		{
			name:       "gets user successfully",
			query:      inboundports.GetUserQuery{ID: userID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			repository: &userRepositoryMock{findByID: user},
			recorder:   &adminEventRecorderMock{},
			actors:     &actorDirectoryMock{},
			assert: func(t *testing.T, result outboundports.UserResult, recorder *adminEventRecorderMock) {
				t.Helper()

				if result.ID != userID.String() {
					t.Fatalf("result.ID = %q", result.ID)
				}

				if result.Username != "alice" {
					t.Fatalf("result.Username = %q, want %q", result.Username, "alice")
				}

				if recorder.command.EventType != getUserAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, getUserAdminEventType)
				}
			},
		},
		{
			name:       "rejects invalid id",
			query:      inboundports.GetUserQuery{ID: "bad-id"},
			repository: &userRepositoryMock{},
			recorder:   &adminEventRecorderMock{},
			actors:     &actorDirectoryMock{},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var domainErr *rootdomain.DomainError
				if !errors.As(err, &domainErr) {
					t.Fatalf("expected domain error, got %v", err)
				}
			},
		},
		{
			name:       "returns not found",
			query:      inboundports.GetUserQuery{ID: userID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: groupID.String()},
			repository: &userRepositoryMock{},
			recorder:   &adminEventRecorderMock{},
			actors:     &actorDirectoryMock{},
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

			useCase := NewGetUserUseCase(tt.repository, tt.recorder, tt.actors)
			result, err := useCase.Execute(context.Background(), tt.query)
			if tt.assertError != nil {
				tt := tt
				tt.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tt := tt
			tt.assert(t, result, tt.recorder)
		})
	}
}
