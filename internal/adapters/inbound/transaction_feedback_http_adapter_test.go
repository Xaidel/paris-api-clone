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
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	"go.uber.org/zap"
)

type upsertFeedbackPortStub struct {
	result  ports.FeedbackResult
	err     error
	command inboundports.UpsertTransactionFeedbackCommand
}

func (s *upsertFeedbackPortStub) Execute(_ context.Context, command inboundports.UpsertTransactionFeedbackCommand) (ports.FeedbackResult, error) {
	s.command = command
	return s.result, s.err
}

type deleteFeedbackPortStub struct {
	err error
}

func (s *deleteFeedbackPortStub) Execute(context.Context, inboundports.DeleteTransactionFeedbackCommand) error {
	return s.err
}

type getFeedbackPortStub struct {
	result *ports.FeedbackResult
	err    error
}

func (s *getFeedbackPortStub) Execute(context.Context, inboundports.GetTransactionFeedbackQuery) (*ports.FeedbackResult, error) {
	return s.result, s.err
}

// TestHttpTransactionFeedbackAdapterRoutes verifies the HTTP transaction feedback adapter routes behavior and the expected outcome asserted below.
func TestHttpTransactionFeedbackAdapterRoutes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	transactionID := "01962b8f-aeb2-7e03-a8ff-1edce1300001"
	actorUserID := "01962b8f-aeb2-7e03-a8ff-1edce1300002"
	actorGroupID := "01962b8f-aeb2-7e03-a8ff-1edce1300003"
	feedbackID := "01962b8f-aeb2-7e03-a8ff-1edce1300301"

	fixedResult := ports.FeedbackResult{
		ID:            feedbackID,
		UserID:        actorUserID,
		TransactionID: transactionID,
		Kind:          "thumbs_up",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	tests := []struct {
		name          string
		method        string
		target        string
		upsertStub    *upsertFeedbackPortStub
		deleteStub    *deleteFeedbackPortStub
		getStub       *getFeedbackPortStub
		wantStatus    int
		wantErrorCode string
		wantEmptyBody bool
		assert        func(t *testing.T, upsertStub *upsertFeedbackPortStub, payload map[string]any)
	}{
		{
			name:       "upsert feedback success",
			method:     http.MethodPut,
			target:     "/api/v1/transactions/" + transactionID + "/feedback?kind=thumbs_up",
			upsertStub: &upsertFeedbackPortStub{result: fixedResult},
			deleteStub: &deleteFeedbackPortStub{},
			getStub:    &getFeedbackPortStub{},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, upsertStub *upsertFeedbackPortStub, payload map[string]any) {
				t.Helper()
				if upsertStub.command.Kind.String() != "thumbs_up" {
					t.Fatalf("command.Kind = %q, want %q", upsertStub.command.Kind.String(), "thumbs_up")
				}
				data := payload["data"].(map[string]any)
				if data["kind"] != "thumbs_up" {
					t.Fatalf("data.kind = %v, want %q", data["kind"], "thumbs_up")
				}
			},
		},
		{
			name:          "delete feedback success",
			method:        http.MethodDelete,
			target:        "/api/v1/transactions/" + transactionID + "/feedback",
			upsertStub:    &upsertFeedbackPortStub{},
			deleteStub:    &deleteFeedbackPortStub{},
			getStub:       &getFeedbackPortStub{},
			wantStatus:    http.StatusNoContent,
			wantEmptyBody: true,
		},
		{
			name:       "get feedback success",
			method:     http.MethodGet,
			target:     "/api/v1/transactions/" + transactionID + "/feedback",
			upsertStub: &upsertFeedbackPortStub{},
			deleteStub: &deleteFeedbackPortStub{},
			getStub:    &getFeedbackPortStub{result: &fixedResult},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, _ *upsertFeedbackPortStub, payload map[string]any) {
				t.Helper()
				data := payload["data"].(map[string]any)
				if data["kind"] != "thumbs_up" {
					t.Fatalf("data.kind = %v, want %q", data["kind"], "thumbs_up")
				}
			},
		},
		{
			name:       "get feedback returns null data when no feedback",
			method:     http.MethodGet,
			target:     "/api/v1/transactions/" + transactionID + "/feedback",
			upsertStub: &upsertFeedbackPortStub{},
			deleteStub: &deleteFeedbackPortStub{},
			getStub:    &getFeedbackPortStub{result: nil},
			wantStatus: http.StatusOK,
		},
		{
			name:          "upsert rejects invalid kind",
			method:        http.MethodPut,
			target:        "/api/v1/transactions/" + transactionID + "/feedback?kind=neutral",
			upsertStub:    &upsertFeedbackPortStub{},
			deleteStub:    &deleteFeedbackPortStub{},
			getStub:       &getFeedbackPortStub{},
			wantStatus:    http.StatusUnprocessableEntity,
			wantErrorCode: "INVALID_FEEDBACK_KIND",
		},
		{
			name:          "upsert rejects missing kind",
			method:        http.MethodPut,
			target:        "/api/v1/transactions/" + transactionID + "/feedback",
			upsertStub:    &upsertFeedbackPortStub{},
			deleteStub:    &deleteFeedbackPortStub{},
			getStub:       &getFeedbackPortStub{},
			wantStatus:    http.StatusUnprocessableEntity,
			wantErrorCode: "INVALID_FEEDBACK_KIND",
		},
		{
			name:          "delete maps not found error",
			method:        http.MethodDelete,
			target:        "/api/v1/transactions/" + transactionID + "/feedback",
			upsertStub:    &upsertFeedbackPortStub{},
			deleteStub:    &deleteFeedbackPortStub{err: &usecases.NotFoundError{Resource: "transaction_feedback", ID: transactionID}},
			getStub:       &getFeedbackPortStub{},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name:          "upsert maps not found when transaction missing",
			method:        http.MethodPut,
			target:        "/api/v1/transactions/" + transactionID + "/feedback?kind=thumbs_up",
			upsertStub:    &upsertFeedbackPortStub{err: &usecases.NotFoundError{Resource: "transaction", ID: transactionID}},
			deleteStub:    &deleteFeedbackPortStub{},
			getStub:       &getFeedbackPortStub{},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewHttpTransactionFeedbackAdapter(tc.upsertStub, tc.deleteStub, tc.getStub)
			router, err := httpserver.NewRouter(zap.NewNop(), adapter)
			if err != nil {
				t.Fatalf("httpserver.NewRouter() error = %v", err)
			}

			req := httptest.NewRequest(tc.method, tc.target, nil)
			req.Header.Set(actorUserIDHeader, actorUserID)
			req.Header.Set(actorGroupIDHeader, actorGroupID)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tc.wantStatus, rec.Body.String())
			}

			if tc.wantEmptyBody {
				if rec.Body.Len() != 0 {
					t.Fatalf("response body = %q, want empty", rec.Body.String())
				}
				return
			}

			var payload map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if tc.wantErrorCode != "" {
				errorPayload := payload["error"].(map[string]any)
				if errorPayload["code"] != tc.wantErrorCode {
					t.Fatalf("error.code = %v, want %q", errorPayload["code"], tc.wantErrorCode)
				}
				return
			}

			if tc.assert != nil {
				tc.assert(t, tc.upsertStub, payload)
			}
		})
	}
}
