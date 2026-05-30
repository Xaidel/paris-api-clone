package valueobjects

import (
	"errors"
	"testing"
)

// TestNewMatchDecisionForcesUnalignedOnReason verifies the new match decision forces unaligned on reason behavior and the expected outcome asserted below.
func TestNewMatchDecisionForcesUnalignedOnReason(t *testing.T) {
	t.Parallel()

	decision := NewMatchDecision(AlignedAlignment(), 0.91, "entry", errors.New("embedding failure"))

	if decision.Alignment().String() != UnalignedAlignment().String() {
		t.Fatalf("decision.Alignment() = %q, want %q", decision.Alignment().String(), UnalignedAlignment().String())
	}

	if decision.Score() != 0 {
		t.Fatalf("decision.Score() = %v, want %v", decision.Score(), 0.0)
	}

	if decision.Reason() == nil {
		t.Fatal("decision.Reason() = nil, want error")
	}
}

// TestNewMatchDecisionWithCandidates verifies the new match decision with candidates behavior and the expected outcome asserted below.
func TestNewMatchDecisionWithCandidates(t *testing.T) {
	t.Parallel()

	decision := NewMatchDecisionWithCandidates(
		AlignedAlignment(),
		0.91,
		"entry-a",
		[]ScoredEntry{
			NewScoredEntry("entry-a", 0.91),
			NewScoredEntry("entry-b", 0.83),
		},
		nil,
	)

	if len(decision.Candidates()) != 2 {
		t.Fatalf("len(decision.Candidates()) = %d, want %d", len(decision.Candidates()), 2)
	}

	if decision.Candidates()[1].EntryText() != "entry-b" {
		t.Fatalf("decision.Candidates()[1].EntryText() = %q, want %q", decision.Candidates()[1].EntryText(), "entry-b")
	}
}
