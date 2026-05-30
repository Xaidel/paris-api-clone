package adapters

import (
	"context"
	"errors"
	"fmt"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/jackc/pgx/v5"
)

const (
	createBugReportQuery = `
INSERT INTO bug_reports (id, user_id, transaction_id, title, description, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`
	findBugReportByIDQuery = `
SELECT id, user_id, transaction_id, title, description, status, created_at, updated_at
FROM bug_reports
WHERE id = $1
`
	listBugReportsQuery = `
SELECT id, user_id, transaction_id, title, description, status, created_at, updated_at
FROM bug_reports
ORDER BY created_at DESC, id ASC
`
	updateBugReportQuery = `
UPDATE bug_reports
SET title = $2, description = $3, status = $4, updated_at = $5
WHERE id = $1
`
	deleteBugReportQuery = `
DELETE FROM bug_reports
WHERE id = $1
`
)

// PostgresBugReportRepository persists bug reports in PostgreSQL.
type PostgresBugReportRepository struct {
	pool pgxQuerier
}

// NewPostgresBugReportRepository builds a PostgresBugReportRepository.
func NewPostgresBugReportRepository(pool pgxQuerier) *PostgresBugReportRepository {
	return &PostgresBugReportRepository{pool: pool}
}

// Create inserts a new bug report.
func (r *PostgresBugReportRepository) Create(ctx context.Context, bugReport *entities.BugReport) error {
	querier := txQuerierFromContext(ctx, r.pool)
	if _, err := querier.Exec(ctx, createBugReportQuery, bugReport.ID().String(), bugReport.UserID(), bugReport.TransactionID(), bugReport.Title(), bugReport.Description(), bugReport.Status(), bugReport.CreatedAt(), bugReport.UpdatedAt()); err != nil {
		return fmt.Errorf("executing create bug report query: %w", err)
	}

	return nil
}

// FindByID returns a bug report by identifier.
func (r *PostgresBugReportRepository) FindByID(ctx context.Context, id valueobjects.BugReportID) (*entities.BugReport, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	bugReport, err := scanBugReport(querier.QueryRow(ctx, findBugReportByIDQuery, id.String()))
	if err != nil {
		return nil, fmt.Errorf("scanning bug report by id: %w", err)
	}

	return bugReport, nil
}

// List returns all bug reports.
func (r *PostgresBugReportRepository) List(ctx context.Context) ([]*entities.BugReport, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	rows, err := querier.Query(ctx, listBugReportsQuery)
	if err != nil {
		return nil, fmt.Errorf("querying bug reports: %w", err)
	}
	defer rows.Close()

	bugReports := make([]*entities.BugReport, 0)
	for rows.Next() {
		bugReport, scanErr := scanBugReport(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scanning listed bug report: %w", scanErr)
		}

		bugReports = append(bugReports, bugReport)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating bug reports: %w", err)
	}

	return bugReports, nil
}

// Update updates an existing bug report.
func (r *PostgresBugReportRepository) Update(ctx context.Context, bugReport *entities.BugReport) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, updateBugReportQuery, bugReport.ID().String(), bugReport.Title(), bugReport.Description(), bugReport.Status(), bugReport.UpdatedAt())
	if err != nil {
		return fmt.Errorf("executing update bug report query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// DeleteByID deletes an existing bug report.
func (r *PostgresBugReportRepository) DeleteByID(ctx context.Context, id valueobjects.BugReportID) error {
	querier := txQuerierFromContext(ctx, r.pool)
	commandTag, err := querier.Exec(ctx, deleteBugReportQuery, id.String())
	if err != nil {
		return fmt.Errorf("executing delete bug report query: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func scanBugReport(row scanner) (*entities.BugReport, error) {
	var (
		bugReportID   string
		userID        string
		transactionID string
		title         string
		description   string
		status        string
		createdAt     time.Time
		updatedAt     time.Time
	)

	if err := row.Scan(&bugReportID, &userID, &transactionID, &title, &description, &status, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	parsedID, err := valueobjects.BugReportIDFromString(bugReportID)
	if err != nil {
		return nil, fmt.Errorf("parsing bug report id: %w", err)
	}

	parsedUserID, err := valueobjects.UserIDFromString(userID)
	if err != nil {
		return nil, fmt.Errorf("parsing bug report user id: %w", err)
	}

	parsedTransactionID, err := valueobjects.TransactionIDFromString(transactionID)
	if err != nil {
		return nil, fmt.Errorf("parsing bug report transaction id: %w", err)
	}

	parsedStatus, err := valueobjects.BugReportStatusFromString(status)
	if err != nil {
		return nil, fmt.Errorf("parsing bug report status: %w", err)
	}

	parsedTitle, err := valueobjects.NewBugReportTitle(title)
	if err != nil {
		return nil, fmt.Errorf("parsing bug report title: %w", err)
	}

	parsedDescription, err := valueobjects.NewBugReportDescription(description)
	if err != nil {
		return nil, fmt.Errorf("parsing bug report description: %w", err)
	}

	return entities.ReconstituteBugReport(parsedID, parsedUserID, parsedTransactionID, parsedTitle, parsedDescription, parsedStatus, createdAt, updatedAt), nil
}
