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

type createSectorPortStub struct {
	result  ports.SectorResult
	err     error
	command inboundports.CreateSectorCommand
}

func (s *createSectorPortStub) Execute(_ context.Context, command inboundports.CreateSectorCommand) (ports.SectorResult, error) {
	s.command = command
	return s.result, s.err
}

type getSectorPortStub struct {
	result ports.SectorResult
	err    error
}

func (s *getSectorPortStub) Execute(context.Context, inboundports.GetSectorQuery) (ports.SectorResult, error) {
	return s.result, s.err
}

type listSectorsPortStub struct {
	result ports.ListSectorsResult
	err    error
	query  inboundports.ListSectorsQuery
}

func (s *listSectorsPortStub) Execute(_ context.Context, query inboundports.ListSectorsQuery) (ports.ListSectorsResult, error) {
	s.query = query
	return s.result, s.err
}

type updateSectorPortStub struct {
	result  ports.SectorResult
	err     error
	command inboundports.UpdateSectorCommand
}

func (s *updateSectorPortStub) Execute(_ context.Context, command inboundports.UpdateSectorCommand) (ports.SectorResult, error) {
	s.command = command
	return s.result, s.err
}

type deleteSectorPortStub struct {
	result  inboundports.DeleteSectorResult
	err     error
	command inboundports.DeleteSectorCommand
}

func (s *deleteSectorPortStub) Execute(_ context.Context, command inboundports.DeleteSectorCommand) (inboundports.DeleteSectorResult, error) {
	s.command = command
	return s.result, s.err
}

// TestHttpSectorAdapterRoutes verifies the HTTP sector adapter routes behavior and the expected outcome asserted below.
func TestHttpSectorAdapterRoutes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		method        string
		target        string
		body          string
		createStub    *createSectorPortStub
		getStub       *getSectorPortStub
		listStub      *listSectorsPortStub
		updateStub    *updateSectorPortStub
		deleteStub    *deleteSectorPortStub
		wantStatus    int
		wantErrorCode string
		wantEmptyBody bool
		assert        func(t *testing.T, createStub *createSectorPortStub, listStub *listSectorsPortStub, updateStub *updateSectorPortStub, deleteStub *deleteSectorPortStub, payload map[string]any)
	}{
		{
			name:       "create sector success",
			method:     http.MethodPost,
			target:     "/api/v1/sectors",
			body:       `{"type":"High Emitting","name":"Steel","description":"Steel production"}`,
			createStub: &createSectorPortStub{result: ports.SectorResult{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", Type: "High Emitting", Name: "Steel", Description: "Steel production", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}},
			getStub:    &getSectorPortStub{},
			listStub:   &listSectorsPortStub{},
			updateStub: &updateSectorPortStub{},
			deleteStub: &deleteSectorPortStub{},
			wantStatus: http.StatusCreated,
			assert: func(t *testing.T, createStub *createSectorPortStub, _ *listSectorsPortStub, _ *updateSectorPortStub, _ *deleteSectorPortStub, payload map[string]any) {
				t.Helper()
				if createStub.command.Type != "High Emitting" {
					t.Fatalf("createStub.command.Type = %q, want %q", createStub.command.Type, "High Emitting")
				}
				if createStub.command.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("createStub.command.ActorUserID = %q, want %q", createStub.command.ActorUserID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
				dataPayload := payload["data"].(map[string]any)
				if dataPayload["name"] != "Steel" {
					t.Fatalf("dataPayload[name] = %v, want %q", dataPayload["name"], "Steel")
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
			name:       "get sector success",
			method:     http.MethodGet,
			target:     "/api/v1/sectors/0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e",
			createStub: &createSectorPortStub{},
			getStub:    &getSectorPortStub{result: ports.SectorResult{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", Type: "High Emitting", Name: "Steel", Description: "Steel production", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}},
			listStub:   &listSectorsPortStub{},
			updateStub: &updateSectorPortStub{},
			deleteStub: &deleteSectorPortStub{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "list sectors success",
			method:     http.MethodGet,
			target:     "/api/v1/sectors",
			createStub: &createSectorPortStub{},
			getStub:    &getSectorPortStub{},
			listStub:   &listSectorsPortStub{result: ports.ListSectorsResult{Sectors: []ports.SectorResult{{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", Type: "PA Aligned", Name: "Renewables", Description: "Renewable power", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}}}},
			updateStub: &updateSectorPortStub{},
			deleteStub: &deleteSectorPortStub{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "update sector success",
			method:     http.MethodPut,
			target:     "/api/v1/sectors/0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e",
			body:       `{"type":"PA Aligned","name":"Renewables","description":"Renewable power"}`,
			createStub: &createSectorPortStub{},
			getStub:    &getSectorPortStub{},
			listStub:   &listSectorsPortStub{},
			updateStub: &updateSectorPortStub{result: ports.SectorResult{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e", Type: "PA Aligned", Name: "Renewables", Description: "Renewable power", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002", CreatedAt: now, UpdatedAt: now}},
			deleteStub: &deleteSectorPortStub{},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *createSectorPortStub, _ *listSectorsPortStub, updateStub *updateSectorPortStub, _ *deleteSectorPortStub, payload map[string]any) {
				t.Helper()
				if updateStub.command.Description != "Renewable power" {
					t.Fatalf("updateStub.command.Description = %q, want %q", updateStub.command.Description, "Renewable power")
				}
				dataPayload := payload["data"].(map[string]any)
				if dataPayload["type"] != "PA Aligned" {
					t.Fatalf("dataPayload[type] = %v, want %q", dataPayload["type"], "PA Aligned")
				}
			},
		},
		{
			name:          "delete sector success",
			method:        http.MethodDelete,
			target:        "/api/v1/sectors/0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e",
			createStub:    &createSectorPortStub{},
			getStub:       &getSectorPortStub{},
			listStub:      &listSectorsPortStub{},
			updateStub:    &updateSectorPortStub{},
			deleteStub:    &deleteSectorPortStub{result: inboundports.DeleteSectorResult{ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e"}},
			wantStatus:    http.StatusNoContent,
			wantEmptyBody: true,
			assert: func(t *testing.T, _ *createSectorPortStub, _ *listSectorsPortStub, _ *updateSectorPortStub, deleteStub *deleteSectorPortStub, _ map[string]any) {
				t.Helper()
				if deleteStub.command.ID != "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e" {
					t.Fatalf("deleteStub.command.ID = %q, want %q", deleteStub.command.ID, "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e")
				}
			},
		},
		{
			name:          "maps not found error",
			method:        http.MethodGet,
			target:        "/api/v1/sectors/0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e",
			createStub:    &createSectorPortStub{},
			getStub:       &getSectorPortStub{err: &usecases.NotFoundError{Resource: "sector entry", ID: "0195e8f0-6b8d-7a11-8c47-9f1a2b3c4d5e"}},
			listStub:      &listSectorsPortStub{},
			updateStub:    &updateSectorPortStub{},
			deleteStub:    &deleteSectorPortStub{},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name:          "rejects malformed create body",
			method:        http.MethodPost,
			target:        "/api/v1/sectors",
			body:          `{"type":"High Emitting"`,
			createStub:    &createSectorPortStub{},
			getStub:       &getSectorPortStub{},
			listStub:      &listSectorsPortStub{},
			updateStub:    &updateSectorPortStub{},
			deleteStub:    &deleteSectorPortStub{},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "bad_request",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewHttpSectorAdapter(tc.createStub, tc.getStub, tc.listStub, tc.updateStub, tc.deleteStub)
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
