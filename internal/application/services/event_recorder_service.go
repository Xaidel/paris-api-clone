package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	domainevents "github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// RecordAdminEventCommand describes an admin event to persist.
type RecordAdminEventCommand struct {
	ActorUserID  string
	ActorGroupID string
	EventType    string
	EventData    json.RawMessage
}

// EventRecorderService records immutable audit events for internal callers.
type EventRecorderService struct {
	adminEventRepository       outboundports.AdminEventRepository
	adminEventOutboxRepository outboundports.AdminEventOutboxRepository
	actorDirectory             outboundports.ActorDirectory
	now                        func() time.Time
	newEventID                 func() (valueobjects.EventID, error)
}

var _ outboundports.EventPublisher = (*EventRecorderService)(nil)

// NewEventRecorderService builds an EventRecorderService.
func NewEventRecorderService(adminEventRepository outboundports.AdminEventRepository, adminEventOutboxRepository outboundports.AdminEventOutboxRepository, actorDirectory outboundports.ActorDirectory) *EventRecorderService {
	return &EventRecorderService{
		adminEventRepository:       adminEventRepository,
		adminEventOutboxRepository: adminEventOutboxRepository,
		actorDirectory:             actorDirectory,
		now:                        time.Now,
		newEventID:                 valueobjects.NewEventID,
	}
}

// RecordAdminEvent persists an immutable admin audit event.
func (s *EventRecorderService) RecordAdminEvent(ctx context.Context, command RecordAdminEventCommand) error {
	if s.actorDirectory != nil {
		if err := s.actorDirectory.ActorExists(ctx, command.ActorUserID, command.ActorGroupID); err != nil {
			return fmt.Errorf("validating actor ids: %w", err)
		}
	}

	eventID, err := s.newEventID()
	if err != nil {
		return fmt.Errorf("generating event id: %w", err)
	}

	event, err := entities.NewAdminEvent(eventID, s.now(), command.ActorUserID, command.ActorGroupID, command.EventType, command.EventData)
	if err != nil {
		return fmt.Errorf("creating admin event entity: %w", err)
	}

	if err := s.adminEventRepository.Create(ctx, event); err != nil {
		return fmt.Errorf("creating admin event: %w", err)
	}

	if err := s.adminEventOutboxRepository.Create(ctx, event); err != nil {
		return fmt.Errorf("creating admin event outbox record: %w", err)
	}

	return nil
}

// Publish persists supported domain events.
func (s *EventRecorderService) Publish(ctx context.Context, events []domain.DomainEvent) error {
	for _, event := range events {
		adminEvent, ok := event.(*domainevents.AdminActionOccurred)
		if !ok {
			return fmt.Errorf("unsupported domain event type %T", event)
		}

		if err := s.RecordAdminEvent(ctx, RecordAdminEventCommand{
			ActorUserID:  adminEvent.ActorUserID(),
			ActorGroupID: adminEvent.ActorGroupID(),
			EventType:    adminEvent.EventType(),
			EventData:    json.RawMessage(adminEvent.EventData()),
		}); err != nil {
			return fmt.Errorf("recording domain event %s: %w", adminEvent.EventType(), err)
		}
	}

	return nil
}
