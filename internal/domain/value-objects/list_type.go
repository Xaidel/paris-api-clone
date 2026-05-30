package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

const (
	listTypeU1Value     = "u1"
	listTypeU2Value     = "u2"
	listTypeSectorValue = "sector"
)

// ListType identifies a classification list.
type ListType struct {
	value string
}

// U1ListType returns the U1 list type.
func U1ListType() ListType {
	return ListType{value: listTypeU1Value}
}

// U2ListType returns the U2 list type.
func U2ListType() ListType {
	return ListType{value: listTypeU2Value}
}

// SectorListType returns the sector list type.
func SectorListType() ListType {
	return ListType{value: listTypeSectorValue}
}

// ListTypeFromString parses and normalizes a list type value.
func ListTypeFromString(value string) (ListType, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case listTypeU1Value:
		return U1ListType(), nil
	case listTypeU2Value:
		return U2ListType(), nil
	case listTypeSectorValue:
		return SectorListType(), nil
	default:
		return ListType{}, domain.NewValidationError([]domain.FieldValidationError{
			domain.NewFieldValidationError("list_type", "invalid_value", "list_type must be one of: u1, u2, sector"),
		})
	}
}

// String returns the canonical list type value.
func (t ListType) String() string {
	return t.value
}

// Equal reports whether two list types are equal.
func (t ListType) Equal(other ListType) bool {
	return t.value == other.value
}
