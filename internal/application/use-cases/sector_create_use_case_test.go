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

// TestCreateSectorUseCaseExecute verifies the create sector use case execute behavior and the expected outcome asserted below.
func TestCreateSectorUseCaseExecute(t *testing.T) {
	t.Parallel()

	fixedID, err := valueobjects.SectorIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("SectorIDFromString() error = %v", err)
	}

	tests := []struct {
		name        string
		repository  *sectorRepositoryMock
		transaction *transactionManagerMock
		recorder    *adminEventRecorderMock
		command     inboundports.CreateSectorCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.SectorResult, repository *sectorRepositoryMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock)
	}{
		{
			name:        "creates sector entry",
			repository:  &sectorRepositoryMock{},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.CreateSectorCommand{Type: "High Emitting", Name: "Steel", Description: "Steel production", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.SectorResult, repository *sectorRepositoryMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock) {
				t.Helper()

				if repository.createdSector == nil {
					t.Fatal("expected created sector")
				}

				if repository.createdBy != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("repository.createdBy = %q, want %q", repository.createdBy, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				if result.CreatedBy != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("result.CreatedBy = %q, want %q", result.CreatedBy, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				if result.Type != "High Emitting" {
					t.Fatalf("result.Type = %q, want %q", result.Type, "High Emitting")
				}

				if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
					t.Fatalf("expected non-zero timestamps, got created_at=%v updated_at=%v", result.CreatedAt, result.UpdatedAt)
				}

				if !transaction.invoked {
					t.Fatal("expected transaction manager to be invoked")
				}

				if recorder.command.EventType != createSectorAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, createSectorAdminEventType)
				}

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}

				if payload["name"] != "Steel" {
					t.Fatalf("payload[name] = %v, want %q", payload["name"], "Steel")
				}

			},
		},
		{
			name:        "wraps repository error",
			repository:  &sectorRepositoryMock{createErr: errors.New("boom")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.CreateSectorCommand{Type: "High Emitting", Name: "Steel", Description: "Steel production", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "creating sector transaction") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewCreateSectorUseCase(tc.repository, tc.transaction, tc.recorder, &actorDirectoryMock{})
			useCase.newID = func() (valueobjects.SectorID, error) { return fixedID, nil }
			useCase.now = testTime

			result, err := useCase.Execute(context.Background(), tc.command)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result, tc.repository, tc.transaction, tc.recorder)
		})
	}
}
