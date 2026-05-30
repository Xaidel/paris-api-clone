package ports

import (
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionStep4Result exposes a persisted transaction step 4 review.
type TransactionStep4Result struct {
	TransactionID     valueobjects.TransactionID
	SectorID          valueobjects.SectorID
	AdditionalContext valueobjects.TransactionStep4AdditionalContext
	IsHighEmitting    bool
	Classification    valueobjects.TransactionClassification
	Status            valueobjects.TransactionStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
