package adapters

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type reactChatModelStub struct {
	messages  []*schema.Message
	errs      []error
	inputs    [][]*schema.Message
	callCount int
}

func (s *reactChatModelStub) Generate(_ context.Context, input []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	if s == nil {
		return nil, nil
	}

	s.callCount++
	s.inputs = append(s.inputs, append([]*schema.Message(nil), input...))

	if len(s.messages) == 0 && len(s.errs) == 0 {
		return nil, nil
	}

	var message *schema.Message
	if len(s.messages) > 0 {
		message = s.messages[0]
		s.messages = s.messages[1:]
	}

	var err error
	if len(s.errs) > 0 {
		err = s.errs[0]
		s.errs = s.errs[1:]
	}

	return message, err
}

type reactTransactionRepositoryStub struct {
	exactMatch *ports.HistoricalTransactionClassificationMatch
	findByID   *entities.Transaction
}

type classificationMetricsStub struct {
	pipelineTotal                  []string
	stepExit                       []struct{ step, alignment string }
	pipelineDurations              []float64
	confidenceStubTotal            int
	reactHistoricalExactMatchTotal int
	reactLLMRouteTotal             int
	reactLLMBatchSizes             []int
	reactLLMLatencies              []float64
	reactInvalidLLMOutputTotal     int
	reactClassification            []string
	reactStatus                    []string
}

func (s *classificationMetricsStub) IncPipelineTotal(status string) {
	s.pipelineTotal = append(s.pipelineTotal, status)
}

func (s *classificationMetricsStub) IncStepExitTotal(step string, alignment string) {
	s.stepExit = append(s.stepExit, struct{ step, alignment string }{step: step, alignment: alignment})
}

func (s *classificationMetricsStub) ObservePipelineDurationSeconds(seconds float64) {
	s.pipelineDurations = append(s.pipelineDurations, seconds)
}

func (s *classificationMetricsStub) IncConfidenceStubTotal() {
	s.confidenceStubTotal++
}

func (s *classificationMetricsStub) IncReactHistoricalExactMatchTotal() {
	s.reactHistoricalExactMatchTotal++
}

func (s *classificationMetricsStub) IncReactLLMRouteTotal() {
	s.reactLLMRouteTotal++
}

func (s *classificationMetricsStub) ObserveReactLLMBatchSize(size int) {
	s.reactLLMBatchSizes = append(s.reactLLMBatchSizes, size)
}

func (s *classificationMetricsStub) ObserveReactLLMLatencySeconds(seconds float64) {
	s.reactLLMLatencies = append(s.reactLLMLatencies, seconds)
}

func (s *classificationMetricsStub) IncReactInvalidLLMOutputTotal() {
	s.reactInvalidLLMOutputTotal++
}

func (s *classificationMetricsStub) IncReactClassificationTotal(classification string) {
	s.reactClassification = append(s.reactClassification, classification)
}

func (s *classificationMetricsStub) IncReactStatusTotal(status string) {
	s.reactStatus = append(s.reactStatus, status)
}

func (s *reactTransactionRepositoryStub) Create(context.Context, *entities.Transaction, string) error {
	return nil
}
func (s *reactTransactionRepositoryStub) CreateMany(context.Context, []*entities.Transaction, string) error {
	return nil
}
func (s *reactTransactionRepositoryStub) Update(context.Context, *entities.Transaction) error {
	return nil
}
func (s *reactTransactionRepositoryStub) FindByID(context.Context, valueobjects.TransactionID) (*entities.Transaction, error) {
	return s.findByID, nil
}
func (s *reactTransactionRepositoryStub) FindHistoricalClassificationByExactGoodsDescription(context.Context, ports.HistoricalTransactionClassificationQuery) (*ports.HistoricalTransactionClassificationMatch, error) {
	return s.exactMatch, nil
}
func (s *reactTransactionRepositoryStub) GetNavigation(context.Context, ports.TransactionNavigationLookup) (*ports.TransactionNavigationResult, error) {
	return nil, nil
}
func (s *reactTransactionRepositoryStub) List(context.Context, ports.TransactionFilter) ([]*entities.Transaction, error) {
	return nil, nil
}
func (s *reactTransactionRepositoryStub) ListByUploadIDs(context.Context, []valueobjects.UploadID) ([]*entities.Transaction, error) {
	return nil, nil
}
func (s *reactTransactionRepositoryStub) HasProcessingByUploadID(context.Context, valueobjects.UploadID) (bool, error) {
	return false, nil
}
func (s *reactTransactionRepositoryStub) DeleteByID(context.Context, valueobjects.TransactionID) error {
	return nil
}
func (s *reactTransactionRepositoryStub) DeleteByUploadID(context.Context, valueobjects.UploadID) error {
	return nil
}

