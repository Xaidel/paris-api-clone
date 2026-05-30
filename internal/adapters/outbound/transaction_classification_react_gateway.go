package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	openaimodel "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	"go.uber.org/zap"
)

const (
	defaultReactClassifierFamily  = "react"
	defaultReactClassifierVersion = "v1"
	defaultReactPromptVersion     = "v1"
	defaultReactBatchSize         = 10
	defaultReactMaxRetries        = 2
	defaultReactRetryBackoff      = 2 * time.Second
)

type reactChatModel interface {
	Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error)
}

type reactSystemPromptBuilder func(context.Context) (string, error)

// ReActTransactionClassificationGatewayOption configures ReActTransactionClassificationGateway.
type ReActTransactionClassificationGatewayOption func(*ReActTransactionClassificationGateway)

// WithReActTransactionClassificationSystemPrompt configures the classifier system prompt template.
func WithReActTransactionClassificationSystemPrompt(systemPrompt string) ReActTransactionClassificationGatewayOption {
	return func(gateway *ReActTransactionClassificationGateway) {
		gateway.systemPrompt = systemPrompt
	}
}

// WithReActTransactionClassificationSystemPromptBuilder configures dynamic system prompt rendering.
func WithReActTransactionClassificationSystemPromptBuilder(builder func(context.Context) (string, error)) ReActTransactionClassificationGatewayOption {
	return func(gateway *ReActTransactionClassificationGateway) {
		gateway.systemPromptBuilder = builder
	}
}

// WithReActTransactionClassificationBatchSize configures the gateway batch size.
func WithReActTransactionClassificationBatchSize(batchSize int) ReActTransactionClassificationGatewayOption {
	return func(gateway *ReActTransactionClassificationGateway) {
		gateway.batchSize = batchSize
	}
}

// WithReActTransactionClassificationClassifier configures lineage metadata.
func WithReActTransactionClassificationClassifier(family string, version string) ReActTransactionClassificationGatewayOption {
	return func(gateway *ReActTransactionClassificationGateway) {
		gateway.classifierFamily = strings.TrimSpace(family)
		gateway.classifierVersion = strings.TrimSpace(version)
	}
}

// WithReActTransactionClassificationPromptVersion configures prompt version metadata.
func WithReActTransactionClassificationPromptVersion(promptVersion string) ReActTransactionClassificationGatewayOption {
	return func(gateway *ReActTransactionClassificationGateway) {
		gateway.promptVersion = strings.TrimSpace(promptVersion)
	}
}

// WithReActTransactionClassificationModelName overrides the recorded model name.
func WithReActTransactionClassificationModelName(modelName string) ReActTransactionClassificationGatewayOption {
	return func(gateway *ReActTransactionClassificationGateway) {
		gateway.modelName = strings.TrimSpace(modelName)
	}
}

// WithReActTransactionClassificationMetrics configures the metrics sink.
func WithReActTransactionClassificationMetrics(metrics ports.ClassificationMetrics) ReActTransactionClassificationGatewayOption {
	return func(gateway *ReActTransactionClassificationGateway) {
		gateway.metrics = metrics
	}
}

// WithReActTransactionClassificationLogger configures the gateway logger.
func WithReActTransactionClassificationLogger(logger *zap.Logger) ReActTransactionClassificationGatewayOption {
	return func(gateway *ReActTransactionClassificationGateway) {
		gateway.logger = logger
	}
}

// WithReActTransactionClassificationRetry configures transient LLM retry behavior.
func WithReActTransactionClassificationRetry(maxRetries int, retryBackoff time.Duration) ReActTransactionClassificationGatewayOption {
	return func(gateway *ReActTransactionClassificationGateway) {
		gateway.maxRetries = maxRetries
		gateway.retryBackoff = retryBackoff
	}
}

// ReActTransactionClassificationGateway classifies transactions using historical reuse and batched LLM calls.
type ReActTransactionClassificationGateway struct {
	transactionRepo     ports.TransactionRepository
	chatModel           reactChatModel
	systemPrompt        string
	systemPromptBuilder reactSystemPromptBuilder
	classifierFamily    string
	classifierVersion   string
	promptVersion       string
	modelName           string
	batchSize           int
	metrics             ports.ClassificationMetrics
	logger              *zap.Logger
	maxRetries          int
	retryBackoff        time.Duration
	sleep               func(context.Context, time.Duration) error
}

