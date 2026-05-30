package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

const (
	transactionClassificationUnclassifiedValue  = "unclassified"
	transactionClassificationAlignedValue       = "aligned"
	transactionClassificationNotAlignedValue    = "not-aligned"
	transactionClassificationNextStepValue      = "next_step"
	transactionClassificationNextStepAliasValue = "next-step"
)

// Active transaction classifications are limited to:
// - unclassified
// - next_step
// - aligned
// - not-aligned

// TransactionClassification describes the review classification assigned to a transaction.
type TransactionClassification struct {
	value string
}

// UnclassifiedTransactionClassification returns the default unclassified value.
func UnclassifiedTransactionClassification() TransactionClassification {
	return TransactionClassification{value: transactionClassificationUnclassifiedValue}
}

// AlignedTransactionClassification returns the aligned value.
func AlignedTransactionClassification() TransactionClassification {
	return TransactionClassification{value: transactionClassificationAlignedValue}
}

// NotAlignedTransactionClassification returns the not-aligned value.
func NotAlignedTransactionClassification() TransactionClassification {
	return TransactionClassification{value: transactionClassificationNotAlignedValue}
}

// NextStepTransactionClassification returns the next_step value.
func NextStepTransactionClassification() TransactionClassification {
	return TransactionClassification{value: transactionClassificationNextStepValue}
}

// TransactionClassificationFromString parses and normalizes a transaction classification value.
func TransactionClassificationFromString(value string) (TransactionClassification, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case transactionClassificationUnclassifiedValue:
		return UnclassifiedTransactionClassification(), nil
	case transactionClassificationAlignedValue:
		return AlignedTransactionClassification(), nil
	case transactionClassificationNotAlignedValue:
		return NotAlignedTransactionClassification(), nil
	case transactionClassificationNextStepValue, transactionClassificationNextStepAliasValue:
		return NextStepTransactionClassification(), nil
	default:
		return TransactionClassification{}, domain.ErrInvalidTransactionClassification
	}
}

// String returns the canonical transaction classification value.
func (c TransactionClassification) String() string {
	return c.value
}

// Equal reports whether two transaction classifications are equal.
func (c TransactionClassification) Equal(other TransactionClassification) bool {
	return c.value == other.value
}

// IsTerminal reports whether the classification is a terminal review outcome.
func (c TransactionClassification) IsTerminal() bool {
	return c.Equal(AlignedTransactionClassification()) || c.Equal(NotAlignedTransactionClassification())
}
