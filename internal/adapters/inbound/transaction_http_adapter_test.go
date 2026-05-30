package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/infrastructure/httpserver"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	"go.uber.org/zap"
)

type createTransactionPortStub struct {
	result  ports.TransactionResult
	err     error
	command inboundports.CreateTransactionCommand
}

func (s *createTransactionPortStub) Execute(_ context.Context, command inboundports.CreateTransactionCommand) (ports.TransactionResult, error) {
	s.command = command
	return s.result, s.err
}

type getTransactionPortStub struct {
	result ports.TransactionResult
	err    error
}

func (s *getTransactionPortStub) Execute(context.Context, inboundports.GetTransactionQuery) (ports.TransactionResult, error) {
	return s.result, s.err
}

type getTransactionNavigationPortStub struct {
	result inboundports.GetTransactionNavigationResult
	err    error
	query  inboundports.GetTransactionNavigationQuery
}

func (s *getTransactionNavigationPortStub) Execute(_ context.Context, query inboundports.GetTransactionNavigationQuery) (inboundports.GetTransactionNavigationResult, error) {
	s.query = query
	return s.result, s.err
}

type listTransactionsPortStub struct {
	result inboundports.ListTransactionsResult
	err    error
	query  inboundports.ListTransactionsQuery
}

func (s *listTransactionsPortStub) Execute(_ context.Context, query inboundports.ListTransactionsQuery) (inboundports.ListTransactionsResult, error) {
	s.query = query
	return s.result, s.err
}

type deleteTransactionPortStub struct {
	result  inboundports.DeleteTransactionResult
	err     error
	command inboundports.DeleteTransactionCommand
}

const (
	lowRiskReasonSuffix  = "This transaction is considered a low-risk transaction."
	highRiskReasonSuffix = "This transaction is considered a high-risk transaction and should proceed to step 4 for further review."
	actorGroupID         = "01962b8f-aeb2-7e03-a8ff-1edce1300003"
	uploadID1            = "01962b8f-aeb2-7e03-a8ff-1edce1300201"
	transactionID1       = "01962b8f-aeb2-7e03-a8ff-1edce1300101"
	transactionID2       = "01962b8f-aeb2-7e03-a8ff-1edce1300102"
	transactionID3       = "01962b8f-aeb2-7e03-a8ff-1edce1300103"
	transactionID4       = "01962b8f-aeb2-7e03-a8ff-1edce1300104"
	transactionID5       = "01962b8f-aeb2-7e03-a8ff-1edce1300105"
	transactionIDLegacy  = "01962b8f-aeb2-7e03-a8ff-1edce1300106"
	transactionIDStep5   = "01962b8f-aeb2-7e03-a8ff-1edce1300107"
)

func (s *deleteTransactionPortStub) Execute(_ context.Context, command inboundports.DeleteTransactionCommand) (inboundports.DeleteTransactionResult, error) {
	s.command = command
	return s.result, s.err
}

func appendReasonSuffix(reason string, suffix string) string {
	if strings.TrimSpace(reason) == "" {
		return suffix
	}

	return reason + ". " + suffix
}

