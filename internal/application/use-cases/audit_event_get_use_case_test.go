package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

type adminEventRepositoryMock struct {
	createdEvent *entities.AdminEvent
	createErr    error
	findByID     *entities.AdminEvent
	findByIDErr  error
	listEvents   []*entities.AdminEvent
	listErr      error
}

func (m *adminEventRepositoryMock) Create(_ context.Context, event *entities.AdminEvent) error {
	m.createdEvent = event
	return m.createErr
}

func (m *adminEventRepositoryMock) FindByID(context.Context, valueobjects.EventID) (*entities.AdminEvent, error) {
	return m.findByID, m.findByIDErr
}

func (m *adminEventRepositoryMock) List(context.Context, outboundports.AuditEventFilter) ([]*entities.AdminEvent, error) {
	return m.listEvents, m.listErr
}

// TestGetAuditEventUseCaseExecute verifies the get audit event use case execute behavior and the expected outcome asserted below.
func TestGetAuditEventUseCaseExecute(t *testing.T) {
	t.Parallel()

	eventID, err := valueobjects.EventIDFromString("018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1")
	if err != nil {
		t.Fatalf("EventIDFromString() error = %v", err)
	}

	event, err := entities.NewAdminEvent(eventID, testTime(), "admin-1", "group-1", "CreateUser", json.RawMessage(`{"resource":"user"}`))
	if err != nil {
		t.Fatalf("NewAdminEvent() error = %v", err)
	}

	tests := []struct {
		name        string
		repository  *adminEventRepositoryMock
		query       inboundports.GetAuditEventQuery
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.AuditEventResult)
	}{
		{
			name:       "returns audit event",
			repository: &adminEventRepositoryMock{findByID: event},
			query:      inboundports.GetAuditEventQuery{ID: eventID.String()},
			assert: func(t *testing.T, result outboundports.AuditEventResult) {
				t.Helper()
				if result.ID != eventID.String() {
					t.Fatalf("result.ID = %q, want %q", result.ID, eventID.String())
				}
				if result.EventOwner != "user" {
					t.Fatalf("result.EventOwner = %q, want %q", result.EventOwner, "user")
				}
			},
		},
		{
			name:       "returns not found",
			repository: &adminEventRepositoryMock{},
			query:      inboundports.GetAuditEventQuery{ID: eventID.String()},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("error = %v, want NotFoundError", err)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			useCase := NewGetAuditEventUseCase(tt.repository)
			result, err := useCase.Execute(context.Background(), tt.query)
			if tt.assertError != nil {
				tt.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tt.assert(t, result)
		})
	}
}
