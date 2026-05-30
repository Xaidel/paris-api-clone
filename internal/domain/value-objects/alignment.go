package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// AlignmentValue is the enum backing an Alignment value object.
type AlignmentValue int

const (
	alignmentUnknownValue AlignmentValue = iota
	AlignmentAlignedValue
	AlignmentUnalignedValue
)

var alignmentValueStrings = map[AlignmentValue]string{
	AlignmentAlignedValue:   "aligned",
	AlignmentUnalignedValue: "unaligned",
}

var alignmentStringValues = map[string]AlignmentValue{
	"aligned":   AlignmentAlignedValue,
	"unaligned": AlignmentUnalignedValue,
}

// String returns the canonical alignment value.
func (v AlignmentValue) String() string {
	return alignmentValueStrings[v]
}

// Alignment describes whether a matcher or pipeline step aligned.
type Alignment struct {
	value AlignmentValue
}

// AlignedAlignment returns the aligned value.
func AlignedAlignment() Alignment {
	return Alignment{value: AlignmentAlignedValue}
}

// UnalignedAlignment returns the unaligned value.
func UnalignedAlignment() Alignment {
	return Alignment{value: AlignmentUnalignedValue}
}

// AlignmentFromString parses and normalizes an alignment value.
func AlignmentFromString(value string) (Alignment, error) {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	alignmentValue, ok := alignmentStringValues[normalizedValue]
	if !ok {
		return Alignment{}, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError("alignment", "invalid_value", "alignment must be one of: aligned, unaligned"),
		})
	}

	return Alignment{value: alignmentValue}, nil
}

// String returns the canonical alignment value.
func (a Alignment) String() string {
	return a.value.String()
}

// Equal reports whether two alignment values are equal.
func (a Alignment) Equal(other Alignment) bool {
	return a.value == other.value
}