// TestHttpTransactionAdapterRoutes verifies the HTTP transaction adapter routes behavior and the expected outcome asserted below.
func TestHttpTransactionAdapterRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		target         string
		body           string
		createStub     *createTransactionPortStub
		getStub        *getTransactionPortStub
		navigationStub *getTransactionNavigationPortStub
		listStub       *listTransactionsPortStub
		deleteStub     *deleteTransactionPortStub
		wantStatus     int
		wantErrorCode  string
		wantEmptyBody  bool
		assert         func(t *testing.T, createStub *createTransactionPortStub, navigationStub *getTransactionNavigationPortStub, listStub *listTransactionsPortStub, deleteStub *deleteTransactionPortStub, payload map[string]any)
	}{
		{
			name:           "create transaction success",
			method:         http.MethodPost,
			target:         "/api/v1/transactions",
			body:           `{"product":"CG","processed_year":2026,"processed_month":4,"dmc_ib":"IB","dmc":"DMC","partner_bank":"Partner Bank","reference_number":"REF-1","transaction_value":"698436.80","transaction_count":1,"goods_description":"Goods","goods_classification":"Classification","applicant_country":"Philippines","beneficiary_country":"Japan","source_country":"Thailand","destination_country":"Philippines","tenor_description":"N","es_category":"","pa_alignment":"PA Aligned"}`,
			createStub:     &createTransactionPortStub{result: ports.TransactionResult{ID: transactionID1, Product: "CG", Classification: "unclassified", Status: "processing", GoodsDescription: "Goods", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002"}},
			getStub:        &getTransactionPortStub{},
			navigationStub: &getTransactionNavigationPortStub{},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusCreated,
			assert: func(t *testing.T, createStub *createTransactionPortStub, _ *getTransactionNavigationPortStub, _ *listTransactionsPortStub, _ *deleteTransactionPortStub, payload map[string]any) {
				t.Helper()
				if createStub.command.Product != "CG" {
					t.Fatalf("createStub.command.Product = %q, want %q", createStub.command.Product, "CG")
				}
				if createStub.command.GoodsDescription != "Goods" {
					t.Fatalf("createStub.command.GoodsDescription = %q, want %q", createStub.command.GoodsDescription, "Goods")
				}
				if createStub.command.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("createStub.command.ActorUserID = %q, want %q", createStub.command.ActorUserID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
				dataPayload := payload["data"].(map[string]any)
				transactionData := requireTransactionData(t, dataPayload)
				if transactionData["product"] != "CG" {
					t.Fatalf("transactionData[product] = %v, want %q", transactionData["product"], "CG")
				}
				if transactionData["goods_description"] != "Goods" {
					t.Fatalf("transactionData[goods_description] = %v, want %q", transactionData["goods_description"], "Goods")
				}
				if transactionData["created_by"] != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("transactionData[created_by] = %v, want %q", transactionData["created_by"], "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
			},
		},
		{
			name:           "list transactions success",
			method:         http.MethodGet,
			target:         "/api/v1/transactions?upload_id=" + uploadID1 + "&created_at_from=2026-04-01&created_at_to=2026-04-30&applicant_country=Philippines&beneficiary_country=Japan&source_country=Thailand&destination_country=Philippines&transaction_count_min=1&transaction_count_max=10&classification=unclassified&status=processing&sort_by=transaction_count&sort_order=asc",
			createStub:     &createTransactionPortStub{},
			getStub:        &getTransactionPortStub{},
			navigationStub: &getTransactionNavigationPortStub{},
			listStub:       &listTransactionsPortStub{result: inboundports.ListTransactionsResult{Transactions: []ports.TransactionResult{{ID: transactionID1, Product: "CG", Classification: "unclassified", Status: "processing", GoodsDescription: "Goods", CreatedBy: "01962b8f-aeb2-7e03-a8ff-1edce1300002"}}}},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusOK,
			assert: func(t *testing.T, _ *createTransactionPortStub, _ *getTransactionNavigationPortStub, listStub *listTransactionsPortStub, _ *deleteTransactionPortStub, payload map[string]any) {
				t.Helper()
				if listStub.query.UploadID != uploadID1 {
					t.Fatalf("listStub.query.UploadID = %q, want %q", listStub.query.UploadID, uploadID1)
				}
				if listStub.query.CreatedAtFrom != "2026-04-01" {
					t.Fatalf("listStub.query.CreatedAtFrom = %q, want %q", listStub.query.CreatedAtFrom, "2026-04-01")
				}
				if listStub.query.CreatedAtTo != "2026-04-30" {
					t.Fatalf("listStub.query.CreatedAtTo = %q, want %q", listStub.query.CreatedAtTo, "2026-04-30")
				}
				if listStub.query.ApplicantCountry != "Philippines" {
					t.Fatalf("listStub.query.ApplicantCountry = %q, want %q", listStub.query.ApplicantCountry, "Philippines")
				}
				if listStub.query.BeneficiaryCountry != "Japan" {
					t.Fatalf("listStub.query.BeneficiaryCountry = %q, want %q", listStub.query.BeneficiaryCountry, "Japan")
				}
				if listStub.query.SourceCountry != "Thailand" {
					t.Fatalf("listStub.query.SourceCountry = %q, want %q", listStub.query.SourceCountry, "Thailand")
				}
				if listStub.query.DestinationCountry != "Philippines" {
					t.Fatalf("listStub.query.DestinationCountry = %q, want %q", listStub.query.DestinationCountry, "Philippines")
				}
				if listStub.query.TransactionCountMin != "1" {
					t.Fatalf("listStub.query.TransactionCountMin = %q, want %q", listStub.query.TransactionCountMin, "1")
				}
				if listStub.query.TransactionCountMax != "10" {
					t.Fatalf("listStub.query.TransactionCountMax = %q, want %q", listStub.query.TransactionCountMax, "10")
				}
				if listStub.query.Classification != "unclassified" {
					t.Fatalf("listStub.query.Classification = %q, want %q", listStub.query.Classification, "unclassified")
				}
				if listStub.query.Status != "processing" {
					t.Fatalf("listStub.query.Status = %q, want %q", listStub.query.Status, "processing")
				}
				if listStub.query.SortBy != "transaction_count" {
					t.Fatalf("listStub.query.SortBy = %q, want %q", listStub.query.SortBy, "transaction_count")
				}
				if listStub.query.SortOrder != "asc" {
					t.Fatalf("listStub.query.SortOrder = %q, want %q", listStub.query.SortOrder, "asc")
				}
				dataPayload := payload["data"].(map[string]any)
				transactions := dataPayload["transactions"].([]any)
				if len(transactions) != 1 {
					t.Fatalf("len(transactions) = %d, want %d", len(transactions), 1)
				}
				transactionPayload := transactions[0].(map[string]any)
				transactionData := requireTransactionData(t, transactionPayload)
				if transactionData["product"] != "CG" {
					t.Fatalf("transactionData[product] = %v, want %q", transactionData["product"], "CG")
				}
				if transactionData["goods_description"] != "Goods" {
					t.Fatalf("transactionData[goods_description] = %v, want %q", transactionData["goods_description"], "Goods")
				}
			},
		},
		{
			name:       "get transaction success",
			method:     http.MethodGet,
			target:     "/api/v1/transactions/" + transactionID1,
			createStub: &createTransactionPortStub{},
			getStub: &getTransactionPortStub{result: ports.TransactionResult{
				ID:               transactionID1,
				UploadID:         uploadID1,
				Product:          "CG",
				Classification:   "aligned",
				Status:           "ai-reviewed",
				TenorDescription: "N",
				GoodsDescription: "Goods",
				CreatedBy:        "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				PipelineResult: &ports.PipelineResultDetails{
					Version:                       "react_v1",
					BatchID:                       "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e62",
					NotAlignedListMatch:           false,
					NotAlignedListMatchConfidence: 6,
					AlignedListMatch:              false,
					AlignedListMatchConfidence:    6,
					Reason:                        "next step",
					ExitStep:                      3,
					FinalClassification:           "aligned",
				},
			}},
			navigationStub: &getTransactionNavigationPortStub{},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusOK,
			assert: func(t *testing.T, _ *createTransactionPortStub, _ *getTransactionNavigationPortStub, _ *listTransactionsPortStub, _ *deleteTransactionPortStub, payload map[string]any) {
				t.Helper()
				dataPayload := payload["data"].(map[string]any)
				transactionData := requireTransactionData(t, dataPayload)
				if dataPayload["exit_classification"] != "aligned" {
					t.Fatalf("dataPayload[exit_classification] = %v, want %q", dataPayload["exit_classification"], "aligned")
				}
				if dataPayload["status"] != "ai-reviewed" {
					t.Fatalf("dataPayload[status] = %v, want %q", dataPayload["status"], "ai-reviewed")
				}
				if transactionData["product"] != "CG" {
					t.Fatalf("transactionData[product] = %v, want %q", transactionData["product"], "CG")
				}
				if transactionData["goods_description"] != "Goods" {
					t.Fatalf("transactionData[goods_description] = %v, want %q", transactionData["goods_description"], "Goods")
				}
				if transactionData["created_by"] != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("transactionData[created_by] = %v, want %q", transactionData["created_by"], "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				if _, exists := dataPayload["row_number"]; exists {
					t.Fatalf("dataPayload[row_number] present = true, want false")
				}

				pipelineResult := requirePipelineResult(t, dataPayload)
				assertKeyAbsent(t, pipelineResult, "version")
				assertKeyAbsent(t, pipelineResult, "source")
				assertKeyAbsent(t, pipelineResult, "confidence_score")
				assertKeyAbsent(t, pipelineResult, "overall_classification")
				assertKeyAbsent(t, pipelineResult, "final_classification")
				assertKeyAbsent(t, pipelineResult, "step1_result")
				assertKeyAbsent(t, pipelineResult, "step2_result")
				assertKeyAbsent(t, pipelineResult, "step3_result")
				assertKeyAbsent(t, pipelineResult, "transaction_id")
				if dataPayload["exit_classification"] != "aligned" {
					t.Fatalf("dataPayload[exit_classification] = %v, want %q", dataPayload["exit_classification"], "aligned")
				}
				if dataPayload["status"] != "ai-reviewed" {
					t.Fatalf("dataPayload[status] = %v, want %q", dataPayload["status"], "ai-reviewed")
				}
				if pipelineResult["batch_id"] != "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e62" {
					t.Fatalf("pipelineResult[batch_id] = %v, want %q", pipelineResult["batch_id"], "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e62")
				}
				if pipelineResult["exit_step_number"] != float64(3) {
					t.Fatalf("pipelineResult[exit_step_number] = %v, want %v", pipelineResult["exit_step_number"], float64(3))
				}
				if pipelineResult["step_4_classification_data"] != nil {
					t.Fatalf("pipelineResult[step_4_classification_data] = %v, want nil", pipelineResult["step_4_classification_data"])
				}
				if pipelineResult["step_5_classification_data"] != nil {
					t.Fatalf("pipelineResult[step_5_classification_data] = %v, want nil", pipelineResult["step_5_classification_data"])
				}

				automated := requireAutomatedClassificationData(t, pipelineResult)
				assertAutomatedStep(t, automated, "step1", step1Question, automatedStepAnswerNo, float64Pointer(6))
				assertAutomatedStep(t, automated, "step2", step2Question, automatedStepAnswerNo, float64Pointer(6))
				assertAutomatedStep(t, automated, "step3", step3Question, automatedStepAnswerYes, nil)
				if automated["result"] != stepAnswerAligned {
					t.Fatalf("automated[result] = %v, want %q", automated["result"], stepAnswerAligned)
				}
				if automated["reason"] != appendReasonSuffix("next step", lowRiskReasonSuffix) {
					t.Fatalf("automated[reason] = %v, want %q", automated["reason"], appendReasonSuffix("next step", lowRiskReasonSuffix))
				}
			},
		},
		{
			name:       "get transaction nulls unreached react steps",
			method:     http.MethodGet,
			target:     "/api/v1/transactions/" + transactionID2,
			createStub: &createTransactionPortStub{},
			getStub: &getTransactionPortStub{result: ports.TransactionResult{
				ID:               transactionID2,
				Product:          "RCF",
				Classification:   "not_aligned",
				Status:           "ai-reviewed",
				GoodsDescription: "Coal",
				CreatedBy:        "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				PipelineResult: &ports.PipelineResultDetails{
					Version:                       "react_v1",
					NotAlignedListMatch:           true,
					NotAlignedListMatchConfidence: 10,
					Reason:                        "coal is excluded",
					ExitStep:                      1,
					FinalClassification:           "not_aligned",
				},
			}},
			navigationStub: &getTransactionNavigationPortStub{},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusOK,
			assert: func(t *testing.T, _ *createTransactionPortStub, _ *getTransactionNavigationPortStub, _ *listTransactionsPortStub, _ *deleteTransactionPortStub, payload map[string]any) {
				t.Helper()

				dataPayload := payload["data"].(map[string]any)
				if dataPayload["exit_classification"] != "not_aligned" {
					t.Fatalf("dataPayload[exit_classification] = %v, want %q", dataPayload["exit_classification"], "not_aligned")
				}
				if dataPayload["status"] != "ai-reviewed" {
					t.Fatalf("dataPayload[status] = %v, want %q", dataPayload["status"], "ai-reviewed")
				}
				pipelineResult := requirePipelineResult(t, dataPayload)
				if pipelineResult["exit_step_number"] != float64(1) {
					t.Fatalf("pipelineResult[exit_step_number] = %v, want %v", pipelineResult["exit_step_number"], float64(1))
				}
				automated := requireAutomatedClassificationData(t, pipelineResult)
				assertAutomatedStep(t, automated, "step1", step1Question, automatedStepAnswerYes, float64Pointer(10))
				assertAutomatedStepNil(t, automated, "step2")
				assertAutomatedStepNil(t, automated, "step3")
				if automated["result"] != stepAnswerNotAligned {
					t.Fatalf("automated[result] = %v, want %q", automated["result"], stepAnswerNotAligned)
				}
			},
		},
		{
			name:       "get transaction marks Y-prefixed tenor as not low risk",
			method:     http.MethodGet,
			target:     "/api/v1/transactions/" + transactionID3,
			createStub: &createTransactionPortStub{},
			getStub: &getTransactionPortStub{result: ports.TransactionResult{
				ID:               transactionID3,
				Product:          "RCF",
				Classification:   "not_aligned",
				Status:           "ai-reviewed",
				TenorDescription: "Y-256",
				GoodsDescription: "Goods",
				CreatedBy:        "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				PipelineResult: &ports.PipelineResultDetails{
					Version:             "react_v1",
					Reason:              "needs more information",
					ExitStep:            3,
					FinalClassification: "not_aligned",
				},
			}},
			navigationStub: &getTransactionNavigationPortStub{},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusOK,
			assert: func(t *testing.T, _ *createTransactionPortStub, _ *getTransactionNavigationPortStub, _ *listTransactionsPortStub, _ *deleteTransactionPortStub, payload map[string]any) {
				t.Helper()

				dataPayload := payload["data"].(map[string]any)
				if dataPayload["exit_classification"] != "next_step" {
					t.Fatalf("dataPayload[exit_classification] = %v, want %q", dataPayload["exit_classification"], "next_step")
				}
				if dataPayload["status"] != "ai-reviewed" {
					t.Fatalf("dataPayload[status] = %v, want %q", dataPayload["status"], "ai-reviewed")
				}
				pipelineResult := requirePipelineResult(t, dataPayload)
				automated := requireAutomatedClassificationData(t, pipelineResult)
				assertAutomatedStep(t, automated, "step3", step3Question, automatedStepAnswerNo, nil)
				if automated["result"] != stepAnswerNextStep {
					t.Fatalf("automated[result] = %v, want %q", automated["result"], stepAnswerNextStep)
				}
			},
		},
		{
			name:       "get transaction standardizes legacy step responses",
			method:     http.MethodGet,
			target:     "/api/v1/transactions/" + transactionIDLegacy,
			createStub: &createTransactionPortStub{},
			getStub: &getTransactionPortStub{result: ports.TransactionResult{
				ID:               transactionIDLegacy,
				Product:          "RCF",
				Classification:   "aligned",
				Status:           "ai-reviewed",
				TenorDescription: "N",
				GoodsDescription: "Goods",
				CreatedBy:        "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				PipelineResult: &ports.PipelineResultDetails{
					ExitStep:            3,
					FinalClassification: "aligned",
					Step1Result: ports.StepResultDetails{
						StepNumber:    1,
						StepAlignment: "unaligned",
						Reason:        "no exclusion match",
					},
					Step2Result: &ports.StepResultDetails{
						StepNumber:    2,
						StepAlignment: "unaligned",
						Reason:        "no u1 match",
					},
					Step3Result: &ports.StepResultDetails{
						StepNumber:    3,
						StepAlignment: "aligned",
						BooleanResult: testBoolPointer(true),
					},
				},
			}},
			navigationStub: &getTransactionNavigationPortStub{},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusOK,
			assert: func(t *testing.T, _ *createTransactionPortStub, _ *getTransactionNavigationPortStub, _ *listTransactionsPortStub, _ *deleteTransactionPortStub, payload map[string]any) {
				t.Helper()

				dataPayload := payload["data"].(map[string]any)
				if dataPayload["exit_classification"] != "aligned" {
					t.Fatalf("dataPayload[exit_classification] = %v, want %q", dataPayload["exit_classification"], "aligned")
				}
				if dataPayload["status"] != "ai-reviewed" {
					t.Fatalf("dataPayload[status] = %v, want %q", dataPayload["status"], "ai-reviewed")
				}

				pipelineResult := requirePipelineResult(t, dataPayload)
				automated := requireAutomatedClassificationData(t, pipelineResult)
				assertAutomatedStep(t, automated, "step1", step1Question, automatedStepAnswerNo, nil)
				assertAutomatedStep(t, automated, "step2", step2Question, automatedStepAnswerNo, nil)
				assertAutomatedStep(t, automated, "step3", step3Question, automatedStepAnswerYes, nil)
				if automated["result"] != stepAnswerAligned {
					t.Fatalf("automated[result] = %v, want %q", automated["result"], stepAnswerAligned)
				}
			},
		},
		{
			name:       "get transaction returns next_step classification unambiguously",
			method:     http.MethodGet,
			target:     "/api/v1/transactions/" + transactionID4,
			createStub: &createTransactionPortStub{},
			getStub: &getTransactionPortStub{result: ports.TransactionResult{
				ID:               transactionID4,
				Product:          "RCF",
				Classification:   "next_step",
				Status:           "ai-reviewed",
				TenorDescription: "Y",
				GoodsDescription: "Goods",
				CreatedBy:        "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				PipelineResult: &ports.PipelineResultDetails{
					Version:             "react_v1",
					Reason:              "needs next review step",
					ExitStep:            3,
					FinalClassification: "next_step",
				},
			}},
			navigationStub: &getTransactionNavigationPortStub{},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusOK,
			assert: func(t *testing.T, _ *createTransactionPortStub, _ *getTransactionNavigationPortStub, _ *listTransactionsPortStub, _ *deleteTransactionPortStub, payload map[string]any) {
				t.Helper()

				dataPayload := payload["data"].(map[string]any)
				transactionData := requireTransactionData(t, dataPayload)
				assertKeyAbsent(t, transactionData, "classification")
				assertKeyAbsent(t, transactionData, "status")
				if dataPayload["exit_classification"] != "next_step" {
					t.Fatalf("dataPayload[exit_classification] = %v, want %q", dataPayload["exit_classification"], "next_step")
				}
				if dataPayload["status"] != "ai-reviewed" {
					t.Fatalf("dataPayload[status] = %v, want %q", dataPayload["status"], "ai-reviewed")
				}

				pipelineResult := requirePipelineResult(t, dataPayload)
				if pipelineResult["exit_step_number"] != float64(0) {
					t.Fatalf("pipelineResult[exit_step_number] = %v, want %v", pipelineResult["exit_step_number"], float64(0))
				}
				automated := requireAutomatedClassificationData(t, pipelineResult)
				if automated["result"] != stepAnswerNextStep {
					t.Fatalf("automated[result] = %v, want %q", automated["result"], stepAnswerNextStep)
				}
			},
		},
		{
			name:       "get transaction reflects terminal review over prior pipeline next_step",
			method:     http.MethodGet,
			target:     "/api/v1/transactions/" + transactionID5,
			createStub: &createTransactionPortStub{},
			getStub: &getTransactionPortStub{result: ports.TransactionResult{
				ID:               transactionID5,
				Product:          "RCF",
				Classification:   "aligned",
				Status:           "professionally-reviewed",
				TenorDescription: "Y",
				GoodsDescription: "Goods",
				CreatedBy:        "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				PipelineResult: &ports.PipelineResultDetails{
					Version:             "react_v1",
					Reason:              "needs next review step",
					ExitStep:            3,
					FinalClassification: "aligned",
				},
				Step4Classification: &ports.TransactionStep4Details{
					IdentifiedSector:      "Energy",
					AdditionalInformation: "Reviewed by analyst",
					Result:                stepAnswerAligned,
				},
			}},
			navigationStub: &getTransactionNavigationPortStub{},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusOK,
			assert: func(t *testing.T, _ *createTransactionPortStub, _ *getTransactionNavigationPortStub, _ *listTransactionsPortStub, _ *deleteTransactionPortStub, payload map[string]any) {
				t.Helper()

				dataPayload := payload["data"].(map[string]any)
				transactionData := requireTransactionData(t, dataPayload)
				assertKeyAbsent(t, transactionData, "classification")
				assertKeyAbsent(t, transactionData, "status")
				if dataPayload["exit_classification"] != "aligned" {
					t.Fatalf("dataPayload[exit_classification] = %v, want %q", dataPayload["exit_classification"], "aligned")
				}
				if dataPayload["status"] != "professionally-reviewed" {
					t.Fatalf("dataPayload[status] = %v, want %q", dataPayload["status"], "professionally-reviewed")
				}

				pipelineResult := requirePipelineResult(t, dataPayload)
				if pipelineResult["exit_step_number"] != float64(4) {
					t.Fatalf("pipelineResult[exit_step_number] = %v, want %v", pipelineResult["exit_step_number"], float64(4))
				}
				automated := requireAutomatedClassificationData(t, pipelineResult)
				if automated["result"] != stepAnswerNextStep {
					t.Fatalf("automated[result] = %v, want %q", automated["result"], stepAnswerNextStep)
				}
				if automated["reason"] != appendReasonSuffix("needs next review step", highRiskReasonSuffix) {
					t.Fatalf("automated[reason] = %v, want %q", automated["reason"], appendReasonSuffix("needs next review step", highRiskReasonSuffix))
				}

				step4Data := pipelineResult["step_4_classification_data"].(map[string]any)
				if step4Data["identified_sector"] != "Energy" {
					t.Fatalf("step4Data[identified_sector] = %v, want %q", step4Data["identified_sector"], "Energy")
				}
				if step4Data["additional_information"] != "Reviewed by analyst" {
					t.Fatalf("step4Data[additional_information] = %v, want %q", step4Data["additional_information"], "Reviewed by analyst")
				}
				if step4Data["result"] != stepAnswerAligned {
					t.Fatalf("step4Data[result] = %v, want %q", step4Data["result"], stepAnswerAligned)
				}
			},
		},
		{
			name:       "get transaction includes step 5 classification data",
			method:     http.MethodGet,
			target:     "/api/v1/transactions/" + transactionIDStep5,
			createStub: &createTransactionPortStub{},
			getStub: &getTransactionPortStub{result: ports.TransactionResult{
				ID:               transactionIDStep5,
				Product:          "RCF",
				Classification:   "not_aligned",
				Status:           "professionally-reviewed",
				TenorDescription: "Y",
				GoodsDescription: "Goods",
				CreatedBy:        "01962b8f-aeb2-7e03-a8ff-1edce1300002",
				PipelineResult: &ports.PipelineResultDetails{
					Version:             "react_v1",
					Reason:              "needs next review step",
					ExitStep:            3,
					FinalClassification: "not_aligned",
				},
				Step4Classification: &ports.TransactionStep4Details{
					IdentifiedSector:      "Energy",
					AdditionalInformation: "Reviewed by analyst",
					Result:                stepAnswerNextStep,
				},
				Step5Classification: &ports.TransactionStep5Details{
					ScreeningQuestion1: ports.TransactionStep5ScreeningQuestionDetails{
						Answer:        true,
						Justification: "question 1 justification",
					},
					ScreeningQuestion2: ports.TransactionStep5ScreeningQuestionDetails{
						Answer:        false,
						Justification: "question 2 justification",
					},
					ReviewerNotes: testStringPointer("optional note"),
					IsFinal:       true,
					Result:        stepAnswerNotAligned,
				},
			}},
			navigationStub: &getTransactionNavigationPortStub{},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusOK,
			assert: func(t *testing.T, _ *createTransactionPortStub, _ *getTransactionNavigationPortStub, _ *listTransactionsPortStub, _ *deleteTransactionPortStub, payload map[string]any) {
				t.Helper()

				dataPayload := payload["data"].(map[string]any)
				if dataPayload["exit_step_number"] != float64(5) {
					t.Fatalf("dataPayload[exit_step_number] = %v, want %v", dataPayload["exit_step_number"], float64(5))
				}
				step5Data := dataPayload["step_5_classification_data"].(map[string]any)
				question1 := step5Data["screening_question_1"].(map[string]any)
				question2 := step5Data["screening_question_2"].(map[string]any)

				if question1["question"] != step5ScreeningQuestion1Text {
					t.Fatalf("question1[question] = %v, want %q", question1["question"], step5ScreeningQuestion1Text)
				}
				if question1["answer"] != step5AnswerYes {
					t.Fatalf("question1[answer] = %v, want %q", question1["answer"], step5AnswerYes)
				}
				if question1["justification"] != "question 1 justification" {
					t.Fatalf("question1[justification] = %v, want %q", question1["justification"], "question 1 justification")
				}
				if question2["answer"] != step5AnswerNo {
					t.Fatalf("question2[answer] = %v, want %q", question2["answer"], step5AnswerNo)
				}
				if step5Data["reviewer_notes"] != "optional note" {
					t.Fatalf("step5Data[reviewer_notes] = %v, want %q", step5Data["reviewer_notes"], "optional note")
				}
				if step5Data["is_final"] != true {
					t.Fatalf("step5Data[is_final] = %v, want true", step5Data["is_final"])
				}
				if step5Data["result"] != stepAnswerNotAligned {
					t.Fatalf("step5Data[result] = %v, want %q", step5Data["result"], stepAnswerNotAligned)
				}
			},
		},
		{
			name:           "get transaction navigation success",
			method:         http.MethodGet,
			target:         "/api/v1/transactions/" + transactionID1 + "/navigation?classification=aligned&step=3&upload_id=" + uploadID1 + "&created_at_from=2026-04-01&created_at_to=2026-04-30&applicant_country=Philippines&beneficiary_country=Japan&source_country=Thailand&destination_country=Philippines&transaction_count_min=1&transaction_count_max=10&status=processing",
			createStub:     &createTransactionPortStub{},
			getStub:        &getTransactionPortStub{},
			navigationStub: &getTransactionNavigationPortStub{result: inboundports.GetTransactionNavigationResult{TransactionID: transactionID1, PreviousID: testStringPointer(transactionID2), NextID: testStringPointer(transactionID3)}},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusOK,
			assert: func(t *testing.T, _ *createTransactionPortStub, navigationStub *getTransactionNavigationPortStub, _ *listTransactionsPortStub, _ *deleteTransactionPortStub, payload map[string]any) {
				t.Helper()
				if navigationStub.query.ID != transactionID1 {
					t.Fatalf("navigationStub.query.ID = %q, want %q", navigationStub.query.ID, transactionID1)
				}
				if navigationStub.query.Classification != "aligned" {
					t.Fatalf("navigationStub.query.Classification = %q, want %q", navigationStub.query.Classification, "aligned")
				}
				if navigationStub.query.Step != "3" {
					t.Fatalf("navigationStub.query.Step = %q, want %q", navigationStub.query.Step, "3")
				}
				if navigationStub.query.UploadID != uploadID1 {
					t.Fatalf("navigationStub.query.UploadID = %q, want %q", navigationStub.query.UploadID, uploadID1)
				}
				if navigationStub.query.CreatedAtFrom != "2026-04-01" {
					t.Fatalf("navigationStub.query.CreatedAtFrom = %q, want %q", navigationStub.query.CreatedAtFrom, "2026-04-01")
				}
				if navigationStub.query.CreatedAtTo != "2026-04-30" {
					t.Fatalf("navigationStub.query.CreatedAtTo = %q, want %q", navigationStub.query.CreatedAtTo, "2026-04-30")
				}
				if navigationStub.query.ApplicantCountry != "Philippines" {
					t.Fatalf("navigationStub.query.ApplicantCountry = %q, want %q", navigationStub.query.ApplicantCountry, "Philippines")
				}
				if navigationStub.query.BeneficiaryCountry != "Japan" {
					t.Fatalf("navigationStub.query.BeneficiaryCountry = %q, want %q", navigationStub.query.BeneficiaryCountry, "Japan")
				}
				if navigationStub.query.SourceCountry != "Thailand" {
					t.Fatalf("navigationStub.query.SourceCountry = %q, want %q", navigationStub.query.SourceCountry, "Thailand")
				}
				if navigationStub.query.DestinationCountry != "Philippines" {
					t.Fatalf("navigationStub.query.DestinationCountry = %q, want %q", navigationStub.query.DestinationCountry, "Philippines")
				}
				if navigationStub.query.TransactionCountMin != "1" {
					t.Fatalf("navigationStub.query.TransactionCountMin = %q, want %q", navigationStub.query.TransactionCountMin, "1")
				}
				if navigationStub.query.TransactionCountMax != "10" {
					t.Fatalf("navigationStub.query.TransactionCountMax = %q, want %q", navigationStub.query.TransactionCountMax, "10")
				}
				if navigationStub.query.Status != "processing" {
					t.Fatalf("navigationStub.query.Status = %q, want %q", navigationStub.query.Status, "processing")
				}
				if navigationStub.query.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("navigationStub.query.ActorUserID = %q, want %q", navigationStub.query.ActorUserID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}
				if navigationStub.query.ActorGroupID != actorGroupID {
					t.Fatalf("navigationStub.query.ActorGroupID = %q, want %q", navigationStub.query.ActorGroupID, actorGroupID)
				}

				dataPayload := payload["data"].(map[string]any)
				if len(dataPayload) != 3 {
					t.Fatalf("len(dataPayload) = %d, want %d", len(dataPayload), 3)
				}
				if dataPayload["transaction_id"] != transactionID1 {
					t.Fatalf("dataPayload[transaction_id] = %v, want %q", dataPayload["transaction_id"], transactionID1)
				}
				if dataPayload["previous_id"] != transactionID2 {
					t.Fatalf("dataPayload[previous_id] = %v, want %q", dataPayload["previous_id"], transactionID2)
				}
				if dataPayload["next_id"] != transactionID3 {
					t.Fatalf("dataPayload[next_id] = %v, want %q", dataPayload["next_id"], transactionID3)
				}
			},
		},
		{
			name:           "maps navigation not found error",
			method:         http.MethodGet,
			target:         "/api/v1/transactions/" + transactionID1 + "/navigation",
			createStub:     &createTransactionPortStub{},
			getStub:        &getTransactionPortStub{},
			navigationStub: &getTransactionNavigationPortStub{err: &usecases.NotFoundError{Resource: "transaction", ID: transactionID1}},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusNotFound,
			wantErrorCode:  "not_found",
		},
		{
			name:           "maps navigation field validation error",
			method:         http.MethodGet,
			target:         "/api/v1/transactions/" + transactionID1 + "/navigation",
			createStub:     &createTransactionPortStub{},
			getStub:        &getTransactionPortStub{},
			navigationStub: &getTransactionNavigationPortStub{err: domain.NewValidationError([]domain.FieldValidationError{domain.NewFieldValidationError("classification", "invalid", "classification is invalid")})},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusUnprocessableEntity,
			assert: func(t *testing.T, _ *createTransactionPortStub, _ *getTransactionNavigationPortStub, _ *listTransactionsPortStub, _ *deleteTransactionPortStub, payload map[string]any) {
				t.Helper()
				errorPayload := payload["error"].(map[string]any)
				if errorPayload["code"] != "validation_error" {
					t.Fatalf("errorPayload[code] = %v, want %q", errorPayload["code"], "validation_error")
				}
			},
		},
		{
			name:           "delete transaction success",
			method:         http.MethodDelete,
			target:         "/api/v1/transactions/" + transactionID1,
			createStub:     &createTransactionPortStub{},
			getStub:        &getTransactionPortStub{},
			navigationStub: &getTransactionNavigationPortStub{},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{result: inboundports.DeleteTransactionResult{ID: transactionID1}},
			wantStatus:     http.StatusNoContent,
			wantEmptyBody:  true,
			assert: func(t *testing.T, _ *createTransactionPortStub, _ *getTransactionNavigationPortStub, _ *listTransactionsPortStub, deleteStub *deleteTransactionPortStub, _ map[string]any) {
				t.Helper()
				if deleteStub.command.ID != transactionID1 {
					t.Fatalf("deleteStub.command.ID = %q, want %q", deleteStub.command.ID, transactionID1)
				}
			},
		},
		{
			name:           "maps not found error",
			method:         http.MethodGet,
			target:         "/api/v1/transactions/" + transactionID1,
			createStub:     &createTransactionPortStub{},
			getStub:        &getTransactionPortStub{err: &usecases.NotFoundError{Resource: "transaction", ID: transactionID1}},
			navigationStub: &getTransactionNavigationPortStub{},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusNotFound,
			wantErrorCode:  "not_found",
		},
		{
			name:           "maps field validation error",
			method:         http.MethodPost,
			target:         "/api/v1/transactions",
			body:           `{"product":"","processed_year":0,"processed_month":13,"dmc_ib":"","dmc":"","partner_bank":"","reference_number":"","transaction_value":"","transaction_count":-1,"goods_description":"","goods_classification":"Classification","applicant_country":"Philippines","beneficiary_country":"Japan","source_country":"Thailand","destination_country":"Philippines","tenor_description":"N","es_category":"","pa_alignment":""}`,
			createStub:     &createTransactionPortStub{err: domain.NewValidationError([]domain.FieldValidationError{domain.NewFieldValidationError("goods_description", "required", "goods_description is required")})},
			getStub:        &getTransactionPortStub{},
			navigationStub: &getTransactionNavigationPortStub{},
			listStub:       &listTransactionsPortStub{},
			deleteStub:     &deleteTransactionPortStub{},
			wantStatus:     http.StatusUnprocessableEntity,
			assert: func(t *testing.T, _ *createTransactionPortStub, _ *getTransactionNavigationPortStub, _ *listTransactionsPortStub, _ *deleteTransactionPortStub, payload map[string]any) {
				t.Helper()
				errorPayload := payload["error"].(map[string]any)
				if errorPayload["code"] != "validation_error" {
					t.Fatalf("errorPayload[code] = %v, want %q", errorPayload["code"], "validation_error")
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewHttpTransactionAdapter(tc.createStub, tc.getStub, tc.navigationStub, tc.listStub, tc.deleteStub)
			router, err := httpserver.NewRouter(zap.NewNop(), adapter)
			if err != nil {
				t.Fatalf("httpserver.NewRouter() error = %v", err)
			}

			request := httptest.NewRequest(tc.method, tc.target, bytes.NewBufferString(tc.body))
			if tc.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			request.Header.Set(actorUserIDHeader, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
			request.Header.Set(actorGroupIDHeader, actorGroupID)
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
					tc.assert(t, tc.createStub, tc.navigationStub, tc.listStub, tc.deleteStub, nil)
				}
				return
			}

			var payload map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				if response.Body.Len() == 0 {
					t.Fatal("expected response body")
				}
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
				tc.assert(t, tc.createStub, tc.navigationStub, tc.listStub, tc.deleteStub, payload)
			}
		})
	}
}

