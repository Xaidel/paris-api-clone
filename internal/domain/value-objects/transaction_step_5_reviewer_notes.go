package valueobjects

import "strings"

// TransactionStep5ReviewerNotes describes optional reviewer notes for step 5.
type TransactionStep5ReviewerNotes struct {
	value    string
	hasValue bool
}

// NewTransactionStep5ReviewerNotes normalizes optional reviewer notes.
func NewTransactionStep5ReviewerNotes(value *string) TransactionStep5ReviewerNotes {
	if value == nil {
		return TransactionStep5ReviewerNotes{}
	}

	normalizedValue := strings.TrimSpace(*value)
	if normalizedValue == "" {
		return TransactionStep5ReviewerNotes{}
	}

	return TransactionStep5ReviewerNotes{value: normalizedValue, hasValue: true}
}

// HasValue reports whether reviewer notes are present.
func (n TransactionStep5ReviewerNotes) HasValue() bool {
	return n.hasValue
}

// String returns reviewer notes or nil when absent.
func (n TransactionStep5ReviewerNotes) String() *string {
	if !n.hasValue {
		return nil
	}

	value := n.value
	return &value
}
