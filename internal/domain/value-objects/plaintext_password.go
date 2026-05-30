package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// PlaintextPassword is a validated user password input.
type PlaintextPassword struct {
	value string
}

// NewPlaintextPassword validates and builds a PlaintextPassword.
func NewPlaintextPassword(value string) (PlaintextPassword, error) {
	normalized := strings.TrimSpace(value)
	if len(normalized) < 8 {
		return PlaintextPassword{}, domain.ErrInvalidPassword
	}

	return PlaintextPassword{value: normalized}, nil
}

// String returns the password value.
func (p PlaintextPassword) String() string {
	return p.value
}
