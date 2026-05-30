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

type createUserPortStub struct {
	result  ports.UserResult
	err     error
	command inboundports.CreateUserCommand
}

type getUserPortStub struct {
	result ports.UserResult
	err    error
}

func (s *getUserPortStub) Execute(context.Context, inboundports.GetUserQuery) (ports.UserResult, error) {
	return s.result, s.err
}

type listUsersPortStub struct {
	result inboundports.ListUsersResult
	err    error
}

func (s *listUsersPortStub) Execute(context.Context, inboundports.ListUsersQuery) (inboundports.ListUsersResult, error) {
	return s.result, s.err
}

type updateUserPortStub struct {
	result  ports.UserResult
	err     error
	command inboundports.UpdateUserCommand
}

func (s *createUserPortStub) Execute(_ context.Context, command inboundports.CreateUserCommand) (ports.UserResult, error) {
	s.command = command
	return s.result, s.err
}

func (s *updateUserPortStub) Execute(_ context.Context, command inboundports.UpdateUserCommand) (ports.UserResult, error) {
	s.command = command
	return s.result, s.err
}

type deleteUserPortStub struct {
	result inboundports.DeleteUserResult
	err    error
}

func (s *deleteUserPortStub) Execute(context.Context, inboundports.DeleteUserCommand) (inboundports.DeleteUserResult, error) {
	return s.result, s.err
}

