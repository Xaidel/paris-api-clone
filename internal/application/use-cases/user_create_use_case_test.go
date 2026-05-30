package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	services "github.com/gyud-adb/paris-api/internal/application/services"
	"github.com/gyud-adb/paris-api/internal/domain"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	domainevents "github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

type passwordHasherMock struct {
	hashResult string
	hashErr    error
	password   string
}

type groupRepositoryMock struct {
	createdGroup *entities.Group
	createErr    error
	findByID     *entities.Group
	findByIDErr  error
	listGroups   []*entities.Group
	listErr      error
	updatedGroup *entities.Group
	updateErr    error
	deletedID    string
	deleteErr    error
}

type actorDirectoryMock struct {
	err     error
	userID  string
	groupID string
}

func (m *passwordHasherMock) Hash(_ context.Context, password string) (string, error) {
	m.password = password
	return m.hashResult, m.hashErr
}

func (m *actorDirectoryMock) ActorExists(_ context.Context, userID string, groupID string) error {
	m.userID = userID
	m.groupID = groupID
	return m.err
}

type userRepositoryMock struct {
	createdUser *entities.User
	createErr   error
	findByID    *entities.User
	findByIDErr error
	listUsers   []*entities.User
	listErr     error
	updatedUser *entities.User
	updateErr   error
	deletedID   string
	deleteErr   error
}

type transactionManagerMock struct {
	err         error
	invoked     bool
	contextSeen context.Context
}

func (m *transactionManagerMock) WithinTransaction(ctx context.Context, operation func(ctx context.Context) error) error {
	m.invoked = true
	m.contextSeen = ctx
	if m.err != nil {
		return m.err
	}

	return operation(ctx)
}

func (m *groupRepositoryMock) Create(_ context.Context, group *entities.Group) error {
	m.createdGroup = group
	return m.createErr
}

func (m *groupRepositoryMock) FindByID(context.Context, valueobjects.GroupID) (*entities.Group, error) {
	return m.findByID, m.findByIDErr
}

func (m *groupRepositoryMock) List(context.Context) ([]*entities.Group, error) {
	return m.listGroups, m.listErr
}

func (m *groupRepositoryMock) Update(_ context.Context, group *entities.Group) error {
	m.updatedGroup = group
	return m.updateErr
}

func (m *groupRepositoryMock) DeleteByID(_ context.Context, id valueobjects.GroupID) error {
	m.deletedID = id.String()
	return m.deleteErr
}

type adminEventRecorderMock struct {
	command services.RecordAdminEventCommand
	err     error
	when    time.Time
}

func (m *adminEventRecorderMock) Publish(_ context.Context, recordedEvents []domain.DomainEvent) error {
	if len(recordedEvents) > 0 {
		adminEvent, ok := recordedEvents[0].(*domainevents.AdminActionOccurred)
		if !ok {
			return errors.New("unexpected domain event type")
		}

		m.command = services.RecordAdminEventCommand{
			ActorUserID:  adminEvent.ActorUserID(),
			ActorGroupID: adminEvent.ActorGroupID(),
			EventType:    adminEvent.EventType(),
			EventData:    json.RawMessage(adminEvent.EventData()),
		}
		m.when = adminEvent.OccurredAt()
	}

	return m.err
}

func (m *userRepositoryMock) Create(_ context.Context, user *entities.User) error {
	m.createdUser = user
	return m.createErr
}

func (m *userRepositoryMock) FindByID(context.Context, valueobjects.UserID) (*entities.User, error) {
	return m.findByID, m.findByIDErr
}

func (m *userRepositoryMock) List(context.Context) ([]*entities.User, error) {
	return m.listUsers, m.listErr
}

func (m *userRepositoryMock) Update(_ context.Context, user *entities.User) error {
	m.updatedUser = user
	return m.updateErr
}

func (m *userRepositoryMock) DeleteByID(_ context.Context, id valueobjects.UserID) error {
	m.deletedID = id.String()
	return m.deleteErr
}

