package usecases

import (
	"context"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestListSectorsUseCaseExecute verifies the list sectors use case execute behavior and the expected outcome asserted below.
func TestListSectorsUseCaseExecute(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.SectorIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("SectorIDFromString() error = %v", err)
	}

	sectorType, err := valueobjects.SectorTypeFromString("PA Aligned")
	if err != nil {
		t.Fatalf("SectorTypeFromString() error = %v", err)
	}

	repository := &sectorRepositoryMock{
		listSectors: []*entities.Sector{entities.ReconstituteSector(id, sectorType, "Renewables", "Renewable power")},
	}
	useCase := NewListSectorsUseCase(repository, &adminEventRecorderMock{})
	useCase.now = testTime

	result, err := useCase.Execute(context.Background(), inboundports.ListSectorsQuery{ActorUserID: "admin-1", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result.Sectors) != 1 {
		t.Fatalf("len(result.Sectors) = %d, want 1", len(result.Sectors))
	}

	if result.Sectors[0].Type != "PA Aligned" {
		t.Fatalf("result.Sectors[0].Type = %q, want %q", result.Sectors[0].Type, "PA Aligned")
	}
}
