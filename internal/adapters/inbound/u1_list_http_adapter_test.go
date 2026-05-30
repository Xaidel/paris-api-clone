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

type createU1ListPortStub struct {
	result  ports.U1ListResult
	err     error
	command inboundports.CreateU1ListCommand
}

func (s *createU1ListPortStub) Execute(_ context.Context, command inboundports.CreateU1ListCommand) (ports.U1ListResult, error) {
	s.command = command
	return s.result, s.err
}

type getU1ListPortStub struct {
	result ports.U1ListResult
	err    error
}

func (s *getU1ListPortStub) Execute(context.Context, inboundports.GetU1ListQuery) (ports.U1ListResult, error) {
	return s.result, s.err
}

type listU1ListPortStub struct {
	result ports.ListU1ListResult
	err    error
	query  inboundports.ListU1ListQuery
}

func (s *listU1ListPortStub) Execute(_ context.Context, query inboundports.ListU1ListQuery) (ports.ListU1ListResult, error) {
	s.query = query
	return s.result, s.err
}

type updateU1ListPortStub struct {
	result  ports.U1ListResult
	err     error
	command inboundports.UpdateU1ListCommand
}

func (s *updateU1ListPortStub) Execute(_ context.Context, command inboundports.UpdateU1ListCommand) (ports.U1ListResult, error) {
	s.command = command
	return s.result, s.err
}

type deleteU1ListPortStub struct {
	result  inboundports.DeleteU1ListResult
	err     error
	command inboundports.DeleteU1ListCommand
}

func (s *deleteU1ListPortStub) Execute(_ context.Context, command inboundports.DeleteU1ListCommand) (inboundports.DeleteU1ListResult, error) {
	s.command = command
	return s.result, s.err
}

