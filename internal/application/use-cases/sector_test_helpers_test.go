package usecases

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

type sectorRepositoryMock struct {
	createdSector *entities.Sector
	createdBy     string
	createErr     error
	findByID      *entities.Sector
	findByIDErr   error
	listSectors   []*entities.Sector
	listErr       error
	updatedSector *entities.Sector
	updateErr     error
	deletedID     string
	deleteErr     error
}

func (m *sectorRepositoryMock) Create(_ context.Context, sector *entities.Sector, createdByUserID string) error {
	m.createdSector = sector
	m.createdBy = createdByUserID
	return m.createErr
}

func (m *sectorRepositoryMock) FindByID(_ context.Context, id valueobjects.SectorID) (*entities.Sector, error) {
	_ = id
	return m.findByID, m.findByIDErr
}

func (m *sectorRepositoryMock) List(context.Context) ([]*entities.Sector, error) {
	return m.listSectors, m.listErr
}

func (m *sectorRepositoryMock) Update(_ context.Context, sector *entities.Sector) error {
	m.updatedSector = sector
	return m.updateErr
}

func (m *sectorRepositoryMock) DeleteByID(_ context.Context, id valueobjects.SectorID) error {
	m.deletedID = id.String()
	return m.deleteErr
}
