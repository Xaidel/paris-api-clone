package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/infrastructure/httpserver"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	"go.uber.org/zap"
)

type createExclusionListPortStub struct {
	result  ports.ExclusionListResult
	err     error
	command inboundports.CreateExclusionListCommand
}

func (s *createExclusionListPortStub) Execute(_ context.Context, command inboundports.CreateExclusionListCommand) (ports.ExclusionListResult, error) {
	s.command = command
	return s.result, s.err
}

type getExclusionListPortStub struct {
	result ports.ExclusionListResult
	err    error
}

func (s *getExclusionListPortStub) Execute(context.Context, inboundports.GetExclusionListQuery) (ports.ExclusionListResult, error) {
	return s.result, s.err
}

type listExclusionListPortStub struct {
	result ports.ListExclusionListResult
	err    error
}

func (s *listExclusionListPortStub) Execute(context.Context, inboundports.ListExclusionListQuery) (ports.ListExclusionListResult, error) {
	return s.result, s.err
}

type updateExclusionListPortStub struct {
	result  ports.ExclusionListResult
	err     error
	command inboundports.UpdateExclusionListCommand
}

func (s *updateExclusionListPortStub) Execute(_ context.Context, command inboundports.UpdateExclusionListCommand) (ports.ExclusionListResult, error) {
	s.command = command
	return s.result, s.err
}

type deleteExclusionListPortStub struct {
	result  inboundports.DeleteExclusionListResult
	err     error
	command inboundports.DeleteExclusionListCommand
}

func (s *deleteExclusionListPortStub) Execute(_ context.Context, command inboundports.DeleteExclusionListCommand) (inboundports.DeleteExclusionListResult, error) {
	s.command = command
	return s.result, s.err
}

