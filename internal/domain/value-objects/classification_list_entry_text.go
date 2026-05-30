package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// ClassificationListEntryText describes canonical text for one classification
// list entry.
type ClassificationListEntryText struct {
	value string
}

// ClassificationListEntryTextFromString parses canonical classification list
// entry text.
func ClassificationListEntryTextFromString(value string) (ClassificationListEntryText, error) {
	normalizedValue := strings.TrimSpace(value)
	if normalizedValue == "" {
		return ClassificationListEntryText{}, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError("entry_text", "required", "entry_text is required"),
		})
	}

	return ClassificationListEntryText{value: normalizedValue}, nil
}

// NewU1ClassificationListEntryText builds canonical text for one U1 list
// entry.
func NewU1ClassificationListEntryText(sector string, eligibleOperationType string, conditionGuidance string) (ClassificationListEntryText, error) {
	normalizedSector := strings.TrimSpace(sector)
	if normalizedSector == "" {
		return ClassificationListEntryText{}, domain.ErrInvalidSector
	}

	normalizedEligibleOperationType := strings.TrimSpace(eligibleOperationType)
	if normalizedEligibleOperationType == "" {
		return ClassificationListEntryText{}, domain.ErrInvalidEligibleOperationType
	}

	normalizedConditionGuidance := strings.TrimSpace(conditionGuidance)
	if normalizedConditionGuidance == "" {
		return ClassificationListEntryText{}, domain.ErrInvalidConditionGuidance
	}

	return ClassificationListEntryText{
		value: "sector: " + normalizedSector + "; eligible_operation_type: " + normalizedEligibleOperationType + "; condition_guidance: " + normalizedConditionGuidance,
	}, nil
}

// NewU2ClassificationListEntryText builds canonical text for one U2 exclusion
// list entry.
func NewU2ClassificationListEntryText(activityType string) (ClassificationListEntryText, error) {
	normalizedActivityType := strings.TrimSpace(activityType)
	if normalizedActivityType == "" {
		return ClassificationListEntryText{}, domain.ErrInvalidActivityType
	}

	return ClassificationListEntryText{value: "activity_type: " + normalizedActivityType}, nil
}

// NewSectorClassificationListEntryText builds canonical text for one sector
// list entry.
func NewSectorClassificationListEntryText(sectorType string, name string, description string) (ClassificationListEntryText, error) {
	normalizedSectorType, err := SectorTypeFromString(sectorType)
	if err != nil {
		return ClassificationListEntryText{}, err
	}

	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return ClassificationListEntryText{}, domain.ErrInvalidSectorName
	}

	normalizedDescription := strings.TrimSpace(description)
	if normalizedDescription == "" {
		return ClassificationListEntryText{}, domain.ErrInvalidSectorDescription
	}

	return ClassificationListEntryText{
		value: "type: " + normalizedSectorType.String() + "; name: " + normalizedName + "; description: " + normalizedDescription,
	}, nil
}

// String returns the canonical classification list entry text.
func (t ClassificationListEntryText) String() string {
	return t.value
}

// Equal reports whether two classification list entry texts are equal.
func (t ClassificationListEntryText) Equal(other ClassificationListEntryText) bool {
	return t.value == other.value
}
