package ports

import "time"

// BugReportResult exposes a bug report to inbound adapters.
type BugReportResult struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	TransactionID string    `json:"transaction_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ListBugReportsResult returns all bug reports.
type ListBugReportsResult struct {
	BugReports []BugReportResult `json:"bug_reports"`
}
