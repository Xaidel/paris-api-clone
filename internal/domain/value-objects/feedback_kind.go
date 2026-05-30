package valueobjects

import (
	"strings"

	"github.com/gyud-adb/paris-api/internal/domain"
)

// FeedbackKindValue is the enum backing a FeedbackKind value object.
type FeedbackKindValue int

const (
	feedbackKindUnknownValue FeedbackKindValue = iota
	FeedbackKindThumbsUpValue
	FeedbackKindThumbsDownValue
)

var feedbackKindValueStrings = map[FeedbackKindValue]string{
	FeedbackKindThumbsUpValue:   "thumbs_up",
	FeedbackKindThumbsDownValue: "thumbs_down",
}

var feedbackKindStringValues = map[string]FeedbackKindValue{
	"thumbs_up":   FeedbackKindThumbsUpValue,
	"thumbs_down": FeedbackKindThumbsDownValue,
}

// String returns the canonical feedback kind value.
func (v FeedbackKindValue) String() string {
	return feedbackKindValueStrings[v]
}

// FeedbackKind describes whether a user gave a thumbs up or thumbs down.
type FeedbackKind struct {
	value FeedbackKindValue
}

// ThumbsUpFeedbackKind returns the thumbs_up feedback kind.
func ThumbsUpFeedbackKind() FeedbackKind {
	return FeedbackKind{value: FeedbackKindThumbsUpValue}
}

// ThumbsDownFeedbackKind returns the thumbs_down feedback kind.
func ThumbsDownFeedbackKind() FeedbackKind {
	return FeedbackKind{value: FeedbackKindThumbsDownValue}
}

// FeedbackKindFromString parses and normalizes a supported feedback kind.
func FeedbackKindFromString(value string) (FeedbackKind, error) {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	kindValue, ok := feedbackKindStringValues[normalizedValue]
	if !ok {
		return FeedbackKind{}, domain.ErrInvalidFeedbackKind
	}

	return FeedbackKind{value: kindValue}, nil
}

// String returns the canonical feedback kind value.
func (k FeedbackKind) String() string {
	return k.value.String()
}

// Equal reports whether two feedback kinds are equal.
func (k FeedbackKind) Equal(other FeedbackKind) bool {
	return k.value == other.value
}
