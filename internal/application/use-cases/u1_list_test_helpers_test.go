package usecases

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

type u1ListRepositoryMock struct {
	createdEntry *entities.U1ListEntry
	createdBy    string
	createErr    error
	findByID     *entities.U1ListEntry
	findByIDErr  error
	listFilter   ports.U1ListFilter
	listEntries  []*entities.U1ListEntry
	listErr      error
	updatedEntry *entities.U1ListEntry
	updateErr    error
	deletedID    string
	deleteErr    error
}

func (m *u1ListRepositoryMock) Create(_ context.Context, entry *entities.U1ListEntry, createdByUserID string) error {
	m.createdEntry = entry
	m.createdBy = createdByUserID
	return m.createErr
}

func (m *u1ListRepositoryMock) FindByID(_ context.Context, id valueobjects.U1ListID) (*entities.U1ListEntry, error) {
	_ = id
	return m.findByID, m.findByIDErr
}

func (m *u1ListRepositoryMock) List(_ context.Context, filter ports.U1ListFilter) ([]*entities.U1ListEntry, error) {
	m.listFilter = filter
	return m.listEntries, m.listErr
}

func (m *u1ListRepositoryMock) Update(_ context.Context, entry *entities.U1ListEntry) error {
	m.updatedEntry = entry
	return m.updateErr
}

func (m *u1ListRepositoryMock) DeleteByID(_ context.Context, id valueobjects.U1ListID) error {
	m.deletedID = id.String()
	return m.deleteErr
}
