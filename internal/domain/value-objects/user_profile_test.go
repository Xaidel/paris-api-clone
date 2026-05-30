package valueobjects

import "testing"

// TestNewUserProfile verifies the new user profile behavior and the expected outcome asserted below.
func TestNewUserProfile(t *testing.T) {
	t.Parallel()

	groupID, err := GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	middleName := "Middle"
	profile, err := NewUserProfile("First", &middleName, "Last", groupID)
	if err != nil {
		t.Fatalf("NewUserProfile() error = %v", err)
	}

	if profile.FirstName() != "First" {
		t.Fatalf("FirstName() = %q, want %q", profile.FirstName(), "First")
	}

	if profile.MiddleName() == nil || *profile.MiddleName() != "Middle" {
		t.Fatalf("MiddleName() = %v, want %q", profile.MiddleName(), "Middle")
	}

	if profile.LastName() != "Last" {
		t.Fatalf("LastName() = %q, want %q", profile.LastName(), "Last")
	}
}
