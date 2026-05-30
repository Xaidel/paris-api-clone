package entities

import (
	"encoding/json"
	"testing"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TestNewAdminEvent verifies the new admin event behavior and the expected outcome asserted below.
func TestNewAdminEvent(t *testing.T) {
	t.Parallel()

	eventID, err := valueobjects.EventIDFromString("018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1")
	if err != nil {
		t.Fatalf("EventIDFromString() error = %v", err)
	}

	event, err := NewAdminEvent(eventID, time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC), "admin-1", "group-1", "CreateUser", json.RawMessage(`{"resource":"user"}`))
	if err != nil {
		t.Fatalf("NewAdminEvent() error = %v", err)
	}

	if event.EventType() != "CreateUser" {
		t.Fatalf("event.EventType() = %q, want %q", event.EventType(), "CreateUser")
	}
}

// TestNewAdminEventRejectsInvalidPayload verifies the new admin event rejects invalid payload behavior and the expected outcome asserted below.
func TestNewAdminEventRejectsInvalidPayload(t *testing.T) {
	t.Parallel()

	eventID, err := valueobjects.EventIDFromString("018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1")
	if err != nil {
		t.Fatalf("EventIDFromString() error = %v", err)
	}

	if _, err := NewAdminEvent(eventID, time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC), "admin-1", "group-1", "CreateUser", []byte(`{"resource":`)); err == nil {
		t.Fatal("expected invalid payload error")
	}
}
