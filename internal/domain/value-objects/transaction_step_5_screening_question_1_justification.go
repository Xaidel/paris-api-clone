package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// TransactionStep5ScreeningQuestion1Justification describes the required justification for screening question 1.
type TransactionStep5ScreeningQuestion1Justification struct {
	value string
}

// NewTransactionStep5ScreeningQuestion1Justification validates and builds a TransactionStep5ScreeningQuestion1Justification.
func NewTransactionStep5ScreeningQuestion1Justification(value string) (TransactionStep5ScreeningQuestion1Justification, error) {
	normalizedValue := strings.TrimSpace(value)
	if normalizedValue == "" {
		return TransactionStep5ScreeningQuestion1Justification{}, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError("screening_question_1_justification", "required", "screening_question_1_justification is required"),
		})
	}

	return TransactionStep5ScreeningQuestion1Justification{value: normalizedValue}, nil
}

// String returns the normalized screening question 1 justification.
func (j TransactionStep5ScreeningQuestion1Justification) String() string {
	return j.value
}
