package usecases

import (
	"context"
	"fmt"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// GetAuditEventUseCase gets a single audit event by identifier.
type GetAuditEventUseCase struct {
	adminEventRepository outboundports.AdminEventRepository
}

// NewGetAuditEventUseCase builds a GetAuditEventUseCase.
func NewGetAuditEventUseCase(adminEventRepository outboundports.AdminEventRepository) *GetAuditEventUseCase {
	return &GetAuditEventUseCase{adminEventRepository: adminEventRepository}
}

// Execute gets a single audit event.
func (uc *GetAuditEventUseCase) Execute(ctx context.Context, query inboundports.GetAuditEventQuery) (outboundports.AuditEventResult, error) {
	eventID, err := valueobjects.EventIDFromString(query.ID)
	if err != nil {
		return outboundports.AuditEventResult{}, fmt.Errorf("parsing event id: %w", err)
	}

	event, err := uc.adminEventRepository.FindByID(ctx, eventID)
	if err != nil {
		return outboundports.AuditEventResult{}, fmt.Errorf("finding audit event by id: %w", err)
	}

	if event == nil {
		return outboundports.AuditEventResult{}, &NotFoundError{Resource: "audit event", ID: query.ID}
	}

	return newAuditEventResult(event), nil
}
