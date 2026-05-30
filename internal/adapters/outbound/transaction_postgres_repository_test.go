package adapters

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	"github.com/pashagolub/pgxmock/v4"
)

// TestPostgresTransactionRepositoryGetNavigation verifies the postgres
// transaction navigation lookup query behavior and the expected outcome
// asserted below.
func TestPostgresTransactionRepositoryGetNavigation(t *testing.T) {
	t.Parallel()

	t.Run("returns navigation with previous and next ids", func(t *testing.T) {
		t.Parallel()

		mock, repository := newNavigationRepository(t)
		lookup := ports.TransactionNavigationLookup{TransactionID: navigationTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61")}

		query, args := buildTransactionNavigationQuery(lookup)
		if !strings.Contains(query, "filtered_transactions AS") {
			t.Fatalf("buildTransactionNavigationQuery() query = %q, want filtered transaction scope", query)
		}
		if !strings.Contains(query, "previous_transaction") {
			t.Fatalf("buildTransactionNavigationQuery() query = %q, want previous_transaction CTE", query)
		}
		if !strings.Contains(query, "next_transaction") {
			t.Fatalf("buildTransactionNavigationQuery() query = %q, want next_transaction CTE", query)
		}

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(args...).
			WillReturnRows(pgxmock.NewRows([]string{"transaction_id", "previous_id", "next_id"}).AddRow(
				lookup.TransactionID.String(),
				"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60",
				"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e62",
			))

		result, err := repository.GetNavigation(context.Background(), lookup)
		if err != nil {
			t.Fatalf("GetNavigation() error = %v", err)
		}

		if result == nil {
			t.Fatal("GetNavigation() = nil, want navigation result")
		}

		if result.TransactionID != lookup.TransactionID.String() {
			t.Fatalf("result.TransactionID = %q, want %q", result.TransactionID, lookup.TransactionID.String())
		}

		if result.PreviousID == nil || *result.PreviousID != "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60" {
			t.Fatalf("result.PreviousID = %v, want previous id", result.PreviousID)
		}

		if result.NextID == nil || *result.NextID != "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e62" {
			t.Fatalf("result.NextID = %v, want next id", result.NextID)
		}
	})

	t.Run("returns nil when current transaction is filtered out", func(t *testing.T) {
		t.Parallel()

		mock, repository := newNavigationRepository(t)
		uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
		if err != nil {
			t.Fatalf("UploadIDFromString() error = %v", err)
		}
		lookup := ports.TransactionNavigationLookup{
			TransactionID: navigationTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"),
			Filter:        ports.TransactionFilter{UploadID: &uploadID},
		}

		query, args := buildTransactionNavigationQuery(lookup)
		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args...).WillReturnRows(
			pgxmock.NewRows([]string{"transaction_id", "previous_id", "next_id"}),
		)

		result, err := repository.GetNavigation(context.Background(), lookup)
		if err != nil {
			t.Fatalf("GetNavigation() error = %v", err)
		}

		if result != nil {
			t.Fatalf("GetNavigation() = %v, want nil", result)
		}
	})

	t.Run("includes classification filter in lookup scope", func(t *testing.T) {
		t.Parallel()

		mock, repository := newNavigationRepository(t)
		classification := valueobjects.AlignedTransactionClassification().String()
		lookup := ports.TransactionNavigationLookup{
			TransactionID:  navigationTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"),
			Classification: &classification,
		}

		query, args := buildTransactionNavigationQuery(lookup)
		if !strings.Contains(query, "t.classification = $2") {
			t.Fatalf("buildTransactionNavigationQuery() query = %q, want classification predicate", query)
		}
		if len(args) != 2 {
			t.Fatalf("len(args) = %d, want %d", len(args), 2)
		}
		if args[1] != classification {
			t.Fatalf("args[1] = %v, want %q", args[1], classification)
		}

		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args...).WillReturnRows(
			pgxmock.NewRows([]string{"transaction_id", "previous_id", "next_id"}).AddRow(lookup.TransactionID.String(), nil, nil),
		)

		result, err := repository.GetNavigation(context.Background(), lookup)
		if err != nil {
			t.Fatalf("GetNavigation() error = %v", err)
		}
		if result == nil || result.TransactionID != lookup.TransactionID.String() {
			t.Fatalf("GetNavigation() = %v, want current transaction result", result)
		}
	})

	t.Run("uses transaction step 5 join path", func(t *testing.T) {
		t.Parallel()

		mock, repository := newNavigationRepository(t)
		step := 5
		lookup := ports.TransactionNavigationLookup{
			TransactionID: navigationTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"),
			Step:          &step,
		}

		query, args := buildTransactionNavigationQuery(lookup)
		if !strings.Contains(query, "JOIN transaction_step_5_data ts5 ON ts5.transaction_id = t.id") {
			t.Fatalf("buildTransactionNavigationQuery() query = %q, want step 5 join", query)
		}
		if strings.Contains(query, "transaction_step_4 ts4") {
			t.Fatalf("buildTransactionNavigationQuery() query = %q, want step 5 scope to avoid step 4 join", query)
		}

		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args...).WillReturnRows(
			pgxmock.NewRows([]string{"transaction_id", "previous_id", "next_id"}).AddRow(lookup.TransactionID.String(), nil, nil),
		)

		result, err := repository.GetNavigation(context.Background(), lookup)
		if err != nil {
			t.Fatalf("GetNavigation() error = %v", err)
		}
		if result == nil || result.TransactionID != lookup.TransactionID.String() {
			t.Fatalf("GetNavigation() = %v, want current transaction result", result)
		}
	})

	t.Run("uses transaction step 4 join predicate shape", func(t *testing.T) {
		t.Parallel()

		mock, repository := newNavigationRepository(t)
		step := 4
		lookup := ports.TransactionNavigationLookup{
			TransactionID: navigationTransactionID(t, "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61"),
			Step:          &step,
		}

		query, args := buildTransactionNavigationQuery(lookup)
		if !strings.Contains(query, "JOIN transaction_step_4 ts4 ON ts4.transaction_id = t.id") {
			t.Fatalf("buildTransactionNavigationQuery() query = %q, want step 4 join", query)
		}
		if !strings.Contains(query, "LEFT JOIN transaction_step_5_data ts5 ON ts5.transaction_id = t.id") {
			t.Fatalf("buildTransactionNavigationQuery() query = %q, want step 5 exclusion join", query)
		}
		if !strings.Contains(query, "ts5.transaction_id IS NULL") {
			t.Fatalf("buildTransactionNavigationQuery() query = %q, want step 5 exclusion predicate", query)
		}

		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args...).WillReturnRows(
			pgxmock.NewRows([]string{"transaction_id", "previous_id", "next_id"}).AddRow(lookup.TransactionID.String(), nil, nil),
		)

		result, err := repository.GetNavigation(context.Background(), lookup)
		if err != nil {
			t.Fatalf("GetNavigation() error = %v", err)
		}
		if result == nil || result.TransactionID != lookup.TransactionID.String() {
			t.Fatalf("GetNavigation() = %v, want current transaction result", result)
		}
	})
}

