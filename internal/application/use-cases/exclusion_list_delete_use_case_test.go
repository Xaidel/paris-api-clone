package usecases

import (
	"context"
	"errors"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestDeleteExclusionListUseCaseExecute verifies the delete exclusion list use case execute behavior and the expected outcome asserted below.
func TestDeleteExclusionListUseCaseExecute(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.ExclusionListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("ExclusionListIDFromString() error = %v", err)
	}

	tests := []struct {
		name        string
		repository  *exclusionListRepositoryMock
		recorder    *adminEventRecorderMock
		command     inboundports.DeleteExclusionListCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result inboundports.DeleteExclusionListResult, repository *exclusionListRepositoryMock, recorder *adminEventRecorderMock)
	}{
		{
			name:       "deletes exclusion list entry",
			repository: &exclusionListRepositoryMock{findByID: entities.ReconstituteExclusionListEntry(id, "agriculture")},
			recorder:   &adminEventRecorderMock{},
			command:    inboundports.DeleteExclusionListCommand{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result inboundports.DeleteExclusionListResult, repository *exclusionListRepositoryMock, recorder *adminEventRecorderMock) {
				t.Helper()
				if repository.deletedID != id.String() {
					t.Fatalf("repository.deletedID = %q, want %q", repository.deletedID, id.String())
				}

				if result.ID != id.String() {
					t.Fatalf("result.ID = %q, want %q", result.ID, id.String())
				}

				if recorder.command.EventType != deleteExclusionListAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, deleteExclusionListAdminEventType)
				}

			},
		},
		{
			name:       "returns not found",
			repository: &exclusionListRepositoryMock{},
			recorder:   &adminEventRecorderMock{},
			command:    inboundports.DeleteExclusionListCommand{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected not found error, got %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewDeleteExclusionListUseCase(tt.repository, tt.recorder)
			useCase.now = testTime
			result, err := useCase.Execute(context.Background(), tt.command)
			if tt.assertError != nil {
				tt.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tt.assert(t, result, tt.repository, tt.recorder)
		})
	}
}
