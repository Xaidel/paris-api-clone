package ports

import "context"

// PasswordHasher hashes plain-text passwords for secure storage.
type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
}
