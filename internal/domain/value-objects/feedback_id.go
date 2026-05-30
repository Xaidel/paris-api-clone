package valueobjects

import (
	"github.com/gyud-adb/paris-api/internal/domain"
)

type feedbackIDTag struct{}

// FeedbackID is a value object representing a UUIDv7 feedback identifier.
type FeedbackID = UUIDv7ID[feedbackIDTag]

// NewFeedbackID creates a new UUIDv7 feedback identifier.
func NewFeedbackID() (FeedbackID, error) {
	return newUUIDv7ID[feedbackIDTag]()
}

// FeedbackIDFromString parses and validates a UUIDv7 feedback identifier.
func FeedbackIDFromString(value string) (FeedbackID, error) {
	parsedValue, err := parseUUIDv7ID[feedbackIDTag](value)
	if err != nil {
		return FeedbackID{}, domain.ErrInvalidFeedbackID
	}

	return parsedValue, nil
}