// TestToTransactionResponseFormatsFlattenedClassificationCases verifies the to transaction response formats flattened classification cases behavior and the expected outcome asserted below.
func TestToTransactionResponseFormatsFlattenedClassificationCases(t *testing.T) {
	t.Parallel()

	transactionID := "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60"

	tests := []struct {
		name             string
		result           *ports.PipelineResultDetails
		tenorDescription string
		wantExitStep     int
		wantResult       string
		wantReason       string
		assert           func(t *testing.T, payload map[string]any)
	}{
		{
			name: "case 1 exclusion match exits at step 1",
			result: &ports.PipelineResultDetails{
				Version:                       "react_v1",
				BatchID:                       "01962b8f-aeb2-7e03-a8ff-1edce1300401",
				ExitStep:                      1,
				NotAlignedListMatch:           true,
				NotAlignedListMatchConfidence: 10,
				FinalClassification:           "not_aligned",
				Reason:                        "coal is excluded",
			},
			wantExitStep: 1,
			wantResult:   stepAnswerNotAligned,
			wantReason:   "coal is excluded",
			assert: func(t *testing.T, payload map[string]any) {
				t.Helper()
				automated := requireAutomatedClassificationData(t, payload)
				assertAutomatedStep(t, automated, "step1", step1Question, automatedStepAnswerYes, float64Pointer(10))
				assertAutomatedStepNil(t, automated, "step2")
				assertAutomatedStepNil(t, automated, "step3")
			},
		},
		{
			name: "case 2 aligned list match exits at step 2",
			result: &ports.PipelineResultDetails{
				Version:                       "react_v1",
				BatchID:                       "01962b8f-aeb2-7e03-a8ff-1edce1300402",
				ExitStep:                      2,
				NotAlignedListMatchConfidence: 10,
				AlignedListMatch:              true,
				AlignedListMatchConfidence:    6,
				FinalClassification:           "aligned",
				Reason:                        "matches universally aligned list",
			},
			wantExitStep: 2,
			wantResult:   stepAnswerAligned,
			wantReason:   "matches universally aligned list",
			assert: func(t *testing.T, payload map[string]any) {
				t.Helper()
				automated := requireAutomatedClassificationData(t, payload)
				assertAutomatedStep(t, automated, "step1", step1Question, automatedStepAnswerNo, float64Pointer(10))
				assertAutomatedStep(t, automated, "step2", step2Question, automatedStepAnswerYes, float64Pointer(6))
				assertAutomatedStepNil(t, automated, "step3")
			},
		},
		{
			name: "case 3 low risk exits aligned at step 3",
			result: &ports.PipelineResultDetails{
				Version:                       "react_v1",
				BatchID:                       "01962b8f-aeb2-7e03-a8ff-1edce1300403",
				ExitStep:                      3,
				NotAlignedListMatchConfidence: 10,
				AlignedListMatchConfidence:    6,
				FinalClassification:           "aligned",
				Reason:                        "meets low risk criteria",
			},
			tenorDescription: "N",
			wantExitStep:     3,
			wantResult:       stepAnswerAligned,
			wantReason:       appendReasonSuffix("meets low risk criteria", lowRiskReasonSuffix),
			assert: func(t *testing.T, payload map[string]any) {
				t.Helper()
				automated := requireAutomatedClassificationData(t, payload)
				assertAutomatedStep(t, automated, "step1", step1Question, automatedStepAnswerNo, float64Pointer(10))
				assertAutomatedStep(t, automated, "step2", step2Question, automatedStepAnswerNo, float64Pointer(6))
				assertAutomatedStep(t, automated, "step3", step3Question, automatedStepAnswerYes, nil)
			},
		},
		{
			name: "case 3 strips debug suffix and appends low risk step 3 messaging",
			result: &ports.PipelineResultDetails{
				Version:                       "react_v1",
				BatchID:                       "batch-3-sanitized",
				ExitStep:                      3,
				NotAlignedListMatchConfidence: 10,
				AlignedListMatchConfidence:    6,
				FinalClassification:           "aligned",
				Reason:                        "Yarn for export-oriented textiles lacks explicit Paris-alignment indicators; further detail required.; legacy step 3 fallback boolean result: true",
			},
			tenorDescription: "N",
			wantExitStep:     3,
			wantResult:       stepAnswerAligned,
			wantReason:       appendReasonSuffix("Yarn for export-oriented textiles lacks explicit Paris-alignment indicators", lowRiskReasonSuffix),
			assert: func(t *testing.T, payload map[string]any) {
				t.Helper()
				automated := requireAutomatedClassificationData(t, payload)
				assertAutomatedStep(t, automated, "step1", step1Question, automatedStepAnswerNo, float64Pointer(10))
				assertAutomatedStep(t, automated, "step2", step2Question, automatedStepAnswerNo, float64Pointer(6))
				assertAutomatedStep(t, automated, "step3", step3Question, automatedStepAnswerYes, nil)
			},
		},
		{
			name: "case 4 unresolved step 3 exits next step",
			result: &ports.PipelineResultDetails{
				Version:                       "react_v1",
				BatchID:                       "01962b8f-aeb2-7e03-a8ff-1edce1300405",
				ExitStep:                      3,
				NotAlignedListMatchConfidence: 10,
				AlignedListMatchConfidence:    6,
				FinalClassification:           "next_step",
				Reason:                        "needs additional review",
			},
			tenorDescription: "Y-256",
			wantExitStep:     0,
			wantResult:       stepAnswerNextStep,
			wantReason:       appendReasonSuffix("needs additional review", highRiskReasonSuffix),
			assert: func(t *testing.T, payload map[string]any) {
				t.Helper()
				automated := requireAutomatedClassificationData(t, payload)
				assertAutomatedStep(t, automated, "step1", step1Question, automatedStepAnswerNo, float64Pointer(10))
				assertAutomatedStep(t, automated, "step2", step2Question, automatedStepAnswerNo, float64Pointer(6))
				assertAutomatedStep(t, automated, "step3", step3Question, automatedStepAnswerNo, nil)
			},
		},
		{
			name: "case 4 step 3 next step appends high risk messaging",
			result: &ports.PipelineResultDetails{
				Version:                       "react_v1",
				BatchID:                       "batch-4-sanitized",
				ExitStep:                      3,
				NotAlignedListMatchConfidence: 10,
				AlignedListMatchConfidence:    6,
				FinalClassification:           "next_step",
				Reason:                        "legacy step 3 fallback boolean result: false",
			},
			tenorDescription: "Y-256",
			wantExitStep:     0,
			wantResult:       stepAnswerNextStep,
			wantReason:       highRiskReasonSuffix,
			assert: func(t *testing.T, payload map[string]any) {
				t.Helper()
				automated := requireAutomatedClassificationData(t, payload)
				assertAutomatedStep(t, automated, "step1", step1Question, automatedStepAnswerNo, float64Pointer(10))
				assertAutomatedStep(t, automated, "step2", step2Question, automatedStepAnswerNo, float64Pointer(6))
				assertAutomatedStep(t, automated, "step3", step3Question, automatedStepAnswerNo, nil)
			},
		},
		{
			name: "case 5 edge case preserves step 1 no with next step result",
			result: &ports.PipelineResultDetails{
				Version:                       "react_v1",
				BatchID:                       "01962b8f-aeb2-7e03-a8ff-1edce1300407",
				ExitStep:                      1,
				NotAlignedListMatchConfidence: 10,
				FinalClassification:           "next_step",
				Reason:                        "requires downstream review despite no exclusion match",
			},
			wantExitStep: 0,
			wantResult:   stepAnswerNextStep,
			wantReason:   "requires downstream review despite no exclusion match",
			assert: func(t *testing.T, payload map[string]any) {
				t.Helper()
				automated := requireAutomatedClassificationData(t, payload)
				assertAutomatedStep(t, automated, "step1", step1Question, automatedStepAnswerNo, float64Pointer(10))
				assertAutomatedStepNil(t, automated, "step2")
				assertAutomatedStepNil(t, automated, "step3")
			},
		},
		{
			name: "case 6 edge case keeps step 1 precedence over step 2",
			result: &ports.PipelineResultDetails{
				Version:                       "react_v1",
				BatchID:                       "01962b8f-aeb2-7e03-a8ff-1edce1300408",
				ExitStep:                      2,
				NotAlignedListMatch:           true,
				NotAlignedListMatchConfidence: 10,
				AlignedListMatch:              true,
				AlignedListMatchConfidence:    6,
				FinalClassification:           "not_aligned",
				Reason:                        "not aligned list takes precedence",
			},
			wantExitStep: 2,
			wantResult:   stepAnswerNotAligned,
			wantReason:   "not aligned list takes precedence",
			assert: func(t *testing.T, payload map[string]any) {
				t.Helper()
				automated := requireAutomatedClassificationData(t, payload)
				assertAutomatedStep(t, automated, "step1", step1Question, automatedStepAnswerYes, float64Pointer(10))
				assertAutomatedStep(t, automated, "step2", step2Question, automatedStepAnswerYes, float64Pointer(6))
				assertAutomatedStepNil(t, automated, "step3")
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			response := toTransactionResponse(ports.TransactionResult{
				ID:               transactionID,
				Classification:   strings.ReplaceAll(tc.wantResult, "-", "_"),
				TenorDescription: tc.tenorDescription,
				PipelineResult:   tc.result,
			})
			payload := mustMarshalToMap(t, response)

			if _, exists := payload["pipeline_result"]; exists {
				t.Fatalf("payload[pipeline_result] present = true, want false")
			}
			assertKeyAbsent(t, payload, "transaction_id")

			if payload["batch_id"] != tc.result.BatchID {
				t.Fatalf("payload[batch_id] = %v, want %q", payload["batch_id"], tc.result.BatchID)
			}
			if payload["exit_step_number"] != float64(tc.wantExitStep) {
				t.Fatalf("payload[exit_step_number] = %v, want %v", payload["exit_step_number"], float64(tc.wantExitStep))
			}
			if payload["exit_classification"] != strings.ReplaceAll(tc.wantResult, "-", "_") {
				t.Fatalf("payload[exit_classification] = %v, want %q", payload["exit_classification"], strings.ReplaceAll(tc.wantResult, "-", "_"))
			}
			if payload["step_4_classification_data"] != nil {
				t.Fatalf("payload[step_4_classification_data] = %v, want nil", payload["step_4_classification_data"])
			}
			if payload["step_5_classification_data"] != nil {
				t.Fatalf("payload[step_5_classification_data] = %v, want nil", payload["step_5_classification_data"])
			}

			automated := requireAutomatedClassificationData(t, payload)
			if automated["result"] != tc.wantResult {
				t.Fatalf("automated[result] = %v, want %q", automated["result"], tc.wantResult)
			}
			if automated["reason"] != tc.wantReason {
				t.Fatalf("automated[reason] = %v, want %q", automated["reason"], tc.wantReason)
			}

			tc.assert(t, payload)
		})
	}
}

