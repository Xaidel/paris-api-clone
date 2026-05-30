package entities

import (
	"testing"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TestNewGroup verifies the new group behavior and the expected outcome asserted below.
func TestNewGroup(t *testing.T) {
	t.Parallel()

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	group, err := NewGroup(groupID, "superadmin")
	if err != nil {
		t.Fatalf("NewGroup() error = %v", err)
	}

	if group.Name() != "superadmin" {
		t.Fatalf("Name() = %q, want %q", group.Name(), "superadmin")
	}
}

// TestGroupRecordCreated verifies the group record created behavior and the expected outcome asserted below.
func TestGroupRecordCreated(t *testing.T) {
	t.Parallel()

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	group := ReconstituteGroup(groupID, "superadmin")
	if err := group.RecordCreated(time.Date(2026, time.April, 10, 12, 0, 0, 0, time.UTC), "01962b8f-aeb2-7e03-a8ff-1edce1300002", "01962b8f-aeb2-7e03-a8ff-1edce1300001"); err != nil {
		t.Fatalf("RecordCreated() error = %v", err)
	}

	if len(group.PullDomainEvents()) != 1 {
		t.Fatal("expected one domain event")
	}
}