// TestInfrastructureRouterHealthz verifies the infrastructure router healthz behavior and the expected outcome asserted below.
func TestInfrastructureRouterHealthz(t *testing.T) {
	t.Parallel()

	router, err := httpserver.NewRouter(zap.NewNop())
	if err != nil {
		t.Fatalf("httpserver.NewRouter() error = %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}
}

// TestHttpUserAdapterRoutes verifies the HTTP user adapter routes behavior and the expected outcome asserted below.
func TestHttpUserAdapterRoutes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name          string
		method        string
		target        string
		body          string
		headers       map[string]string
		createStub    *createUserPortStub
		getStub       *getUserPortStub
		listStub      *listUsersPortStub
		updateStub    *updateUserPortStub
		deleteStub    *deleteUserPortStub
		wantStatus    int
		wantErrorCode string
		wantEmptyBody bool
		wantID        string
		wantUsername  string
		assert        func(t *testing.T, createStub *createUserPortStub, updateStub *updateUserPortStub)
	}{
		{name: "create user success", method: http.MethodPost, target: "/api/v1/users", body: `{"username":"alice","password":"supersecret","firstname":"Alice","middlename":null,"lastname":"Admin","group_id":"01962b8f-aeb2-7e03-a8ff-1edce1300001"}`, headers: map[string]string{actorUserIDHeader: "01962b8f-aeb2-7e03-a8ff-1edce1300002", actorGroupIDHeader: "01962b8f-aeb2-7e03-a8ff-1edce1300001"}, createStub: &createUserPortStub{result: ports.UserResult{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", Username: "alice", FirstName: "Alice", LastName: "Admin", GroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001", CreatedAt: now, UpdatedAt: now}}, getStub: &getUserPortStub{}, listStub: &listUsersPortStub{}, updateStub: &updateUserPortStub{}, deleteStub: &deleteUserPortStub{}, wantStatus: http.StatusCreated, wantID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", wantUsername: "alice", assert: func(t *testing.T, createStub *createUserPortStub, _ *updateUserPortStub) {
			t.Helper()
			if createStub.command.Username != "alice" || createStub.command.Password != "supersecret" || createStub.command.FirstName != "Alice" || createStub.command.LastName != "Admin" || createStub.command.GroupID != "01962b8f-aeb2-7e03-a8ff-1edce1300001" {
				t.Fatalf("create command = %+v", createStub.command)
			}
			if createStub.command.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" || createStub.command.ActorGroupID != "01962b8f-aeb2-7e03-a8ff-1edce1300001" {
				t.Fatalf("create command actor fields = %+v", createStub.command)
			}
		}},
		{name: "get user success", method: http.MethodGet, target: "/api/v1/users/01962b8f-aeb2-7e03-a8ff-1edce1300002", headers: map[string]string{actorUserIDHeader: "01962b8f-aeb2-7e03-a8ff-1edce1300002", actorGroupIDHeader: "01962b8f-aeb2-7e03-a8ff-1edce1300001"}, createStub: &createUserPortStub{}, getStub: &getUserPortStub{result: ports.UserResult{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", Username: "alice", FirstName: "Alice", LastName: "Admin", GroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001", CreatedAt: now, UpdatedAt: now}}, listStub: &listUsersPortStub{}, updateStub: &updateUserPortStub{}, deleteStub: &deleteUserPortStub{}, wantStatus: http.StatusOK, wantID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", wantUsername: "alice"},
		{name: "list users success", method: http.MethodGet, target: "/api/v1/users", headers: map[string]string{actorUserIDHeader: "01962b8f-aeb2-7e03-a8ff-1edce1300002", actorGroupIDHeader: "01962b8f-aeb2-7e03-a8ff-1edce1300001"}, createStub: &createUserPortStub{}, getStub: &getUserPortStub{}, listStub: &listUsersPortStub{result: inboundports.ListUsersResult{Users: []ports.UserResult{{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", Username: "alice", FirstName: "Alice", LastName: "Admin", GroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001", CreatedAt: now, UpdatedAt: now}}}}, updateStub: &updateUserPortStub{}, deleteStub: &deleteUserPortStub{}, wantStatus: http.StatusOK},
		{name: "update user success", method: http.MethodPut, target: "/api/v1/users/01962b8f-aeb2-7e03-a8ff-1edce1300002", body: `{"username":"alice","password":"supersecret","firstname":"Alice","middlename":null,"lastname":"Admin","group_id":"01962b8f-aeb2-7e03-a8ff-1edce1300001"}`, headers: map[string]string{actorUserIDHeader: "01962b8f-aeb2-7e03-a8ff-1edce1300002", actorGroupIDHeader: "01962b8f-aeb2-7e03-a8ff-1edce1300001"}, createStub: &createUserPortStub{}, getStub: &getUserPortStub{}, listStub: &listUsersPortStub{}, updateStub: &updateUserPortStub{result: ports.UserResult{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", Username: "alice", FirstName: "Alice", LastName: "Admin", GroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001", CreatedAt: now, UpdatedAt: now}}, deleteStub: &deleteUserPortStub{}, wantStatus: http.StatusOK, wantID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", wantUsername: "alice", assert: func(t *testing.T, _ *createUserPortStub, updateStub *updateUserPortStub) {
			t.Helper()
			if updateStub.command.ID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" || updateStub.command.Username != "alice" || updateStub.command.Password != "supersecret" || updateStub.command.FirstName != "Alice" || updateStub.command.LastName != "Admin" || updateStub.command.GroupID != "01962b8f-aeb2-7e03-a8ff-1edce1300001" {
				t.Fatalf("update command = %+v", updateStub.command)
			}
			if updateStub.command.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" || updateStub.command.ActorGroupID != "01962b8f-aeb2-7e03-a8ff-1edce1300001" {
				t.Fatalf("update command actor fields = %+v", updateStub.command)
			}
		}},
		{name: "delete user success", method: http.MethodDelete, target: "/api/v1/users/01962b8f-aeb2-7e03-a8ff-1edce1300002", headers: map[string]string{actorUserIDHeader: "01962b8f-aeb2-7e03-a8ff-1edce1300002", actorGroupIDHeader: "01962b8f-aeb2-7e03-a8ff-1edce1300001"}, createStub: &createUserPortStub{}, getStub: &getUserPortStub{}, listStub: &listUsersPortStub{}, updateStub: &updateUserPortStub{}, deleteStub: &deleteUserPortStub{result: inboundports.DeleteUserResult{ID: "01962b8f-aeb2-7e03-a8ff-1edce1300002"}}, wantStatus: http.StatusNoContent, wantEmptyBody: true},
		{name: "maps not found error", method: http.MethodGet, target: "/api/v1/users/01962b8f-aeb2-7e03-a8ff-1edce1300002", createStub: &createUserPortStub{}, getStub: &getUserPortStub{err: &usecases.NotFoundError{Resource: "user", ID: "01962b8f-aeb2-7e03-a8ff-1edce1300002"}}, listStub: &listUsersPortStub{}, updateStub: &updateUserPortStub{}, deleteStub: &deleteUserPortStub{}, wantStatus: http.StatusNotFound, wantErrorCode: "not_found"},
		{name: "rejects malformed create body", method: http.MethodPost, target: "/api/v1/users", body: `{"username":"alice"`, createStub: &createUserPortStub{}, getStub: &getUserPortStub{}, listStub: &listUsersPortStub{}, updateStub: &updateUserPortStub{}, deleteStub: &deleteUserPortStub{}, wantStatus: http.StatusBadRequest, wantErrorCode: "bad_request"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewHttpUserAdapter(tt.createStub, tt.getStub, tt.listStub, tt.updateStub, tt.deleteStub)
			router, err := httpserver.NewRouter(zap.NewNop(), adapter)
			if err != nil {
				t.Fatalf("httpserver.NewRouter() error = %v", err)
			}

			request := httptest.NewRequest(tt.method, tt.target, bytes.NewBufferString(tt.body))
			if tt.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			for key, value := range tt.headers {
				request.Header.Set(key, value)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status code = %d, want %d", response.Code, tt.wantStatus)
			}

			if tt.wantEmptyBody {
				if response.Body.Len() != 0 {
					t.Fatalf("response body = %q, want empty body", response.Body.String())
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
					t.Fatalf("error code = %v, want %q", errorPayload["code"], tt.wantErrorCode)
				}
				return
			}

			if tt.wantID != "" {
				dataPayload := payload["data"].(map[string]any)
				if deleteID, exists := dataPayload["id"]; exists {
					if deleteID != tt.wantID {
						t.Fatalf("id = %v, want %q", deleteID, tt.wantID)
					}
				} else if dataPayload["ID"] != tt.wantID {
					t.Fatalf("ID = %v, want %q", dataPayload["ID"], tt.wantID)
				}

				if tt.wantUsername != "" && dataPayload["username"] != tt.wantUsername {
					t.Fatalf("username = %v, want %q", dataPayload["username"], tt.wantUsername)
				}
			}

			if tt.assert != nil {
				tt.assert(t, tt.createStub, tt.updateStub)
			}
		})
	}
}
