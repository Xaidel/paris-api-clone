package entities

import (
	"testing"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TestNewU1ListEntry verifies the new U1 list entry behavior and the expected outcome asserted below.
func TestNewU1ListEntry(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.NewU1ListID()
	if err != nil {
		t.Fatalf("NewU1ListID() error = %v", err)
	}

	entry, err := NewU1ListEntry(id, "  energy ", "  grant ", "  must satisfy rule 1 ")
	if err != nil {
		t.Fatalf("NewU1ListEntry() error = %v", err)
	}

	if entry.Sector() != "energy" {
		t.Fatalf("entry.Sector() = %q, want %q", entry.Sector(), "energy")
	}

	if entry.EligibleOperationType() != "grant" {
		t.Fatalf("entry.EligibleOperationType() = %q, want %q", entry.EligibleOperationType(), "grant")
	}

	if entry.ConditionGuidance() != "must satisfy rule 1" {
		t.Fatalf("entry.ConditionGuidance() = %q, want %q", entry.ConditionGuidance(), "must satisfy rule 1")
	}
}

// TestU1ListEntryUpdate verifies the U1 list entry update behavior and the expected outcome asserted below.
func TestU1ListEntryUpdate(t *testing.T) {
	t.Parallel()

	id, err := valueobjects.NewU1ListID()
	if err != nil {
		t.Fatalf("NewU1ListID() error = %v", err)
	}

	entry := ReconstituteU1ListEntry(id, "energy", "grant", "rule 1")
	if err := entry.Update(" agriculture ", " loan ", " rule 2 "); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if entry.Sector() != "agriculture" {
		t.Fatalf("entry.Sector() = %q, want %q", entry.Sector(), "agriculture")
	}

	if entry.EligibleOperationType() != "loan" {
		t.Fatalf("entry.EligibleOperationType() = %q, want %q", entry.EligibleOperationType(), "loan")
	}

	if entry.ConditionGuidance() != "rule 2" {
		t.Fatalf("entry.ConditionGuidance() = %q, want %q", entry.ConditionGuidance(), "rule 2")
	}
}
