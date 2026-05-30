package usecases

import (
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func newFeedbackResult(f *entities.Feedback) ports.FeedbackResult {
	return ports.FeedbackResult{
		ID:            f.ID().String(),
		UserID:        f.UserID(),
		TransactionID: f.TransactionID(),
		Kind:          f.Kind(),
		CreatedAt:     f.CreatedAt(),
		UpdatedAt:     f.UpdatedAt(),
	}
}