// TestCreateUserUseCaseExecute verifies the create user use case execute behavior and the expected outcome asserted below.
func TestCreateUserUseCaseExecute(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		repository  *userRepositoryMock
		groups      *groupRepositoryMock
		hasher      *passwordHasherMock
		transaction *transactionManagerMock
		recorder    *adminEventRecorderMock
		actors      *actorDirectoryMock
		command     inboundports.CreateUserCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result outboundports.UserResult, repository *userRepositoryMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock, actors *actorDirectoryMock)
	}{
		{
			name:        "creates user",
			repository:  &userRepositoryMock{},
			groups:      &groupRepositoryMock{findByID: entities.ReconstituteGroup(mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300001"), "superadmin")},
			hasher:      &passwordHasherMock{hashResult: "hashed-password"},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			actors:      &actorDirectoryMock{},
			command:     inboundports.CreateUserCommand{Username: "alice", Password: "supersecret", FirstName: "Alice", LastName: "Admin", GroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001"},
			assert: func(t *testing.T, result outboundports.UserResult, repository *userRepositoryMock, transaction *transactionManagerMock, recorder *adminEventRecorderMock, actors *actorDirectoryMock) {
				t.Helper()

				if repository.createdUser == nil {
					t.Fatal("expected created user")
				}

				if result.ID == "" {
					t.Fatal("expected user id")
				}

				if result.Username != "alice" {
					t.Fatalf("result.Username = %q, want %q", result.Username, "alice")
				}

				if result.FirstName != "Alice" {
					t.Fatalf("result.FirstName = %q, want %q", result.FirstName, "Alice")
				}

				if !transaction.invoked {
					t.Fatal("expected transaction manager to be invoked")
				}

				if recorder.command.ActorUserID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" {
					t.Fatalf("recorder.command.ActorUserID = %q, want %q", recorder.command.ActorUserID, "01962b8f-aeb2-7e03-a8ff-1edce1300002")
				}

				if recorder.command.ActorGroupID != "01962b8f-aeb2-7e03-a8ff-1edce1300001" {
					t.Fatalf("recorder.command.ActorGroupID = %q, want %q", recorder.command.ActorGroupID, "01962b8f-aeb2-7e03-a8ff-1edce1300001")
				}

				if actors.userID != "01962b8f-aeb2-7e03-a8ff-1edce1300002" || actors.groupID != "01962b8f-aeb2-7e03-a8ff-1edce1300001" {
					t.Fatalf("actor validation = (%q, %q)", actors.userID, actors.groupID)
				}

				if recorder.command.EventType != createUserAdminEventType {
					t.Fatalf("recorder.command.EventType = %q, want %q", recorder.command.EventType, createUserAdminEventType)
				}

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}

				if payload["target_username"] != "alice" {
					t.Fatalf("payload[target_username] = %v, want %q", payload["target_username"], "alice")
				}
			},
		},
		{
			name:        "wraps repository error",
			repository:  &userRepositoryMock{createErr: errors.New("boom")},
			groups:      &groupRepositoryMock{findByID: entities.ReconstituteGroup(mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300001"), "superadmin")},
			hasher:      &passwordHasherMock{hashResult: "hashed-password"},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			actors:      &actorDirectoryMock{},
			command:     inboundports.CreateUserCommand{Username: "alice", Password: "supersecret", FirstName: "Alice", LastName: "Admin", GroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001"},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				if err == nil {
					t.Fatal("expected error")
				}

				if !strings.Contains(err.Error(), "creating user transaction") {
					t.Fatalf("err.Error() = %q, want substring %q", err.Error(), "creating user transaction")
				}
			},
		},
		{
			name:        "rejects short password",
			repository:  &userRepositoryMock{},
			groups:      &groupRepositoryMock{},
			hasher:      &passwordHasherMock{},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{},
			actors:      &actorDirectoryMock{},
			command:     inboundports.CreateUserCommand{Username: "alice", Password: "short", FirstName: "Alice", LastName: "Admin", GroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001"},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				if err == nil {
					t.Fatal("expected error")
				}
			},
		},
		{
			name:        "wraps event recorder error",
			repository:  &userRepositoryMock{},
			groups:      &groupRepositoryMock{findByID: entities.ReconstituteGroup(mustGroupID(t, "01962b8f-aeb2-7e03-a8ff-1edce1300001"), "superadmin")},
			hasher:      &passwordHasherMock{hashResult: "hashed-password"},
			transaction: &transactionManagerMock{},
			recorder:    &adminEventRecorderMock{err: errors.New("audit failed")},
			actors:      &actorDirectoryMock{},
			command:     inboundports.CreateUserCommand{Username: "alice", Password: "supersecret", FirstName: "Alice", LastName: "Admin", GroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "01962b8f-aeb2-7e03-a8ff-1edce1300001"},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				if err == nil {
					t.Fatal("expected error")
				}

				if !strings.Contains(err.Error(), "creating user transaction") {
					t.Fatalf("err.Error() = %q, want substring %q", err.Error(), "creating user transaction")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewCreateUserUseCase(tt.repository, tt.groups, tt.hasher, tt.recorder, tt.transaction, tt.actors)
			useCase.now = testTime

			result, err := useCase.Execute(context.Background(), tt.command)
			if tt.assertError != nil {
				tt := tt
				tt.assertError(t, err)
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tt := tt
			tt.assert(t, result, tt.repository, tt.transaction, tt.recorder, tt.actors)
		})
	}
}

func mustGroupID(t *testing.T, value string) valueobjects.GroupID {
	t.Helper()

	groupID, err := valueobjects.GroupIDFromString(value)
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	return groupID
}
