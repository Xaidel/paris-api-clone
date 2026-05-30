package usecases

import (
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

func newBugReportResult(bugReport *entities.BugReport) ports.BugReportResult {
	return ports.BugReportResult{
		ID:            bugReport.ID().String(),
		UserID:        bugReport.UserID(),
		TransactionID: bugReport.TransactionID(),
		Title:         bugReport.Title(),
		Description:   bugReport.Description(),
		Status:        bugReport.Status(),
		CreatedAt:     bugReport.CreatedAt(),
		UpdatedAt:     bugReport.UpdatedAt(),
	}
}