// TestReActTransactionClassificationGatewayUsesUUIDv7BatchID verifies the ReAct act transaction classification gateway uses uui dv 7 batch ID behavior and the expected outcome asserted below.
func TestReActTransactionClassificationGatewayUsesUUIDv7BatchID(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	repo := &reactTransactionRepositoryStub{}
	chatModel := &reactChatModelStub{messages: []*schema.Message{{Content: `[{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","not_aligned_list_match":false,"not_aligned_list_match_confidence":10,"aligned_list_match":true,"aligned_list_match_confidence":9,"overall_classification":"aligned","reason":"matches renewable energy manufacturing."}]`}}}

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("test prompt"),
		WithReActTransactionClassificationLogger(zap.NewNop()),
	)

	decisions, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "solar panels",
	}})
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("len(decisions) = %d, want 1", len(decisions))
	}

	react := decisions[0].ReviewResult.React()
	if react == nil {
		t.Fatal("decisions[0].ReviewResult.React() = nil, want non-nil")
	}

	batchID := react.BatchID()
	if batchID == nil {
		t.Fatal("react.BatchID() = nil, want non-nil")
	}

	parsedBatchID, err := uuid.Parse(*batchID)
	if err != nil {
		t.Fatalf("uuid.Parse(%q) error = %v", *batchID, err)
	}

	if parsedBatchID.Version() != 7 {
		t.Fatalf("parsedBatchID.Version() = %d, want 7", parsedBatchID.Version())
	}
}

// TestReActTransactionClassificationGatewayRepairsInvalidOutputAndLogsRepair verifies the ReAct act transaction classification gateway repairs invalid output and logs repair behavior and the expected outcome asserted below.
func TestReActTransactionClassificationGatewayRepairsInvalidOutputAndLogsRepair(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	repo := &reactTransactionRepositoryStub{}
	metrics := &classificationMetricsStub{}
	observedCore, observedLogs := observer.New(zapcore.InfoLevel)
	chatModel := &reactChatModelStub{messages: []*schema.Message{
		{Content: `{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60"}`},
		{Content: `[{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","not_aligned_list_match":false,"not_aligned_list_match_confidence":2,"aligned_list_match":true,"aligned_list_match_confidence":8,"overall_classification":"aligned","reason":"repair succeeded"}]`},
	}}

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("test prompt"),
		WithReActTransactionClassificationClassifier("react", "v1"),
		WithReActTransactionClassificationPromptVersion("prompt-v2"),
		WithReActTransactionClassificationModelName("gpt-4.1-mini"),
		WithReActTransactionClassificationMetrics(metrics),
		WithReActTransactionClassificationLogger(zap.New(observedCore)),
	)

	decisions, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "solar panels",
	}})
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("len(decisions) = %d, want 1", len(decisions))
	}

	entries := observedLogs.All()
	if len(entries) < 3 {
		t.Fatalf("len(entries) = %d, want at least 3", len(entries))
	}

	if entries[0].Message != "react classification llm batch returned invalid output" {
		t.Fatalf("entries[0].Message = %q", entries[0].Message)
	}

	if entries[1].Message != "react classification llm batch requesting repair" {
		t.Fatalf("entries[1].Message = %q", entries[1].Message)
	}

	if entries[2].Message != "react classification llm batch repair succeeded" {
		t.Fatalf("entries[2].Message = %q", entries[2].Message)
	}

	if metrics.reactInvalidLLMOutputTotal != 1 {
		t.Fatalf("metrics.reactInvalidLLMOutputTotal = %d, want 1", metrics.reactInvalidLLMOutputTotal)
	}
}

