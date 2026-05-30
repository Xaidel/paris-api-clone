package entities

import (
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionStep5 stores the persisted screening answers for classification step 5.
type TransactionStep5 struct {
	aggregateRoot
	transactionID                   valueobjects.TransactionID
	screeningQuestion1Answer        bool
	screeningQuestion1Justification valueobjects.TransactionStep5ScreeningQuestion1Justification
	screeningQuestion2Answer        bool
	screeningQuestion2Justification valueobjects.TransactionStep5ScreeningQuestion2Justification
	reviewerNotes                   valueobjects.TransactionStep5ReviewerNotes
	isFinal                         bool
	createdAt                       time.Time
	updatedAt                       time.Time
}

// NewTransactionStep5 creates a valid step 5 screening record.
func NewTransactionStep5(
	transactionID valueobjects.TransactionID,
	screeningQuestion1Answer bool,
	screeningQuestion1Justification valueobjects.TransactionStep5ScreeningQuestion1Justification,
	screeningQuestion2Answer bool,
	screeningQuestion2Justification valueobjects.TransactionStep5ScreeningQuestion2Justification,
	reviewerNotes valueobjects.TransactionStep5ReviewerNotes,
	isFinal bool,
	now time.Time,
) (*TransactionStep5, error) {
	if now.IsZero() {
		return nil, domain.ErrInvalidTimestamp
	}

	return &TransactionStep5{
		transactionID:                   transactionID,
		screeningQuestion1Answer:        screeningQuestion1Answer,
		screeningQuestion1Justification: screeningQuestion1Justification,
		screeningQuestion2Answer:        screeningQuestion2Answer,
		screeningQuestion2Justification: screeningQuestion2Justification,
		reviewerNotes:                   reviewerNotes,
		isFinal:                         isFinal,
		createdAt:                       now,
		updatedAt:                       now,
	}, nil
}

// ReconstituteTransactionStep5 rebuilds a step 5 screening record from storage.
func ReconstituteTransactionStep5(
	transactionID valueobjects.TransactionID,
	screeningQuestion1Answer bool,
	screeningQuestion1Justification valueobjects.TransactionStep5ScreeningQuestion1Justification,
	screeningQuestion2Answer bool,
	screeningQuestion2Justification valueobjects.TransactionStep5ScreeningQuestion2Justification,
	reviewerNotes valueobjects.TransactionStep5ReviewerNotes,
	isFinal bool,
	createdAt time.Time,
	updatedAt time.Time,
) *TransactionStep5 {
	return &TransactionStep5{
		transactionID:                   transactionID,
		screeningQuestion1Answer:        screeningQuestion1Answer,
		screeningQuestion1Justification: screeningQuestion1Justification,
		screeningQuestion2Answer:        screeningQuestion2Answer,
		screeningQuestion2Justification: screeningQuestion2Justification,
		reviewerNotes:                   reviewerNotes,
		isFinal:                         isFinal,
		createdAt:                       createdAt,
		updatedAt:                       updatedAt,
	}
}

// TransactionID returns the related transaction identifier.
func (s *TransactionStep5) TransactionID() valueobjects.TransactionID {
	return s.transactionID
}

// ScreeningQuestion1Answer returns the screening question 1 answer.
func (s *TransactionStep5) ScreeningQuestion1Answer() bool {
	return s.screeningQuestion1Answer
}

// ScreeningQuestion1Justification returns the screening question 1 justification.
func (s *TransactionStep5) ScreeningQuestion1Justification() valueobjects.TransactionStep5ScreeningQuestion1Justification {
	return s.screeningQuestion1Justification
}

// ScreeningQuestion2Answer returns the screening question 2 answer.
func (s *TransactionStep5) ScreeningQuestion2Answer() bool {
	return s.screeningQuestion2Answer
}

// ScreeningQuestion2Justification returns the screening question 2 justification.
func (s *TransactionStep5) ScreeningQuestion2Justification() valueobjects.TransactionStep5ScreeningQuestion2Justification {
	return s.screeningQuestion2Justification
}

// ReviewerNotes returns the optional reviewer notes.
func (s *TransactionStep5) ReviewerNotes() valueobjects.TransactionStep5ReviewerNotes {
	return s.reviewerNotes
}

// IsFinal reports whether the step 5 review should finalize the transaction.
func (s *TransactionStep5) IsFinal() bool {
	return s.isFinal
}

// Classification returns the deterministic step 5 classification outcome.
func (s *TransactionStep5) Classification() valueobjects.TransactionClassification {
	if s.screeningQuestion1Answer || s.screeningQuestion2Answer {
		return valueobjects.NotAlignedTransactionClassification()
	}

	return valueobjects.AlignedTransactionClassification()
}

// CreatedAt returns the creation timestamp.
func (s *TransactionStep5) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt returns the update timestamp.
func (s *TransactionStep5) UpdatedAt() time.Time {
	return s.updatedAt
}
