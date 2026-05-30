package valueobjects

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var errInvalidCanonicalUUID = errors.New("invalid canonical uuid")

// UUIDv7ID is a generic UUIDv7-backed identifier distinguished by its phantom tag type.
type UUIDv7ID[T any] struct {
	value uuid.UUID
}

func newUUIDv7ID[T any]() (UUIDv7ID[T], error) {
	value, err := uuid.NewV7()
	if err != nil {
		return UUIDv7ID[T]{}, err
	}

	return UUIDv7ID[T]{value: value}, nil
}

func parseUUIDv7ID[T any](value string) (UUIDv7ID[T], error) {
	parsedValue, err := parseCanonicalUUID(value)
	if err != nil {
		return UUIDv7ID[T]{}, err
	}
	if parsedValue.Version() != uuid.Version(7) || parsedValue.Variant() != uuid.RFC4122 {
		return UUIDv7ID[T]{}, errInvalidCanonicalUUID
	}

	return UUIDv7ID[T]{value: parsedValue}, nil
}

// String returns the canonical identifier string.
func (id UUIDv7ID[T]) String() string {
	return id.value.String()
}

// Equal reports whether two identifiers are equal.
func (id UUIDv7ID[T]) Equal(other UUIDv7ID[T]) bool {
	return id.value == other.value
}

// IsZero reports whether the identifier was left unset.
func (id UUIDv7ID[T]) IsZero() bool {
	return id.value == uuid.Nil
}

// parseCanonicalUUID accepts only trimmed, lowercase canonical UUID text.
func parseCanonicalUUID(value string) (uuid.UUID, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	parsedValue, err := uuid.Parse(normalized)
	if err != nil {
		return uuid.Nil, err
	}
	if parsedValue.String() != normalized {
		return uuid.Nil, errInvalidCanonicalUUID
	}

	return parsedValue, nil
}
