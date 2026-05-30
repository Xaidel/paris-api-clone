package valueobjects

import "testing"

// TestFeedbackKindFromStringAcceptsValidValues verifies the feedback kind from string accepts valid values behavior and the expected outcome asserted below.
func TestFeedbackKindFromStringAcceptsValidValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input string
		want  string
	}{
		{"thumbs_up", "thumbs_up"},
		{"thumbs_down", "thumbs_down"},
		{"THUMBS_UP", "thumbs_up"},
		{"THUMBS_DOWN", "thumbs_down"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			kind, err := FeedbackKindFromString(tc.input)
			if err != nil {
				t.Fatalf("FeedbackKindFromString(%q) error = %v", tc.input, err)
			}

			if kind.String() != tc.want {
				t.Fatalf("kind.String() = %q, want %q", kind.String(), tc.want)
			}
		})
	}
}

// TestFeedbackKindFromStringRejectsInvalidValue verifies the feedback kind from string rejects invalid value behavior and the expected outcome asserted below.
func TestFeedbackKindFromStringRejectsInvalidValue(t *testing.T) {
	t.Parallel()

	if _, err := FeedbackKindFromString("neutral"); err == nil {
		t.Fatal("expected error for invalid feedback kind")
	}
}

// TestFeedbackKindEqual verifies the feedback kind equal behavior and the expected outcome asserted below.
func TestFeedbackKindEqual(t *testing.T) {
	t.Parallel()

	up := ThumbsUpFeedbackKind()
	down := ThumbsDownFeedbackKind()

	if !up.Equal(ThumbsUpFeedbackKind()) {
		t.Fatal("expected thumbs_up to equal thumbs_up")
	}

	if up.Equal(down) {
		t.Fatal("expected thumbs_up to not equal thumbs_down")
	}
}
