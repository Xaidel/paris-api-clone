package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

const (
	sectorTypePAAlignedValue    = "PA Aligned"
	sectorTypeHighEmittingValue = "High Emitting"
)

// SectorType is a value object describing a supported sector classification type.
type SectorType struct {
	value string
}

// PAAlignedSectorType returns the canonical PA Aligned sector type.
func PAAlignedSectorType() SectorType {
	return SectorType{value: sectorTypePAAlignedValue}
}

// HighEmittingSectorType returns the canonical High Emitting sector type.
func HighEmittingSectorType() SectorType {
	return SectorType{value: sectorTypeHighEmittingValue}
}

// SectorTypeFromString parses and normalizes a supported sector type.
func SectorTypeFromString(value string) (SectorType, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case strings.ToLower(sectorTypePAAlignedValue):
		return PAAlignedSectorType(), nil
	case strings.ToLower(sectorTypeHighEmittingValue):
		return HighEmittingSectorType(), nil
	default:
		return SectorType{}, domain.ErrInvalidSectorType
	}
}

// String returns the canonical sector type value.
func (t SectorType) String() string {
	return t.value
}

// Equal reports whether two sector types are equal.
func (t SectorType) Equal(other SectorType) bool {
	return t.value == other.value
}