// NewReActTransactionClassificationGateway builds a ReActTransactionClassificationGateway.
func NewReActTransactionClassificationGateway(
	transactionRepo ports.TransactionRepository,
	chatModel reactChatModel,
	opts ...ReActTransactionClassificationGatewayOption,
) *ReActTransactionClassificationGateway {
	gateway := &ReActTransactionClassificationGateway{
		transactionRepo:   transactionRepo,
		chatModel:         chatModel,
		classifierFamily:  defaultReactClassifierFamily,
		classifierVersion: defaultReactClassifierVersion,
		promptVersion:     defaultReactPromptVersion,
		batchSize:         defaultReactBatchSize,
		maxRetries:        defaultReactMaxRetries,
		retryBackoff:      defaultReactRetryBackoff,
		sleep:             sleepWithContext,
	}

	for _, opt := range opts {
		opt(gateway)
	}

	return gateway
}

// Classify classifies one or more transactions behind the outbound gateway boundary.
func (g *ReActTransactionClassificationGateway) Classify(ctx context.Context, candidates []ports.TransactionClassificationCandidate) ([]ports.TransactionClassificationDecision, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	groups, order := dedupeClassificationCandidates(candidates)
	decisionsByDescription := make(map[string]ports.TransactionClassificationDecision, len(groups))
	unresolved := make([]dedupedClassificationCandidate, 0, len(groups))

	for _, description := range order {
		group := groups[description]
		decision, resolved, err := g.resolveHistorically(ctx, group)
		if err != nil {
			return nil, err
		}

		if resolved {
			decisionsByDescription[description] = decision
			continue
		}

		unresolved = append(unresolved, group)
	}

	for start := 0; start < len(unresolved); start += g.classificationBatchSize() {
		end := start + g.classificationBatchSize()
		if end > len(unresolved) {
			end = len(unresolved)
		}

		batchDecisions, err := g.classifyBatch(ctx, unresolved[start:end])
		if err != nil {
			return nil, err
		}

		for description, decision := range batchDecisions {
			decisionsByDescription[description] = decision
		}
	}

	results := make([]ports.TransactionClassificationDecision, 0, len(candidates))
	for _, candidate := range candidates {
		decision, ok := decisionsByDescription[strings.TrimSpace(candidate.GoodsDescription)]
		if !ok {
			return nil, fmt.Errorf("missing classification decision for goods description %q", candidate.GoodsDescription)
		}

		result := cloneClassificationDecision(decision)
		result.TransactionID = candidate.TransactionID

		reviewResult := result.ReviewResult
		if react := reviewResult.React(); react != nil {
			updatedReact := valueobjects.NewReactReviewResult(
				react.Source(),
				react.ClassifierFamily(),
				react.ClassifierVersion(),
				react.PromptVersion(),
				react.Model(),
				candidate.TransactionID.String(),
				transactionIDPointerString(result.MatchedTransactionID),
				react.MatchedGoodsDescription(),
				react.BatchID(),
				react.BatchSize(),
				react.NotAlignedListMatch(),
				react.NotAlignedListMatchConfidence(),
				react.AlignedListMatch(),
				react.AlignedListMatchConfidence(),
				react.ExitStep(),
				react.OverallClassification(),
				react.Reason(),
			)
			result.ReviewResult = valueobjects.NewReactPipelineResult(updatedReact)
		}

		results = append(results, result)
	}

	return results, nil
}

type dedupedClassificationCandidate struct {
	description string
	items       []ports.TransactionClassificationCandidate
}

type llmBatchInputItem struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Sector      string `json:"sector,omitempty"`
}

