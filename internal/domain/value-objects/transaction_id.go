package valueobjects

import (
	"github.com/gyud-adb/paris-api/internal/domain"
)

type transactionIDTag struct{}

// TransactionID is a value object representing a UUIDv7 transaction identifier.
type TransactionID = UUIDv7ID[transactionIDTag]

// NewTransactionID creates a new UUIDv7 transaction identifier.
func NewTransactionID() (TransactionID, error) {
	return newUUIDv7ID[transactionIDTag]()
}

// TransactionIDFromString parses and validates a UUIDv7 transaction identifier.
func TransactionIDFromString(value string) (TransactionID, error) {
	parsedValue, err := parseUUIDv7ID[transactionIDTag](value)
	if err != nil {
		return TransactionID{}, domain.ErrInvalidTransactionID
	}

	return parsedValue, nil
}