func TestToTransactionResponseDerivesExitClassificationFromAutomatedSteps(t *testing.T) {
	t.Parallel()

	response := toTransactionResponse(ports.TransactionResult{
		ID:               "tx-derived-exit-classification",
		Classification:   "aligned",
		Status:           "ai-reviewed",
		TenorDescription: "Y-256",
		PipelineResult: &ports.PipelineResultDetails{
			Version:                       "react_v1",
			BatchID:                       "batch-derived-exit-classification",
			ExitStep:                      3,
			NotAlignedListMatch:           false,
			NotAlignedListMatchConfidence: 10,
			AlignedListMatch:              false,
			AlignedListMatchConfidence:    6,
			FinalClassification:           "aligned",
			Reason:                        "awaiting low risk determination",
		},
	})

	payload := mustMarshalToMap(t, response)
	if payload["exit_classification"] != "next_step" {
		t.Fatalf("payload[exit_classification] = %v, want %q", payload["exit_classification"], "next_step")
	}

	automated := requireAutomatedClassificationData(t, payload)
	assertAutomatedStep(t, automated, "step1", step1Question, automatedStepAnswerNo, float64Pointer(10))
	assertAutomatedStep(t, automated, "step2", step2Question, automatedStepAnswerNo, float64Pointer(6))
	assertAutomatedStep(t, automated, "step3", step3Question, automatedStepAnswerNo, nil)
	if automated["result"] != stepAnswerNextStep {
		t.Fatalf("automated[result] = %v, want %q", automated["result"], stepAnswerNextStep)
	}
}