func newNavigationRepository(t *testing.T) (pgxmock.PgxPoolIface, *PostgresTransactionRepository) {
	t.Helper()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	t.Cleanup(func() { mock.Close() })

	return mock, NewPostgresTransactionRepository(mock)
}

func navigationTransactionID(t *testing.T, raw string) valueobjects.TransactionID {
	t.Helper()

	id, err := valueobjects.TransactionIDFromString(raw)
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	return id
}

// This test guards the JSON round trip used for persisted pipeline results so
// classifier metadata survives storage and later rehydration unchanged.
func TestMarshalAndParsePipelineResultReactPreservesClassifierVersion(t *testing.T) {
	t.Parallel()

	result := valueobjects.NewReactPipelineResult(valueobjects.NewReactReviewResult(
		valueobjects.PipelineResultSourceLLMBatch,
		"react",
		"classifier-v7",
		"prompt-v3",
		"gpt-4.1-mini",
		"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60",
		nil,
		nil,
		nil,
		nil,
		false,
		2,
		true,
		8,
		2,
		valueobjects.AlignedTransactionClassification(),
		"classifier produced aligned result",
	))

	encoded, err := marshalPipelineResult(&result)
	if err != nil {
		t.Fatalf("marshalPipelineResult() error = %v", err)
	}

	parsed, err := parsePipelineResult(encoded)
	if err != nil {
		t.Fatalf("parsePipelineResult() error = %v", err)
	}

	if parsed == nil || parsed.React() == nil {
		t.Fatal("parsePipelineResult() = nil react result, want react result")
	}

	react := parsed.React()
	if react.ClassifierVersion() != "classifier-v7" {
		t.Fatalf("react.ClassifierVersion() = %q, want %q", react.ClassifierVersion(), "classifier-v7")
	}

	if react.PromptVersion() != "prompt-v3" {
		t.Fatalf("react.PromptVersion() = %q, want %q", react.PromptVersion(), "prompt-v3")
	}
}

