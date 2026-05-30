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

// TestUpdateSectorUseCaseExecute verifies the update sector use case execute behavior and the expected outcome asserted below.
func TestUpdateSectorUseCaseExecute(t *testing.T) {
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
		command     inboundports.UpdateSectorCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.SectorResult, repository *sectorRepositoryMock, transaction *transactionManagerMock)
	}{
		{
			name:        "updates sector entry",
			repository:  &sectorRepositoryMock{findByID: entities.ReconstituteSectorWithAudit(id, sectorType, "Steel", "Steel production", valueobjects.UserID{}, testTime(), testTime())},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.UpdateSectorCommand{ID: id.String(), Type: "PA Aligned", Name: "Renewables", Description: "Renewable power", ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.SectorResult, repository *sectorRepositoryMock, transaction *transactionManagerMock) {
				t.Helper()
				if repository.updatedSector == nil {
					t.Fatal("expected updated sector")
				}
				if !transaction.invoked {
					t.Fatal("expected transaction manager invocation")
				}
				if result.Type != "PA Aligned" {
					t.Fatalf("result.Type = %q, want %q", result.Type, "PA Aligned")
				}
				if !result.UpdatedAt.Equal(testTime()) {
					t.Fatalf("result.UpdatedAt = %v, want %v", result.UpdatedAt, testTime())
				}

			},
		},
		{
			name:        "wraps repository update error",
			repository:  &sectorRepositoryMock{findByID: entities.ReconstituteSectorWithAudit(id, sectorType, "Steel", "Steel production", valueobjects.UserID{}, testTime(), testTime()), updateErr: errors.New("boom")},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			command:     inboundports.UpdateSectorCommand{ID: id.String(), Type: "PA Aligned", Name: "Renewables", Description: "Renewable power", ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "updating sector transaction") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewUpdateSectorUseCase(tc.repository, tc.transaction, tc.recorder)
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