func TestResolveAutomatedClassificationReasonKeepsNonStep3ReasonUnchanged(t *testing.T) {
	t.Parallel()

	reason := resolveAutomatedClassificationReason(
		&ports.PipelineResultDetails{
			Version:             "react_v1",
			ExitStep:            2,
			FinalClassification: "aligned",
			Reason:              "matches universally aligned list",
		},
		2,
		stepAnswerAligned,
	)

	if reason != "matches universally aligned list" {
		t.Fatalf("resolveAutomatedClassificationReason() = %q, want %q", reason, "matches universally aligned list")
	}
}

func TestResolveAutomatedClassificationReasonDoesNotDuplicateExistingStep3Suffix(t *testing.T) {
	t.Parallel()

	reason := resolveAutomatedClassificationReason(
		&ports.PipelineResultDetails{
			Version:             "react_v1",
			ExitStep:            3,
			FinalClassification: "aligned",
			Reason:              appendReasonSuffix("meets low risk criteria", lowRiskReasonSuffix),
		},
		3,
		stepAnswerAligned,
	)

	if reason != appendReasonSuffix("meets low risk criteria", lowRiskReasonSuffix) {
		t.Fatalf("resolveAutomatedClassificationReason() = %q, want %q", reason, appendReasonSuffix("meets low risk criteria", lowRiskReasonSuffix))
	}
}

