package ports

import (
	"context"
	"time"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// ListTransactionUploadsQuery requests upload history.
type ListTransactionUploadsQuery struct {
	FileName     string
	StartedAt    *time.Time
	EndedAt      *time.Time
	ActorUserID  string
	ActorGroupID string
}

// ListTransactionUploadsResult returns upload history with accepted transactions.
type ListTransactionUploadsResult struct {
	Uploads []outboundports.TransactionUploadDetailsResult
}

// ListTransactionUploadsPort lists upload history.
type ListTransactionUploadsPort interface {
	Execute(ctx context.Context, query ListTransactionUploadsQuery) (ListTransactionUploadsResult, error)
}