// TestHttpExclusionListAdapterRoutes verifies the HTTP exclusion list adapter routes behavior and the expected outcome asserted below.
func TestHttpExclusionListAdapterRoutes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		method        string
		target        string
		body          string
		createStub    *createExclusionListPortStub
		getStub       *getExclusionListPortStub
		listStub      *listExclusionListPortStub
		updateStub    *updateExclusionListPortStub
		deleteStub    *deleteExclusionListPortStub
		wantStatus    int
		wantErrorCode string
		wantEmptyBody bool
		assert        func(t *testing.T, createStub *createExclusionListPortStub, updateStub *updateExclusionListPortStub, deleteStub *deleteExclusionListPortStub, payload map[string]any)
	}{
		{
			name:       "create entry success",
			method:     http.MethodPost,
			target:     "/api/v1/u2-exclusion-list",
			body:       `{"activity_type":"agriculture"}`,
			createStub: &createExclusionListPortStub{result: ports.ExclusionListResult{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", ActivityType: "agriculture", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}},
			getStub:    &getExclusionListPortStub{},
			listStub:   &listExclusionListPortStub{},
			updateStub: &updateExclusionListPortStub{},
			deleteStub: &deleteExclusionListPortStub{},
			wantStatus: http.StatusCreated,
			assert: func(t *testing.T, createStub *createExclusionListPortStub, _ *updateExclusionListPortStub, _ *deleteExclusionListPortStub, payload map[string]any) {
				t.Helper()
				if createStub.command.ActivityType != "agriculture" {
					t.Fatalf("createStub.command.ActivityType = %q, want %q", createStub.command.ActivityType, "agriculture")
				}
				if createStub.command.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("createStub.command.ActorUserID = %q, want %q", createStub.command.ActorUserID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
				dataPayload := payload["data"].(map[string]any)
				if dataPayload["activity_type"] != "agriculture" {
					t.Fatalf("dataPayload[activity_type] = %v, want %q", dataPayload["activity_type"], "agriculture")
				}
				if dataPayload["created_by"] != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("dataPayload[created_by] = %v, want %q", dataPayload["created_by"], "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
				if dataPayload["created_at"] == nil || dataPayload["updated_at"] == nil {
					t.Fatalf("expected created_at and updated_at in response, got %v", dataPayload)
				}
			},
		},
		{
			name:       "get entry success",
			method:     http.MethodGet,
			target:     "/api/v1/u2-exclusion-list/0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e",
			createStub: &createExclusionListPortStub{},
			getStub:    &getExclusionListPortStub{result: ports.ExclusionListResult{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", ActivityType: "agriculture", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}},
			listStub:   &listExclusionListPortStub{},
			updateStub: &updateExclusionListPortStub{},
			deleteStub: &deleteExclusionListPortStub{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "list entries success",
			method:     http.MethodGet,
			target:     "/api/v1/u2-exclusion-list",
			createStub: &createExclusionListPortStub{},
			getStub:    &getExclusionListPortStub{},
			listStub:   &listExclusionListPortStub{result: ports.ListExclusionListResult{Entries: []ports.ExclusionListResult{{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", ActivityType: "agriculture", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}}}},
			updateStub: &updateExclusionListPortStub{},
			deleteStub: &deleteExclusionListPortStub{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "update entry success",
			method:     http.MethodPut,
			target:     "/api/v1/u2-exclusion-list/0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e",
			body:       `{"activity_type":"energy"}`,
			createStub: &createExclusionListPortStub{},
			getStub:    &getExclusionListPortStub{},
			listStub:   &listExclusionListPortStub{},
			updateStub: &updateExclusionListPortStub{result: ports.ExclusionListResult{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", ActivityType: "energy", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}},
			deleteStub: &deleteExclusionListPortStub{},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *createExclusionListPortStub, updateStub *updateExclusionListPortStub, _ *deleteExclusionListPortStub, payload map[string]any) {
				t.Helper()
				if updateStub.command.ActivityType != "energy" {
					t.Fatalf("updateStub.command.ActivityType = %q, want %q", updateStub.command.ActivityType, "energy")
				}
				dataPayload := payload["data"].(map[string]any)
				if dataPayload["activity_type"] != "energy" {
					t.Fatalf("dataPayload[activity_type] = %v, want %q", dataPayload["activity_type"], "energy")
				}
			},
		},
		{
			name:          "delete entry success",
			method:        http.MethodDelete,
			target:        "/api/v1/u2-exclusion-list/0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e",
			createStub:    &createExclusionListPortStub{},
			getStub:       &getExclusionListPortStub{},
			listStub:      &listExclusionListPortStub{},
			updateStub:    &updateExclusionListPortStub{},
			deleteStub:    &deleteExclusionListPortStub{result: inboundports.DeleteExclusionListResult{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e"}},
			wantStatus:    http.StatusNoContent,
			wantEmptyBody: true,
			assert: func(t *testing.T, _ *createExclusionListPortStub, _ *updateExclusionListPortStub, deleteStub *deleteExclusionListPortStub, _ map[string]any) {
				t.Helper()
				if deleteStub.command.ID != "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e" {
					t.Fatalf("deleteStub.command.ID = %q, want %q", deleteStub.command.ID, "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
				}
			},
		},
		{
			name:          "maps not found error",
			method:        http.MethodGet,
			target:        "/api/v1/u2-exclusion-list/0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e",
			createStub:    &createExclusionListPortStub{},
			getStub:       &getExclusionListPortStub{err: &usecases.NotFoundError{Resource: "exclusion list entry", ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e"}},
			listStub:      &listExclusionListPortStub{},
			updateStub:    &updateExclusionListPortStub{},
			deleteStub:    &deleteExclusionListPortStub{},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name:          "rejects malformed create body",
			method:        http.MethodPost,
			target:        "/api/v1/u2-exclusion-list",
			body:          `{"activity_type":"agriculture"`,
			createStub:    &createExclusionListPortStub{},
			getStub:       &getExclusionListPortStub{},
			listStub:      &listExclusionListPortStub{},
			updateStub:    &updateExclusionListPortStub{},
			deleteStub:    &deleteExclusionListPortStub{},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "bad_request",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewHttpExclusionListAdapter(tt.createStub, tt.getStub, tt.listStub, tt.updateStub, tt.deleteStub)
			router, err := httpserver.NewRouter(zap.NewNop(), adapter)
			if err != nil {
				t.Fatalf("httpserver.NewRouter() error = %v", err)
			}

			request := httptest.NewRequest(tt.method, tt.target, bytes.NewBufferString(tt.body))
			if tt.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			request.Header.Set(actorUserIDHeader, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
			request.Header.Set(actorGroupIDHeader, "01962b8f-aeb2-7e03-a8ff-1edce1300003")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status code = %d, want %d", response.Code, tt.wantStatus)
			}

			if tt.wantEmptyBody {
				if response.Body.Len() != 0 {
					t.Fatalf("response body = %q, want empty body", response.Body.String())
				}
				if tt.assert != nil {
					tt.assert(t, tt.createStub, tt.updateStub, tt.deleteStub, nil)
				}
				return
			}

			var payload map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if tt.wantErrorCode != "" {
				errorPayload := payload["error"].(map[string]any)
				if errorPayload["code"] != tt.wantErrorCode {
					t.Fatalf("errorPayload[code] = %v, want %q", errorPayload["code"], tt.wantErrorCode)
				}
				return
			}

			if tt.assert != nil {
				tt.assert(t, tt.createStub, tt.updateStub, tt.deleteStub, payload)
			}
		})
	}
}
