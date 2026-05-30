package entities

import (
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
)

func validateTimestamp(value time.Time) error {
	if value.IsZero() {
		return domain.ErrInvalidTimestamp
	}

	return nil
}
