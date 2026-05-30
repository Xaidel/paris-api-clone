package usecases

import (
	"context"
	"encoding/json"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TestListAuditEventsUseCaseExecute verifies the list audit events use case execute behavior and the expected outcome asserted below.
func TestListAuditEventsUseCaseExecute(t *testing.T) {
	t.Parallel()

	eventID, err := valueobjects.EventIDFromString("018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1")
	if err != nil {
		t.Fatalf("EventIDFromString() error = %v", err)
	}

	event, err := entities.NewAdminEvent(eventID, testTime(), "admin-1", "group-1", "CreateUser", json.RawMessage(`{"resource":"user"}`))
	if err != nil {
		t.Fatalf("NewAdminEvent() error = %v", err)
	}

	repository := &adminEventRepositoryMock{listEvents: []*entities.AdminEvent{event}}
	useCase := NewListAuditEventsUseCase(repository)

	result, err := useCase.Execute(context.Background(), inboundports.ListAuditEventsQuery{EventOwner: "user", UserID: "admin-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(result.Events) != 1 {
		t.Fatalf("len(result.Events) = %d, want 1", len(result.Events))
	}

	if result.Events[0].ID != eventID.String() {
		t.Fatalf("result.Events[0].ID = %q, want %q", result.Events[0].ID, eventID.String())
	}
}
