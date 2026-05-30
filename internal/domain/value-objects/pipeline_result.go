package valueobjects

// PipelineResultVersionLegacy identifies the legacy pipeline result payload.
const PipelineResultVersionLegacy = "legacy_pipeline"

// PipelineResultVersionReactV1 identifies the ReAct review-result payload.
const PipelineResultVersionReactV1 = "react_v1"

// PipelineResultSourceLLMBatch identifies direct LLM classification output.
const PipelineResultSourceLLMBatch = "llm_batch"

// PipelineResultSourceExactHistoricalMatch identifies exact historical reuse.
const PipelineResultSourceExactHistoricalMatch = "exact_historical_match"

// PipelineResultSourceLegacyStep3Fallback identifies legacy step 3 fallback output.
const PipelineResultSourceLegacyStep3Fallback = "legacy_step3_fallback"

// PipelineResult describes the persisted transaction review output envelope.
type PipelineResult struct {
	version string
	legacy  *LegacyPipelineResult
	react   *ReactReviewResult
}

// NewPipelineResult builds a legacy pipeline review-result envelope.
func NewPipelineResult(
	transactionID string,
	step1Result StepResult,
	step2Result *StepResult,
	step3Result *StepResult,
	exitStep int,
	finalClassification TransactionClassification,
) PipelineResult {
	legacy := NewLegacyPipelineResult(
		transactionID,
		step1Result,
		step2Result,
		step3Result,
		exitStep,
		finalClassification,
	)

	return PipelineResult{
		version: PipelineResultVersionLegacy,
		legacy:  &legacy,
	}
}

// NewReactPipelineResult builds a ReAct review-result envelope.
func NewReactPipelineResult(result ReactReviewResult) PipelineResult {
	resultCopy := result
	return PipelineResult{
		version: PipelineResultVersionReactV1,
		react:   &resultCopy,
	}
}

// Version returns the persisted envelope version.
func (r PipelineResult) Version() string {
	if r.version == "" {
		if r.react != nil {
			return PipelineResultVersionReactV1
		}

		return PipelineResultVersionLegacy
	}

	return r.version
}

// Legacy returns the legacy pipeline payload when present.
func (r PipelineResult) Legacy() *LegacyPipelineResult {
	if r.legacy == nil {
		return nil
	}

	resultCopy := *r.legacy
	return &resultCopy
}

// React returns the ReAct review payload when present.
func (r PipelineResult) React() *ReactReviewResult {
	if r.react == nil {
		return nil
	}

	resultCopy := *r.react
	return &resultCopy
}

// IsReact reports whether this envelope stores a ReAct result.
func (r PipelineResult) IsReact() bool {
	return r.React() != nil
}

// IsLegacy reports whether this envelope stores a legacy pipeline result.
func (r PipelineResult) IsLegacy() bool {
	return r.Legacy() != nil
}

// TransactionID returns the transaction identifier.
func (r PipelineResult) TransactionID() string {
	if react := r.React(); react != nil {
		return react.TransactionID()
	}

	if legacy := r.Legacy(); legacy != nil {
		return legacy.TransactionID()
	}

	return ""
}

// Step1Result returns the legacy step 1 result.
func (r PipelineResult) Step1Result() StepResult {
	if legacy := r.Legacy(); legacy != nil {
		return legacy.Step1Result()
	}

	return StepResult{}
}

// Step2Result returns the legacy step 2 result when present.
func (r PipelineResult) Step2Result() *StepResult {
	if legacy := r.Legacy(); legacy != nil {
		return legacy.Step2Result()
	}

	return nil
}

// Step3Result returns the legacy step 3 result when present.
func (r PipelineResult) Step3Result() *StepResult {
	if legacy := r.Legacy(); legacy != nil {
		return legacy.Step3Result()
	}

	return nil
}

