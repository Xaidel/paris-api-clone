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

// TestGetU1ListUseCaseExecute verifies the get U1 list use case execute behavior and the expected outcome asserted below.
func TestGetU1ListUseCaseExecute(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.U1ListIDFromString("0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
	if err != nil {
		t.Fatalf("U1ListIDFromString() error = %v", err)
	}

	tests := []struct {
		name        string
		repository  *u1ListRepositoryMock
		recorder    *adminEventRecorderMock
		query       inboundports.GetU1ListQuery
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.U1ListResult)
	}{
		{
			name:       "gets u1 list entry",
			repository: &u1ListRepositoryMock{findByID: entities.ReconstituteU1ListEntry(id, "energy", "grant", "rule 1")},
			recorder:   &adminEventRecorderMock{},
			query:      inboundports.GetU1ListQuery{ID: id.String(), ActorUserID: "admin-1", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result outboundports.U1ListResult) {
				t.Helper()
				if result.EligibleOperationType != "grant" {
					t.Fatalf("result.EligibleOperationType = %q, want %q", result.EligibleOperationType, "grant")
				}
			},
		},
		{
			name:       "returns not found",
			repository: &u1ListRepositoryMock{},
			recorder:   &adminEventRecorderMock{},
			query:      inboundports.GetU1ListQuery{ID: id.String()},
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
			repository: &u1ListRepositoryMock{findByIDErr: errors.New("boom")},
			recorder:   &adminEventRecorderMock{},
			query:      inboundports.GetU1ListQuery{ID: id.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "finding u1 list entry by id") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewGetU1ListUseCase(tc.repository, tc.recorder)
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