type llmBatchOutputItem struct {
	ID                            string `json:"id"`
	NotAlignedListMatch           bool   `json:"not_aligned_list_match"`
	NotAlignedListMatchConfidence int    `json:"not_aligned_list_match_confidence"`
	AlignedListMatch              bool   `json:"aligned_list_match"`
	AlignedListMatchConfidence    int    `json:"aligned_list_match_confidence"`
	OverallClassification         string `json:"overall_classification"`
	Reason                        string `json:"reason"`
}

func dedupeClassificationCandidates(candidates []ports.TransactionClassificationCandidate) (map[string]dedupedClassificationCandidate, []string) {
	groups := make(map[string]dedupedClassificationCandidate, len(candidates))
	order := make([]string, 0, len(candidates))

	for _, candidate := range candidates {
		description := strings.TrimSpace(candidate.GoodsDescription)
		group, exists := groups[description]
		if !exists {
			group = dedupedClassificationCandidate{description: description}
			order = append(order, description)
		}

		group.items = append(group.items, candidate)
		groups[description] = group
	}

	return groups, order
}

func (g *ReActTransactionClassificationGateway) resolveHistorically(ctx context.Context, candidate dedupedClassificationCandidate) (ports.TransactionClassificationDecision, bool, error) {
	exactMatch, err := g.transactionRepo.FindHistoricalClassificationByExactGoodsDescription(ctx, ports.HistoricalTransactionClassificationQuery{
		ClassifierFamily:  g.classifierFamily,
		ClassifierVersion: g.promptVersion,
		ResultVersion:     valueobjects.PipelineResultVersionReactV1,
		GoodsDescription:  candidate.description,
	})
	if err != nil {
		return ports.TransactionClassificationDecision{}, false, fmt.Errorf("looking up exact historical classification for %q: %w", candidate.description, err)
	}

	if exactMatch != nil {
		if !g.isEligibleHistoricalReviewResult(exactMatch.ReviewResult.React()) {
			return ports.TransactionClassificationDecision{}, false, nil
		}

		g.recordExactHistoricalMatch()
		decision := newHistoricalClassificationDecision(*exactMatch, valueobjects.PipelineResultSourceExactHistoricalMatch)
		g.logHistoricalMatch(decision, candidate)
		return decision, true, nil
	}

	return ports.TransactionClassificationDecision{}, false, nil
}

