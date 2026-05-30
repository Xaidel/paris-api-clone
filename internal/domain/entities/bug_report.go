// Package entities
package entities

import (
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// BugReport is a domain entity representing a user-submitted bug report.
type BugReport struct {
	aggregateRoot
	id            valueobjects.BugReportID
	userID        valueobjects.UserID
	transactionID valueobjects.TransactionID
	title         valueobjects.BugReportTitle
	description   valueobjects.BugReportDescription
	status        valueobjects.BugReportStatus
	createdAt     time.Time
	updatedAt     time.Time
}

// NewBugReport creates a new valid bug report entity.
func NewBugReport(id valueobjects.BugReportID, userID valueobjects.UserID, transactionID valueobjects.TransactionID, title valueobjects.BugReportTitle, description valueobjects.BugReportDescription, status valueobjects.BugReportStatus, now time.Time) (*BugReport, error) {
	if now.IsZero() {
		return nil, domain.ErrInvalidTimestamp
	}

	return &BugReport{
		id:            id,
		userID:        userID,
		transactionID: transactionID,
		title:         title,
		description:   description,
		status:        status,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

// ReconstituteBugReport rebuilds a bug report from storage.
func ReconstituteBugReport(id valueobjects.BugReportID, userID valueobjects.UserID, transactionID valueobjects.TransactionID, title valueobjects.BugReportTitle, description valueobjects.BugReportDescription, status valueobjects.BugReportStatus, createdAt, updatedAt time.Time) *BugReport {
	return &BugReport{
		id:            id,
		userID:        userID,
		transactionID: transactionID,
		title:         title,
		description:   description,
		status:        status,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

// RecordCreated records the bug report creation event.
func (r *BugReport) RecordCreated(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.CreateBugReportEventType, map[string]any{
		"action":         "create",
		"resource":       "bug_report",
		"target_id":      r.ID().String(),
		"transaction_id": r.TransactionID(),
		"title":          r.Title(),
		"description":    r.Description(),
		"status":         r.Status(),
	})
	if err != nil {
		return err
	}

	r.recordDomainEvent(event)
	return nil
}

// RecordRead records the bug report read event.
func (r *BugReport) RecordRead(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.GetBugReportEventType, map[string]any{
		"action":    "read",
		"resource":  "bug_report",
		"target_id": r.ID().String(),
	})
	if err != nil {
		return err
	}

	r.recordDomainEvent(event)
	return nil
}

// RecordUpdated records the bug report update event.
func (r *BugReport) RecordUpdated(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.UpdateBugReportEventType, map[string]any{
		"action":         "update",
		"resource":       "bug_report",
		"target_id":      r.ID().String(),
		"transaction_id": r.TransactionID(),
		"title":          r.Title(),
		"description":    r.Description(),
		"status":         r.Status(),
	})
	if err != nil {
		return err
	}

	r.recordDomainEvent(event)
	return nil
}

// RecordDeleted records the bug report deletion event.
func (r *BugReport) RecordDeleted(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.DeleteBugReportEventType, map[string]any{
		"action":    "delete",
		"resource":  "bug_report",
		"target_id": r.ID().String(),
	})
	if err != nil {
		return err
	}

	r.recordDomainEvent(event)
	return nil
}

// ID returns the bug report identifier.
func (r *BugReport) ID() valueobjects.BugReportID {
	return r.id
}

// UserID returns the submitter user identifier.
func (r *BugReport) UserID() string {
	return r.userID.String()
}

// TransactionID returns the tagged transaction identifier.
func (r *BugReport) TransactionID() string {
	return r.transactionID.String()
}

// Title returns the bug report title.
func (r *BugReport) Title() string {
	return r.title.String()
}

// Description returns the bug report description.
func (r *BugReport) Description() string {
	return r.description.String()
}

// Status returns the bug report status.
func (r *BugReport) Status() string {
	return r.status.String()
}

// CreatedAt returns the creation timestamp.
func (r *BugReport) CreatedAt() time.Time {
	return r.createdAt
}

// UpdatedAt returns the update timestamp.
func (r *BugReport) UpdatedAt() time.Time {
	return r.updatedAt
}

// Touch updates the entity timestamp.
func (r *BugReport) Touch(now time.Time) error {
	if now.IsZero() {
		return domain.ErrInvalidTimestamp
	}

	r.updatedAt = now

	return nil
}

// Update updates the mutable bug report fields.
func (r *BugReport) Update(title valueobjects.BugReportTitle, description valueobjects.BugReportDescription, status valueobjects.BugReportStatus, now time.Time) error {
	if err := r.Touch(now); err != nil {
		return err
	}

	r.title = title
	r.description = description
	r.status = status

	return nil
}

// Equal reports whether two reports share the same identity.
func (r *BugReport) Equal(other *BugReport) bool {
	if r == nil || other == nil {
		return false
	}

	return r.id.Equal(other.id)
}
