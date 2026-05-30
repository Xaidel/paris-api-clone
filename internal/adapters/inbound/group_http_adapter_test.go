package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/infrastructure/httpserver"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	"go.uber.org/zap"
)

type createGroupPortStub struct {
	result  ports.GroupResult
	err     error
	command inboundports.CreateGroupCommand
}

func (s *createGroupPortStub) Execute(_ context.Context, command inboundports.CreateGroupCommand) (ports.GroupResult, error) {
	s.command = command
	return s.result, s.err
}

type getGroupPortStub struct {
	result ports.GroupResult
	err    error
}

func (s *getGroupPortStub) Execute(context.Context, inboundports.GetGroupQuery) (ports.GroupResult, error) {
	return s.result, s.err
}

type listGroupsPortStub struct {
	result ports.ListGroupsResult
	err    error
}

func (s *listGroupsPortStub) Execute(context.Context, inboundports.ListGroupsQuery) (ports.ListGroupsResult, error) {
	return s.result, s.err
}

type updateGroupPortStub struct {
	result  ports.GroupResult
	err     error
	command inboundports.UpdateGroupCommand
}

func (s *updateGroupPortStub) Execute(_ context.Context, command inboundports.UpdateGroupCommand) (ports.GroupResult, error) {
	s.command = command
	return s.result, s.err
}

type deleteGroupPortStub struct {
	result inboundports.DeleteGroupResult
	err    error
}

func (s *deleteGroupPortStub) Execute(context.Context, inboundports.DeleteGroupCommand) (inboundports.DeleteGroupResult, error) {
	return s.result, s.err
}

// TestHttpGroupAdapterRoutes verifies the HTTP group adapter routes behavior and the expected outcome asserted below.
func TestHttpGroupAdapterRoutes(t *testing.T) {
	t.Parallel()

	adapter := NewHttpGroupAdapter(
		&createGroupPortStub{result: ports.GroupResult{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300001", Name: "superadmin"}},
		&getGroupPortStub{result: ports.GroupResult{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300001", Name: "superadmin"}},
		&listGroupsPortStub{result: ports.ListGroupsResult{Groups: []ports.GroupResult{{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300001", Name: "superadmin"}}}},
		&updateGroupPortStub{result: ports.GroupResult{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300001", Name: "admins"}},
		&deleteGroupPortStub{result: inboundports.DeleteGroupResult{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300001"}},
	)

	router, err := httpserver.NewRouter(zap.NewNop(), adapter)
	if err != nil {
		t.Fatalf("httpserver.NewRouter() error = %v", err)
	}

	tests := []struct {
		name       string
		method     string
		target     string
		body       string
		wantStatus int
	}{
		{name: "create", method: http.MethodPost, target: "/api/v1/groups", body: `{"name":"superadmin"}`, wantStatus: http.StatusCreated},
		{name: "get", method: http.MethodGet, target: "/api/v1/groups/01962b8f-aeb2-7e03-a8ff-1edce1300001", wantStatus: http.StatusOK},
		{name: "list", method: http.MethodGet, target: "/api/v1/groups", wantStatus: http.StatusOK},
		{name: "update", method: http.MethodPut, target: "/api/v1/groups/01962b8f-aeb2-7e03-a8ff-1edce1300001", body: `{"name":"admins"}`, wantStatus: http.StatusOK},
		{name: "delete", method: http.MethodDelete, target: "/api/v1/groups/01962b8f-aeb2-7e03-a8ff-1edce1300001", wantStatus: http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.target, bytes.NewBufferString(tt.body))
			if tt.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			request.Header.Set(actorUserIDHeader, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
			request.Header.Set(actorGroupIDHeader, "01962b8f-aeb2-7e03-a8ff-1edce1300001")

			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status code = %d, want %d", response.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusNoContent {
				return
			}

			var payload map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if payload["data"] == nil {
				t.Fatal("expected data payload")
			}
		})
	}
}

// TestHttpGroupAdapterMapsNotFound verifies the HTTP group adapter maps not-found behavior and the expected outcome asserted below.
func TestHttpGroupAdapterMapsNotFound(t *testing.T) {
	t.Parallel()

	adapter := NewHttpGroupAdapter(&createGroupPortStub{}, &getGroupPortStub{err: &usecases.NotFoundError{Resource: "group", ID: "missing"}}, &listGroupsPortStub{}, &updateGroupPortStub{}, &deleteGroupPortStub{})
	router, err := httpserver.NewRouter(zap.NewNop(), adapter)
	if err != nil {
		t.Fatalf("httpserver.NewRouter() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/groups/missing", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusNotFound)
	}
}