// This test ensures the persisted pipeline payload keeps the authoritative exit
// step, which downstream code uses to interpret partial and step-3 outcomes.
func TestMarshalAndParsePipelineResultReactPreservesExitStep(t *testing.T) {
	t.Parallel()

	result := valueobjects.NewReactPipelineResult(valueobjects.NewReactReviewResult(
		valueobjects.PipelineResultSourceLLMBatch,
		"react",
		"classifier-v7",
		"prompt-v3",
		"gpt-4.1-mini",
		"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60",
		nil,
		nil,
		nil,
		nil,
		false,
		1,
		true,
		8,
		2,
		valueobjects.AlignedTransactionClassification(),
		"aligned in step 2",
	))

	encoded, err := marshalPipelineResult(&result)
	if err != nil {
		t.Fatalf("marshalPipelineResult() error = %v", err)
	}

	parsed, err := parsePipelineResult(encoded)
	if err != nil {
		t.Fatalf("parsePipelineResult() error = %v", err)
	}

	if parsed == nil || parsed.React() == nil {
		t.Fatal("parsePipelineResult() = nil react result, want react result")
	}

	if parsed.React().ExitStep() != 2 {
		t.Fatalf("parsed.React().ExitStep() = %d, want %d", parsed.React().ExitStep(), 2)
	}

	if parsed.ExitStep() != 2 {
		t.Fatalf("parsed.ExitStep() = %d, want %d", parsed.ExitStep(), 2)
	}
}

// This test preserves backward compatibility for older stored payloads that
// only recorded prompt_version instead of classifier_version.
func TestParsePipelineResultReactFallsBackClassifierVersionToPromptVersion(t *testing.T) {
	t.Parallel()

	encoded := []byte(`{"version":"react_v1","source":"llm_batch","classifier_family":"react","prompt_version":"prompt-v3","model":"gpt-4.1-mini","transaction_id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","not_aligned_list_match":false,"not_aligned_list_match_confidence":1,"aligned_list_match":true,"aligned_list_match_confidence":8,"overall_classification":"aligned","reason":"fallback works"}`)

	parsed, err := parsePipelineResult(encoded)
	if err != nil {
		t.Fatalf("parsePipelineResult() error = %v", err)
	}

	if parsed == nil || parsed.React() == nil {
		t.Fatal("parsePipelineResult() = nil react result, want react result")
	}

	react := parsed.React()
	if react.ClassifierVersion() != "prompt-v3" {
		t.Fatalf("react.ClassifierVersion() = %q, want %q", react.ClassifierVersion(), "prompt-v3")
	}
}

// This test captures the Phase 3 write-path cleanup: new ReAct payloads must
// stop emitting embedding_similarity even if older in-memory shapes still carry it.
func TestMarshalPipelineResultReactOmitsEmbeddingSimilarity(t *testing.T) {
	t.Parallel()

	result := valueobjects.NewReactPipelineResult(valueobjects.NewReactReviewResult(
		valueobjects.PipelineResultSourceExactHistoricalMatch,
		"react",
		"classifier-v7",
		"prompt-v3",
		"gpt-4.1-mini",
		"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60",
		resultMapperStringPointerTransactionRepo("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e99"),
		resultMapperStringPointerTransactionRepo("solar panels"),
		nil,
		nil,
		false,
		1,
		true,
		8,
		2,
		valueobjects.AlignedTransactionClassification(),
		"historical exact reuse",
	))

	encoded, err := marshalPipelineResult(&result)
	if err != nil {
		t.Fatalf("marshalPipelineResult() error = %v", err)
	}

	if bytes.Contains(encoded, []byte("embedding_similarity")) {
		t.Fatalf("marshalPipelineResult() encoded %s, want no embedding_similarity field", string(encoded))
	}
}

// This test preserves legacy payload compatibility while Phase 3 stops writing
// embedding similarity for new ReAct results.
func TestParsePipelineResultReactLegacyEmbeddingPayloadStillParses(t *testing.T) {
	t.Parallel()

	encoded := []byte(`{"version":"react_v1","source":"exact_historical_match","classifier_family":"react","classifier_version":"classifier-v7","prompt_version":"prompt-v3","model":"gpt-4.1-mini","transaction_id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","matched_transaction_id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e99","matched_goods_description":"solar panels","embedding_similarity":0.9999,"not_aligned_list_match":false,"not_aligned_list_match_confidence":1,"aligned_list_match":true,"aligned_list_match_confidence":8,"exit_step":2,"overall_classification":"aligned","reason":"old payload still decodes"}`)

	parsed, err := parsePipelineResult(encoded)
	if err != nil {
		t.Fatalf("parsePipelineResult() error = %v", err)
	}

	if parsed == nil || parsed.React() == nil {
		t.Fatal("parsePipelineResult() = nil react result, want react result")
	}

	react := parsed.React()
	if react.Source() != valueobjects.PipelineResultSourceExactHistoricalMatch {
		t.Fatalf("react.Source() = %q, want %q", react.Source(), valueobjects.PipelineResultSourceExactHistoricalMatch)
	}

	if react.MatchedGoodsDescription() == nil || *react.MatchedGoodsDescription() != "solar panels" {
		t.Fatalf("react.MatchedGoodsDescription() = %v, want %q", react.MatchedGoodsDescription(), "solar panels")
	}

	if react.ExitStep() != 2 {
		t.Fatalf("react.ExitStep() = %d, want %d", react.ExitStep(), 2)
	}
}

