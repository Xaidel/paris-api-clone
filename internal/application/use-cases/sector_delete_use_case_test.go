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

// TestDeleteSectorUseCaseExecute verifies the delete sector use case execute behavior and the expected outcome asserted below.
func TestDeleteSectorUseCaseExecute(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.SectorIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("SectorIDFromString() error = %v", err)
	}

	sectorType, err := valueobjects.SectorTypeFromString("High Emitting")
	if err != nil {
		t.Fatalf("SectorTypeFromString() error = %v", err)
	}

	tests := []struct {
		name        string
		repository  *sectorRepositoryMock
		transaction *transactionManagerMock
		recorder    *adminEventRecorderMock
		command     inboundports.DeleteSectorCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result inboundports.DeleteSectorResult, repository *sectorRepositoryMock, transaction *transactionManagerMock)
	}{
		{
			name:        "deletes sector entry",
			repository:  &sectorRepositoryMock{findByID: entities.ReconstituteSector(id, sectorType, "Steel", "Steel production")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.DeleteSectorCommand{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result inboundports.DeleteSectorResult, repository *sectorRepositoryMock, transaction *transactionManagerMock) {
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
			repository:  &sectorRepositoryMock{findByID: entities.ReconstituteSector(id, sectorType, "Steel", "Steel production"), deleteErr: errors.New("boom")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.DeleteSectorCommand{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "deleting sector transaction") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewDeleteSectorUseCase(tc.repository, tc.transaction, tc.recorder)
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