func (g *ReActTransactionClassificationGateway) classifyBatch(ctx context.Context, candidates []dedupedClassificationCandidate) (map[string]ports.TransactionClassificationDecision, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	if g.chatModel == nil {
		return nil, fmt.Errorf("react chat model is required for unresolved classifications")
	}

	batchID, err := g.nextBatchID()
	if err != nil {
		return nil, err
	}
	g.recordLLMRoute(len(candidates))

	items := make([]llmBatchInputItem, 0, len(candidates))
	for index, candidate := range candidates {
		item := llmBatchInputItem{
			ID:          candidate.items[0].TransactionID.String(),
			Description: candidate.description,
		}
		if sector := candidate.items[0].Sector; sector != nil && strings.TrimSpace(*sector) != "" {
			item.Sector = strings.TrimSpace(*sector)
		}
		if item.ID == "" {
			item.ID = fmt.Sprintf("batch-item-%d", index)
		}
		items = append(items, item)
	}

	batchPayload, err := json.Marshal(items)
	if err != nil {
		g.logBatchFailure(batchID, items, fmt.Errorf("marshalling react classification batch payload: %w", err))
		return nil, fmt.Errorf("marshalling react classification batch payload: %w", err)
	}

	systemPrompt, err := g.classificationSystemPrompt(ctx)
	if err != nil {
		g.logBatchFailure(batchID, items, fmt.Errorf("building react classification system prompt: %w", err))
		return nil, fmt.Errorf("building react classification system prompt: %w", err)
	}

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(string(batchPayload)),
	}
	message, err := g.generateWithRetry(ctx, messages, batchID, items)
	if err != nil {
		g.logBatchFailure(batchID, items, fmt.Errorf("executing react classification batch %s: %w", batchID, err))
		return nil, fmt.Errorf("executing react classification batch %s: %w", batchID, err)
	}

	parsed, err := g.parseAndValidateLLMOutput(message, items)
	if err != nil {
		g.recordInvalidLLMOutput()
		g.logInvalidBatchOutput(batchID, items, err)
		g.logRepairAttempt(batchID, items)

		repairMessage, repairErr := g.requestRepair(ctx, batchID, items, batchPayload, message)
		if repairErr != nil {
			g.logBatchFailure(batchID, items, fmt.Errorf("repairing invalid react classification batch %s output: %w", batchID, repairErr))
			return nil, fmt.Errorf("repairing invalid react classification batch %s output: %w", batchID, repairErr)
		}

		parsed, err = g.parseAndValidateLLMOutput(repairMessage, items)
		if err != nil {
			g.recordInvalidLLMOutput()
			g.logBatchFailure(batchID, items, fmt.Errorf("validating repaired react classification batch %s output: %w", batchID, err))
			return nil, fmt.Errorf("validating repaired react classification batch %s output: %w", batchID, err)
		}

		g.logRepairSuccess(batchID, items)
	}

	decisions := make(map[string]ports.TransactionClassificationDecision, len(candidates))
	for index, output := range parsed {
		classification, err := mapReactClassification(output.OverallClassification)
		if err != nil {
			g.logBatchFailure(batchID, items, err)
			return nil, err
		}

		reviewResult := valueobjects.NewReactReviewResult(
			valueobjects.PipelineResultSourceLLMBatch,
			g.classifierFamily,
			g.classifierVersion,
			g.promptVersion,
			g.modelName,
			items[index].ID,
			nil,
			nil,
			&batchID,
			intPointer(len(items)),
			output.NotAlignedListMatch,
			output.NotAlignedListMatchConfidence,
			output.AlignedListMatch,
			output.AlignedListMatchConfidence,
			reactClassificationExitStep(output, classification),
			classification,
			output.Reason,
		)

		decision := ports.TransactionClassificationDecision{
			TransactionID:  candidates[index].items[0].TransactionID,
			Classification: classification,
			Status:         valueobjects.AIReviewedTransactionStatus(),
			ReviewResult:   valueobjects.NewReactPipelineResult(reviewResult),
			Source:         valueobjects.PipelineResultSourceLLMBatch,
		}
		decisions[candidates[index].description] = decision
		g.recordClassification(decision)
	}

	g.logBatchCompletion(batchID, decisions)
	return decisions, nil
}

func (g *ReActTransactionClassificationGateway) requestRepair(ctx context.Context, batchID string, items []llmBatchInputItem, batchPayload []byte, invalidMessage *schema.Message) (*schema.Message, error) {
	if invalidMessage == nil {
		return nil, fmt.Errorf("invalid llm response is required for repair")
	}

	systemPrompt, err := g.classificationSystemPrompt(ctx)
	if err != nil {
		return nil, fmt.Errorf("building react classification repair system prompt: %w", err)
	}

	repairPrompt := fmt.Sprintf("The previous response was invalid. Return only a JSON array matching the required schema and preserving the same order and IDs. Original input: %s\nInvalid response: %s", string(batchPayload), invalidMessage.Content)
	return g.generateWithRetry(ctx, []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(repairPrompt),
	}, batchID, items)
}

func (g *ReActTransactionClassificationGateway) generateWithRetry(ctx context.Context, messages []*schema.Message, batchID string, items []llmBatchInputItem) (*schema.Message, error) {
	attempts := g.totalAttempts()
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		startedAt := time.Now()
		message, err := g.chatModel.Generate(ctx, messages)
		g.observeLLMLatency(time.Since(startedAt).Seconds())
		if err == nil {
			g.logBatchTokenUsage(batchID, items, message)
			if attempt > 1 {
				g.logRetrySuccess(batchID, items, attempt)
			}
			return message, nil
		}

		lastErr = err
		if !isRetryableReActChatError(err) || attempt == attempts {
			return nil, err
		}

		backoff := g.retryDelay(attempt)
		g.logRetryAttempt(batchID, items, attempt, backoff, err)
		if sleepErr := g.sleep(ctx, backoff); sleepErr != nil {
			return nil, sleepErr
		}
	}

	return nil, lastErr
}

