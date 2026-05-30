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

// TestGetSectorUseCaseExecute verifies the get sector use case execute behavior and the expected outcome asserted below.
func TestGetSectorUseCaseExecute(t *testing.T) {
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
		recorder    *adminEventRecorderMock
		query       inboundports.GetSectorQuery
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.SectorResult)
	}{
		{
			name:       "gets sector entry",
			repository: &sectorRepositoryMock{findByID: entities.ReconstituteSector(id, sectorType, "Steel", "Steel production")},
			recorder:   &adminEventRecorderMock{},
			query:      inboundports.GetSectorQuery{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.SectorResult) {
				t.Helper()
				if result.Name != "Steel" {
					t.Fatalf("result.Name = %q, want %q", result.Name, "Steel")
				}
			},
		},
		{
			name:       "returns not found",
			repository: &sectorRepositoryMock{},
			recorder:   &adminEventRecorderMock{},
			query:      inboundports.GetSectorQuery{ID: id.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
			},
		},
		{
			name:       "wraps repository error",
			repository: &sectorRepositoryMock{findByIDErr: errors.New("boom")},
			recorder:   &adminEventRecorderMock{},
			query:      inboundports.GetSectorQuery{ID: id.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "finding sector entry by id") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewGetSectorUseCase(tc.repository, tc.recorder)
			useCase.now = testTime
			result, err := useCase.Execute(context.Background(), tc.query)
			if tc.assertError != nil {
				tc.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result)
		})
	}
}
