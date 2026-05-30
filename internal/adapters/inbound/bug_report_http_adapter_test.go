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

type createReportPortStub struct {
	result  ports.BugReportResult
	err     error
	command inboundports.CreateBugReportCommand
}

func (s *createReportPortStub) Execute(_ context.Context, command inboundports.CreateBugReportCommand) (ports.BugReportResult, error) {
	s.command = command
	return s.result, s.err
}

type getReportPortStub struct {
	result ports.BugReportResult
	err    error
}

func (s *getReportPortStub) Execute(context.Context, inboundports.GetBugReportQuery) (ports.BugReportResult, error) {
	return s.result, s.err
}

type listReportsPortStub struct {
	result ports.ListBugReportsResult
	err    error
}

func (s *listReportsPortStub) Execute(context.Context, inboundports.ListBugReportsQuery) (ports.ListBugReportsResult, error) {
	return s.result, s.err
}

type updateReportPortStub struct {
	result  ports.BugReportResult
	err     error
	command inboundports.UpdateBugReportCommand
}

func (s *updateReportPortStub) Execute(_ context.Context, command inboundports.UpdateBugReportCommand) (ports.BugReportResult, error) {
	s.command = command
	return s.result, s.err
}

type deleteReportPortStub struct {
	result inboundports.DeleteBugReportResult
	err    error
}

func (s *deleteReportPortStub) Execute(context.Context, inboundports.DeleteBugReportCommand) (inboundports.DeleteBugReportResult, error) {
	return s.result, s.err
}

