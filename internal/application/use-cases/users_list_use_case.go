package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain/events"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// ListUsersUseCase lists users.
type ListUsersUseCase struct {
	userRepository outboundports.UserRepository
	eventRecorder  adminEventRecorder
	actorDirectory outboundports.ActorDirectory
	now            func() time.Time
}

// NewListUsersUseCase builds a ListUsersUseCase.
func NewListUsersUseCase(userRepository outboundports.UserRepository, eventRecorder adminEventRecorder, actorDirectory outboundports.ActorDirectory) *ListUsersUseCase {
	return &ListUsersUseCase{userRepository: userRepository, eventRecorder: eventRecorder, actorDirectory: actorDirectory, now: time.Now}
}

// Execute lists all users.
func (uc *ListUsersUseCase) Execute(ctx context.Context, query inboundports.ListUsersQuery) (inboundports.ListUsersResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, query.ActorUserID, query.ActorGroupID); err != nil {
		return inboundports.ListUsersResult{}, err
	}

	users, err := uc.userRepository.List(ctx)
	if err != nil {
		return inboundports.ListUsersResult{}, fmt.Errorf("listing users: %w", err)
	}

	results := make([]outboundports.UserResult, 0, len(users))
	for _, user := range users {
		results = append(results, newUserResult(user))
	}

	if err := recordAdminEvent(ctx, uc.eventRecorder, uc.now(), query.ActorUserID, query.ActorGroupID, events.ListUsersEventType, map[string]any{
		"action":       "list",
		"resource":     "user",
		"result_count": len(results),
	}); err != nil {
		return inboundports.ListUsersResult{}, err
	}

	return inboundports.ListUsersResult{Users: results}, nil
}
