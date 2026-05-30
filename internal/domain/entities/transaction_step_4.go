package entities

import (
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionStep4 stores the aggregated reviewer decision for classification step 4.
type TransactionStep4 struct {
	aggregateRoot
	transactionID     valueobjects.TransactionID
	sectorID          valueobjects.SectorID
	additionalContext valueobjects.TransactionStep4AdditionalContext
	isHighEmitting    bool
	createdAt         time.Time
	updatedAt         time.Time
}

// NewTransactionStep4 creates a valid step 4 review record.
func NewTransactionStep4(
	transactionID valueobjects.TransactionID,
	sectorID valueobjects.SectorID,
	additionalContext valueobjects.TransactionStep4AdditionalContext,
	isHighEmitting bool,
	now time.Time,
) (*TransactionStep4, error) {
	if now.IsZero() {
		return nil, domain.ErrInvalidTimestamp
	}

	return &TransactionStep4{
		transactionID:     transactionID,
		sectorID:          sectorID,
		additionalContext: additionalContext,
		isHighEmitting:    isHighEmitting,
		createdAt:         now,
		updatedAt:         now,
	}, nil
}

// ReconstituteTransactionStep4 rebuilds a step 4 review record from storage.
func ReconstituteTransactionStep4(
	transactionID valueobjects.TransactionID,
	sectorID valueobjects.SectorID,
	additionalContext valueobjects.TransactionStep4AdditionalContext,
	isHighEmitting bool,
	createdAt time.Time,
	updatedAt time.Time,
) *TransactionStep4 {
	return &TransactionStep4{
		transactionID:     transactionID,
		sectorID:          sectorID,
		additionalContext: additionalContext,
		isHighEmitting:    isHighEmitting,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}
}

// TransactionID returns the related transaction identifier.
func (s *TransactionStep4) TransactionID() valueobjects.TransactionID {
	return s.transactionID
}

// SectorID returns the selected sector identifier.
func (s *TransactionStep4) SectorID() valueobjects.SectorID {
	return s.sectorID
}

// AdditionalContext returns the reviewer justification.
func (s *TransactionStep4) AdditionalContext() valueobjects.TransactionStep4AdditionalContext {
	return s.additionalContext
}

// IsHighEmitting reports whether the reviewer marked the transaction as high emitting.
func (s *TransactionStep4) IsHighEmitting() bool {
	return s.isHighEmitting
}

// CreatedAt returns the creation timestamp.
func (s *TransactionStep4) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt returns the update timestamp.
func (s *TransactionStep4) UpdatedAt() time.Time {
	return s.updatedAt
}
