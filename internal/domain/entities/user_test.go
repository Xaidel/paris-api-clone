package entities

import (
	"errors"
	"testing"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TestNewUser verifies the new user behavior and the expected outcome asserted below.
func TestNewUser(t *testing.T) {
	t.Parallel()

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
	user, err := NewUser(userID, "alice", "$2a$10$abcdefghijklmnopqrstuuuuuuuuuuuuuuuuuuuuuuuuuu", profile, now)
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	if user.Username() != "alice" {
		t.Fatalf("Username() = %q, want %q", user.Username(), "alice")
	}

	if !user.CreatedAt().Equal(now) {
		t.Fatalf("CreatedAt() = %v, want %v", user.CreatedAt(), now)
	}

	if !user.UpdatedAt().Equal(now) {
		t.Fatalf("UpdatedAt() = %v, want %v", user.UpdatedAt(), now)
	}
}

// TestNewUserRequiresTimestamp verifies the new user requires timestamp behavior and the expected outcome asserted below.
func TestNewUserRequiresTimestamp(t *testing.T) {
	t.Parallel()

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

	_, err = NewUser(userID, "alice", "$2a$10$abcdefghijklmnopqrstuuuuuuuuuuuuuuuuuuuuuuuuuu", profile, time.Time{})
	if !errors.Is(err, domain.ErrInvalidTimestamp) {
		t.Fatalf("NewUser() error = %v, want %v", err, domain.ErrInvalidTimestamp)
	}
}

// TestNewUserRequiresUsernameAndPasswordHash verifies the new user requires username and password hash behavior and the expected outcome asserted below.
func TestNewUserRequiresUsernameAndPasswordHash(t *testing.T) {
	t.Parallel()

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

	_, err = NewUser(userID, "", "hash", profile, testTime())
	if !errors.Is(err, domain.ErrInvalidUsername) {
		t.Fatalf("NewUser() error = %v, want %v", err, domain.ErrInvalidUsername)
	}

	_, err = NewUser(userID, "alice", "", profile, testTime())
	if !errors.Is(err, domain.ErrInvalidPasswordHash) {
		t.Fatalf("NewUser() error = %v, want %v", err, domain.ErrInvalidPasswordHash)
	}
}

// TestReconstituteUser verifies the reconstitute user behavior and the expected outcome asserted below.
func TestReconstituteUser(t *testing.T) {
	t.Parallel()

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
	user := ReconstituteUser(userID, "alice", "hash", profile, now, now)

	if !user.Equal(ReconstituteUser(userID, "bob", "other-hash", profile, now, now)) {
		t.Fatal("Equal() should compare identities")
	}
}

// TestUserTouch verifies the user touch behavior and the expected outcome asserted below.
func TestUserTouch(t *testing.T) {
	t.Parallel()

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

	createdAt := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	user := ReconstituteUser(userID, "alice", "hash", profile, createdAt, createdAt)

	if err := user.Touch(updatedAt); err != nil {
		t.Fatalf("Touch() error = %v", err)
	}

	if !user.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("UpdatedAt() = %v, want %v", user.UpdatedAt(), updatedAt)
	}

	if err := user.Touch(time.Time{}); !errors.Is(err, domain.ErrInvalidTimestamp) {
		t.Fatalf("Touch() error = %v, want %v", err, domain.ErrInvalidTimestamp)
	}
}

// TestUserUpdate verifies the user update behavior and the expected outcome asserted below.
func TestUserUpdate(t *testing.T) {
	t.Parallel()

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

	newGroupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300002")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	updatedProfile, err := valueobjects.NewUserProfile("Bob", nil, "Builder", newGroupID)
	if err != nil {
		t.Fatalf("NewUserProfile() error = %v", err)
	}

	createdAt := time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	user := ReconstituteUser(userID, "alice", "old-hash", profile, createdAt, createdAt)

	if err := user.Update("bob", "new-hash", updatedProfile, updatedAt); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if user.Username() != "bob" {
		t.Fatalf("Username() = %q, want %q", user.Username(), "bob")
	}

	if user.PasswordHash() != "new-hash" {
		t.Fatalf("PasswordHash() = %q, want %q", user.PasswordHash(), "new-hash")
	}

	if user.Profile().FirstName() != "Bob" {
		t.Fatalf("Profile().FirstName() = %q, want %q", user.Profile().FirstName(), "Bob")
	}
}

func testTime() time.Time {
	return time.Date(2026, time.April, 2, 12, 0, 0, 0, time.UTC)
}