// TestHttpU1ListAdapterRoutes verifies the HTTP U1 list adapter routes behavior and the expected outcome asserted below.
func TestHttpU1ListAdapterRoutes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		method        string
		target        string
		body          string
		createStub    *createU1ListPortStub
		getStub       *getU1ListPortStub
		listStub      *listU1ListPortStub
		updateStub    *updateU1ListPortStub
		deleteStub    *deleteU1ListPortStub
		wantStatus    int
		wantErrorCode string
		wantEmptyBody bool
		assert        func(t *testing.T, createStub *createU1ListPortStub, listStub *listU1ListPortStub, updateStub *updateU1ListPortStub, deleteStub *deleteU1ListPortStub, payload map[string]any)
	}{
		{
			name:       "create entry success",
			method:     http.MethodPost,
			target:     "/api/v1/u1-list",
			body:       `{"sector":"energy","eligible_operation_type":"grant","condition_guidance":"rule 1"}`,
			createStub: &createU1ListPortStub{result: ports.U1ListResult{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", Sector: "energy", EligibleOperationType: "grant", ConditionGuidance: "rule 1", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}},
			getStub:    &getU1ListPortStub{},
			listStub:   &listU1ListPortStub{},
			updateStub: &updateU1ListPortStub{},
			deleteStub: &deleteU1ListPortStub{},
			wantStatus: http.StatusCreated,
			assert: func(t *testing.T, createStub *createU1ListPortStub, _ *listU1ListPortStub, _ *updateU1ListPortStub, _ *deleteU1ListPortStub, payload map[string]any) {
				t.Helper()
				if createStub.command.Sector != "energy" {
					t.Fatalf("createStub.command.Sector = %q, want %q", createStub.command.Sector, "energy")
				}
				if createStub.command.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("createStub.command.ActorUserID = %q, want %q", createStub.command.ActorUserID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
				dataPayload := payload["data"].(map[string]any)
				if dataPayload["eligible_operation_type"] != "grant" {
					t.Fatalf("dataPayload[eligible_operation_type] = %v, want %q", dataPayload["eligible_operation_type"], "grant")
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
			target:     "/api/v1/u1-list/0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e",
			createStub: &createU1ListPortStub{},
			getStub:    &getU1ListPortStub{result: ports.U1ListResult{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", Sector: "energy", EligibleOperationType: "grant", ConditionGuidance: "rule 1", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}},
			listStub:   &listU1ListPortStub{},
			updateStub: &updateU1ListPortStub{},
			deleteStub: &deleteU1ListPortStub{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "list entries success",
			method:     http.MethodGet,
			target:     "/api/v1/u1-list",
			createStub: &createU1ListPortStub{},
			getStub:    &getU1ListPortStub{},
			listStub:   &listU1ListPortStub{result: ports.ListU1ListResult{Entries: []ports.U1ListResult{{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", Sector: "energy", EligibleOperationType: "grant", ConditionGuidance: "rule 1", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}}}},
			updateStub: &updateU1ListPortStub{},
			deleteStub: &deleteU1ListPortStub{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "list entries filtered by sector",
			method:     http.MethodGet,
			target:     "/api/v1/u1-list?sector=EnErGy",
			createStub: &createU1ListPortStub{},
			getStub:    &getU1ListPortStub{},
			listStub:   &listU1ListPortStub{result: ports.ListU1ListResult{Entries: []ports.U1ListResult{{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", Sector: "energy", EligibleOperationType: "grant", ConditionGuidance: "rule 1", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}}}},
			updateStub: &updateU1ListPortStub{},
			deleteStub: &deleteU1ListPortStub{},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *createU1ListPortStub, listStub *listU1ListPortStub, _ *updateU1ListPortStub, _ *deleteU1ListPortStub, _ map[string]any) {
				t.Helper()
				if listStub.query.Sector != "EnErGy" {
					t.Fatalf("listStub.query.Sector = %q, want %q", listStub.query.Sector, "EnErGy")
				}
			},
		},
		{
			name:       "update entry success",
			method:     http.MethodPut,
			target:     "/api/v1/u1-list/0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e",
			body:       `{"sector":"agriculture","eligible_operation_type":"loan","condition_guidance":"rule 2"}`,
			createStub: &createU1ListPortStub{},
			getStub:    &getU1ListPortStub{},
			listStub:   &listU1ListPortStub{},
			updateStub: &updateU1ListPortStub{result: ports.U1ListResult{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", Sector: "agriculture", EligibleOperationType: "loan", ConditionGuidance: "rule 2", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}},
			deleteStub: &deleteU1ListPortStub{},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *createU1ListPortStub, _ *listU1ListPortStub, updateStub *updateU1ListPortStub, _ *deleteU1ListPortStub, payload map[string]any) {
				t.Helper()
				if updateStub.command.ConditionGuidance != "rule 2" {
					t.Fatalf("updateStub.command.ConditionGuidance = %q, want %q", updateStub.command.ConditionGuidance, "rule 2")
				}
				dataPayload := payload["data"].(map[string]any)
				if dataPayload["sector"] != "agriculture" {
					t.Fatalf("dataPayload[sector] = %v, want %q", dataPayload["sector"], "agriculture")
				}
			},
		},
		{
			name:          "delete entry success",
			method:        http.MethodDelete,
			target:        "/api/v1/u1-list/0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e",
			createStub:    &createU1ListPortStub{},
			getStub:       &getU1ListPortStub{},
			listStub:      &listU1ListPortStub{},
			updateStub:    &updateU1ListPortStub{},
			deleteStub:    &deleteU1ListPortStub{result: inboundports.DeleteU1ListResult{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e"}},
			wantStatus:    http.StatusNoContent,
			wantEmptyBody: true,
			assert: func(t *testing.T, _ *createU1ListPortStub, _ *listU1ListPortStub, _ *updateU1ListPortStub, deleteStub *deleteU1ListPortStub, _ map[string]any) {
				t.Helper()
				if deleteStub.command.ID != "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e" {
					t.Fatalf("deleteStub.command.ID = %q, want %q", deleteStub.command.ID, "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
				}
			},
		},
		{
			name:          "maps not found error",
			method:        http.MethodGet,
			target:        "/api/v1/u1-list/0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e",
			createStub:    &createU1ListPortStub{},
			getStub:       &getU1ListPortStub{err: &usecases.NotFoundError{Resource: "u1 list entry", ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e"}},
			listStub:      &listU1ListPortStub{},
			updateStub:    &updateU1ListPortStub{},
			deleteStub:    &deleteU1ListPortStub{},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name:          "rejects malformed create body",
			method:        http.MethodPost,
			target:        "/api/v1/u1-list",
			body:          `{"sector":"energy"`,
			createStub:    &createU1ListPortStub{},
			getStub:       &getU1ListPortStub{},
			listStub:      &listU1ListPortStub{},
			updateStub:    &updateU1ListPortStub{},
			deleteStub:    &deleteU1ListPortStub{},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "bad_request",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewHttpU1ListAdapter(tc.createStub, tc.getStub, tc.listStub, tc.updateStub, tc.deleteStub)
			router, err := httpserver.NewRouter(zap.NewNop(), adapter)
			if err != nil {
				t.Fatalf("httpserver.NewRouter() error = %v", err)
			}

			request := httptest.NewRequest(tc.method, tc.target, bytes.NewBufferString(tc.body))
			if tc.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			request.Header.Set(actorUserIDHeader, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
			request.Header.Set(actorGroupIDHeader, "01962b8f-aeb2-7e03-a8ff-1edce1300003")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tc.wantStatus {
				t.Fatalf("status code = %d, want %d", response.Code, tc.wantStatus)
			}

			if tc.wantEmptyBody {
				if response.Body.Len() != 0 {
					t.Fatalf("response body = %q, want empty body", response.Body.String())
				}
				if tc.assert != nil {
					tc.assert(t, tc.createStub, tc.listStub, tc.updateStub, tc.deleteStub, nil)
				}
				return
			}

			var payload map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if tc.wantErrorCode != "" {
				errorPayload := payload["error"].(map[string]any)
				if errorPayload["code"] != tc.wantErrorCode {
					t.Fatalf("errorPayload[code] = %v, want %q", errorPayload["code"], tc.wantErrorCode)
				}
				return
			}

			if tc.assert != nil {
				tc.assert(t, tc.createStub, tc.listStub, tc.updateStub, tc.deleteStub, payload)
			}
		})
	}
}
