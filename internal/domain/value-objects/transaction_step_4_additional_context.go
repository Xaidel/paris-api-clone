package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// TransactionStep4AdditionalContext describes the reviewer justification for step 4.
type TransactionStep4AdditionalContext struct {
	value string
}

// NewTransactionStep4AdditionalContext validates and builds a TransactionStep4AdditionalContext.
func NewTransactionStep4AdditionalContext(value string) (TransactionStep4AdditionalContext, error) {
	normalizedValue := strings.TrimSpace(value)
	if normalizedValue == "" {
		return TransactionStep4AdditionalContext{}, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError("additional_context", "required", "additional_context is required"),
		})
	}

	return TransactionStep4AdditionalContext{value: normalizedValue}, nil
}

// String returns the reviewer justification.
func (c TransactionStep4AdditionalContext) String() string {
	return c.value
}
