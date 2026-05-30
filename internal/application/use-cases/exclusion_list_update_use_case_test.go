package usecases

import (
	"context"
	"errors"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestUpdateExclusionListUseCaseExecute verifies the update exclusion list use case execute behavior and the expected outcome asserted below.
func TestUpdateExclusionListUseCaseExecute(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.ExclusionListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("ExclusionListIDFromString() error = %v", err)
	}

	tests := []struct {
		name        string
		repository  *exclusionListRepositoryMock
		recorder    *adminEventRecorderMock
		command     inboundports.UpdateExclusionListCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.ExclusionListResult, repository *exclusionListRepositoryMock, recorder *adminEventRecorderMock)
	}{
		{
			name:       "updates exclusion list entry",
			repository: &exclusionListRepositoryMock{findByID: entities.ReconstituteExclusionListEntryWithAudit(id, "agriculture", valueobjects.UserID{}, testTime(), testTime())},
			recorder:   &adminEventRecorderMock{},
			command:    inboundports.UpdateExclusionListCommand{ID: id.String(), ActivityType: "energy", ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.ExclusionListResult, repository *exclusionListRepositoryMock, recorder *adminEventRecorderMock) {
				t.Helper()

				if repository.updatedEntry == nil {
					t.Fatal("expected updated entry")
				}

				if result.ActivityType != "energy" {
					t.Fatalf("result.ActivityType = %q, want %q", result.ActivityType, "energy")
				}

				if !result.UpdatedAt.Equal(testTime()) {
					t.Fatalf("result.UpdatedAt = %v, want %v", result.UpdatedAt, testTime())
				}

				if recorder.command.EventType != updateExclusionListAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, updateExclusionListAdminEventType)
				}

			},
		},
		{
			name:       "returns not found",
			repository: &exclusionListRepositoryMock{},
			recorder:   &adminEventRecorderMock{},
			command:    inboundports.UpdateExclusionListCommand{ID: id.String(), ActivityType: "energy", ActorUserID: "admin-1", ActorGroupID: "group-1"},
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

			useCase := NewUpdateExclusionListUseCase(tt.repository, tt.recorder)
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
