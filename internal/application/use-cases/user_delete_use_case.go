package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// DeleteUserUseCase deletes a user.
type DeleteUserUseCase struct {
	userRepository outboundports.UserRepository
	eventRecorder  adminEventRecorder
	actorDirectory outboundports.ActorDirectory
	now            func() time.Time
}

// NewDeleteUserUseCase builds a DeleteUserUseCase.
func NewDeleteUserUseCase(userRepository outboundports.UserRepository, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *DeleteUserUseCase {
	return &DeleteUserUseCase{userRepository: userRepository, eventRecorder: eventRecorder, actorDirectory: actorDirectory, now: time.Now}
}

// Execute deletes an existing user.
func (uc *DeleteUserUseCase) Execute(ctx context.Context, command inboundports.DeleteUserCommand) (inboundports.DeleteUserResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, command.ActorUserID, command.ActorGroupID); err != nil {
		return inboundports.DeleteUserResult{}, err
	}

	userID, err := valueobjects.UserIDFromString(command.ID)
	if err != nil {
		return inboundports.DeleteUserResult{}, fmt.Errorf("parsing user id: %w", err)
	}

	user, err := uc.userRepository.FindByID(ctx, userID)
	if err != nil {
		return inboundports.DeleteUserResult{}, fmt.Errorf("finding user by id: %w", err)
	}

	if user == nil {
		return inboundports.DeleteUserResult{}, &NotFoundError{Resource: "user", ID: command.ID}
	}

	if err := uc.userRepository.DeleteByID(ctx, userID); err != nil {
		return inboundports.DeleteUserResult{}, fmt.Errorf("deleting user: %w", err)
	}

	if err := user.RecordDeleted(uc.now(), command.ActorUserID, command.ActorGroupID); err != nil {
		return inboundports.DeleteUserResult{}, fmt.Errorf("recording user deletion event: %w", err)
	}

	if err := publishDomainEvents(ctx, uc.eventRecorder, user.PullDomainEvents()); err != nil {
		return inboundports.DeleteUserResult{}, fmt.Errorf("publishing user events: %w", err)
	}

	return inboundports.DeleteUserResult{ID: command.ID}, nil
}
