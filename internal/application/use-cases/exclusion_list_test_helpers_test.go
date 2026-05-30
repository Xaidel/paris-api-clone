package usecases

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

type exclusionListRepositoryMock struct {
	createdEntry *entities.ExclusionListEntry
	createdBy    string
	createErr    error
	findByID     *entities.ExclusionListEntry
	findByIDErr  error
	listEntries  []*entities.ExclusionListEntry
	listErr      error
	updatedEntry *entities.ExclusionListEntry
	updateErr    error
	deletedID    string
	deleteErr    error
}

func (m *exclusionListRepositoryMock) Create(_ context.Context, entry *entities.ExclusionListEntry, createdByUserID string) error {
	m.createdEntry = entry
	m.createdBy = createdByUserID
	return m.createErr
}

func (m *exclusionListRepositoryMock) FindByID(_ context.Context, id valueobjects.ExclusionListID) (*entities.ExclusionListEntry, error) {
	_ = id
	return m.findByID, m.findByIDErr
}

func (m *exclusionListRepositoryMock) List(context.Context) ([]*entities.ExclusionListEntry, error) {
	return m.listEntries, m.listErr
}

func (m *exclusionListRepositoryMock) Update(_ context.Context, entry *entities.ExclusionListEntry) error {
	m.updatedEntry = entry
	return m.updateErr
}

func (m *exclusionListRepositoryMock) DeleteByID(_ context.Context, id valueobjects.ExclusionListID) error {
	m.deletedID = id.String()
	return m.deleteErr
}
