package usecases

import (
	"context"
	"fmt"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// ListAuditEventsUseCase lists audit events.
type ListAuditEventsUseCase struct {
	adminEventRepository outboundports.AdminEventRepository
}

// NewListAuditEventsUseCase builds a ListAuditEventsUseCase.
func NewListAuditEventsUseCase(adminEventRepository outboundports.AdminEventRepository) *ListAuditEventsUseCase {
	return &ListAuditEventsUseCase{adminEventRepository: adminEventRepository}
}

// Execute lists audit events using the supplied filters.
func (uc *ListAuditEventsUseCase) Execute(ctx context.Context, query inboundports.ListAuditEventsQuery) (outboundports.ListAuditEventsResult, error) {
	events, err := uc.adminEventRepository.List(ctx, outboundports.AuditEventFilter(query))
	if err != nil {
		return outboundports.ListAuditEventsResult{}, fmt.Errorf("listing audit events: %w", err)
	}

	results := make([]outboundports.AuditEventResult, 0, len(events))
	for _, event := range events {
		results = append(results, newAuditEventResult(event))
	}

	return outboundports.ListAuditEventsResult{Events: results}, nil
}
