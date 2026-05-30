package main

import (
	"strings"
	"testing"
)

func TestParseAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{name: "defaults to up", args: nil, want: migrationActionUp},
		{name: "down action", args: []string{"down"}, want: migrationActionDown},
		{name: "seed reference data action", args: []string{"seed-reference-data"}, want: migrationActionSeedReferenceData},
		{name: "seed reference data action normalized", args: []string{" SeEd-ReFeReNcE-Data "}, want: migrationActionSeedReferenceData},
		{name: "version action with spaces", args: []string{" version "}, want: migrationActionVersion},
		{name: "invalid action", args: []string{"reset"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseAction(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("parseAction() error = nil, want error")
				}

				return
			}

			if err != nil {
				t.Fatalf("parseAction() error = %v", err)
			}

			if got != tt.want {
				t.Fatalf("parseAction() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseActionInvalidActionMessageMentionsSeedReferenceData(t *testing.T) {
	t.Parallel()

	_, err := parseAction([]string{"reset"})
	if err == nil {
		t.Fatal("parseAction() error = nil, want error")
	}

	if !strings.Contains(err.Error(), migrationActionSeedReferenceData) {
		t.Fatalf("parseAction() error = %q, want message to mention %q", err.Error(), migrationActionSeedReferenceData)
	}
}
