package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestDeleteU1ListUseCaseExecute verifies the delete U1 list use case execute behavior and the expected outcome asserted below.
func TestDeleteU1ListUseCaseExecute(t *testing.T) {
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
		command     inboundports.DeleteU1ListCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result inboundports.DeleteU1ListResult, repository *u1ListRepositoryMock, transaction *transactionManagerMock)
	}{
		{
			name:        "deletes u1 list entry",
			repository:  &u1ListRepositoryMock{findByID: entities.ReconstituteU1ListEntry(id, "energy", "grant", "rule 1")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.DeleteU1ListCommand{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result inboundports.DeleteU1ListResult, repository *u1ListRepositoryMock, transaction *transactionManagerMock) {
				t.Helper()
				if repository.deletedID != id.String() {
					t.Fatalf("repository.deletedID = %q, want %q", repository.deletedID, id.String())
				}
				if !transaction.invoked {
					t.Fatal("expected transaction manager invocation")
				}
				if result.ID != id.String() {
					t.Fatalf("result.ID = %q, want %q", result.ID, id.String())
				}

			},
		},
		{
			name:        "wraps repository delete error",
			repository:  &u1ListRepositoryMock{findByID: entities.ReconstituteU1ListEntry(id, "energy", "grant", "rule 1"), deleteErr: errors.New("boom")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.DeleteU1ListCommand{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "deleting u1 list transaction") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewDeleteU1ListUseCase(tc.repository, tc.transaction, tc.recorder)
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
