package usecases

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

type reportRepositoryMock struct {
	createdReport *entities.BugReport
	createErr     error
	findByID      *entities.BugReport
	findByIDErr   error
	listReports   []*entities.BugReport
	listErr       error
	updatedReport *entities.BugReport
	updateErr     error
	deletedID     string
	deleteErr     error
}

func (m *reportRepositoryMock) Create(_ context.Context, bugReport *entities.BugReport) error {
	m.createdReport = bugReport
	return m.createErr
}

func (m *reportRepositoryMock) FindByID(_ context.Context, _ valueobjects.BugReportID) (*entities.BugReport, error) {
	return m.findByID, m.findByIDErr
}

func (m *reportRepositoryMock) List(_ context.Context) ([]*entities.BugReport, error) {
	return m.listReports, m.listErr
}

func (m *reportRepositoryMock) Update(_ context.Context, bugReport *entities.BugReport) error {
	m.updatedReport = bugReport
	return m.updateErr
}

func (m *reportRepositoryMock) DeleteByID(_ context.Context, id valueobjects.BugReportID) error {
	m.deletedID = id.String()
	return m.deleteErr
}
