package ports

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// UserRepository persists user aggregates.
type UserRepository interface {
	// Create stores a newly created user.
	Create(ctx context.Context, user *entities.User) error
	// FindByID returns one user or nil when it does not exist.
	FindByID(ctx context.Context, id valueobjects.UserID) (*entities.User, error)
	// List returns all persisted users in repository order.
	List(ctx context.Context) ([]*entities.User, error)
	// Update persists changes to an existing user.
	Update(ctx context.Context, user *entities.User) error
	// DeleteByID removes the identified user.
	DeleteByID(ctx context.Context, id valueobjects.UserID) error
}
