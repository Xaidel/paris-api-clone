package usecases

import (
	"context"
	"errors"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestGetExclusionListUseCaseExecute verifies the get exclusion list use case execute behavior and the expected outcome asserted below.
func TestGetExclusionListUseCaseExecute(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.ExclusionListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("ExclusionListIDFromString() error = %v", err)
	}

	entry := entities.ReconstituteExclusionListEntry(id, "agriculture")
	repository := &exclusionListRepositoryMock{findByID: entry}
	recorder := &adminEventRecorderMock{}

	useCase := NewGetExclusionListUseCase(repository, recorder)
	useCase.now = testTime
	result, err := useCase.Execute(context.Background(), inboundports.GetExclusionListQuery{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if result.ActivityType != "agriculture" {
		t.Fatalf("result.ActivityType = %q, want %q", result.ActivityType, "agriculture")
	}

	if recorder.command.EventType != getExclusionListAdminEventType {
		t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, getExclusionListAdminEventType)
	}

	if !recorder.when.Equal(testTime()) {
		t.Fatalf("recorder.when = %v, want %v", recorder.when, testTime())
	}

	repository.findByID = nil
	_, err = useCase.Execute(context.Background(), inboundports.GetExclusionListQuery{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"})
	var notFoundErr *NotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Fatalf("expected not found error, got %v", err)
	}
}
