package valueobjects

import "slices"

// MatchDecision describes the outcome of a single scoring run.
type MatchDecision struct {
	alignment    Alignment
	score        float64
	matchedEntry string
	candidates   []ScoredEntry
	reason       error
}

// NewMatchDecision builds a MatchDecision value.
func NewMatchDecision(alignment Alignment, score float64, matchedEntry string, reason error) MatchDecision {
	return NewMatchDecisionWithCandidates(alignment, score, matchedEntry, nil, reason)
}

// NewMatchDecisionWithCandidates builds a MatchDecision value with candidate
// scores.
func NewMatchDecisionWithCandidates(alignment Alignment, score float64, matchedEntry string, candidates []ScoredEntry, reason error) MatchDecision {
	if reason != nil {
		return MatchDecision{
			alignment:    UnalignedAlignment(),
			score:        0,
			matchedEntry: matchedEntry,
			candidates:   nil,
			reason:       reason,
		}
	}

	normalizedCandidates := slices.Clone(candidates)
	normalizedMatchedEntry := matchedEntry
	if len(normalizedCandidates) == 0 && normalizedMatchedEntry != "" {
		normalizedCandidates = []ScoredEntry{NewScoredEntry(normalizedMatchedEntry, score)}
	}

	if normalizedMatchedEntry == "" && len(normalizedCandidates) > 0 {
		normalizedMatchedEntry = normalizedCandidates[0].EntryText()
	}

	return MatchDecision{
		alignment:    alignment,
		score:        score,
		matchedEntry: normalizedMatchedEntry,
		candidates:   normalizedCandidates,
		reason:       nil,
	}
}

// Alignment returns the alignment decision.
func (d MatchDecision) Alignment() Alignment {
	return d.alignment
}

// Score returns the raw normalized score.
func (d MatchDecision) Score() float64 {
	return d.score
}

// MatchedEntry returns the best matching entry.
func (d MatchDecision) MatchedEntry() string {
	return d.matchedEntry
}

// Candidates returns the scored candidates in descending rank order when
// available.
func (d MatchDecision) Candidates() []ScoredEntry {
	return slices.Clone(d.candidates)
}

// Reason returns the domain-level reliability signal.
func (d MatchDecision) Reason() error {
	return d.reason
}
