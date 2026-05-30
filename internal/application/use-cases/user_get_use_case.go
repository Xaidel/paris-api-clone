package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// GetUserUseCase gets a user by identifier.
type GetUserUseCase struct {
	userRepository outboundports.UserRepository
	eventRecorder  adminEventRecorder
	actorDirectory outboundports.ActorDirectory
	now            func() time.Time
}

// NewGetUserUseCase builds a GetUserUseCase.
func NewGetUserUseCase(userRepository outboundports.UserRepository, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *GetUserUseCase {
	return &GetUserUseCase{userRepository: userRepository, eventRecorder: eventRecorder, actorDirectory: actorDirectory, now: time.Now}
}

// Execute loads a user by identifier.
func (uc *GetUserUseCase) Execute(ctx context.Context, query inboundports.GetUserQuery) (outboundports.UserResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, query.ActorUserID, query.ActorGroupID); err != nil {
		return outboundports.UserResult{}, err
	}

	userID, err := valueobjects.UserIDFromString(query.ID)
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("parsing user id: %w", err)
	}

	user, err := uc.userRepository.FindByID(ctx, userID)
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("finding user by id: %w", err)
	}

	if user == nil {
		return outboundports.UserResult{}, &NotFoundError{Resource: "user", ID: query.ID}
	}

	if err := user.RecordRead(uc.now(), query.ActorUserID, query.ActorGroupID); err != nil {
		return outboundports.UserResult{}, fmt.Errorf("recording user read event: %w", err)
	}

	if err := publishDomainEvents(ctx, uc.eventRecorder, user.PullDomainEvents()); err != nil {
		return outboundports.UserResult{}, fmt.Errorf("publishing user events: %w", err)
	}

	return newUserResult(user), nil
}