func TestResolveAutomatedClassificationReasonRemovesVerificationNeededForAlignedStep3(t *testing.T) {
	t.Parallel()

	reason := resolveAutomatedClassificationReason(
		&ports.PipelineResultDetails{
			Version:             "react_v1",
			ExitStep:            3,
			FinalClassification: "aligned",
			Reason:              "Brand New Capital Machinery for footwear production does not map to any explicit aligned activity in the lists; further verification needed",
		},
		3,
		stepAnswerAligned,
	)

	want := appendReasonSuffix("Brand New Capital Machinery for footwear production does not map to any explicit aligned activity in the lists", lowRiskReasonSuffix)
	if reason != want {
		t.Fatalf("resolveAutomatedClassificationReason() = %q, want %q", reason, want)
	}
}

func TestSanitizeAutomatedReasonRemovesOnlyKnownDebugTrailer(t *testing.T) {
	t.Parallel()

	reason := sanitizeAutomatedReason("business fallback capacity remains relevant; legacy controls remain documented; legacy step 3 fallback boolean result: true", stepAnswerNextStep)
	if reason != "business fallback capacity remains relevant; legacy controls remain documented" {
		t.Fatalf("sanitizeAutomatedReason() = %q, want %q", reason, "business fallback capacity remains relevant; legacy controls remain documented")
	}
}

