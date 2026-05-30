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
	"github.com/gyud-adb/paris-api/internal/domain"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/gyud-adb/paris-api/internal/infrastructure/httpserver"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	"go.uber.org/zap"
)

type createTransactionStep5PortStub struct {
	result  ports.TransactionStep5Result
	err     error
	command inboundports.CreateTransactionStep5Command
}

func (s *createTransactionStep5PortStub) Execute(_ context.Context, command inboundports.CreateTransactionStep5Command) (ports.TransactionStep5Result, error) {
	s.command = command
	return s.result, s.err
}

// TestHttpTransactionStep5AdapterRoutes verifies the HTTP transaction step 5 adapter routes behavior and the expected outcome asserted below.
func TestHttpTransactionStep5AdapterRoutes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 4, 12, 0, 0, 0, time.UTC)
	transactionID := "01962b8f-aeb2-7e03-a8ff-1edce1300401"
	actorUserID := "01962b8f-aeb2-7e03-a8ff-1edce1300002"

	transactionIDValue, err := valueobjects.TransactionIDFromString(transactionID)
	if err != nil {
		t.Fatalf("TransactionIDFromString(%q) error = %v", transactionID, err)
	}

	question1Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion1Justification("question 1 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion1Justification() error = %v", err)
	}

	question2Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion2Justification("question 2 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion2Justification() error = %v", err)
	}

	fixedPreviewResult := ports.TransactionStep5Result{
		TransactionID:                   transactionIDValue,
		ScreeningQuestion1Answer:        false,
		ScreeningQuestion1Justification: question1Justification,
		ScreeningQuestion2Answer:        false,
		ScreeningQuestion2Justification: question2Justification,
		ReviewerNotes:                   valueobjects.NewTransactionStep5ReviewerNotes(step5AdapterStringPointer("  optional note  ")),
		IsFinal:                         false,
		Classification:                  valueobjects.AlignedTransactionClassification(),
		Detail:                          "Auto-determined based on your screening answer",
	}

	fixedFinalResult := ports.TransactionStep5Result{
		TransactionID:                   transactionIDValue,
		ScreeningQuestion1Answer:        true,
		ScreeningQuestion1Justification: question1Justification,
		ScreeningQuestion2Answer:        false,
		ScreeningQuestion2Justification: question2Justification,
		ReviewerNotes:                   valueobjects.NewTransactionStep5ReviewerNotes(step5AdapterStringPointer("  optional note  ")),
		IsFinal:                         true,
		Classification:                  valueobjects.NotAlignedTransactionClassification(),
		Detail:                          "Auto-determined based on your screening answer",
		CreatedAt:                       &now,
		UpdatedAt:                       &now,
	}

	tests := []struct {
		name          string
		target        string
		body          string
		stub          *createTransactionStep5PortStub
		wantStatus    int
		wantErrorCode string
		assert        func(t *testing.T, stub *createTransactionStep5PortStub, payload map[string]any)
	}{
		{
			name:       "preview step 5 success without persistence",
			target:     "/api/v1/transactions/" + transactionID + "/step-5",
			body:       `{"screening_question_1":{"answer":false,"justification":"question 1 justification"},"screening_question_2":{"answer":false,"justification":"question 2 justification"},"reviewer_notes":"optional note","is_final":false}`,
			stub:       &createTransactionStep5PortStub{result: fixedPreviewResult},
			wantStatus: http.StatusOK,
			assert: func(t *testing.T, stub *createTransactionStep5PortStub, payload map[string]any) {
				t.Helper()
				if stub.command.TransactionID.String() != transactionID {
					t.Fatalf("command.TransactionID.String() = %q, want %q", stub.command.TransactionID.String(), transactionID)
				}
				if stub.command.ScreeningQuestion1Answer == nil || *stub.command.ScreeningQuestion1Answer {
					t.Fatalf("command.ScreeningQuestion1Answer = %v, want false", stub.command.ScreeningQuestion1Answer)
				}
				if stub.command.ActorUserID != actorUserID {
					t.Fatalf("command.ActorUserID = %q, want %q", stub.command.ActorUserID, actorUserID)
				}

				dataPayload := payload["data"].(map[string]any)
				if dataPayload["classification"] != valueobjects.AlignedTransactionClassification().String() {
					t.Fatalf("data.classification = %v, want %q", dataPayload["classification"], valueobjects.AlignedTransactionClassification().String())
				}
				if dataPayload["detail"] != "Auto-determined based on your screening answer" {
					t.Fatalf("data.detail = %v, want %q", dataPayload["detail"], "Auto-determined based on your screening answer")
				}
				if _, exists := dataPayload["transaction_id"]; exists {
					t.Fatalf("data.transaction_id present = true, want false")
				}
				if _, exists := dataPayload["screening_question_1"]; exists {
					t.Fatalf("data.screening_question_1 present = true, want false")
				}
				if _, exists := dataPayload["is_final"]; exists {
					t.Fatalf("data.is_final present = true, want false")
				}
			},
		},
		{
			name:       "final step 5 success persists result",
			target:     "/api/v1/transactions/" + transactionID + "/step-5",
			body:       `{"screening_question_1":{"answer":true,"justification":"question 1 justification"},"screening_question_2":{"answer":false,"justification":"question 2 justification"},"reviewer_notes":"optional note","is_final":true}`,
			stub:       &createTransactionStep5PortStub{result: fixedFinalResult},
			wantStatus: http.StatusCreated,
			assert: func(t *testing.T, _ *createTransactionStep5PortStub, payload map[string]any) {
				t.Helper()
				dataPayload := payload["data"].(map[string]any)
				if dataPayload["classification"] != valueobjects.NotAlignedTransactionClassification().String() {
					t.Fatalf("data.classification = %v, want %q", dataPayload["classification"], valueobjects.NotAlignedTransactionClassification().String())
				}
				if dataPayload["is_final"] != true {
					t.Fatalf("data.is_final = %v, want true", dataPayload["is_final"])
				}
				if dataPayload["created_at"] == nil {
					t.Fatal("data.created_at = nil, want timestamp")
				}
			},
		},
		{
			name:          "rejects malformed request body",
			target:        "/api/v1/transactions/" + transactionID + "/step-5",
			body:          `{"screening_question_1":`,
			stub:          &createTransactionStep5PortStub{},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "bad_request",
		},
		{
			name:       "maps validation error for missing required fields",
			target:     "/api/v1/transactions/" + transactionID + "/step-5",
			body:       `{}`,
			stub:       &createTransactionStep5PortStub{err: domain.NewValidationError([]domain.FieldValidationError{domain.NewFieldValidationError("screening_question_1.answer", "required", "screening_question_1.answer is required")})},
			wantStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, _ *createTransactionStep5PortStub, payload map[string]any) {
				t.Helper()
				errorPayload := payload["error"].(map[string]any)
				if errorPayload["code"] != "validation_error" {
					t.Fatalf("error.code = %v, want %q", errorPayload["code"], "validation_error")
				}
			},
		},
		{
			name:          "maps not found error",
			target:        "/api/v1/transactions/" + transactionID + "/step-5",
			body:          `{"screening_question_1":{"answer":false,"justification":"question 1 justification"},"screening_question_2":{"answer":false,"justification":"question 2 justification"},"is_final":false}`,
			stub:          &createTransactionStep5PortStub{err: &usecases.NotFoundError{Resource: "transaction", ID: transactionID}},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name:          "maps conflict error when step 4 is not eligible",
			target:        "/api/v1/transactions/" + transactionID + "/step-5",
			body:          `{"screening_question_1":{"answer":false,"justification":"question 1 justification"},"screening_question_2":{"answer":false,"justification":"question 2 justification"},"is_final":true}`,
			stub:          &createTransactionStep5PortStub{err: &usecases.ConflictError{Resource: "transaction_step_5", Reason: "transaction step 4 result must be next_step"}},
			wantStatus:    http.StatusConflict,
			wantErrorCode: "conflict",
		},
		{
			name:          "maps domain error for invalid transaction id",
			target:        "/api/v1/transactions/not-a-valid-id/step-5",
			body:          `{"screening_question_1":{"answer":false,"justification":"question 1 justification"},"screening_question_2":{"answer":false,"justification":"question 2 justification"},"is_final":false}`,
			stub:          &createTransactionStep5PortStub{err: domain.ErrInvalidTransactionID},
			wantStatus:    http.StatusUnprocessableEntity,
			wantErrorCode: domain.ErrInvalidTransactionID.Code,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewHttpTransactionStep5Adapter(tc.stub)
			router, err := httpserver.NewRouter(zap.NewNop(), adapter)
			if err != nil {
				t.Fatalf("httpserver.NewRouter() error = %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, tc.target, bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set(actorUserIDHeader, actorUserID)
			req.Header.Set(actorGroupIDHeader, "01962b8f-aeb2-7e03-a8ff-1edce1300003")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tc.wantStatus, rec.Body.String())
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
				tc.assert(t, tc.stub, payload)
			}
		})
	}
}

func step5AdapterStringPointer(value string) *string {
	return &value
}
