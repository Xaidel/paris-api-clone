package valueobjects

import (
	"errors"
	"testing"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// TestNewU1ClassificationListEntryText verifies the new U1 classification list entry text behavior and the expected outcome asserted below.
func TestNewU1ClassificationListEntryText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		sector                string
		eligibleOperationType string
		conditionGuidance     string
		want                  string
		wantErr               error
	}{
		{
			name:                  "normalizes fields",
			sector:                " energy ",
			eligibleOperationType: " loan ",
			conditionGuidance:     " solar ",
			want:                  "sector: energy; eligible_operation_type: loan; condition_guidance: solar",
		},
		{
			name:                  "rejects empty sector",
			sector:                " ",
			eligibleOperationType: "loan",
			conditionGuidance:     "solar",
			wantErr:               domain.ErrInvalidSector,
		},
		{
			name:                  "rejects empty eligible operation type",
			sector:                "energy",
			eligibleOperationType: " ",
			conditionGuidance:     "solar",
			wantErr:               domain.ErrInvalidEligibleOperationType,
		},
		{
			name:                  "rejects empty condition guidance",
			sector:                "energy",
			eligibleOperationType: "loan",
			conditionGuidance:     " ",
			wantErr:               domain.ErrInvalidConditionGuidance,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewU1ClassificationListEntryText(tc.sector, tc.eligibleOperationType, tc.conditionGuidance)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("NewU1ClassificationListEntryText() error = %v, want %v", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewU1ClassificationListEntryText() error = %v", err)
			}

			if got.String() != tc.want {
				t.Fatalf("NewU1ClassificationListEntryText().String() = %q, want %q", got.String(), tc.want)
			}
		})
	}
}

// TestNewU2ClassificationListEntryText verifies the new u 2 classification list entry text behavior and the expected outcome asserted below.
func TestNewU2ClassificationListEntryText(t *testing.T) {
	t.Parallel()

	got, err := NewU2ClassificationListEntryText(" coal ")
	if err != nil {
		t.Fatalf("NewU2ClassificationListEntryText() error = %v", err)
	}

	if got.String() != "activity_type: coal" {
		t.Fatalf("NewU2ClassificationListEntryText().String() = %q, want %q", got.String(), "activity_type: coal")
	}
}

// TestNewSectorClassificationListEntryText verifies the new sector classification list entry text behavior and the expected outcome asserted below.
func TestNewSectorClassificationListEntryText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		sectorType  string
		nameValue   string
		description string
		want        string
		wantErr     error
	}{
		{
			name:        "normalizes fields",
			sectorType:  " pa aligned ",
			nameValue:   " steel ",
			description: " green steel ",
			want:        "type: PA Aligned; name: steel; description: green steel",
		},
		{
			name:        "rejects invalid sector type",
			sectorType:  "unknown",
			nameValue:   "steel",
			description: "green steel",
			wantErr:     domain.ErrInvalidSectorType,
		},
		{
			name:        "rejects empty name",
			sectorType:  "PA Aligned",
			nameValue:   " ",
			description: "green steel",
			wantErr:     domain.ErrInvalidSectorName,
		},
		{
			name:        "rejects empty description",
			sectorType:  "PA Aligned",
			nameValue:   "steel",
			description: " ",
			wantErr:     domain.ErrInvalidSectorDescription,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewSectorClassificationListEntryText(tc.sectorType, tc.nameValue, tc.description)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("NewSectorClassificationListEntryText() error = %v, want %v", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewSectorClassificationListEntryText() error = %v", err)
			}

			if got.String() != tc.want {
				t.Fatalf("NewSectorClassificationListEntryText().String() = %q, want %q", got.String(), tc.want)
			}
		})
	}
}
