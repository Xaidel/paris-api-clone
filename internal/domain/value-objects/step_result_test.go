package valueobjects

import "testing"

// TestNewStepResultUsesORLogic verifies the new step result uses or logic behavior and the expected outcome asserted below.
func TestNewStepResultUsesORLogic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		keywordDecision   MatchDecision
		semanticDecision  MatchDecision
		wantStepAlignment Alignment
	}{
		{
			name:              "semantic aligned makes step aligned",
			keywordDecision:   NewMatchDecision(UnalignedAlignment(), 0.2, "u1-a", nil),
			semanticDecision:  NewMatchDecision(AlignedAlignment(), 0.8, "u1-b", nil),
			wantStepAlignment: AlignedAlignment(),
		},
		{
			name:              "both unaligned keeps step unaligned",
			keywordDecision:   NewMatchDecision(UnalignedAlignment(), 0.2, "u1-a", nil),
			semanticDecision:  NewMatchDecision(UnalignedAlignment(), 0.4, "u1-b", nil),
			wantStepAlignment: UnalignedAlignment(),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := NewStepResult(1, tc.keywordDecision, tc.semanticDecision)

			if result.StepAlignment().String() != tc.wantStepAlignment.String() {
				t.Fatalf("result.StepAlignment() = %q, want %q", result.StepAlignment().String(), tc.wantStepAlignment.String())
			}
		})
	}
}
