package valueobjects

import (
	"reflect"
	"testing"
)

// This test captures the Phase 4 domain cleanup: legacy pipeline results keep
// only the remaining step summaries and final classification state.
func TestPipelineResultLegacyOmitsDeprecatedScoreAndConfidenceState(t *testing.T) {
	t.Parallel()

	step1 := NewStepResult(
		1,
		NewMatchDecision(AlignedAlignment(), 0.8, "u1-a", nil),
		NewMatchDecision(UnalignedAlignment(), 0.6, "u1-b", nil),
	)
	step2 := NewStepResult(
		2,
		NewMatchDecision(UnalignedAlignment(), 0.5, "u2-a", nil),
		NewMatchDecision(AlignedAlignment(), 0.25, "u2-b", nil),
	)
	step3 := NewBooleanStepResult(3, false)

	result := NewPipelineResult(
		"tx-1",
		step1,
		&step2,
		&step3,
		3,
		NextStepTransactionClassification(),
	)

	legacy := result.Legacy()
	if legacy == nil {
		t.Fatal("result.Legacy() = nil, want legacy result")
	}

	legacyType := reflect.TypeOf(*legacy)
	for _, fieldName := range []string{"combinedKeywordScore", "combinedSemanticScore", "confidenceScore"} {
		if _, exists := legacyType.FieldByName(fieldName); exists {
			t.Fatalf("LegacyPipelineResult field %q present = true, want false", fieldName)
		}
	}
}

func stepResultPointer(result StepResult) *StepResult {
	resultCopy := result
	return &resultCopy
}
