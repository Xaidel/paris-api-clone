package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// TransactionStep5ScreeningQuestion2Justification describes the required justification for screening question 2.
type TransactionStep5ScreeningQuestion2Justification struct {
	value string
}

// NewTransactionStep5ScreeningQuestion2Justification validates and builds a TransactionStep5ScreeningQuestion2Justification.
func NewTransactionStep5ScreeningQuestion2Justification(value string) (TransactionStep5ScreeningQuestion2Justification, error) {
	normalizedValue := strings.TrimSpace(value)
	if normalizedValue == "" {
		return TransactionStep5ScreeningQuestion2Justification{}, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError("screening_question_2_justification", "required", "screening_question_2_justification is required"),
		})
	}

	return TransactionStep5ScreeningQuestion2Justification{value: normalizedValue}, nil
}

// String returns the normalized screening question 2 justification.
func (j TransactionStep5ScreeningQuestion2Justification) String() string {
	return j.value
}