// ExitStep returns the legacy exit step.
func (r PipelineResult) ExitStep() int {
	if react := r.React(); react != nil {
		return react.ExitStep()
	}

	if legacy := r.Legacy(); legacy != nil {
		return legacy.ExitStep()
	}

	return 0
}

// FinalClassification returns the final transaction classification.
func (r PipelineResult) FinalClassification() TransactionClassification {
	if react := r.React(); react != nil {
		return react.OverallClassification()
	}

	if legacy := r.Legacy(); legacy != nil {
		return legacy.FinalClassification()
	}

	return TransactionClassification{}
}

// Reason returns the persisted classifier rationale when present.
func (r PipelineResult) Reason() string {
	if react := r.React(); react != nil {
		return react.Reason()
	}

	return ""
}

// LegacyPipelineResult describes the full legacy classification pipeline output.
type LegacyPipelineResult struct {
	transactionID       string
	step1Result         StepResult
	step2Result         *StepResult
	step3Result         *StepResult
	exitStep            int
	finalClassification TransactionClassification
}

// NewLegacyPipelineResult builds a LegacyPipelineResult value.
func NewLegacyPipelineResult(
	transactionID string,
	step1Result StepResult,
	step2Result *StepResult,
	step3Result *StepResult,
	exitStep int,
	finalClassification TransactionClassification,
) LegacyPipelineResult {
	result := LegacyPipelineResult{
		transactionID:       transactionID,
		step1Result:         step1Result,
		exitStep:            exitStep,
		finalClassification: finalClassification,
	}

	if step2Result != nil {
		step2ResultCopy := *step2Result
		result.step2Result = &step2ResultCopy
	}

	if step3Result != nil {
		step3ResultCopy := *step3Result
		result.step3Result = &step3ResultCopy
	}


	return result
}

// TransactionID returns the transaction identifier.
func (r LegacyPipelineResult) TransactionID() string {
	return r.transactionID
}

// Step1Result returns the step 1 result.
func (r LegacyPipelineResult) Step1Result() StepResult {
	return r.step1Result
}

// Step2Result returns the step 2 result when present.
func (r LegacyPipelineResult) Step2Result() *StepResult {
	if r.step2Result == nil {
		return nil
	}

	resultCopy := *r.step2Result
	return &resultCopy
}

// Step3Result returns the step 3 result when present.
func (r LegacyPipelineResult) Step3Result() *StepResult {
	if r.step3Result == nil {
		return nil
	}

	resultCopy := *r.step3Result
	return &resultCopy
}

// ExitStep returns the exit step.
func (r LegacyPipelineResult) ExitStep() int {
	return r.exitStep
}

// FinalClassification returns the final classification.
func (r LegacyPipelineResult) FinalClassification() TransactionClassification {
	return r.finalClassification
}

// ReactReviewResult describes one structured ReAct classification payload.
type ReactReviewResult struct {
	version                       string
	source                        string
	classifierFamily              string
	classifierVersion             string
	promptVersion                 string
	model                         string
	transactionID                 string
	matchedTransactionID          *string
	matchedGoodsDescription       *string
	batchID                       *string
	batchSize                     *int
	notAlignedListMatch           bool
	notAlignedListMatchConfidence int
	alignedListMatch              bool
	alignedListMatchConfidence    int
	exitStep                      int
	overallClassification         TransactionClassification
	reason                        string
}

