package usecases

import (
	"context"
	"fmt"
	"slices"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// GetTransactionUploadPreviewUseCase gets one persisted transaction upload preview.
type GetTransactionUploadPreviewUseCase struct {
	uploadRepository  outboundports.TransactionUploadRepository
	previewRepository outboundports.TransactionUploadPreviewRepository
	eventRecorder     adminEventRecorder
	actorDirectory    outboundports.ActorDirectory
	now               func() time.Time
}

// NewGetTransactionUploadPreviewUseCase builds a GetTransactionUploadPreviewUseCase.
func NewGetTransactionUploadPreviewUseCase(uploadRepository outboundports.TransactionUploadRepository, previewRepository outboundports.TransactionUploadPreviewRepository, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *GetTransactionUploadPreviewUseCase {
	return &GetTransactionUploadPreviewUseCase{
		uploadRepository:  uploadRepository,
		previewRepository: previewRepository,
		eventRecorder:     eventRecorder,
		actorDirectory:    actorDirectory,
		now:               time.Now,
	}
}

// Execute loads one transaction upload preview by identifier.
func (uc *GetTransactionUploadPreviewUseCase) Execute(ctx context.Context, query inboundports.GetTransactionUploadPreviewQuery) (inboundports.GetTransactionUploadPreviewResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, query.ActorUserID, query.ActorGroupID); err != nil {
		return inboundports.GetTransactionUploadPreviewResult{}, err
	}

	uploadID, err := valueobjects.UploadIDFromString(query.ID)
	if err != nil {
		return inboundports.GetTransactionUploadPreviewResult{}, fmt.Errorf("parsing upload id: %w", err)
	}

	upload, err := uc.uploadRepository.FindByID(ctx, uploadID)
	if err != nil {
		return inboundports.GetTransactionUploadPreviewResult{}, fmt.Errorf("finding upload by id: %w", err)
	}
	if upload == nil {
		return inboundports.GetTransactionUploadPreviewResult{}, &NotFoundError{Resource: "transaction upload", ID: query.ID}
	}

	if upload.GroupID().String() != query.ActorGroupID {
		return inboundports.GetTransactionUploadPreviewResult{}, &ForbiddenError{Resource: "transaction upload", Reason: "upload belongs to a different group"}
	}

	preview, err := uc.previewRepository.FindByUploadID(ctx, uploadID)
	if err != nil {
		return inboundports.GetTransactionUploadPreviewResult{}, fmt.Errorf("finding upload preview by id: %w", err)
	}
	if preview == nil {
		return inboundports.GetTransactionUploadPreviewResult{}, &NotFoundError{Resource: "transaction upload preview", ID: query.ID}
	}

	if err := upload.RecordPreviewed(uc.now(), query.ActorUserID, query.ActorGroupID); err != nil {
		return inboundports.GetTransactionUploadPreviewResult{}, fmt.Errorf("recording transaction upload preview event: %w", err)
	}

	if err := publishDomainEvents(ctx, uc.eventRecorder, upload.PullDomainEvents()); err != nil {
		return inboundports.GetTransactionUploadPreviewResult{}, fmt.Errorf("publishing transaction upload preview events: %w", err)
	}

	return inboundports.GetTransactionUploadPreviewResult{
		FileID:           upload.ID().String(),
		FileName:         upload.FileName(),
		Columns:          slices.Clone(preview.Columns),
		Rows:             cloneTransactionUploadPreviewRows(preview.Rows),
		TotalRows:        preview.TotalRows,
		ValidationErrors: slices.Clone(preview.ValidationErrors),
	}, nil
}
