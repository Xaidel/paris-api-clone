package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestUpdateU1ListUseCaseExecute verifies the update U1 list use case execute behavior and the expected outcome asserted below.
func TestUpdateU1ListUseCaseExecute(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.U1ListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("U1ListIDFromString() error = %v", err)
	}

	tests := []struct {
		name        string
		repository  *u1ListRepositoryMock
		transaction *transactionManagerMock
		recorder    *adminEventRecorderMock
		command     inboundports.UpdateU1ListCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.U1ListResult, repository *u1ListRepositoryMock, transaction *transactionManagerMock)
	}{
		{
			name:        "updates u1 list entry",
			repository:  &u1ListRepositoryMock{findByID: entities.ReconstituteU1ListEntryWithAudit(id, "energy", "grant", "rule 1", valueobjects.UserID{}, testTime(), testTime())},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.UpdateU1ListCommand{ID: id.String(), Sector: "agriculture", EligibleOperationType: "loan", ConditionGuidance: "rule 2", ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.U1ListResult, repository *u1ListRepositoryMock, transaction *transactionManagerMock) {
				t.Helper()
				if repository.updatedEntry == nil {
					t.Fatal("expected updated entry")
				}
				if !transaction.invoked {
					t.Fatal("expected transaction manager invocation")
				}
				if result.Sector != "agriculture" {
					t.Fatalf("result.Sector = %q, want %q", result.Sector, "agriculture")
				}
				if !result.UpdatedAt.Equal(testTime()) {
					t.Fatalf("result.UpdatedAt = %v, want %v", result.UpdatedAt, testTime())
				}

			},
		},
		{
			name:        "wraps repository update error",
			repository:  &u1ListRepositoryMock{findByID: entities.ReconstituteU1ListEntryWithAudit(id, "energy", "grant", "rule 1", valueobjects.UserID{}, testTime(), testTime()), updateErr: errors.New("boom")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.UpdateU1ListCommand{ID: id.String(), Sector: "agriculture", EligibleOperationType: "loan", ConditionGuidance: "rule 2", ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "updating u1 list transaction") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewUpdateU1ListUseCase(tc.repository, tc.transaction, tc.recorder)
			useCase.now = testTime
			result, err := useCase.Execute(context.Background(), tc.command)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result, tc.repository, tc.transaction)
		})
	}
}