func (g *ReActTransactionClassificationGateway) totalAttempts() int {
	if g == nil || g.maxRetries < 0 {
		return 1
	}

	return g.maxRetries + 1
}

func (g *ReActTransactionClassificationGateway) retryDelay(attempt int) time.Duration {
	if g == nil || g.retryBackoff <= 0 {
		return 0
	}

	if attempt <= 1 {
		return g.retryBackoff
	}

	return g.retryBackoff * time.Duration(1<<(attempt-1))
}

func isRetryableReActChatError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		if urlErr.Timeout() {
			return true
		}
		if isRetryableReActChatError(urlErr.Err) {
			return true
		}
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "timeout") ||
		strings.Contains(message, "deadline exceeded") ||
		strings.Contains(message, "too many requests") ||
		strings.Contains(message, "status code: 429") ||
		strings.Contains(message, "status code: 500") ||
		strings.Contains(message, "status code: 502") ||
		strings.Contains(message, "status code: 503") ||
		strings.Contains(message, "status code: 504") ||
		strings.Contains(message, http.StatusText(http.StatusTooManyRequests)) ||
		strings.Contains(message, http.StatusText(http.StatusBadGateway)) ||
		strings.Contains(message, http.StatusText(http.StatusServiceUnavailable)) ||
		strings.Contains(message, http.StatusText(http.StatusGatewayTimeout))
}

func sleepWithContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (g *ReActTransactionClassificationGateway) parseAndValidateLLMOutput(message *schema.Message, items []llmBatchInputItem) ([]llmBatchOutputItem, error) {
	if message == nil {
		return nil, fmt.Errorf("llm response is required")
	}

	var output []llmBatchOutputItem
	if err := json.Unmarshal([]byte(strings.TrimSpace(message.Content)), &output); err != nil {
		return nil, fmt.Errorf("unmarshalling llm output: %w", err)
	}

	if len(output) != len(items) {
		return nil, fmt.Errorf("llm output length = %d, want %d", len(output), len(items))
	}

	byID := make(map[string]llmBatchOutputItem, len(output))
	for index, item := range output {
		if strings.TrimSpace(item.ID) == "" {
			return nil, fmt.Errorf("llm output id at index %d is required", index)
		}

		if _, exists := byID[item.ID]; exists {
			return nil, fmt.Errorf("llm output id %q is duplicated", item.ID)
		}
		byID[item.ID] = item

		if item.NotAlignedListMatchConfidence < 0 || item.NotAlignedListMatchConfidence > 10 {
			return nil, fmt.Errorf("not_aligned_list_match_confidence for %s = %d, want 0..10", item.ID, item.NotAlignedListMatchConfidence)
		}

		if item.AlignedListMatchConfidence < 0 || item.AlignedListMatchConfidence > 10 {
			return nil, fmt.Errorf("aligned_list_match_confidence for %s = %d, want 0..10", item.ID, item.AlignedListMatchConfidence)
		}

		if _, err := mapReactClassification(item.OverallClassification); err != nil {
			return nil, fmt.Errorf("invalid overall_classification for %s: %w", item.ID, err)
		}
	}

	reordered := make([]llmBatchOutputItem, 0, len(items))
	for index, input := range items {
		item, ok := byID[input.ID]
		if !ok {
			return nil, fmt.Errorf("llm output id %q missing at index %d", input.ID, index)
		}

		reordered = append(reordered, item)
	}

	return reordered, nil
}

func mapReactClassification(value string) (valueobjects.TransactionClassification, error) {
	switch strings.TrimSpace(value) {
	case "aligned":
		return valueobjects.AlignedTransactionClassification(), nil
	case "not_aligned":
		return valueobjects.NotAlignedTransactionClassification(), nil
	case "next_step", "next-step":
		return valueobjects.NextStepTransactionClassification(), nil
	default:
		return valueobjects.TransactionClassification{}, fmt.Errorf("unsupported react classification %q", value)
	}
}

