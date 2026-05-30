package adapters

import (
	"context"
	"regexp"
	"testing"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/pashagolub/pgxmock/v4"
)

// TestPostgresUserRepositoryCreate verifies the postgres user repository create behavior and the expected outcome asserted below.
func TestPostgresUserRepositoryCreate(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	userID, err := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("UserIDFromString() error = %v", err)
	}

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	profile, err := valueobjects.NewUserProfile("Alice", nil, "Admin", groupID)
	if err != nil {
		t.Fatalf("NewUserProfile() error = %v", err)
	}

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	user, err := entities.NewUser(userID, "alice", "hashed-password", profile, now)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	repository := NewPostgresUserRepository(mock)
	mock.ExpectExec(regexp.QuoteMeta(createUserQuery)).
		WithArgs(user.ID().String(), user.Username(), user.PasswordHash(), user.CreatedAt(), user.UpdatedAt()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectExec(regexp.QuoteMeta(createUserProfileQuery)).
		WithArgs(user.ID().String(), user.Profile().FirstName(), nil, user.Profile().LastName(), user.Profile().GroupID().String()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	if err := repository.Create(context.Background(), user); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

// TestPostgresUserRepositoryFindByID verifies the postgres user repository find by ID behavior and the expected outcome asserted below.
func TestPostgresUserRepositoryFindByID(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		rows    *pgxmock.Rows
		wantNil bool
	}{
		{name: "returns user", rows: pgxmock.NewRows([]string{"id", "username", "password_hash", "first_name", "middle_name", "last_name", "group_id", "created_at", "updated_at"}).AddRow("01962b8f-aeb2-7e03-a8ff-1edce1300002", "alice", "hashed-password", "Alice", nil, "Admin", "01962b8f-aeb2-7e03-a8ff-1edce1300001", now, now)},
		{name: "returns nil when not found", rows: pgxmock.NewRows([]string{"id", "username", "password_hash", "first_name", "middle_name", "last_name", "group_id", "created_at", "updated_at"}), wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock, err := pgxmock.NewPool()
			if err != nil {
				t.Fatalf("pgxmock.NewPool() error = %v", err)
			}
			defer mock.Close()

			userID, err := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
			if err != nil {
				t.Fatalf("UserIDFromString() error = %v", err)
			}

			mock.ExpectQuery(regexp.QuoteMeta(findUserByIDQuery)).WithArgs(userID.String()).WillReturnRows(tt.rows)
			repository := NewPostgresUserRepository(mock)

			user, err := repository.FindByID(context.Background(), userID)
			if err != nil {
				t.Fatalf("FindByID() error = %v", err)
			}

			if tt.wantNil {
				if user != nil {
					t.Fatal("expected nil user")
				}
			} else if user.ID().String() != userID.String() {
				t.Fatalf("user.ID() = %q, want %q", user.ID().String(), userID.String())
			} else if user.Username() != "alice" {
				t.Fatalf("user.Username() = %q, want %q", user.Username(), "alice")
			}
		})
	}
}

// TestPostgresUserRepositoryListUpdateDelete verifies the postgres user repository list update delete behavior and the expected outcome asserted below.
func TestPostgresUserRepositoryListUpdateDelete(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	repository := NewPostgresUserRepository(mock)
	userID, err := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("UserIDFromString() error = %v", err)
	}

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	profile, err := valueobjects.NewUserProfile("Alice", nil, "Admin", groupID)
	if err != nil {
		t.Fatalf("NewUserProfile() error = %v", err)
	}

	user := entities.ReconstituteUser(userID, "alice", "hashed-password", profile, now, now)

	mock.ExpectQuery(regexp.QuoteMeta(listUsersQuery)).WillReturnRows(
		pgxmock.NewRows([]string{"id", "username", "password_hash", "first_name", "middle_name", "last_name", "group_id", "created_at", "updated_at"}).AddRow(userID.String(), "alice", "hashed-password", "Alice", nil, "Admin", groupID.String(), now, now),
	)

	users, err := repository.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("len(users) = %d, want 1", len(users))
	}

	mock.ExpectExec(regexp.QuoteMeta(updateUserQuery)).WithArgs(userID.String(), user.Username(), user.PasswordHash(), user.UpdatedAt()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	mock.ExpectExec(regexp.QuoteMeta(updateUserProfileQuery)).WithArgs(userID.String(), user.Profile().FirstName(), nil, user.Profile().LastName(), user.Profile().GroupID().String()).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	if err := repository.Update(context.Background(), user); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(deleteUserQuery)).WithArgs(userID.String()).WillReturnResult(pgxmock.NewResult("DELETE", 1))
	if err := repository.DeleteByID(context.Background(), userID); err != nil {
		t.Fatalf("DeleteByID() error = %v", err)
	}
}

// TestScanUserInvalidData verifies the scan user invalid data behavior and the expected outcome asserted below.
func TestScanUserInvalidData(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	userID, err := valueobjects.UserIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("UserIDFromString() error = %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(findUserByIDQuery)).WithArgs(userID.String()).WillReturnRows(
		pgxmock.NewRows([]string{"id", "username", "password_hash", "first_name", "middle_name", "last_name", "group_id", "created_at", "updated_at"}).AddRow("bad-id", "alice", "hashed-password", "Alice", nil, "Admin", "01962b8f-aeb2-7e03-a8ff-1edce1300001", time.Now(), time.Now()),
	)
	repository := NewPostgresUserRepository(mock)

	_, err = repository.FindByID(context.Background(), userID)
	if err == nil {
		t.Fatal("expected error for invalid scanned data")
	}
}
