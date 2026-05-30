package valueobjects

import "github.com/gyud-adb/paris-api/internal/domain"

// ColumnIndex is a zero-based column position within a tabular upload.
type ColumnIndex struct {
	value uint
}

// NewColumnIndex builds a ColumnIndex from a zero-based column position.
func NewColumnIndex(value int) (ColumnIndex, error) {
	if value < 0 {
		return ColumnIndex{}, domain.ErrInvalidColumnIndex
	}

	return ColumnIndex{value: uint(value)}, nil
}

// Int returns the zero-based column position as an int.
func (i ColumnIndex) Int() int {
	return int(i.value)
}

// Equal reports whether two column indexes are equal.
func (i ColumnIndex) Equal(other ColumnIndex) bool {
	return i.value == other.value
}