// TestReActTransactionClassificationGatewayReordersRepairedOutputByID verifies the ReAct act transaction classification gateway reorders repaired output by ID behavior and the expected outcome asserted below.
func TestReActTransactionClassificationGatewayReordersRepairedOutputByID(t *testing.T) {
	t.Parallel()

	firstTransactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	secondTransactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61")
	repo := &reactTransactionRepositoryStub{}
	chatModel := &reactChatModelStub{messages: []*schema.Message{
		{Content: `{"id":"invalid"}`},
		{Content: `[
			{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61","not_aligned_list_match":false,"not_aligned_list_match_confidence":10,"aligned_list_match":true,"aligned_list_match_confidence":8,"overall_classification":"aligned","reason":"wind result"},
			{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","not_aligned_list_match":false,"not_aligned_list_match_confidence":10,"aligned_list_match":true,"aligned_list_match_confidence":9,"overall_classification":"aligned","reason":"solar result"}
		]`},
	}}

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("test prompt"),
		WithReActTransactionClassificationLogger(zap.NewNop()),
	)

	decisions, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    firstTransactionID,
		GoodsDescription: "solar panels",
	}, {
		TransactionID:    secondTransactionID,
		GoodsDescription: "wind turbine",
	}})
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	if len(decisions) != 2 {
		t.Fatalf("len(decisions) = %d, want 2", len(decisions))
	}

	if decisions[0].TransactionID.String() != firstTransactionID.String() {
		t.Fatalf("decisions[0].TransactionID = %q, want %q", decisions[0].TransactionID.String(), firstTransactionID.String())
	}

	if decisions[0].ReviewResult.React().Reason() != "solar result" {
		t.Fatalf("decisions[0].ReviewResult.React().Reason() = %q, want %q", decisions[0].ReviewResult.React().Reason(), "solar result")
	}

	if decisions[0].ReviewResult.React().ExitStep() != 2 {
		t.Fatalf("decisions[0].ReviewResult.React().ExitStep() = %d, want %d", decisions[0].ReviewResult.React().ExitStep(), 2)
	}

	if decisions[1].TransactionID.String() != secondTransactionID.String() {
		t.Fatalf("decisions[1].TransactionID = %q, want %q", decisions[1].TransactionID.String(), secondTransactionID.String())
	}

	if decisions[1].ReviewResult.React().Reason() != "wind result" {
		t.Fatalf("decisions[1].ReviewResult.React().Reason() = %q, want %q", decisions[1].ReviewResult.React().Reason(), "wind result")
	}

	if decisions[1].ReviewResult.React().ExitStep() != 2 {
		t.Fatalf("decisions[1].ReviewResult.React().ExitStep() = %d, want %d", decisions[1].ReviewResult.React().ExitStep(), 2)
	}
}