// This test captures the Phase 4 write-path cleanup for legacy payloads: new
// records should keep only the step summaries still needed by the read model.
func TestMarshalPipelineResultLegacyOmitsDeprecatedMatcherAndScoreFields(t *testing.T) {
	t.Parallel()

	step1 := valueobjects.NewStepResult(
		1,
		valueobjects.NewMatchDecision(valueobjects.AlignedAlignment(), 0.9, "entry-1", nil),
		valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.4, "entry-2", nil),
	)
	step2 := valueobjects.NewStepResult(
		2,
		valueobjects.NewMatchDecision(valueobjects.UnalignedAlignment(), 0.3, "entry-3", nil),
		valueobjects.NewMatchDecision(valueobjects.AlignedAlignment(), 0.8, "entry-4", nil),
	)
	step3 := valueobjects.NewBooleanStepResultWithReason(3, false, errors.New("legacy step 3 fallback boolean result: false"))
	result := valueobjects.NewPipelineResult(
		"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60",
		step1,
		&step2,
		&step3,
		3,
		valueobjects.NextStepTransactionClassification(),
	)

	encoded, err := marshalPipelineResult(&result)
	if err != nil {
		t.Fatalf("marshalPipelineResult() error = %v", err)
	}

	for _, forbiddenField := range [][]byte{
		[]byte("keyword_decision"),
		[]byte("semantic_decision"),
		[]byte("combined_keyword_score"),
		[]byte("combined_semantic_score"),
		[]byte("confidence_score"),
	} {
		if bytes.Contains(encoded, forbiddenField) {
			t.Fatalf("marshalPipelineResult() encoded %s, want no %q field", string(encoded), string(forbiddenField))
		}
	}

	parsed, err := parsePipelineResult(encoded)
	if err != nil {
		t.Fatalf("parsePipelineResult() error = %v", err)
	}

	if parsed == nil {
		t.Fatal("parsePipelineResult() = nil, want pipeline result")
	}

	if !parsed.Step1Result().StepAlignment().Equal(valueobjects.AlignedAlignment()) {
		t.Fatalf("parsed.Step1Result().StepAlignment() = %q, want %q", parsed.Step1Result().StepAlignment().String(), valueobjects.AlignedAlignment().String())
	}

	if parsed.Step2Result() == nil {
		t.Fatal("parsed.Step2Result() = nil, want step 2 result")
	}

	if !parsed.Step2Result().StepAlignment().Equal(valueobjects.AlignedAlignment()) {
		t.Fatalf("parsed.Step2Result().StepAlignment() = %q, want %q", parsed.Step2Result().StepAlignment().String(), valueobjects.AlignedAlignment().String())
	}

	if parsed.Step3Result() == nil {
		t.Fatal("parsed.Step3Result() = nil, want step 3 result")
	}

	if parsed.Step3Result().BooleanResult() == nil || *parsed.Step3Result().BooleanResult() {
		t.Fatalf("parsed.Step3Result().BooleanResult() = %v, want %v", parsed.Step3Result().BooleanResult(), false)
	}

	if parsed.Step3Result().Reason() == nil || parsed.Step3Result().Reason().Error() != "legacy step 3 fallback boolean result: false" {
		t.Fatalf("parsed.Step3Result().Reason() = %v, want %q", parsed.Step3Result().Reason(), "legacy step 3 fallback boolean result: false")
	}

	pipelineRecordType := reflect.TypeOf(pipelineResultRecord{})
	for _, fieldName := range []string{"EmbeddingSimilarity", "CombinedKeywordScore", "CombinedSemanticScore", "ConfidenceScore"} {
		if _, exists := pipelineRecordType.FieldByName(fieldName); exists {
			t.Fatalf("pipelineResultRecord field %q present = true, want false", fieldName)
		}
	}

	stepRecordType := reflect.TypeOf(stepResultRecord{})
	for _, fieldName := range []string{"KeywordDecision", "SemanticDecision"} {
		if _, exists := stepRecordType.FieldByName(fieldName); exists {
			t.Fatalf("stepResultRecord field %q present = true, want false", fieldName)
		}
	}
}

func resultMapperStringPointerTransactionRepo(value string) *string {
	return &value
}
