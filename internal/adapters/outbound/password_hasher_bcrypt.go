package adapters

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const defaultBcryptCost = bcrypt.DefaultCost

// BcryptPasswordHasher hashes passwords using bcrypt.
type BcryptPasswordHasher struct {
	cost int
}

// NewBcryptPasswordHasher builds a BcryptPasswordHasher.
func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{cost: defaultBcryptCost}
}

// Hash returns a bcrypt hash for the provided password.
func (h *BcryptPasswordHasher) Hash(_ context.Context, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", fmt.Errorf("generating bcrypt hash: %w", err)
	}

	return string(hash), nil
}
