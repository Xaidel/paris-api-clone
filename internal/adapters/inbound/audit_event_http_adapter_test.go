package adapters

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/infrastructure/httpserver"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	"go.uber.org/zap"
)

type listAuditEventsPortStub struct {
	result outboundports.ListAuditEventsResult
	err    error
	query  inboundports.ListAuditEventsQuery
}

func (s *listAuditEventsPortStub) Execute(_ context.Context, query inboundports.ListAuditEventsQuery) (outboundports.ListAuditEventsResult, error) {
	s.query = query
	return s.result, s.err
}

type getAuditEventPortStub struct {
	result outboundports.AuditEventResult
	err    error
	query  inboundports.GetAuditEventQuery
}

func (s *getAuditEventPortStub) Execute(_ context.Context, query inboundports.GetAuditEventQuery) (outboundports.AuditEventResult, error) {
	s.query = query
	return s.result, s.err
}

// TestHttpAuditEventAdapterRoutes verifies the HTTP audit event adapter routes behavior and the expected outcome asserted below.
func TestHttpAuditEventAdapterRoutes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	auditActorUserID := "01962b8f-aeb2-7e03-a8ff-1edce1300002"
	auditActorGroupID := "01962b8f-aeb2-7e03-a8ff-1edce1300003"
	tests := []struct {
		name          string
		target        string
		listStub      *listAuditEventsPortStub
		getStub       *getAuditEventPortStub
		wantStatus    int
		wantErrorCode string
		assert        func(t *testing.T, listStub *listAuditEventsPortStub, getStub *getAuditEventPortStub, payload map[string]any)
	}{
		{
			name:       "lists audit events",
			target:     "/api/v1/audit/events?event_owner=user&event_type=CreateUser&user_id=" + auditActorUserID + "&start_date=2026-04-01T00:00:00Z&end_date=2026-04-03T00:00:00Z",
			listStub:   &listAuditEventsPortStub{result: outboundports.ListAuditEventsResult{Events: []outboundports.AuditEventResult{{ID: "018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1", Timestamp: now, EventOwner: "user", EventType: "CreateUser", UserID: auditActorUserID, GroupID: auditActorGroupID, EventData: json.RawMessage(`{"resource":"user"}`)}}}},
			getStub:    &getAuditEventPortStub{},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, listStub *listAuditEventsPortStub, _ *getAuditEventPortStub, payload map[string]any) {
				t.Helper()
				if listStub.query.EventOwner != "user" {
					t.Fatalf("listStub.query.EventOwner = %q, want %q", listStub.query.EventOwner, "user")
				}
				dataPayload := payload["data"].(map[string]any)
				events := dataPayload["events"].([]any)
				if len(events) != 1 {
					t.Fatalf("len(events) = %d, want 1", len(events))
				}
			},
		},
		{
			name:       "gets audit event",
			target:     "/api/v1/audit/events/018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1",
			listStub:   &listAuditEventsPortStub{},
			getStub:    &getAuditEventPortStub{result: outboundports.AuditEventResult{ID: "018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1", Timestamp: now, EventOwner: "user", EventType: "CreateUser", UserID: auditActorUserID, GroupID: auditActorGroupID, EventData: json.RawMessage(`{"resource":"user"}`)}},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *listAuditEventsPortStub, getStub *getAuditEventPortStub, payload map[string]any) {
				t.Helper()
				if getStub.query.ID != "018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1" {
					t.Fatalf("getStub.query.ID = %q, want %q", getStub.query.ID, "018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1")
				}
				dataPayload := payload["data"].(map[string]any)
				if dataPayload["id"] != "018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1" {
					t.Fatalf("dataPayload[id] = %v, want %q", dataPayload["id"], "018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1")
				}
			},
		},
		{
			name:          "maps audit not found error",
			target:        "/api/v1/audit/events/018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1",
			listStub:      &listAuditEventsPortStub{},
			getStub:       &getAuditEventPortStub{err: &usecases.NotFoundError{Resource: "audit event", ID: "018f0f61-8d52-7cc0-bf0a-8b8d2d5b7ea1"}},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewHttpAuditEventAdapter(tt.listStub, tt.getStub)
			router, err := httpserver.NewRouter(zap.NewNop(), adapter)
			if err != nil {
				t.Fatalf("httpserver.NewRouter() error = %v", err)
			}

			request := httptest.NewRequest(http.MethodGet, tt.target, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status code = %d, want %d", response.Code, tt.wantStatus)
			}

			var payload map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if tt.wantErrorCode != "" {
				errorPayload := payload["error"].(map[string]any)
				if errorPayload["code"] != tt.wantErrorCode {
					t.Fatalf("error code = %v, want %q", errorPayload["code"], tt.wantErrorCode)
				}
				return
			}

			if tt.assert != nil {
				tt.assert(t, tt.listStub, tt.getStub, payload)
			}
		})
	}
}
