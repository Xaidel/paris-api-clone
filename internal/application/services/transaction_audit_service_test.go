package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	domainevents "github.com/gyud-adb/paris-api/internal/domain/events"
)

type eventPublisherStub struct {
	published bool
	err       error
	count     int
	lastType  string
	lastEvent *domainevents.AdminActionOccurred
}

func (s *eventPublisherStub) Publish(_ context.Context, events []domain.DomainEvent) error {
	s.published = true
	s.count = len(events)
	if len(events) > 0 {
		if event, ok := events[0].(*domainevents.AdminActionOccurred); ok {
			s.lastType = event.EventType()
			s.lastEvent = event
		}
	}
	return s.err
}

// TestTransactionAuditServiceRecordListTransactions verifies the transaction audit service record list transactions behavior and the expected outcome asserted below.
func TestTransactionAuditServiceRecordListTransactions(t *testing.T) {
	t.Parallel()

	publisher := &eventPublisherStub{}
	service := NewTransactionAuditService(publisher)
	service.now = func() time.Time { return time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC) }

	if err := service.RecordListTransactions(context.Background(), "admin-1", "group-1", "upload-1", 3); err != nil {
		t.Fatalf("RecordListTransactions() error = %v", err)
	}

	if !publisher.published {
		t.Fatal("expected events to be published")
	}

	if publisher.count != 1 {
		t.Fatalf("publisher.count = %d, want %d", publisher.count, 1)
	}

	if publisher.lastType != "ListTransactions" {
		t.Fatalf("publisher.lastType = %q, want %q", publisher.lastType, "ListTransactions")
	}

	if publisher.lastEvent == nil {
		t.Fatal("expected admin audit event")
	}

	var got map[string]any
	if err := json.Unmarshal(publisher.lastEvent.EventData(), &got); err != nil {
		t.Fatalf("json.Unmarshal(EventData()) error = %v", err)
	}

	if got["action"] != "read" {
		t.Fatalf("payload action = %v, want %q", got["action"], "read")
	}

	if got["resource"] != "transaction" {
		t.Fatalf("payload resource = %v, want %q", got["resource"], "transaction")
	}

	if got["upload_id"] != "upload-1" {
		t.Fatalf("payload upload_id = %v, want %q", got["upload_id"], "upload-1")
	}

	if got["result_count"] != float64(3) {
		t.Fatalf("payload result_count = %v, want %v", got["result_count"], float64(3))
	}
}
