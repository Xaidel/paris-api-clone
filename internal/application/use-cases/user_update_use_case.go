package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// UpdateUserUseCase updates a user.
type UpdateUserUseCase struct {
	userRepository     outboundports.UserRepository
	groupRepository    outboundports.GroupRepository
	passwordHasher     outboundports.PasswordHasher
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	actorDirectory     outboundports.ActorDirectory
	now                func() time.Time
}

// NewUpdateUserUseCase builds an UpdateUserUseCase.
func NewUpdateUserUseCase(userRepository outboundports.UserRepository, groupRepository outboundports.GroupRepository, passwordHasher outboundports.PasswordHasher, transactionManager outboundports.TransactionManager, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *UpdateUserUseCase {
	return &UpdateUserUseCase{userRepository: userRepository, groupRepository: groupRepository, passwordHasher: passwordHasher, transactionManager: transactionManager, eventRecorder: eventRecorder, actorDirectory: actorDirectory, now: time.Now}
}

// Execute updates an existing user.
func (uc *UpdateUserUseCase) Execute(ctx context.Context, command inboundports.UpdateUserCommand) (outboundports.UserResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, command.ActorUserID, command.ActorGroupID); err != nil {
		return outboundports.UserResult{}, err
	}

	password, err := valueobjects.NewPlaintextPassword(command.Password)
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("validating password: %w", err)
	}

	groupID, err := valueobjects.GroupIDFromString(command.GroupID)
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("parsing group id: %w", err)
	}

	userID, err := valueobjects.UserIDFromString(command.ID)
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("parsing user id: %w", err)
	}

	user, err := uc.userRepository.FindByID(ctx, userID)
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("finding user by id: %w", err)
	}

	if user == nil {
		return outboundports.UserResult{}, &NotFoundError{Resource: "user", ID: command.ID}
	}

	group, err := uc.groupRepository.FindByID(ctx, groupID)
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("finding group by id: %w", err)
	}

	if group == nil {
		return outboundports.UserResult{}, &NotFoundError{Resource: "group", ID: command.GroupID}
	}

	passwordHash, err := uc.passwordHasher.Hash(ctx, password.String())
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("hashing password: %w", err)
	}

	profile, err := valueobjects.NewUserProfile(command.FirstName, command.MiddleName, command.LastName, group.ID())
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("creating user profile: %w", err)
	}

	if err := user.Update(command.Username, passwordHash, profile, uc.now()); err != nil {
		return outboundports.UserResult{}, fmt.Errorf("updating user: %w", err)
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.userRepository.Update(txCtx, user); err != nil {
			return fmt.Errorf("updating user: %w", err)
		}

		if err := user.RecordUpdated(command.ActorUserID, command.ActorGroupID); err != nil {
			return fmt.Errorf("recording user update event: %w", err)
		}

		if err := publishDomainEvents(txCtx, uc.eventRecorder, user.PullDomainEvents()); err != nil {
			return fmt.Errorf("publishing user events: %w", err)
		}

		return nil
	}); err != nil {
		return outboundports.UserResult{}, fmt.Errorf("updating user transaction: %w", err)
	}

	return newUserResult(user), nil
}