func newHistoricalClassificationDecision(match ports.HistoricalTransactionClassificationMatch, source string) ports.TransactionClassificationDecision {
	react := match.ReviewResult.React()
	if react == nil {
		reviewResult := valueobjects.NewReactReviewResult(
			source,
			defaultReactClassifierFamily,
			defaultReactClassifierVersion,
			defaultReactPromptVersion,
			"",
			match.TransactionID.String(),
			stringPointer(match.TransactionID.String()),
			stringPointer(match.GoodsDescription),
			nil,
			nil,
			false,
			0,
			false,
			0,
			historicalClassificationExitStep(match.Classification),
			match.Classification,
			"",
		)
		return ports.TransactionClassificationDecision{
			TransactionID:        match.TransactionID,
			Classification:       match.Classification,
			Status:               valueobjects.FromPreviousTransactionsTransactionStatus(),
			ReviewResult:         valueobjects.NewReactPipelineResult(reviewResult),
			MatchedTransactionID: &match.TransactionID,
			Source:               source,
		}
	}

	reviewResult := valueobjects.NewReactReviewResult(
		source,
		react.ClassifierFamily(),
		react.ClassifierVersion(),
		react.PromptVersion(),
		react.Model(),
		match.TransactionID.String(),
		stringPointer(match.TransactionID.String()),
		stringPointer(match.GoodsDescription),
		react.BatchID(),
		react.BatchSize(),
		react.NotAlignedListMatch(),
		react.NotAlignedListMatchConfidence(),
		react.AlignedListMatch(),
		react.AlignedListMatchConfidence(),
		historicalClassificationExitStep(match.Classification),
		match.Classification,
		react.Reason(),
	)

	return ports.TransactionClassificationDecision{
		TransactionID:        match.TransactionID,
		Classification:       match.Classification,
		Status:               valueobjects.FromPreviousTransactionsTransactionStatus(),
		ReviewResult:         valueobjects.NewReactPipelineResult(reviewResult),
		MatchedTransactionID: &match.TransactionID,
		Source:               source,
	}
}

func cloneClassificationDecision(decision ports.TransactionClassificationDecision) ports.TransactionClassificationDecision {
	clone := decision
	if decision.MatchedTransactionID != nil {
		matched := *decision.MatchedTransactionID
		clone.MatchedTransactionID = &matched
	}
	return clone
}

func stringPointer(value string) *string {
	copyValue := value
	return &copyValue
}

func intPointer(value int) *int {
	copyValue := value
	return &copyValue
}

func transactionIDPointerString(value *valueobjects.TransactionID) *string {
	if value == nil {
		return nil
	}

	return stringPointer(value.String())
}

func (g *ReActTransactionClassificationGateway) nextBatchID() (string, error) {
	batchID, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate uuidv7 batch id: %w", err)
	}

	return batchID.String(), nil
}

func (g *ReActTransactionClassificationGateway) classificationBatchSize() int {
	if g == nil || g.batchSize <= 0 {
		return defaultReactBatchSize
	}

	return g.batchSize
}

func (g *ReActTransactionClassificationGateway) classificationSystemPrompt(ctx context.Context) (string, error) {
	if g == nil {
		return "", nil
	}

	var prompt string
	var err error
	if g.systemPromptBuilder != nil {
		prompt, err = g.systemPromptBuilder(ctx)
	} else {
		prompt = g.systemPrompt
	}
	if err != nil {
		return "", err
	}

	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return "", fmt.Errorf("react classification system prompt is required")
	}

	return prompt, nil
}

func (g *ReActTransactionClassificationGateway) recordExactHistoricalMatch() {
	if g == nil || g.metrics == nil {
		return
	}

	g.metrics.IncReactHistoricalExactMatchTotal()
}

func (g *ReActTransactionClassificationGateway) recordLLMRoute(batchSize int) {
	if g == nil || g.metrics == nil {
		return
	}

	g.metrics.IncReactLLMRouteTotal()
	g.metrics.ObserveReactLLMBatchSize(batchSize)
}

func (g *ReActTransactionClassificationGateway) observeLLMLatency(seconds float64) {
	if g == nil || g.metrics == nil {
		return
	}

	g.metrics.ObserveReactLLMLatencySeconds(seconds)
}

