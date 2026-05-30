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

// TestCreateU1ListUseCaseExecute verifies the create U1 list use case execute behavior and the expected outcome asserted below.
func TestCreateU1ListUseCaseExecute(t *testing.T) {
	t.Parallel()

	fixedID, err := valueobjects.U1ListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("U1ListIDFromString() error = %v", err)
	}

	tests := []struct {
		name        string
		repository  *u1ListRepositoryMock
		transaction *transactionManagerMock
		recorder    *adminEventRecorderMock
		command     inboundports.CreateU1ListCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.U1ListResult, repository *u1ListRepositoryMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock)
	}{
		{
			name:        "creates u1 list entry",
			repository:  &u1ListRepositoryMock{},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.CreateU1ListCommand{Sector: "energy", EligibleOperationType: "grant", ConditionGuidance: "rule 1", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.U1ListResult, repository *u1ListRepositoryMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock) {
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

				if result.Sector != "energy" {
					t.Fatalf("result.Sector = %q, want %q", result.Sector, "energy")
				}

				if result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() {
					t.Fatalf("expected non-zero timestamps, got created_at=%v updated_at=%v", result.CreatedAt, result.UpdatedAt)
				}

				if !transaction.invoked {
					t.Fatal("expected transaction manager to be invoked")
				}

				if recorder.command.EventType != createU1ListAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, createU1ListAdminEventType)
				}

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}

				if payload["sector"] != "energy" {
					t.Fatalf("payload[sector] = %v, want %q", payload["sector"], "energy")
				}

			},
		},
		{
			name:        "wraps repository error",
			repository:  &u1ListRepositoryMock{createErr: errors.New("boom")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.CreateU1ListCommand{Sector: "energy", EligibleOperationType: "grant", ConditionGuidance: "rule 1", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "creating u1 list transaction") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewCreateU1ListUseCase(tc.repository, tc.transaction, tc.recorder, &actorDirectoryMock{})
			useCase.newID = func() (valueobjects.U1ListID, error) { return fixedID, nil }
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
