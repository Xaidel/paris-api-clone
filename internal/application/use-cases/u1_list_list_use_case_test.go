package usecases

import (
	"context"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestListU1ListUseCaseExecute verifies the list U1 list use case execute behavior and the expected outcome asserted below.
func TestListU1ListUseCaseExecute(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.U1ListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("U1ListIDFromString() error = %v", err)
	}

	tests := []struct {
		name       string
		query      inboundports.ListU1ListQuery
		wantSector string
	}{
		{name: "lists all entries without filter", query: inboundports.ListU1ListQuery{ActorUserID: "admin-1", ActorGroupID: "group-1"}},
		{name: "lists entries filtered by sector", query: inboundports.ListU1ListQuery{Sector: " EnErGy ", ActorUserID: "admin-1", ActorGroupID: "group-1"}, wantSector: "energy"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repository := &u1ListRepositoryMock{
				listEntries: []*entities.U1ListEntry{entities.ReconstituteU1ListEntry(id, "energy", "grant", "rule 1")},
			}
			useCase := NewListU1ListUseCase(repository, &adminEventRecorderMock{})
			useCase.now = testTime

			result, err := useCase.Execute(context.Background(), tc.query)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			if repository.listFilter.Sector != tc.wantSector {
				t.Fatalf("repository.listFilter.Sector = %q, want %q", repository.listFilter.Sector, tc.wantSector)
			}

			if len(result.Entries) != 1 {
				t.Fatalf("len(result.Entries) = %d, want 1", len(result.Entries))
			}

			if result.Entries[0].ConditionGuidance != "rule 1" {
				t.Fatalf("result.Entries[0].ConditionGuidance = %q, want %q", result.Entries[0].ConditionGuidance, "rule 1")
			}
		})
	}
}