func testIntPointer(value int) *int {
	return &value
}

func testBoolPointer(value bool) *bool {
	return &value
}

func float64Pointer(value float64) *float64 {
	return &value
}

func testStringPointer(value string) *string {
	return &value
}

func requirePipelineResult(t *testing.T, payload map[string]any) map[string]any {
	t.Helper()

	if _, exists := payload["pipeline_result"]; exists {
		t.Fatalf("payload[pipeline_result] present = true, want false")
	}

	return payload
}

func requireTransactionData(t *testing.T, payload map[string]any) map[string]any {
	t.Helper()

	transactionData, ok := payload["transaction_data"].(map[string]any)
	if !ok {
		t.Fatalf("payload[transaction_data] = %T, want map[string]any", payload["transaction_data"])
	}

	return transactionData
}

func requireAutomatedClassificationData(t *testing.T, payload map[string]any) map[string]any {
	t.Helper()

	automated, ok := payload["automated_classification_data"].(map[string]any)
	if !ok {
		t.Fatalf("payload[automated_classification_data] = %T, want map[string]any", payload["automated_classification_data"])
	}

	return automated
}

func assertAutomatedStep(t *testing.T, automated map[string]any, field, question, answer string, confidence *float64) {
	t.Helper()

	step, ok := automated[field].(map[string]any)
	if !ok {
		t.Fatalf("automated[%s] = %T, want map[string]any", field, automated[field])
	}

	if step["question"] != question {
		t.Fatalf("automated[%s][question] = %v, want %q", field, step["question"], question)
	}
	if step["answer"] != answer {
		t.Fatalf("automated[%s][answer] = %v, want %q", field, step["answer"], answer)
	}

	if confidence == nil {
		if _, exists := step["confidence"]; exists {
			t.Fatalf("automated[%s][confidence] present = true, want false", field)
		}
		return
	}

	if step["confidence"] != *confidence {
		t.Fatalf("automated[%s][confidence] = %v, want %v", field, step["confidence"], *confidence)
	}
}

func assertAutomatedStepNil(t *testing.T, automated map[string]any, field string) {
	t.Helper()

	if automated[field] != nil {
		t.Fatalf("automated[%s] = %v, want nil", field, automated[field])
	}
}

func assertKeyAbsent(t *testing.T, payload map[string]any, key string) {
	t.Helper()

	if _, exists := payload[key]; exists {
		t.Fatalf("payload[%s] present = true, want false", key)
	}
}

func mustMarshalToMap(t *testing.T, value any) map[string]any {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(encoded, &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	return payload
}
