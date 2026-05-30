package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestCreateExclusionListUseCaseExecute verifies the create exclusion list use case execute behavior and the expected outcome asserted below.
func TestCreateExclusionListUseCaseExecute(t *testing.T) {
	t.Parallel()

	fixedID, err := valueobjects.ExclusionListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("ExclusionListIDFromString() error = %v", err)
	}

	tests := []struct {
		name        string
		repository  *exclusionListRepositoryMock
		transaction *transactionManagerMock
		recorder    *adminEventRecorderMock
		command     inboundports.CreateExclusionListCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.ExclusionListResult, repository *exclusionListRepositoryMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock)
	}{
		{
			name:        "creates exclusion list entry",
			repository:  &exclusionListRepositoryMock{},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.CreateExclusionListCommand{ActivityType: "agriculture", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.ExclusionListResult, repository *exclusionListRepositoryMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock) {
				t.Helper()

				if repository.createdEntry == nil {
					t.Fatal("expected created entry")
				}

				if repository.createdBy != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("repository.createdBy = %q, want %q", repository.createdBy, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				if result.CreatedBy != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("result.CreatedBy = %q, want %q", result.CreatedBy, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				if result.ActivityType != "agriculture" {
					t.Fatalf("result.ActivityType = %q, want %q", result.ActivityType, "agriculture")
				}

				if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
					t.Fatalf("expected non-zero timestamps, got created_at=%v updated_at=%v", result.CreatedAt, result.UpdatedAt)
				}

				if !transaction.invoked {
					t.Fatal("expected transaction manager to be invoked")
				}

				if recorder.command.EventType != createExclusionListAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, createExclusionListAdminEventType)
				}

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}

				if payload["activity_type"] != "agriculture" {
					t.Fatalf("payload[activity_type] = %v, want %q", payload["activity_type"], "agriculture")
				}

			},
		},
		{
			name:        "wraps repository error",
			repository:  &exclusionListRepositoryMock{createErr: errors.New("boom")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.CreateExclusionListCommand{ActivityType: "agriculture", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "creating exclusion list transaction") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewCreateExclusionListUseCase(tt.repository, tt.transaction, tt.recorder, &actorDirectoryMock{})
			useCase.newID = func() (valueobjects.ExclusionListID, error) { return fixedID, nil }
			useCase.now = testTime

			result, err := useCase.Execute(context.Background(), tt.command)
			if tt.assertError != nil {
				tt.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tt.assert(t, result, tt.repository, tt.transaction, tt.recorder)
		})
	}
}
