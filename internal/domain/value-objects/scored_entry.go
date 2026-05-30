package valueobjects

// ScoredEntry describes one candidate entry score.
type ScoredEntry struct {
	entryText string
	score     float64
}

// NewScoredEntry builds a ScoredEntry value.
func NewScoredEntry(entryText string, score float64) ScoredEntry {
	return ScoredEntry{entryText: entryText, score: score}
}

// EntryText returns the candidate entry text.
func (e ScoredEntry) EntryText() string {
	return e.entryText
}

// Score returns the candidate score.
func (e ScoredEntry) Score() float64 {
	return e.score
}