// TestReActTransactionClassificationGatewayUsesDynamicSystemPromptForRepair verifies the ReAct act transaction classification gateway uses dynamic system prompt for repair behavior and the expected outcome asserted below.
func TestReActTransactionClassificationGatewayUsesDynamicSystemPromptForRepair(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	repo := &reactTransactionRepositoryStub{}
	chatModel := &reactChatModelStub{messages: []*schema.Message{
		{Content: `{"id":"invalid"}`},
		{Content: `[{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","not_aligned_list_match":false,"not_aligned_list_match_confidence":2,"aligned_list_match":true,"aligned_list_match_confidence":8,"overall_classification":"aligned","reason":"repair succeeded"}]`},
	}}

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPromptBuilder(func(context.Context) (string, error) {
			return "dynamic repair prompt", nil
		}),
		WithReActTransactionClassificationLogger(zap.NewNop()),
	)

	_, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "solar panels",
	}})
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	if len(chatModel.inputs) != 2 {
		t.Fatalf("len(chatModel.inputs) = %d, want 2", len(chatModel.inputs))
	}

	if len(chatModel.inputs[1]) != 2 {
		t.Fatalf("len(chatModel.inputs[1]) = %d, want 2", len(chatModel.inputs[1]))
	}

	if chatModel.inputs[1][0] == nil {
		t.Fatal("chatModel.inputs[1][0] = nil, want system message")
	}

	if chatModel.inputs[1][0].Content != "dynamic repair prompt" {
		t.Fatalf("chatModel.inputs[1][0].Content = %q, want %q", chatModel.inputs[1][0].Content, "dynamic repair prompt")
	}
}

// TestReActTransactionClassificationGatewayRejectsBlankSystemPrompt verifies the ReAct act transaction classification gateway rejects blank system prompt behavior and the expected outcome asserted below.
func TestReActTransactionClassificationGatewayRejectsBlankSystemPrompt(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	repo := &reactTransactionRepositoryStub{}
	chatModel := &reactChatModelStub{}

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("   "),
		WithReActTransactionClassificationLogger(zap.NewNop()),
	)

	_, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "solar panels",
	}})
	if err == nil {
		t.Fatal("Classify() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "react classification system prompt is required") {
		t.Fatalf("Classify() error = %v, want system prompt required error", err)
	}

	if chatModel.callCount != 0 {
		t.Fatalf("chatModel.callCount = %d, want 0", chatModel.callCount)
	}
}

// TestReActTransactionClassificationGatewayTracksStep1ForNotAlignedResult verifies the ReAct act transaction classification gateway tracks step 1 for not aligned result behavior and the expected outcome asserted below.
func TestReActTransactionClassificationGatewayTracksStep1ForNotAlignedResult(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	repo := &reactTransactionRepositoryStub{}
	chatModel := &reactChatModelStub{messages: []*schema.Message{{Content: `[{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","not_aligned_list_match":true,"not_aligned_list_match_confidence":10,"aligned_list_match":false,"aligned_list_match_confidence":0,"overall_classification":"not_aligned","reason":"coal is excluded"}]`}}}

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("test prompt"),
		WithReActTransactionClassificationLogger(zap.NewNop()),
	)

	decisions, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "thermal coal",
	}})
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("len(decisions) = %d, want 1", len(decisions))
	}

	if decisions[0].ReviewResult.React().ExitStep() != 1 {
		t.Fatalf("decisions[0].ReviewResult.React().ExitStep() = %d, want %d", decisions[0].ReviewResult.React().ExitStep(), 1)
	}
}

// TestReActTransactionClassificationGatewayLeavesExitStepUnsetForNextStep verifies the gateway keeps step-3 restoration out of the adapter layer.
func TestReActTransactionClassificationGatewayLeavesExitStepUnsetForNextStep(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	repo := &reactTransactionRepositoryStub{}
	chatModel := &reactChatModelStub{messages: []*schema.Message{{Content: `[{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","not_aligned_list_match":false,"not_aligned_list_match_confidence":10,"aligned_list_match":false,"aligned_list_match_confidence":4,"overall_classification":"next_step","reason":"unclear"}]`}}}

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("test prompt"),
		WithReActTransactionClassificationLogger(zap.NewNop()),
	)

	decisions, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "modified cationic starch",
	}})
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("len(decisions) = %d, want 1", len(decisions))
	}

	if decisions[0].ReviewResult.React().ExitStep() != 0 {
		t.Fatalf("decisions[0].ReviewResult.React().ExitStep() = %d, want %d", decisions[0].ReviewResult.React().ExitStep(), 0)
	}
}