func (g *ReActTransactionClassificationGateway) recordInvalidLLMOutput() {
	if g == nil || g.metrics == nil {
		return
	}

	g.metrics.IncReactInvalidLLMOutputTotal()
}

func (g *ReActTransactionClassificationGateway) recordClassification(decision ports.TransactionClassificationDecision) {
	if g == nil || g.metrics == nil {
		return
	}

	g.metrics.IncReactClassificationTotal(decision.Classification.String())
	g.metrics.IncReactStatusTotal(decision.Status.String())
}

func (g *ReActTransactionClassificationGateway) logHistoricalMatch(decision ports.TransactionClassificationDecision, candidate dedupedClassificationCandidate) {
	if g == nil || g.logger == nil {
		return
	}

	fields := []zap.Field{
		zap.String("source", decision.Source),
		zap.String("goods_description", candidate.description),
		zap.String("transaction_id", candidate.items[0].TransactionID.String()),
		zap.String("prompt_version", g.promptVersion),
		zap.String("model", g.modelName),
	}
	if decision.MatchedTransactionID != nil {
		fields = append(fields, zap.String("matched_transaction_id", decision.MatchedTransactionID.String()))
	}

	g.logger.Info("react classification historical reuse applied", fields...)
}

func (g *ReActTransactionClassificationGateway) logInvalidBatchOutput(batchID string, items []llmBatchInputItem, err error) {
	if g == nil || g.logger == nil {
		return
	}

	g.logger.Warn("react classification llm batch returned invalid output",
		zap.String("batch_id", batchID),
		zap.Strings("transaction_ids", llmBatchInputIDs(items)),
		zap.String("prompt_version", g.promptVersion),
		zap.String("model", g.modelName),
		zap.Error(err),
	)
}

func (g *ReActTransactionClassificationGateway) logRepairAttempt(batchID string, items []llmBatchInputItem) {
	if g == nil || g.logger == nil {
		return
	}

	g.logger.Info("react classification llm batch requesting repair",
		zap.String("batch_id", batchID),
		zap.Strings("transaction_ids", llmBatchInputIDs(items)),
		zap.String("prompt_version", g.promptVersion),
		zap.String("model", g.modelName),
	)
}

func (g *ReActTransactionClassificationGateway) logRepairSuccess(batchID string, items []llmBatchInputItem) {
	if g == nil || g.logger == nil {
		return
	}

	g.logger.Info("react classification llm batch repair succeeded",
		zap.String("batch_id", batchID),
		zap.Strings("transaction_ids", llmBatchInputIDs(items)),
		zap.String("prompt_version", g.promptVersion),
		zap.String("model", g.modelName),
	)
}

func (g *ReActTransactionClassificationGateway) logBatchFailure(batchID string, items []llmBatchInputItem, err error) {
	if g == nil || g.logger == nil {
		return
	}

	g.logger.Warn("react classification llm batch failed",
		zap.String("batch_id", batchID),
		zap.Strings("transaction_ids", llmBatchInputIDs(items)),
		zap.String("prompt_version", g.promptVersion),
		zap.String("model", g.modelName),
		zap.Error(err),
	)
}

func (g *ReActTransactionClassificationGateway) logRetryAttempt(batchID string, items []llmBatchInputItem, attempt int, backoff time.Duration, err error) {
	if g == nil || g.logger == nil {
		return
	}

	g.logger.Warn("react classification llm batch retrying after transient failure",
		zap.String("batch_id", batchID),
		zap.Strings("transaction_ids", llmBatchInputIDs(items)),
		zap.Int("attempt", attempt),
		zap.Duration("retry_backoff", backoff),
		zap.String("prompt_version", g.promptVersion),
		zap.String("model", g.modelName),
		zap.Error(err),
	)
}

func (g *ReActTransactionClassificationGateway) logRetrySuccess(batchID string, items []llmBatchInputItem, attempt int) {
	if g == nil || g.logger == nil {
		return
	}

	g.logger.Info("react classification llm batch retry succeeded",
		zap.String("batch_id", batchID),
		zap.Strings("transaction_ids", llmBatchInputIDs(items)),
		zap.Int("attempt", attempt),
		zap.String("prompt_version", g.promptVersion),
		zap.String("model", g.modelName),
	)
}