// NewReactReviewResult builds a ReactReviewResult value.
func NewReactReviewResult(
	source string,
	classifierFamily string,
	classifierVersion string,
	promptVersion string,
	model string,
	transactionID string,
	matchedTransactionID *string,
	matchedGoodsDescription *string,
	batchID *string,
	batchSize *int,
	notAlignedListMatch bool,
	notAlignedListMatchConfidence int,
	alignedListMatch bool,
	alignedListMatchConfidence int,
	exitStep int,
	overallClassification TransactionClassification,
	reason string,
) ReactReviewResult {
	return ReactReviewResult{
		version:                       PipelineResultVersionReactV1,
		source:                        source,
		classifierFamily:              classifierFamily,
		classifierVersion:             classifierVersion,
		promptVersion:                 promptVersion,
		model:                         model,
		transactionID:                 transactionID,
		matchedTransactionID:          cloneStringPointer(matchedTransactionID),
		matchedGoodsDescription:       cloneStringPointer(matchedGoodsDescription),
		batchID:                       cloneStringPointer(batchID),
		batchSize:                     cloneIntPointer(batchSize),
		notAlignedListMatch:           notAlignedListMatch,
		notAlignedListMatchConfidence: notAlignedListMatchConfidence,
		alignedListMatch:              alignedListMatch,
		alignedListMatchConfidence:    alignedListMatchConfidence,
		exitStep:                      exitStep,
		overallClassification:         overallClassification,
		reason:                        reason,
	}
}

// Version returns the ReAct envelope version.
func (r ReactReviewResult) Version() string {
	if r.version == "" {
		return PipelineResultVersionReactV1
	}

	return r.version
}

// Source returns the result source label.
func (r ReactReviewResult) Source() string {
	return r.source
}

// ClassifierFamily returns the classifier lineage identifier.
func (r ReactReviewResult) ClassifierFamily() string {
	return r.classifierFamily
}

// ClassifierVersion returns the classifier lineage version identifier.
func (r ReactReviewResult) ClassifierVersion() string {
	return r.classifierVersion
}

// PromptVersion returns the prompt version identifier.
func (r ReactReviewResult) PromptVersion() string {
	return r.promptVersion
}

// Model returns the model name.
func (r ReactReviewResult) Model() string {
	return r.model
}

// TransactionID returns the classified transaction identifier.
func (r ReactReviewResult) TransactionID() string {
	return r.transactionID
}

// MatchedTransactionID returns the reused historical transaction identifier.
func (r ReactReviewResult) MatchedTransactionID() *string {
	return cloneStringPointer(r.matchedTransactionID)
}

// MatchedGoodsDescription returns the matched historical description.
func (r ReactReviewResult) MatchedGoodsDescription() *string {
	return cloneStringPointer(r.matchedGoodsDescription)
}

// BatchID returns the source batch identifier when present.
func (r ReactReviewResult) BatchID() *string {
	return cloneStringPointer(r.batchID)
}

// BatchSize returns the executed batch size when present.
func (r ReactReviewResult) BatchSize() *int {
	return cloneIntPointer(r.batchSize)
}

// NotAlignedListMatch reports whether the description matches a not-aligned list.
func (r ReactReviewResult) NotAlignedListMatch() bool {
	return r.notAlignedListMatch
}

// NotAlignedListMatchConfidence returns the not-aligned list match confidence.
func (r ReactReviewResult) NotAlignedListMatchConfidence() int {
	return r.notAlignedListMatchConfidence
}

// AlignedListMatch reports whether the description matches an aligned list.
func (r ReactReviewResult) AlignedListMatch() bool {
	return r.alignedListMatch
}

// AlignedListMatchConfidence returns the aligned list match confidence.
func (r ReactReviewResult) AlignedListMatchConfidence() int {
	return r.alignedListMatchConfidence
}

// ExitStep returns the step that produced the final classification.
func (r ReactReviewResult) ExitStep() int {
	if r.exitStep > 0 {
		return r.exitStep
	}

	if r.notAlignedListMatch {
		return 1
	}

	if r.alignedListMatch {
		return 2
	}

	return 0
}

// OverallClassification returns the final transaction classification.
func (r ReactReviewResult) OverallClassification() TransactionClassification {
	return r.overallClassification
}

// Reason returns the free-text classifier rationale.
func (r ReactReviewResult) Reason() string {
	return r.reason
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}

	copyValue := *value
	return &copyValue
}

func cloneIntPointer(value *int) *int {
	if value == nil {
		return nil
	}

	copyValue := *value
	return &copyValue
}
