package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// GetSectorUseCase gets a sector entry by identifier.
type GetSectorUseCase struct {
	repository    outboundports.SectorRepository
	eventRecorder adminEventRecorder
	now           func() time.Time
}

// NewGetSectorUseCase builds a GetSectorUseCase.
func NewGetSectorUseCase(repository outboundports.SectorRepository, eventRecorder adminEventRecorder) *GetSectorUseCase {
	return &GetSectorUseCase{repository: repository, eventRecorder: eventRecorder, now: time.Now}
}

// Execute loads a sector entry by identifier.
func (uc *GetSectorUseCase) Execute(ctx context.Context, query inboundports.GetSectorQuery) (outboundports.SectorResult, error) {
	id, err := valueobjects.SectorIDFromString(query.ID)
	if err != nil {
		return outboundports.SectorResult{}, fmt.Errorf("parsing sector id: %w", err)
	}

	sector, err := uc.repository.FindByID(ctx, id)
	if err != nil {
		return outboundports.SectorResult{}, fmt.Errorf("finding sector entry by id: %w", err)
	}

	if sector == nil {
		return outboundports.SectorResult{}, &NotFoundError{Resource: "sector entry", ID: query.ID}
	}

	if err := recordAdminEvent(ctx, uc.eventRecorder, uc.now(), query.ActorUserID, query.ActorGroupID, getSectorAdminEventType, map[string]any{
		"action":      "read",
		"resource":    "sector",
		"target_id":   sector.ID().String(),
		"type":        sector.Type(),
		"name":        sector.Name(),
		"description": sector.Description(),
	}); err != nil {
		return outboundports.SectorResult{}, err
	}

	return newSectorResult(sector), nil
}
