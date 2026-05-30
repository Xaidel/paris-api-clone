package ports

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// BugReportRepository persists bug reports.
type BugReportRepository interface {
	Create(ctx context.Context, bugReport *entities.BugReport) error
	FindByID(ctx context.Context, id valueobjects.BugReportID) (*entities.BugReport, error)
	List(ctx context.Context) ([]*entities.BugReport, error)
	Update(ctx context.Context, bugReport *entities.BugReport) error
	DeleteByID(ctx context.Context, id valueobjects.BugReportID) error
}
