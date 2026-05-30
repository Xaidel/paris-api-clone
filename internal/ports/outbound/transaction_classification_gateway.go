package ports

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionClassificationCandidate describes one transaction awaiting classification.
type TransactionClassificationCandidate struct {
	// TransactionID identifies the transaction to classify.
	TransactionID valueobjects.TransactionID
	// GoodsDescription is the primary text input used by the classifier.
	GoodsDescription string
	// Sector carries an optional precomputed sector hint from earlier steps.
	Sector *string
}

// TransactionClassificationDecision describes one classification decision.
type TransactionClassificationDecision struct {
	// TransactionID identifies the classified transaction.
	TransactionID valueobjects.TransactionID
	// Classification and Status contain the resulting review outcome.
	Classification valueobjects.TransactionClassification
	Status         valueobjects.TransactionStatus
	// ReviewResult preserves the detailed evidence and step outputs.
	ReviewResult valueobjects.PipelineResult
	// MatchedTransactionID references a reused historical decision when present.
	MatchedTransactionID *valueobjects.TransactionID
	// Source records which classifier path produced the decision.
	Source string
}

// TransactionClassificationGateway classifies transactions behind an outbound boundary.
type TransactionClassificationGateway interface {
	// Classify evaluates the supplied candidates and returns one decision per
	// classified transaction.
	Classify(ctx context.Context, candidates []TransactionClassificationCandidate) ([]TransactionClassificationDecision, error)
}
