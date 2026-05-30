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

// CreateUserUseCase creates a new user.
type CreateUserUseCase struct {
	userRepository     outboundports.UserRepository
	groupRepository    outboundports.GroupRepository
	passwordHasher     outboundports.PasswordHasher
	eventPublisher     outboundports.EventPublisher
	transactionManager outboundports.TransactionManager
	actorDirectory     outboundports.ActorDirectory
	now                func() time.Time
}

// NewCreateUserUseCase builds a CreateUserUseCase.
func NewCreateUserUseCase(userRepository outboundports.UserRepository, groupRepository outboundports.GroupRepository, passwordHasher outboundports.PasswordHasher, eventPublisher outboundports.EventPublisher, transactionManager outboundports.TransactionManager, actorDirectory outboundports.ActorDirectory) *CreateUserUseCase {
	return &CreateUserUseCase{
		userRepository:     userRepository,
		groupRepository:    groupRepository,
		passwordHasher:     passwordHasher,
		eventPublisher:     eventPublisher,
		transactionManager: transactionManager,
		actorDirectory:     actorDirectory,
		now:                time.Now,
	}
}

// Execute creates and persists a user.
func (uc *CreateUserUseCase) Execute(ctx context.Context, command inboundports.CreateUserCommand) (outboundports.UserResult, error) {
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

	group, err := uc.groupRepository.FindByID(ctx, groupID)
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("finding group by id: %w", err)
	}

	if group == nil {
		return outboundports.UserResult{}, &NotFoundError{Resource: "group", ID: command.GroupID}
	}

	userID, err := valueobjects.NewUserID()
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("generating user id: %w", err)
	}

	passwordHash, err := uc.passwordHasher.Hash(ctx, password.String())
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("hashing password: %w", err)
	}

	profile, err := valueobjects.NewUserProfile(command.FirstName, command.MiddleName, command.LastName, group.ID())
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("creating user profile: %w", err)
	}

	user, err := entities.NewUser(userID, command.Username, passwordHash, profile, uc.now())
	if err != nil {
		return outboundports.UserResult{}, fmt.Errorf("creating user entity: %w", err)
	}

	if err := user.RecordCreated(command.ActorUserID, command.ActorGroupID); err != nil {
		return outboundports.UserResult{}, fmt.Errorf("recording user creation event: %w", err)
	}

	if err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := uc.userRepository.Create(txCtx, user); err != nil {
			return fmt.Errorf("creating user: %w", err)
		}

		if err := publishDomainEvents(txCtx, uc.eventPublisher, user.PullDomainEvents()); err != nil {
			return fmt.Errorf("publishing user events: %w", err)
		}

		return nil
	}); err != nil {
		return outboundports.UserResult{}, fmt.Errorf("creating user transaction: %w", err)
	}

	return newUserResult(user), nil
}
