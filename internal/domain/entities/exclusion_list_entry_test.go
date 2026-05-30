package entities

import (
	"testing"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TestNewExclusionListEntry verifies the new exclusion list entry behavior and the expected outcome asserted below.
func TestNewExclusionListEntry(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.NewExclusionListID()
	if err != nil {
		t.Fatalf("NewExclusionListID() error = %v", err)
	}

	entry, err := NewExclusionListEntry(id, "  agriculture ")
	if err != nil {
		t.Fatalf("NewExclusionListEntry() error = %v", err)
	}

	if entry.ActivityType() != "agriculture" {
		t.Fatalf("entry.ActivityType() = %q, want %q", entry.ActivityType(), "agriculture")
	}
}

// TestExclusionListEntryUpdateActivityType verifies the exclusion list entry update activity type behavior and the expected outcome asserted below.
func TestExclusionListEntryUpdateActivityType(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.NewExclusionListID()
	if err != nil {
		t.Fatalf("NewExclusionListID() error = %v", err)
	}

	entry := ReconstituteExclusionListEntry(id, "agriculture")
	if err := entry.UpdateActivityType(" energy "); err != nil {
		t.Fatalf("UpdateActivityType() error = %v", err)
	}

	if entry.ActivityType() != "energy" {
		t.Fatalf("entry.ActivityType() = %q, want %q", entry.ActivityType(), "energy")
	}
}