func (g *ReActTransactionClassificationGateway) logBatchTokenUsage(batchID string, items []llmBatchInputItem, message *schema.Message) {
	if g == nil || g.logger == nil || !g.logger.Core().Enabled(zap.DebugLevel) {
		return
	}

	if message == nil || message.ResponseMeta == nil || message.ResponseMeta.Usage == nil {
		return
	}

	usage := message.ResponseMeta.Usage
	g.logger.Debug("react classification llm batch token usage",
		zap.String("batch_id", batchID),
		zap.Strings("transaction_ids", llmBatchInputIDs(items)),
		zap.Int("batch_size", len(items)),
		zap.String("prompt_version", g.promptVersion),
		zap.String("model", g.modelName),
		zap.Int("read_tokens", usage.PromptTokens),
		zap.Int("write_tokens", usage.CompletionTokens),
		zap.Int("total_tokens", usage.TotalTokens),
		zap.Int("cached_read_tokens", usage.PromptTokenDetails.CachedTokens),
		zap.Int("reasoning_write_tokens", usage.CompletionTokensDetails.ReasoningTokens),
	)
}

func (g *ReActTransactionClassificationGateway) logBatchCompletion(batchID string, decisions map[string]ports.TransactionClassificationDecision) {
	if g == nil || g.logger == nil {
		return
	}

	transactionIDs := make([]string, 0, len(decisions))
	for _, decision := range decisions {
		transactionIDs = append(transactionIDs, decision.TransactionID.String())
	}

	g.logger.Info("react classification llm batch completed",
		zap.String("batch_id", batchID),
		zap.Strings("transaction_ids", transactionIDs),
		zap.String("prompt_version", g.promptVersion),
		zap.String("model", g.modelName),
	)
}

func (g *ReActTransactionClassificationGateway) isEligibleHistoricalReviewResult(react *valueobjects.ReactReviewResult) bool {
	if react == nil {
		return false
	}

	return react.Version() == valueobjects.PipelineResultVersionReactV1 &&
		react.ClassifierFamily() == g.classifierFamily &&
		react.ClassifierVersion() == g.promptVersion &&
		(react.OverallClassification().Equal(valueobjects.AlignedTransactionClassification()) ||
			react.OverallClassification().Equal(valueobjects.NotAlignedTransactionClassification()))
}

func llmBatchInputIDs(items []llmBatchInputItem) []string {
	transactionIDs := make([]string, 0, len(items))
	for _, item := range items {
		transactionIDs = append(transactionIDs, item.ID)
	}

	return transactionIDs
}

func reactClassificationExitStep(output llmBatchOutputItem, classification valueobjects.TransactionClassification) int {
	if output.NotAlignedListMatch {
		return 1
	}

	if output.AlignedListMatch && classification.Equal(valueobjects.AlignedTransactionClassification()) {
		return 2
	}

	return 0
}

func historicalClassificationExitStep(classification valueobjects.TransactionClassification) int {
	if classification.Equal(valueobjects.NotAlignedTransactionClassification()) {
		return 1
	}

	if classification.Equal(valueobjects.AlignedTransactionClassification()) {
		return 2
	}

	return 0
}

func mustTransactionClassification(raw string) valueobjects.TransactionClassification {
	classification, err := valueobjects.TransactionClassificationFromString(raw)
	if err != nil {
		return valueobjects.UnclassifiedTransactionClassification()
	}

	return classification
}

func mustTransactionStatus(raw string) valueobjects.TransactionStatus {
	status, err := valueobjects.TransactionStatusFromString(raw)
	if err != nil {
		return valueobjects.FailedTransactionStatus()
	}

	return status
}

// NewOpenAIReActChatModel builds the OpenAI-backed Eino chat model used by the gateway.
func NewOpenAIReActChatModel(ctx context.Context, config *openaimodel.ChatModelConfig) (*openaimodel.ChatModel, error) {
	return openaimodel.NewChatModel(ctx, config)
}

var _ ports.TransactionClassificationGateway = (*ReActTransactionClassificationGateway)(nil)
