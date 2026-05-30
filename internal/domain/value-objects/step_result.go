package valueobjects

// StepResult describes the outcome of one pipeline step.
type StepResult struct {
	stepNumber       int
	keywordDecision  *MatchDecision
	semanticDecision *MatchDecision
	stepAlignment    Alignment
	booleanResult    *bool
	reason           error
}

// NewStepResult builds a scored step result for step 1 or 2.
func NewStepResult(stepNumber int, keywordDecision MatchDecision, semanticDecision MatchDecision) StepResult {
	stepAlignment := UnalignedAlignment()
	if keywordDecision.Alignment().Equal(AlignedAlignment()) || semanticDecision.Alignment().Equal(AlignedAlignment()) {
		stepAlignment = AlignedAlignment()
	}

	keywordDecisionCopy := keywordDecision
	semanticDecisionCopy := semanticDecision

	return StepResult{
		stepNumber:       stepNumber,
		keywordDecision:  &keywordDecisionCopy,
		semanticDecision: &semanticDecisionCopy,
		stepAlignment:    stepAlignment,
		booleanResult:    nil,
		reason:           nil,
	}
}

// NewBooleanStepResult builds the step 3 boolean result.
func NewBooleanStepResult(stepNumber int, booleanResult bool) StepResult {
	return NewBooleanStepResultWithReason(stepNumber, booleanResult, nil)
}

// NewBooleanStepResultWithReason builds the step 3 boolean result with an
// optional explanatory reason.
func NewBooleanStepResultWithReason(stepNumber int, booleanResult bool, reason error) StepResult {
	stepAlignment := UnalignedAlignment()
	if booleanResult {
		stepAlignment = AlignedAlignment()
	}

	booleanResultCopy := booleanResult

	return StepResult{
		stepNumber:       stepNumber,
		keywordDecision:  nil,
		semanticDecision: nil,
		stepAlignment:    stepAlignment,
		booleanResult:    &booleanResultCopy,
		reason:           reason,
	}
}

// StepNumber returns the pipeline step number.
func (r StepResult) StepNumber() int {
	return r.stepNumber
}

// KeywordDecision returns the keyword match decision when present.
func (r StepResult) KeywordDecision() *MatchDecision {
	if r.keywordDecision == nil {
		return nil
	}

	decisionCopy := *r.keywordDecision
	return &decisionCopy
}

// SemanticDecision returns the semantic match decision when present.
func (r StepResult) SemanticDecision() *MatchDecision {
	if r.semanticDecision == nil {
		return nil
	}

	decisionCopy := *r.semanticDecision
	return &decisionCopy
}

// StepAlignment returns the combined step alignment decision.
func (r StepResult) StepAlignment() Alignment {
	return r.stepAlignment
}

// BooleanResult returns the step 3 boolean result when present.
func (r StepResult) BooleanResult() *bool {
	if r.booleanResult == nil {
		return nil
	}

	resultCopy := *r.booleanResult
	return &resultCopy
}

// Reason returns the optional explanation attached to the step result.
func (r StepResult) Reason() error {
	return r.reason
}
