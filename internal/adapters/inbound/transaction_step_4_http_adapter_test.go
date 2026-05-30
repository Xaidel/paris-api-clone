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

type createTransactionStep4PortStub struct {
	result  ports.TransactionStep4Result
	err     error
	command inboundports.CreateTransactionStep4Command
}

func (s *createTransactionStep4PortStub) Execute(_ context.Context, command inboundports.CreateTransactionStep4Command) (ports.TransactionStep4Result, error) {
	s.command = command
	return s.result, s.err
}

// TestHttpTransactionStep4AdapterRoutes verifies the HTTP transaction step 4 adapter routes behavior and the expected outcome asserted below.
func TestHttpTransactionStep4AdapterRoutes(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	transactionID := "01962b8f-aeb2-7e03-a8ff-1edce1300001"
	sectorID := "01962b8f-aeb2-7e03-a8ff-1edce1300002"
	actorUserID := "01962b8f-aeb2-7e03-a8ff-1edce1300002"

	falseValue := false
	transactionIDValue, err := valueobjects.TransactionIDFromString(transactionID)
	if err != nil {
		t.Fatalf("TransactionIDFromString(%q) error = %v", transactionID, err)
	}

	sectorIDValue, err := valueobjects.SectorIDFromString(sectorID)
	if err != nil {
		t.Fatalf("SectorIDFromString(%q) error = %v", sectorID, err)
	}

	additionalContextValue, err := valueobjects.NewTransactionStep4AdditionalContext("Reviewed by analyst")
	if err != nil {
		t.Fatalf("NewTransactionStep4AdditionalContext() error = %v", err)
	}

	fixedResult := ports.TransactionStep4Result{
		TransactionID:     transactionIDValue,
		SectorID:          sectorIDValue,
		AdditionalContext: additionalContextValue,
		IsHighEmitting:    false,
		Classification:    valueobjects.AlignedTransactionClassification(),
		Status:            valueobjects.ProfessionallyReviewedTransactionStatus(),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	tests := []struct {
		name          string
		target        string
		body          string
		stub          *createTransactionStep4PortStub
		wantStatus    int
		wantErrorCode string
		assert        func(t *testing.T, stub *createTransactionStep4PortStub, payload map[string]any)
	}{
		{
			name:       "create step 4 success",
			target:     "/api/v1/transactions/" + transactionID + "/step-4",
			body:       `{"sector_id":"` + sectorID + `","additional_context":"Reviewed by analyst","is_high_emitting":false}`,
			stub:       &createTransactionStep4PortStub{result: fixedResult},
			wantStatus: http.StatusCreated,
			assert: func(t *testing.T, stub *createTransactionStep4PortStub, payload map[string]any) {
				t.Helper()
				if stub.command.TransactionID.String() != transactionID {
					t.Fatalf("command.TransactionID.String() = %q, want %q", stub.command.TransactionID.String(), transactionID)
				}
				if stub.command.SectorID.String() != sectorID {
					t.Fatalf("command.SectorID.String() = %q, want %q", stub.command.SectorID.String(), sectorID)
				}
				if stub.command.AdditionalContext.String() != "Reviewed by analyst" {
					t.Fatalf("command.AdditionalContext.String() = %q, want %q", stub.command.AdditionalContext.String(), "Reviewed by analyst")
				}
				if stub.command.IsHighEmitting == nil {
					t.Fatal("command.IsHighEmitting = nil, want non-nil")
				}
				if *stub.command.IsHighEmitting {
					t.Fatal("command.IsHighEmitting = true, want false")
				}
				if stub.command.ActorUserID != actorUserID {
					t.Fatalf("command.ActorUserID = %q, want %q", stub.command.ActorUserID, actorUserID)
				}

				dataPayload := payload["data"].(map[string]any)
				if dataPayload["question"] != step4Question {
					t.Fatalf("data.question = %v, want %q", dataPayload["question"], step4Question)
				}
				if dataPayload["answer"] != stepAnswerAligned {
					t.Fatalf("data.answer = %v, want %q", dataPayload["answer"], stepAnswerAligned)
				}
				if dataPayload["classification"] != "aligned" {
					t.Fatalf("data.classification = %v, want %q", dataPayload["classification"], "aligned")
				}
				if dataPayload["status"] != "professionally-reviewed" {
					t.Fatalf("data.status = %v, want %q", dataPayload["status"], "professionally-reviewed")
				}
			},
		},
		{
			name:          "rejects malformed request body",
			target:        "/api/v1/transactions/" + transactionID + "/step-4",
			body:          `{"sector_id":"` + sectorID + `"`,
			stub:          &createTransactionStep4PortStub{},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: "bad_request",
		},
		{
			name:       "maps validation error for missing required fields",
			target:     "/api/v1/transactions/" + transactionID + "/step-4",
			body:       `{}`,
			stub:       &createTransactionStep4PortStub{err: domain.NewValidationError([]domain.FieldValidationError{domain.NewFieldValidationError("sector_id", "required", "sector_id is required"), domain.NewFieldValidationError("additional_context", "required", "additional_context is required"), domain.NewFieldValidationError("is_high_emitting", "required", "is_high_emitting is required")})},
			wantStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, _ *createTransactionStep4PortStub, payload map[string]any) {
				t.Helper()
				errorPayload := payload["error"].(map[string]any)
				if errorPayload["code"] != "validation_error" {
					t.Fatalf("error.code = %v, want %q", errorPayload["code"], "validation_error")
				}
				fields := errorPayload["fields"].([]any)
				if len(fields) != 3 {
					t.Fatalf("len(error.fields) = %d, want %d", len(fields), 3)
				}
			},
		},
		{
			name:          "maps not found error",
			target:        "/api/v1/transactions/" + transactionID + "/step-4",
			body:          `{"sector_id":"` + sectorID + `","additional_context":"Reviewed by analyst","is_high_emitting":false}`,
			stub:          &createTransactionStep4PortStub{err: &usecases.NotFoundError{Resource: "transaction", ID: transactionID}},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: "not_found",
		},
		{
			name:          "maps conflict error when step 3 is not eligible",
			target:        "/api/v1/transactions/" + transactionID + "/step-4",
			body:          `{"sector_id":"` + sectorID + `","additional_context":"Reviewed by analyst","is_high_emitting":false}`,
			stub:          &createTransactionStep4PortStub{err: &usecases.ConflictError{Resource: "transaction_step_4", Reason: "transaction step 3 result must be next_step"}},
			wantStatus:    http.StatusConflict,
			wantErrorCode: "conflict",
		},
		{
			name:          "maps domain error for invalid transaction id",
			target:        "/api/v1/transactions/not-a-valid-id/step-4",
			body:          `{"sector_id":"` + sectorID + `","additional_context":"Reviewed by analyst","is_high_emitting":false}`,
			stub:          &createTransactionStep4PortStub{err: domain.ErrInvalidTransactionID},
			wantStatus:    http.StatusUnprocessableEntity,
			wantErrorCode: domain.ErrInvalidTransactionID.Code,
		},
		{
			name:       "passes explicit false boolean value",
			target:     "/api/v1/transactions/" + transactionID + "/step-4",
			body:       `{"sector_id":"` + sectorID + `","additional_context":"Reviewed by analyst","is_high_emitting":false}`,
			stub:       &createTransactionStep4PortStub{result: fixedResult},
			wantStatus: http.StatusCreated,
			assert: func(t *testing.T, stub *createTransactionStep4PortStub, _ map[string]any) {
				t.Helper()
				if stub.command.IsHighEmitting == nil {
					t.Fatal("command.IsHighEmitting = nil, want non-nil")
				}
				if *stub.command.IsHighEmitting != falseValue {
					t.Fatalf("command.IsHighEmitting = %v, want %v", *stub.command.IsHighEmitting, falseValue)
				}
			},
		},
		{
			name:   "create step 4 returns not-aligned answer for high-emitting",
			target: "/api/v1/transactions/" + transactionID + "/step-4",
			body:   `{"sector_id":"` + sectorID + `","additional_context":"Reviewed by analyst","is_high_emitting":true}`,
			stub: &createTransactionStep4PortStub{result: ports.TransactionStep4Result{
				TransactionID:     transactionIDValue,
				SectorID:          sectorIDValue,
				AdditionalContext: additionalContextValue,
				IsHighEmitting:    true,
				Classification:    valueobjects.NextStepTransactionClassification(),
				Status:            valueobjects.ProfessionallyReviewedTransactionStatus(),
				CreatedAt:         now,
				UpdatedAt:         now,
			}},
			wantStatus: http.StatusCreated,
			assert: func(t *testing.T, _ *createTransactionStep4PortStub, payload map[string]any) {
				t.Helper()

				dataPayload := payload["data"].(map[string]any)
				if dataPayload["question"] != step4Question {
					t.Fatalf("data.question = %v, want %q", dataPayload["question"], step4Question)
				}
				if dataPayload["answer"] != stepAnswerNotAligned {
					t.Fatalf("data.answer = %v, want %q", dataPayload["answer"], stepAnswerNotAligned)
				}
				if dataPayload["classification"] != valueobjects.NextStepTransactionClassification().String() {
					t.Fatalf("data.classification = %v, want %q", dataPayload["classification"], valueobjects.NextStepTransactionClassification().String())
				}
				if dataPayload["status"] != valueobjects.ProfessionallyReviewedTransactionStatus().String() {
					t.Fatalf("data.status = %v, want %q", dataPayload["status"], valueobjects.ProfessionallyReviewedTransactionStatus().String())
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewHttpTransactionStep4Adapter(tc.stub)
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