// TestReActTransactionClassificationGatewaySkipsHistoricalNextStepReuse verifies the ReAct act transaction classification gateway skips historical next step reuse behavior and the expected outcome asserted below.
func TestReActTransactionClassificationGatewaySkipsHistoricalNextStepReuse(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	matchedTransactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e99")
	repo := &reactTransactionRepositoryStub{exactMatch: &ports.HistoricalTransactionClassificationMatch{
		TransactionID:    matchedTransactionID,
		GoodsDescription: "modified cationic starch",
		Classification:   valueobjects.NextStepTransactionClassification(),
		Status:           valueobjects.AIReviewedTransactionStatus(),
		ReviewResult: valueobjects.NewReactPipelineResult(valueobjects.NewReactReviewResult(
			valueobjects.PipelineResultSourceLLMBatch,
			"react",
			"v1",
			"v1",
			"gpt-4o-mini",
			matchedTransactionID.String(),
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
			"historical non terminal",
		)),
	}}
	chatModel := &reactChatModelStub{messages: []*schema.Message{{Content: `[{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","not_aligned_list_match":false,"not_aligned_list_match_confidence":10,"aligned_list_match":false,"aligned_list_match_confidence":4,"overall_classification":"next_step","reason":"unclear"}]`}}}

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("test prompt"),
		WithReActTransactionClassificationLogger(zap.NewNop()),
	)

	decisions, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "modified cationic starch",
	}})
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("len(decisions) = %d, want 1", len(decisions))
	}

	if chatModel.callCount != 1 {
		t.Fatalf("chatModel.callCount = %d, want %d", chatModel.callCount, 1)
	}

	if decisions[0].Source != valueobjects.PipelineResultSourceLLMBatch {
		t.Fatalf("decisions[0].Source = %q, want %q", decisions[0].Source, valueobjects.PipelineResultSourceLLMBatch)
	}
}

// TestReActTransactionClassificationGatewayReusesExactHistoricalMatchWithoutLLM verifies exact historical reuse remains active without the embedding fallback.
func TestReActTransactionClassificationGatewayReusesExactHistoricalMatchWithoutLLM(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	matchedTransactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e99")
	repo := &reactTransactionRepositoryStub{exactMatch: &ports.HistoricalTransactionClassificationMatch{
		TransactionID:    matchedTransactionID,
		GoodsDescription: "solar panels",
		Classification:   valueobjects.AlignedTransactionClassification(),
		Status:           valueobjects.AIReviewedTransactionStatus(),
		ReviewResult: valueobjects.NewReactPipelineResult(valueobjects.NewReactReviewResult(
			valueobjects.PipelineResultSourceLLMBatch,
			"react",
			"v1",
			"v1",
			"gpt-4o-mini",
			matchedTransactionID.String(),
			nil,
			nil,
			nil,
			nil,
			false,
			10,
			true,
			8,
			2,
			valueobjects.AlignedTransactionClassification(),
			"historical aligned match",
		)),
	}}
	chatModel := &reactChatModelStub{}

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("test prompt"),
		WithReActTransactionClassificationClassifier("react", "v1"),
		WithReActTransactionClassificationPromptVersion("v1"),
		WithReActTransactionClassificationModelName("gpt-4o-mini"),
		WithReActTransactionClassificationLogger(zap.NewNop()),
	)

	decisions, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "solar panels",
	}})
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("len(decisions) = %d, want 1", len(decisions))
	}

	if chatModel.callCount != 0 {
		t.Fatalf("chatModel.callCount = %d, want 0", chatModel.callCount)
	}

	if decisions[0].Source != valueobjects.PipelineResultSourceExactHistoricalMatch {
		t.Fatalf("decisions[0].Source = %q, want %q", decisions[0].Source, valueobjects.PipelineResultSourceExactHistoricalMatch)
	}
}

