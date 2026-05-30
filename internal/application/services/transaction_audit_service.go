package services

import (
	"context"
	"fmt"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/domain/events"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// TransactionAuditService records transaction-related list audit events.
type TransactionAuditService struct {
	publisher ports.EventPublisher
	now       func() time.Time
}

// NewTransactionAuditService builds a TransactionAuditService.
func NewTransactionAuditService(publisher ports.EventPublisher) *TransactionAuditService {
	return &TransactionAuditService{publisher: publisher, now: time.Now}
}

// RecordListTransactions persists a transaction list audit event.
func (s *TransactionAuditService) RecordListTransactions(ctx context.Context, actorUserID, actorGroupID, uploadID string, resultCount int) error {
	if s == nil || s.publisher == nil {
		return nil
	}

	payload := map[string]any{
		"action":       "read",
		"resource":     "transaction",
		"upload_id":    uploadID,
		"result_count": resultCount,
	}

	adminEvent, err := events.NewAdminActionOccurred(s.now(), actorUserID, actorGroupID, events.ListTransactionsEventType, payload)
	if err != nil {
		return fmt.Errorf("creating list transactions audit event: %w", err)
	}

	if err := s.publisher.Publish(ctx, []domain.DomainEvent{adminEvent}); err != nil {
		return fmt.Errorf("publishing list transactions audit event: %w", err)
	}

	return nil
}
