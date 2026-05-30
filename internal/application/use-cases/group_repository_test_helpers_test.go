package usecases

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

type groupRepositoryUseCaseMock struct {
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

func (m *groupRepositoryUseCaseMock) Create(_ context.Context, group *entities.Group) error {
	m.createdGroup = group
	return m.createErr
}

func (m *groupRepositoryUseCaseMock) FindByID(context.Context, valueobjects.GroupID) (*entities.Group, error) {
	return m.findByID, m.findByIDErr
}

func (m *groupRepositoryUseCaseMock) List(context.Context) ([]*entities.Group, error) {
	return m.listGroups, m.listErr
}

func (m *groupRepositoryUseCaseMock) Update(_ context.Context, group *entities.Group) error {
	m.updatedGroup = group
	return m.updateErr
}

func (m *groupRepositoryUseCaseMock) DeleteByID(_ context.Context, id valueobjects.GroupID) error {
	m.deletedID = id.String()
	return m.deleteErr
}
