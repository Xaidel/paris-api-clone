// This code is not used in production.
package valueobjects

// ConfidenceScore describes the stubbed logistic confidence output.
type ConfidenceScore struct {
	combinedKeywordScore  float64
	combinedSemanticScore float64
	beta0                 float64
	beta1                 float64
	beta2                 float64
	probability           float64
	certaintyScore        float64
	isStub                bool
}

// NewConfidenceScore builds a ConfidenceScore value.
func NewConfidenceScore(combinedKeywordScore, combinedSemanticScore, beta0, beta1, beta2 float64) ConfidenceScore {
	return ConfidenceScore{
		combinedKeywordScore:  combinedKeywordScore,
		combinedSemanticScore: combinedSemanticScore,
		beta0:                 beta0,
		beta1:                 beta1,
		beta2:                 beta2,
	}
}

// Compute returns the stubbed confidence score.
func (s ConfidenceScore) Compute() ConfidenceScore {
	// TODO: implement sigmoid confidence once logistic regression weights are trained.
	// C = 1 / (1 + exp(-(β0 + β1*S_kw_combined + β2*S_sem_combined)))
	// S_C = 200 * |C - 0.5|
	// Training requires a labeled dataset of human-classified transactions with their
	// combined keyword and semantic scores as features.
	s.isStub = true
	s.probability = 0
	s.certaintyScore = 0
	return s
}

// CombinedKeywordScore returns the combined keyword score.
func (s ConfidenceScore) CombinedKeywordScore() float64 {
	return s.combinedKeywordScore
}

// CombinedSemanticScore returns the combined semantic score.
func (s ConfidenceScore) CombinedSemanticScore() float64 {
	return s.combinedSemanticScore
}

// Beta0 returns the logistic intercept.
func (s ConfidenceScore) Beta0() float64 {
	return s.beta0
}

// Beta1 returns the keyword weight.
func (s ConfidenceScore) Beta1() float64 {
	return s.beta1
}

// Beta2 returns the semantic weight.
func (s ConfidenceScore) Beta2() float64 {
	return s.beta2
}

// Probability returns the confidence probability.
func (s ConfidenceScore) Probability() float64 {
	return s.probability
}

// CertaintyScore returns the scaled certainty score.
func (s ConfidenceScore) CertaintyScore() float64 {
	return s.certaintyScore
}

// IsStub reports whether the confidence score is still stubbed.
func (s ConfidenceScore) IsStub() bool {
	return s.isStub
}
