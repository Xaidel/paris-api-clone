package usecases

import (
	"reflect"
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// This test keeps non-terminal pipeline classifications intact in the outward
// result so pending review flows do not look prematurely finalized.
func TestNewTransactionResultPreservesNonTerminalClassification(t *testing.T) {
	t.Parallel()

	transaction := mustNewTransactionResultMapperTransaction(t)
	pipelineResult := valueobjects.NewReactPipelineResult(valueobjects.NewReactReviewResult(
		valueobjects.PipelineResultSourceLLMBatch,
		"react",
		"v1",
		"v1",
		"gpt-4o-mini",
		transaction.ID().String(),
		nil,
		nil,
		nil,
		nil,
		false,
		10,
		false,
		4,
		0,
		valueobjects.NextStepTransactionClassification(),
		"needs next review step",
	))

	updatedAt := time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)
	if err := transaction.MarkClassified(valueobjects.NextStepTransactionClassification(), pipelineResult, updatedAt); err != nil {
		t.Fatalf("transaction.MarkClassified() error = %v", err)
	}

	result := newTransactionResult(transaction, nil, nil, nil)
	if result.Classification != valueobjects.NextStepTransactionClassification().String() {
		t.Fatalf("result.Classification = %q, want %q", result.Classification, valueobjects.NextStepTransactionClassification().String())
	}

	if result.PipelineResult == nil {
		t.Fatal("result.PipelineResult = nil, want pipeline result")
	}

	if result.PipelineResult.FinalClassification != valueobjects.NextStepTransactionClassification().String() {
		t.Fatalf("result.PipelineResult.FinalClassification = %q, want %q", result.PipelineResult.FinalClassification, valueobjects.NextStepTransactionClassification().String())
	}

}

