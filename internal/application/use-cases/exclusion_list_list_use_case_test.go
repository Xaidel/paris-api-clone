package usecases

import (
	"context"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestListExclusionListUseCaseExecute verifies the list exclusion list use case execute behavior and the expected outcome asserted below.
func TestListExclusionListUseCaseExecute(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.ExclusionListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("ExclusionListIDFromString() error = %v", err)
	}

	repository := &exclusionListRepositoryMock{listEntries: []*entities.ExclusionListEntry{entities.ReconstituteExclusionListEntry(id, "agriculture")}}
	recorder := &adminEventRecorderMock{}

	useCase := NewListExclusionListUseCase(repository, recorder)
	useCase.now = testTime
	result, err := useCase.Execute(context.Background(), inboundports.ListExclusionListQuery{ActorUserID: "admin-1", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result.Entries) != 1 {
		t.Fatalf("len(result.Entries) = %d, want 1", len(result.Entries))
	}

	if result.Entries[0].ActivityType != "agriculture" {
		t.Fatalf("result.Entries[0].ActivityType = %q, want %q", result.Entries[0].ActivityType, "agriculture")
	}

	if recorder.command.EventType != listExclusionListAdminEventType {
		t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, listExclusionListAdminEventType)
	}

	if !recorder.when.Equal(testTime()) {
		t.Fatalf("recorder.when = %v, want %v", recorder.when, testTime())
	}
}
