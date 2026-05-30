package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// GetU1ListUseCase gets a U1 list entry by identifier.
type GetU1ListUseCase struct {
	repository    outboundports.U1ListRepository
	eventRecorder adminEventRecorder
	now           func() time.Time
}

// NewGetU1ListUseCase builds a GetU1ListUseCase.
func NewGetU1ListUseCase(repository outboundports.U1ListRepository, eventRecorder adminEventRecorder) *GetU1ListUseCase {
	return &GetU1ListUseCase{repository: repository, eventRecorder: eventRecorder, now: time.Now}
}

// Execute loads a U1 list entry by identifier.
func (uc *GetU1ListUseCase) Execute(ctx context.Context, query inboundports.GetU1ListQuery) (outboundports.U1ListResult, error) {
	id, err := valueobjects.U1ListIDFromString(query.ID)
	if err != nil {
		return outboundports.U1ListResult{}, fmt.Errorf("parsing u1 list id: %w", err)
	}

	entry, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return outboundports.U1ListResult{}, fmt.Errorf("finding u1 list entry by id: %w", err)
	}

	if entry == nil {
		return outboundports.U1ListResult{}, &NotFoundError{Resource: "u1 list entry", ID: query.ID}
	}

	if err := recordAdminEvent(ctx, uc.eventRecorder, uc.now(), query.ActorUserID, query.ActorGroupID, getU1ListAdminEventType, map[string]any{
		"action":                  "read",
		"resource":                "u1_list",
		"target_id":               entry.ID().String(),
		"sector":                  entry.Sector(),
		"eligible_operation_type": entry.EligibleOperationType(),
		"condition_guidance":      entry.ConditionGuidance(),
	}); err != nil {
		return outboundports.U1ListResult{}, err
	}

	return newU1ListResult(entry), nil
}
