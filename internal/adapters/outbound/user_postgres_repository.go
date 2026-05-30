package adapters

import (
	"context"
	"errors"
	"fmt"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	createUserQuery = `
INSERT INTO "user" (id, username, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
`
	createUserProfileQuery = `
INSERT INTO user_profile (user_id, first_name, middle_name, last_name, group_id)
VALUES ($1, $2, $3, $4, $5)
`
	findUserByIDQuery = `
SELECT u.id, u.username, u.password_hash, up.first_name, up.middle_name, up.last_name, up.group_id, u.created_at, u.updated_at
FROM "user" u
JOIN user_profile up ON up.user_id = u.id
WHERE u.id = $1
`
	listUsersQuery = `
SELECT u.id, u.username, u.password_hash, up.first_name, up.middle_name, up.last_name, up.group_id, u.created_at, u.updated_at
FROM "user" u
JOIN user_profile up ON up.user_id = u.id
ORDER BY u.created_at ASC
`
	updateUserQuery = `
UPDATE "user"
SET username = $2,
	password_hash = $3,
	updated_at = $4
WHERE id = $1
`
	updateUserProfileQuery = `
UPDATE user_profile
SET first_name = $2,
	middle_name = $3,
	last_name = $4,
	group_id = $5
WHERE user_id = $1
`
	deleteUserQuery = `
DELETE FROM "user"
WHERE id = $1
`
)

type pgxQuerier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type scanner interface {
	Scan(dest ...any) error
}

// PostgresUserRepository persists users in PostgreSQL.
type PostgresUserRepository struct {
	pool pgxQuerier
}

// NewPostgresUserRepository builds a PostgresUserRepository.
func NewPostgresUserRepository(pool pgxQuerier) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

// Create inserts a new user.
func (r *PostgresUserRepository) Create(ctx context.Context, user *entities.User) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(ctx, createUserQuery, user.ID().String(), user.Username(), user.PasswordHash(), user.CreatedAt(), user.UpdatedAt()); err != nil {
		return fmt.Errorf("executing create user query: %w", err)
	}

	middleNameValue := any(nil)
	if middleName := user.Profile().MiddleName(); middleName != nil {
		middleNameValue = *middleName
	}

	if _, err := querier.Exec(ctx, createUserProfileQuery, user.ID().String(), user.Profile().FirstName(), middleNameValue, user.Profile().LastName(), user.Profile().GroupID().String()); err != nil {
		return fmt.Errorf("executing create user profile query: %w", err)
	}

	return nil
}

// FindByID returns a user by identifier.
func (r *PostgresUserRepository) FindByID(ctx context.Context, id valueobjects.UserID) (*entities.User, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	user, err := scanUser(querier.QueryRow(ctx, findUserByIDQuery, id.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning user by id: %w", err)
	}

	return user, nil
}

// List returns all users.
func (r *PostgresUserRepository) List(ctx context.Context) ([]*entities.User, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	rows, err := querier.Query(ctx, listUsersQuery)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	users := make([]*entities.User, 0)
	for rows.Next() {
		user, scanErr := scanUser(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scanning listed user: %w", scanErr)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating listed users: %w", err)
	}

	return users, nil
}

// Update updates an existing user.
func (r *PostgresUserRepository) Update(ctx context.Context, user *entities.User) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, updateUserQuery, user.ID().String(), user.Username(), user.PasswordHash(), user.UpdatedAt())
	if err != nil {
		return fmt.Errorf("executing update user query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	middleNameValue := any(nil)
	if middleName := user.Profile().MiddleName(); middleName != nil {
		middleNameValue = *middleName
	}

	profileTag, err := querier.Exec(ctx, updateUserProfileQuery, user.ID().String(), user.Profile().FirstName(), middleNameValue, user.Profile().LastName(), user.Profile().GroupID().String())
	if err != nil {
		return fmt.Errorf("executing update user profile query: %w", err)
	}

	if profileTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// DeleteByID deletes an existing user.
func (r *PostgresUserRepository) DeleteByID(ctx context.Context, id valueobjects.UserID) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, deleteUserQuery, id.String())
	if err != nil {
		return fmt.Errorf("executing delete user query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func scanUser(row scanner) (*entities.User, error) {
	var (
		userID       string
		username     string
		passwordHash string
		firstName    string
		middleName   *string
		lastName     string
		groupID      string
		createdAt    time.Time
		updatedAt    time.Time
	)

	if err := row.Scan(&userID, &username, &passwordHash, &firstName, &middleName, &lastName, &groupID, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	parsedUserID, err := valueobjects.UserIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("parsing user id: %w", err)
	}

	parsedGroupID, err := valueobjects.GroupIDFromString(groupID)
	if err != nil {
		return nil, fmt.Errorf("parsing group id: %w", err)
	}

	profile, err := valueobjects.NewUserProfile(firstName, middleName, lastName, parsedGroupID)
	if err != nil {
		return nil, fmt.Errorf("creating user profile: %w", err)
	}

	return entities.ReconstituteUser(parsedUserID, username, passwordHash, profile, createdAt, updatedAt), nil
}
