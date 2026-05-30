package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

type adminEventRepositoryStub struct {
	createdEvent *entities.AdminEvent
	createErr    error
}

func (s *adminEventRepositoryStub) Create(_ context.Context, event *entities.AdminEvent) error {
	s.createdEvent = event
	return s.createErr
}

func (s *adminEventRepositoryStub) FindByID(context.Context, valueobjects.EventID) (*entities.AdminEvent, error) {
	return nil, nil
}

func (s *adminEventRepositoryStub) List(context.Context, ports.AuditEventFilter) ([]*entities.AdminEvent, error) {
	return nil, nil
}

type adminEventOutboxRepositoryStub struct {
	createdEvent *entities.AdminEvent
	createErr    error
}

func (s *adminEventOutboxRepositoryStub) Create(_ context.Context, event *entities.AdminEvent) error {
	s.createdEvent = event
	return s.createErr
}

type actorDirectoryServiceStub struct {
	err error
}

func (s *actorDirectoryServiceStub) ActorExists(context.Context, string, string) error {
	return s.err
}

// TestEventRecorderServiceRecordAdminEvent verifies the event recorder service record admin event behavior and the expected outcome asserted below.
func TestEventRecorderServiceRecordAdminEvent(t *testing.T) {
	t.Parallel()

	eventID, err := valueobjects.EventIDFromString("018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1")
	if err != nil {
		t.Fatalf("EventIDFromString() error = %v", err)
	}

	tests := []struct {
		name        string
		repository  *adminEventRepositoryStub
		outbox      *adminEventOutboxRepositoryStub
		command     RecordAdminEventCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, repository *adminEventRepositoryStub, outbox *adminEventOutboxRepositoryStub)
	}{
		{
			name:       "records admin event and outbox",
			repository: &adminEventRepositoryStub{},
			outbox:     &adminEventOutboxRepositoryStub{},
			command:    RecordAdminEventCommand{ActorUserID: "admin-1", ActorGroupID: "group-1", EventType: "CreateUser", EventData: json.RawMessage(`{"resource":"user"}`)},
			assert: func(t *testing.T, repository *adminEventRepositoryStub, outbox *adminEventOutboxRepositoryStub) {
				t.Helper()
				if repository.createdEvent == nil {
					t.Fatal("expected admin event to be created")
				}
				if outbox.createdEvent == nil {
					t.Fatal("expected outbox event to be created")
				}
				if !repository.createdEvent.ID().Equal(outbox.createdEvent.ID()) {
					t.Fatal("expected outbox event to reuse admin event id")
				}
			},
		},
		{
			name:       "wraps outbox errors",
			repository: &adminEventRepositoryStub{},
			outbox:     &adminEventOutboxRepositoryStub{createErr: errors.New("boom")},
			command:    RecordAdminEventCommand{ActorUserID: "admin-1", ActorGroupID: "group-1", EventType: "CreateUser", EventData: json.RawMessage(`{"resource":"user"}`)},
			assertError: func(t *testing.T, err error) {
				t.Helper()
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), "creating admin event outbox record") {
					t.Fatalf("err.Error() = %q, want substring %q", err.Error(), "creating admin event outbox record")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			service := NewEventRecorderService(tt.repository, tt.outbox, &actorDirectoryServiceStub{})
			service.now = func() time.Time { return time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC) }
			service.newEventID = func() (valueobjects.EventID, error) { return eventID, nil }

			err := service.RecordAdminEvent(context.Background(), tt.command)
			if tt.assertError != nil {
				tt.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("RecordAdminEvent() error = %v", err)
			}

			tt.assert(t, tt.repository, tt.outbox)
		})
	}
}