// TestReActTransactionClassificationGatewayFallsBackToLLMAfterExactMiss verifies the LLM path remains active when no exact history exists.
func TestReActTransactionClassificationGatewayFallsBackToLLMAfterExactMiss(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	repo := &reactTransactionRepositoryStub{}
	chatModel := &reactChatModelStub{messages: []*schema.Message{{Content: `[{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","not_aligned_list_match":false,"not_aligned_list_match_confidence":2,"aligned_list_match":true,"aligned_list_match_confidence":8,"overall_classification":"aligned","reason":"llm path used"}]`}}}

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("test prompt"),
		WithReActTransactionClassificationLogger(zap.NewNop()),
	)

	decisions, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "wind turbine",
	}})
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("len(decisions) = %d, want 1", len(decisions))
	}

	if chatModel.callCount != 1 {
		t.Fatalf("chatModel.callCount = %d, want %d", chatModel.callCount, 1)
	}

	if decisions[0].Source != valueobjects.PipelineResultSourceLLMBatch {
		t.Fatalf("decisions[0].Source = %q, want %q", decisions[0].Source, valueobjects.PipelineResultSourceLLMBatch)
	}
}

// TestReActTransactionClassificationGatewayPreservesClassifierVersionInReactPayload verifies the ReAct act transaction classification gateway preserves classifier version in ReAct payload behavior and the expected outcome asserted below.
func TestReActTransactionClassificationGatewayPreservesClassifierVersionInReactPayload(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	repo := &reactTransactionRepositoryStub{}
	chatModel := &reactChatModelStub{messages: []*schema.Message{{Content: `[{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","not_aligned_list_match":false,"not_aligned_list_match_confidence":10,"aligned_list_match":true,"aligned_list_match_confidence":9,"overall_classification":"aligned","reason":"matches renewable energy manufacturing."}]`}}}

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("test prompt"),
		WithReActTransactionClassificationClassifier("react", "classifier-v7"),
		WithReActTransactionClassificationPromptVersion("prompt-v3"),
		WithReActTransactionClassificationModelName("gpt-4o-mini"),
	)

	decisions, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "solar panels",
	}})
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	react := decisions[0].ReviewResult.React()
	if react == nil {
		t.Fatal("decisions[0].ReviewResult.React() = nil, want react review result")
	}

	if react.ClassifierVersion() != "classifier-v7" {
		t.Fatalf("react.ClassifierVersion() = %q, want %q", react.ClassifierVersion(), "classifier-v7")
	}

	if react.PromptVersion() != "prompt-v3" {
		t.Fatalf("react.PromptVersion() = %q, want %q", react.PromptVersion(), "prompt-v3")
	}
}

// TestReActTransactionClassificationGatewayLogsBatchTokenUsageAtDebug verifies the ReAct act transaction classification gateway logs batch token usage at debug behavior and the expected outcome asserted below.
func TestReActTransactionClassificationGatewayLogsBatchTokenUsageAtDebug(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	repo := &reactTransactionRepositoryStub{}
	observedCore, observedLogs := observer.New(zap.DebugLevel)
	chatModel := &reactChatModelStub{messages: []*schema.Message{{
		Content: `[{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","not_aligned_list_match":false,"not_aligned_list_match_confidence":10,"aligned_list_match":true,"aligned_list_match_confidence":9,"overall_classification":"aligned","reason":"matches renewable energy manufacturing."}]`,
		ResponseMeta: &schema.ResponseMeta{Usage: &schema.TokenUsage{
			PromptTokens:     120,
			CompletionTokens: 40,
			TotalTokens:      160,
			PromptTokenDetails: schema.PromptTokenDetails{
				CachedTokens: 8,
			},
			CompletionTokensDetails: schema.CompletionTokensDetails{
				ReasoningTokens: 11,
			},
		}},
	}}}

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("test prompt"),
		WithReActTransactionClassificationLogger(zap.New(observedCore)),
	)

	_, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "solar panels",
	}})
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	entries := observedLogs.FilterMessage("react classification llm batch token usage").All()
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1", len(entries))
	}

	contextMap := entries[0].ContextMap()
	if contextMap["read_tokens"] != int64(120) {
		t.Fatalf("read_tokens = %v, want %d", contextMap["read_tokens"], 120)
	}
	if contextMap["write_tokens"] != int64(40) {
		t.Fatalf("write_tokens = %v, want %d", contextMap["write_tokens"], 40)
	}
	if contextMap["total_tokens"] != int64(160) {
		t.Fatalf("total_tokens = %v, want %d", contextMap["total_tokens"], 160)
	}
}

