package adapters

import (
	"context"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestBcryptPasswordHasherHash verifies the bcrypt password hasher hash behavior and the expected outcome asserted below.
func TestBcryptPasswordHasherHash(t *testing.T) {
	t.Parallel()

	hasher := NewBcryptPasswordHasher()
	hash, err := hasher.Hash(context.Background(), "supersecret")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	if hash == "supersecret" {
		t.Fatal("Hash() returned plain-text password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("supersecret")); err != nil {
		t.Fatalf("CompareHashAndPassword() error = %v", err)
	}
}
