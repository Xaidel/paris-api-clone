package ports

import (
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// TransactionStep5Result exposes a computed or persisted transaction step 5 screening result.
type TransactionStep5Result struct {
	TransactionID                   valueobjects.TransactionID
	ScreeningQuestion1Answer        bool
	ScreeningQuestion1Justification valueobjects.TransactionStep5ScreeningQuestion1Justification
	ScreeningQuestion2Answer        bool
	ScreeningQuestion2Justification valueobjects.TransactionStep5ScreeningQuestion2Justification
	ReviewerNotes                   valueobjects.TransactionStep5ReviewerNotes
	IsFinal                         bool
	Classification                  valueobjects.TransactionClassification
	Detail                          string
	CreatedAt                       *time.Time
	UpdatedAt                       *time.Time
}
