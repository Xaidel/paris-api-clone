package adapters

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/pashagolub/pgxmock/v4"
)

const (
	adminEventActorUserID  = "01962b8f-aeb2-7e03-a8ff-1edce1300002"
	adminEventActorGroupID = "01962b8f-aeb2-7e03-a8ff-1edce1300003"
)

// TestPostgresAdminEventRepositoryCreate verifies the Postgres admin event repository create behavior and the expected outcome asserted below.
func TestPostgresAdminEventRepositoryCreate(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	eventID, err := valueobjects.EventIDFromString("018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1")
	if err != nil {
		t.Fatalf("EventIDFromString() error = %v", err)
	}

	event, err := entities.NewAdminEvent(eventID, time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC), adminEventActorUserID, adminEventActorGroupID, "CreateUser", json.RawMessage(`{"resource":"user"}`))
	if err != nil {
		t.Fatalf("NewAdminEvent() error = %v", err)
	}

	repository := NewPostgresAdminEventRepository(mock)
	mock.ExpectExec(regexp.QuoteMeta(createAdminEventQuery)).
		WithArgs(event.ID().String(), event.OccurredAt(), event.UserID(), event.GroupID(), event.EventType(), string(event.EventData())).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Create(context.Background(), event); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

// TestPostgresAdminEventRepositoryFindByIDAndList verifies the Postgres admin event repository find-by-ID and list behavior and the expected outcome asserted below.
func TestPostgresAdminEventRepositoryFindByIDAndList(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	eventID, err := valueobjects.EventIDFromString("018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1")
	if err != nil {
		t.Fatalf("EventIDFromString() error = %v", err)
	}

	repository := NewPostgresAdminEventRepository(mock)
	mock.ExpectQuery(regexp.QuoteMeta(findAdminEventByIDQuery)).WithArgs(eventID.String()).WillReturnRows(
		pgxmock.NewRows([]string{"id", "timestamp", "user_id", "group_id", "event_type", "event_data"}).AddRow(eventID.String(), now, adminEventActorUserID, adminEventActorGroupID, "CreateUser", []byte(`{"resource":"user"}`)),
	)

	event, err := repository.FindByID(context.Background(), eventID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}

	if event == nil {
		t.Fatal("expected event")
	}

	filter := ports.AuditEventFilter{EventOwner: "user", EventType: "CreateUser", UserID: adminEventActorUserID, StartedAt: &now, EndedAt: &now}
	listQuery, args := buildListAdminEventsQuery(filter)
	mock.ExpectQuery(regexp.QuoteMeta(listQuery)).WithArgs(args...).WillReturnRows(
		pgxmock.NewRows([]string{"id", "timestamp", "user_id", "group_id", "event_type", "event_data"}).AddRow(eventID.String(), now, adminEventActorUserID, adminEventActorGroupID, "CreateUser", []byte(`{"resource":"user"}`)),
	)

	events, err := repository.List(context.Background(), filter)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
}

// TestPostgresAdminEventOutboxRepositoryCreate verifies the Postgres admin event outbox repository create behavior and the expected outcome asserted below.
func TestPostgresAdminEventOutboxRepositoryCreate(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	eventID, err := valueobjects.EventIDFromString("018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1")
	if err != nil {
		t.Fatalf("EventIDFromString() error = %v", err)
	}

	event, err := entities.NewAdminEvent(eventID, time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC), adminEventActorUserID, adminEventActorGroupID, "CreateUser", json.RawMessage(`{"resource":"user"}`))
	if err != nil {
		t.Fatalf("NewAdminEvent() error = %v", err)
	}

	repository := NewPostgresAdminEventOutboxRepository(mock)
	mock.ExpectExec(regexp.QuoteMeta(createAdminEventOutboxQuery)).
		WithArgs(event.ID().String(), event.EventType(), string(event.EventData()), event.OccurredAt()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Create(context.Background(), event); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}