// This test ensures a human review overrides the pipeline classification shown
// in the outward DTO while still retaining pipeline context.
func TestNewTransactionResultOverridesPipelineClassificationAfterProfessionalReview(t *testing.T) {
	t.Parallel()

	transaction := mustNewTransactionResultMapperTransaction(t)
	pipelineResult := valueobjects.NewReactPipelineResult(valueobjects.NewReactReviewResult(
		valueobjects.PipelineResultSourceLegacyStep3Fallback,
		"react",
		"v1",
		"v1",
		"gpt-4o-mini",
		transaction.ID().String(),
		nil,
		nil,
		nil,
		nil,
		false,
		10,
		false,
		4,
		3,
		valueobjects.NextStepTransactionClassification(),
		"needs next review step",
	))

	if err := transaction.MarkClassified(valueobjects.NextStepTransactionClassification(), pipelineResult, time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("transaction.MarkClassified() error = %v", err)
	}

	if err := transaction.MarkProfessionallyReviewed(valueobjects.AlignedTransactionClassification(), time.Date(2026, time.April, 4, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("transaction.MarkProfessionallyReviewed() error = %v", err)
	}

	result := newTransactionResult(transaction, nil, nil, nil)
	if result.Classification != valueobjects.AlignedTransactionClassification().String() {
		t.Fatalf("result.Classification = %q, want %q", result.Classification, valueobjects.AlignedTransactionClassification().String())
	}

	if result.Status != valueobjects.ProfessionallyReviewedTransactionStatus().String() {
		t.Fatalf("result.Status = %q, want %q", result.Status, valueobjects.ProfessionallyReviewedTransactionStatus().String())
	}

	if result.PipelineResult == nil {
		t.Fatal("result.PipelineResult = nil, want pipeline result")
	}

	if result.PipelineResult.FinalClassification != valueobjects.AlignedTransactionClassification().String() {
		t.Fatalf("result.PipelineResult.FinalClassification = %q, want %q", result.PipelineResult.FinalClassification, valueobjects.AlignedTransactionClassification().String())
	}

}

// This test captures the Phase 4 contract cleanup: outward transaction results
// no longer expose deprecated legacy score/detail fields or the old embedding
// similarity field.
func TestNewTransactionResultOmitsDeprecatedLegacyAndEmbeddingPipelineFields(t *testing.T) {
	t.Parallel()

	transaction := mustNewTransactionResultMapperTransaction(t)
	pipelineResult := valueobjects.NewReactPipelineResult(valueobjects.NewReactReviewResult(
		valueobjects.PipelineResultSourceExactHistoricalMatch,
		"react",
		"v1",
		"v1",
		"gpt-4o-mini",
		transaction.ID().String(),
		resultMapperStringPointer(transaction.ID().String()),
		resultMapperStringPointer("solar panels"),
		nil,
		nil,
		false,
		10,
		true,
		8,
		2,
		valueobjects.AlignedTransactionClassification(),
		"historical exact reuse",
	))

	if err := transaction.MarkClassifiedFromPreviousTransaction(valueobjects.AlignedTransactionClassification(), pipelineResult, time.Date(2026, time.April, 3, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("transaction.MarkClassifiedFromPreviousTransaction() error = %v", err)
	}

	result := newTransactionResult(transaction, nil, nil, nil)
	if result.PipelineResult == nil {
		t.Fatal("result.PipelineResult = nil, want pipeline result")
	}

	pipelineResultType := reflect.TypeOf(*result.PipelineResult)
	for _, fieldName := range []string{
		"Source",
		"ClassifierFamily",
		"PromptVersion",
		"Model",
		"MatchedTransactionID",
		"MatchedGoodsDescription",
		"BatchSize",
		"TransactionID",
		"OverallClassification",
		"EmbeddingSimilarity",
		"CombinedKeywordScore",
		"CombinedSemanticScore",
		"ConfidenceScore",
	} {
		if _, exists := pipelineResultType.FieldByName(fieldName); exists {
			t.Fatalf("PipelineResultDetails field %q present = true, want false", fieldName)
		}
	}

	stepResultType := reflect.TypeOf(result.PipelineResult.Step1Result)
	for _, fieldName := range []string{"KeywordDecision", "SemanticDecision"} {
		if _, exists := stepResultType.FieldByName(fieldName); exists {
			t.Fatalf("StepResultDetails field %q present = true, want false", fieldName)
		}
	}
}

// This test verifies the mapper attaches step-4 classification details and the
// resolved sector name to the outward result.
func TestNewTransactionResultIncludesStep4Classification(t *testing.T) {
	t.Parallel()

	transaction := mustNewTransactionResultMapperTransaction(t)
	if err := transaction.MarkProfessionallyReviewed(valueobjects.AlignedTransactionClassification(), time.Date(2026, time.April, 4, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("transaction.MarkProfessionallyReviewed() error = %v", err)
	}

	sectorID, err := valueobjects.SectorIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("SectorIDFromString() error = %v", err)
	}

	additionalContext, err := valueobjects.NewTransactionStep4AdditionalContext("Reviewed by analyst")
	if err != nil {
		t.Fatalf("NewTransactionStep4AdditionalContext() error = %v", err)
	}

	step4, err := entities.NewTransactionStep4(transaction.ID(), sectorID, additionalContext, false, time.Date(2026, time.April, 4, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("entities.NewTransactionStep4() error = %v", err)
	}

	sector, err := entities.NewSector(sectorID, "High Emitting", "Energy", "High emitting energy sector")
	if err != nil {
		t.Fatalf("entities.NewSector() error = %v", err)
	}

	result := newTransactionResult(transaction, step4, sector, nil)
	if result.Step4Classification == nil {
		t.Fatal("result.Step4Classification = nil, want step 4 classification")
	}

	if result.Step4Classification.IdentifiedSector != "Energy" {
		t.Fatalf("result.Step4Classification.IdentifiedSector = %q, want %q", result.Step4Classification.IdentifiedSector, "Energy")
	}

	if result.Step4Classification.AdditionalInformation != "Reviewed by analyst" {
		t.Fatalf("result.Step4Classification.AdditionalInformation = %q, want %q", result.Step4Classification.AdditionalInformation, "Reviewed by analyst")
	}

	if result.Step4Classification.Result != "aligned" {
		t.Fatalf("result.Step4Classification.Result = %q, want %q", result.Step4Classification.Result, "aligned")
	}
}

// This test verifies the mapper attaches step-5 screening answers and trimmed
// reviewer notes to the outward result.
func TestNewTransactionResultIncludesStep5Classification(t *testing.T) {
	t.Parallel()

	transaction := mustNewTransactionResultMapperTransaction(t)
	if err := transaction.MarkProfessionallyReviewed(valueobjects.NotAlignedTransactionClassification(), time.Date(2026, time.April, 4, 12, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("transaction.MarkProfessionallyReviewed() error = %v", err)
	}

	question1Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion1Justification("question 1 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion1Justification() error = %v", err)
	}

	question2Justification, err := valueobjects.NewTransactionStep5ScreeningQuestion2Justification("question 2 justification")
	if err != nil {
		t.Fatalf("NewTransactionStep5ScreeningQuestion2Justification() error = %v", err)
	}

	step5, err := entities.NewTransactionStep5(
		transaction.ID(),
		true,
		question1Justification,
		false,
		question2Justification,
		valueobjects.NewTransactionStep5ReviewerNotes(resultMapperStringPointer("  optional note  ")),
		true,
		time.Date(2026, time.April, 4, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("entities.NewTransactionStep5() error = %v", err)
	}

	result := newTransactionResult(transaction, nil, nil, step5)
	if result.Step5Classification == nil {
		t.Fatal("result.Step5Classification = nil, want step 5 classification")
	}

	if !result.Step5Classification.ScreeningQuestion1.Answer {
		t.Fatal("result.Step5Classification.ScreeningQuestion1.Answer = false, want true")
	}

	if result.Step5Classification.ScreeningQuestion1.Justification != "question 1 justification" {
		t.Fatalf("result.Step5Classification.ScreeningQuestion1.Justification = %q, want %q", result.Step5Classification.ScreeningQuestion1.Justification, "question 1 justification")
	}

	if result.Step5Classification.Result != valueobjects.NotAlignedTransactionClassification().String() {
		t.Fatalf("result.Step5Classification.Result = %q, want %q", result.Step5Classification.Result, valueobjects.NotAlignedTransactionClassification().String())
	}

	if result.Step5Classification.ReviewerNotes == nil || *result.Step5Classification.ReviewerNotes != "optional note" {
		t.Fatalf("result.Step5Classification.ReviewerNotes = %v, want %q", result.Step5Classification.ReviewerNotes, "optional note")
	}
}

func resultMapperStringPointer(value string) *string {
	return &value
}

func mustNewTransactionResultMapperTransaction(t *testing.T) *entities.Transaction {
	t.Helper()

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transaction, err := entities.NewTransaction(
		transactionID,
		"CG",
		2026,
		4,
		"IB",
		"DMC",
		"Partner Bank",
		"REF-1",
		"698436.80",
		1,
		"Goods",
		"Classification",
		"Philippines",
		"Japan",
		"Thailand",
		"Philippines",
		"Y",
		"",
		"PA Aligned",
		time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("entities.NewTransaction() error = %v", err)
	}

	return transaction
}