// TestReActTransactionClassificationGatewayRetriesTransientLLMFailures verifies the ReAct act transaction classification gateway retries transient LLM failures behavior and the expected outcome asserted below.
func TestReActTransactionClassificationGatewayRetriesTransientLLMFailures(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	repo := &reactTransactionRepositoryStub{}
	observedCore, observedLogs := observer.New(zapcore.InfoLevel)
	chatModel := &reactChatModelStub{
		errs: []error{&url.Error{Op: "Post", URL: "https://example.test", Err: context.DeadlineExceeded}},
		messages: []*schema.Message{
			{Content: `temporary timeout response`},
			{Content: `[{"id":"0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60","not_aligned_list_match":false,"not_aligned_list_match_confidence":10,"aligned_list_match":true,"aligned_list_match_confidence":9,"overall_classification":"aligned","reason":"matches renewable energy manufacturing."}]`},
		},
	}
	sleepCalls := 0

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("test prompt"),
		WithReActTransactionClassificationLogger(zap.New(observedCore)),
		WithReActTransactionClassificationRetry(2, time.Millisecond),
	)
	gateway.sleep = func(context.Context, time.Duration) error {
		sleepCalls++
		return nil
	}

	decisions, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "solar panels",
	}})
	if err != nil {
		t.Fatalf("Classify() error = %v", err)
	}

	if len(decisions) != 1 {
		t.Fatalf("len(decisions) = %d, want 1", len(decisions))
	}

	if sleepCalls != 1 {
		t.Fatalf("sleepCalls = %d, want 1", sleepCalls)
	}

	entries := observedLogs.All()
	if len(entries) < 2 {
		t.Fatalf("len(entries) = %d, want at least 2", len(entries))
	}

	if entries[0].Message != "react classification llm batch retrying after transient failure" {
		t.Fatalf("entries[0].Message = %q", entries[0].Message)
	}

	if entries[1].Message != "react classification llm batch retry succeeded" {
		t.Fatalf("entries[1].Message = %q", entries[1].Message)
	}
}

// TestReActTransactionClassificationGatewayDoesNotRetryNonTransientLLMFailures verifies the ReAct act transaction classification gateway does not retry non transient LLM failures behavior and the expected outcome asserted below.
func TestReActTransactionClassificationGatewayDoesNotRetryNonTransientLLMFailures(t *testing.T) {
	t.Parallel()

	transactionID, _ := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	repo := &reactTransactionRepositoryStub{}
	chatModel := &reactChatModelStub{
		errs:     []error{errors.New("invalid_request_error: prompt too long")},
		messages: []*schema.Message{{Content: `[]`}},
	}
	sleepCalls := 0

	gateway := NewReActTransactionClassificationGateway(
		repo,
		chatModel,
		WithReActTransactionClassificationSystemPrompt("test prompt"),
		WithReActTransactionClassificationRetry(2, time.Millisecond),
	)
	gateway.sleep = func(context.Context, time.Duration) error {
		sleepCalls++
		return nil
	}

	_, err := gateway.Classify(context.Background(), []ports.TransactionClassificationCandidate{{
		TransactionID:    transactionID,
		GoodsDescription: "solar panels",
	}})
	if err == nil {
		t.Fatal("Classify() error = nil, want error")
	}

	if sleepCalls != 0 {
		t.Fatalf("sleepCalls = %d, want 0", sleepCalls)
	}

	if chatModel.callCount != 1 {
		t.Fatalf("chatModel.callCount = %d, want 1", chatModel.callCount)
	}
}