// TestHttpReportAdapterRoutes verifies the HTTP report adapter routes behavior and the expected outcome asserted below.
func TestHttpReportAdapterRoutes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	reportID := "01962b8f-aeb2-7e03-a8ff-1edce1300002"
	transactionID := "01962b8f-aeb2-7e03-a8ff-1edce1300001"
	actorUserID := "01962b8f-aeb2-7e03-a8ff-1edce1300002"

	fixedResult := ports.BugReportResult{
		ID:            reportID,
		UserID:        actorUserID,
		TransactionID: transactionID,
		Title:         "Incorrect classification",
		Description:   "Transaction was incorrectly classified",
		Status:        "Open",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	tests := []struct {
		name          string
		method        string
		target        string
		body          string
		createStub    *createReportPortStub
		getStub       *getReportPortStub
		listStub      *listReportsPortStub
		updateStub    *updateReportPortStub
		deleteStub    *deleteReportPortStub
		wantStatus    int
		wantErrorCode string
		wantEmptyBody bool
		assert        func(t *testing.T, createStub *createReportPortStub, updateStub *updateReportPortStub, deleteStub *deleteReportPortStub, payload map[string]any)
	}{
		{
			name:       "create bug report success",
			method:     http.MethodPost,
			target:     "/api/v1/bug-reports",
			body:       `{"transaction_id":"01962b8f-aeb2-7e03-a8ff-1edce1300001","title":"Incorrect classification","description":"Transaction was incorrectly classified"}`,
			createStub: &createReportPortStub{result: fixedResult},
			getStub:    &getReportPortStub{},
			listStub:   &listReportsPortStub{},
			updateStub: &updateReportPortStub{},
			deleteStub: &deleteReportPortStub{},
			wantStatus: http.StatusCreated,
			assert: func(t *testing.T, createStub *createReportPortStub, _ *updateReportPortStub, _ *deleteReportPortStub, payload map[string]any) {
				t.Helper()

				if createStub.command.TransactionID.String() != transactionID {
					t.Fatalf("command.TransactionID = %q, want %q", createStub.command.TransactionID, transactionID)
				}

				if createStub.command.ActorUserID != actorUserID {
					t.Fatalf("command.ActorUserID = %q, want %q", createStub.command.ActorUserID, actorUserID)
				}

				dataPayload := payload["data"].(map[string]any)
				if dataPayload["user_id"] != actorUserID {
					t.Fatalf("data.user_id = %v, want %q", dataPayload["user_id"], actorUserID)
				}

				if dataPayload["status"] != "Open" {
					t.Fatalf("data.status = %v, want %q", dataPayload["status"], "Open")
				}
			},
		},
		{
			name:       "get bug report success",
			method:     http.MethodGet,
			target:     "/api/v1/bug-reports/" + reportID,
			createStub: &createReportPortStub{},
			getStub:    &getReportPortStub{result: fixedResult},
			listStub:   &listReportsPortStub{},
			updateStub: &updateReportPortStub{},
			deleteStub: &deleteReportPortStub{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "list bug reports success",
			method:     http.MethodGet,
			target:     "/api/v1/bug-reports",
			createStub: &createReportPortStub{},
			getStub:    &getReportPortStub{},
			listStub:   &listReportsPortStub{result: ports.ListBugReportsResult{BugReports: []ports.BugReportResult{fixedResult}}},
			updateStub: &updateReportPortStub{},
			deleteStub: &deleteReportPortStub{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "update bug report success",
			method:     http.MethodPut,
			target:     "/api/v1/bug-reports/" + reportID,
			body:       `{"title":"Updated title","description":"Updated description","status":"Closed"}`,
			createStub: &createReportPortStub{},
			getStub:    &getReportPortStub{},
			listStub:   &listReportsPortStub{},
			updateStub: &updateReportPortStub{result: ports.BugReportResult{ID: reportID, UserID: actorUserID, TransactionID: transactionID, Title: "Updated title", Description: "Updated description", Status: "Closed", CreatedAt: now, UpdatedAt: now}},
			deleteStub: &deleteReportPortStub{},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *createReportPortStub, updateStub *updateReportPortStub, _ *deleteReportPortStub, payload map[string]any) {
				t.Helper()
				if updateStub.command.Title.String() != "Updated title" {
					t.Fatalf("updateStub.command.Title = %q, want %q", updateStub.command.Title.String(), "Updated title")
				}
				dataPayload := payload["data"].(map[string]any)
				if dataPayload["status"] != "Closed" {
					t.Fatalf("data.status = %v, want %q", dataPayload["status"], "Closed")
				}
			},
		},
		{
			name:          "delete bug report success",
			method:        http.MethodDelete,
			target:        "/api/v1/bug-reports/" + reportID,
			createStub:    &createReportPortStub{},
			getStub:       &getReportPortStub{},
			listStub:      &listReportsPortStub{},
			updateStub:    &updateReportPortStub{},
			deleteStub:    &deleteReportPortStub{result: inboundports.DeleteBugReportResult{ID: reportID}},
			wantStatus:    http.StatusNoContent,
			wantEmptyBody: true,
		},
		{
			name:          "maps not found error",
			method:        http.MethodGet,
			target:        "/api/v1/bug-reports/" + reportID,
			createStub:    &createReportPortStub{},
			getStub:       &getReportPortStub{err: &usecases.NotFoundError{Resource: "bug_report", ID: reportID}},
			listStub:      &listReportsPortStub{},
			updateStub:    &updateReportPortStub{},
			deleteStub:    &deleteReportPortStub{},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name:          "rejects malformed create body",
			method:        http.MethodPost,
			target:        "/api/v1/bug-reports",
			body:          `{"transaction_id":"01962b8f-aeb2-7e03-a8ff-1edce1300001"`,
			createStub:    &createReportPortStub{},
			getStub:       &getReportPortStub{},
			listStub:      &listReportsPortStub{},
			updateStub:    &updateReportPortStub{},
			deleteStub:    &deleteReportPortStub{},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "bad_request",
		},
		{
			name:       "create: user_id comes from header not body",
			method:     http.MethodPost,
			target:     "/api/v1/bug-reports",
			body:       `{"transaction_id":"01962b8f-aeb2-7e03-a8ff-1edce1300001","title":"Title","description":"Desc"}`,
			createStub: &createReportPortStub{result: fixedResult},
			getStub:    &getReportPortStub{},
			listStub:   &listReportsPortStub{},
			updateStub: &updateReportPortStub{},
			deleteStub: &deleteReportPortStub{},
			wantStatus: http.StatusCreated,
			assert: func(t *testing.T, createStub *createReportPortStub, _ *updateReportPortStub, _ *deleteReportPortStub, _ map[string]any) {
				t.Helper()
				if createStub.command.ActorUserID != actorUserID {
					t.Fatalf("ActorUserID = %q, want %q", createStub.command.ActorUserID, actorUserID)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewHttpBugReportAdapter(tc.createStub, tc.getStub, tc.listStub, tc.updateStub, tc.deleteStub)
			router, err := httpserver.NewRouter(zap.NewNop(), adapter)
			if err != nil {
				t.Fatalf("httpserver.NewRouter() error = %v", err)
			}

			request := httptest.NewRequest(tc.method, tc.target, bytes.NewBufferString(tc.body))
			if tc.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			request.Header.Set(actorUserIDHeader, actorUserID)
			request.Header.Set(actorGroupIDHeader, "01962b8f-aeb2-7e03-a8ff-1edce1300003")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tc.wantStatus {
				t.Fatalf("status code = %d, want %d", response.Code, tc.wantStatus)
			}

			if tc.wantEmptyBody {
				if response.Body.Len() != 0 {
					t.Fatalf("response body = %q, want empty", response.Body.String())
				}
				if tc.assert != nil {
					tc.assert(t, tc.createStub, tc.updateStub, tc.deleteStub, nil)
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
					t.Fatalf("error.code = %v, want %q", errorPayload["code"], tc.wantErrorCode)
				}
				return
			}

			if payload["data"] == nil {
				t.Fatal("expected data in response payload")
			}

			if tc.assert != nil {
				tc.assert(t, tc.createStub, tc.updateStub, tc.deleteStub, payload)
			}
		})
	}
}
