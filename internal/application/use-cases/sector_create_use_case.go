package usecases

import (
	"context"
	"fmt"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// CreateSectorUseCase creates a new sector entry.
type CreateSectorUseCase struct {
	repository         outboundports.SectorRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	actorDirectory     outboundports.ActorDirectory
	newID              func() (valueobjects.SectorID, error)
	now                func() time.Time
}

// NewCreateSectorUseCase builds a CreateSectorUseCase.
func NewCreateSectorUseCase(repository outboundports.SectorRepository, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *CreateSectorUseCase {
	return &CreateSectorUseCase{
		repository:         repository,
		transactionManager: transactionManager,
		eventRecorder:      eventRecorder,
		actorDirectory:     actorDirectory,
		newID:              valueobjects.NewSectorID,
		now:                time.Now,
	}
}

// Execute creates and persists a sector entry.
func (uc *CreateSectorUseCase) Execute(ctx context.Context, command inboundports.CreateSectorCommand) (outboundports.SectorResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, command.ActorUserID, command.ActorGroupID); err != nil {
		return outboundports.SectorResult{}, err
	}

	createdBy, err := valueobjects.UserIDFromString(command.ActorUserID)
	if err != nil {
		return outboundports.SectorResult{}, fmt.Errorf("parsing actor user id: %w", err)
	}

	id, err := uc.newID()
	if err != nil {
		return outboundports.SectorResult{}, fmt.Errorf("generating sector id: %w", err)
	}

	sector, err := entities.NewSector(id, command.Type, command.Name, command.Description)
	if err != nil {
		return outboundports.SectorResult{}, fmt.Errorf("creating sector entry: %w", err)
	}

	now := uc.now()
	sector.SetCreatedBy(createdBy)
	if err := sector.SetAuditTimestamps(now, now); err != nil {
		return outboundports.SectorResult{}, fmt.Errorf("setting sector audit timestamps: %w", err)
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.repository.Create(txCtx, sector, command.ActorUserID); err != nil {
			return fmt.Errorf("creating sector entry: %w", err)
		}

		if err := recordAdminEvent(txCtx, uc.eventRecorder, now, command.ActorUserID, command.ActorGroupID, createSectorAdminEventType, map[string]any{
			"action":      "create",
			"resource":    "sector",
			"target_id":   sector.ID().String(),
			"type":        sector.Type(),
			"name":        sector.Name(),
			"description": sector.Description(),
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return outboundports.SectorResult{}, fmt.Errorf("creating sector transaction: %w", err)
	}

	return newSectorResult(sector), nil
}
