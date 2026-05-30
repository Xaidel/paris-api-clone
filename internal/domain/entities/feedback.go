// Package entities
package entities

import (
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/domain/events"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

// Feedback is a domain entity representing a user's thumbs up/down on a transaction.
type Feedback struct {
	aggregateRoot
	id            valueobjects.FeedbackID
	userID        valueobjects.UserID
	transactionID valueobjects.TransactionID
	kind          valueobjects.FeedbackKind
	createdAt     time.Time
	updatedAt     time.Time
}

// NewFeedback creates a new valid feedback entity.
func NewFeedback(id valueobjects.FeedbackID, userID valueobjects.UserID, transactionID valueobjects.TransactionID, kind valueobjects.FeedbackKind, now time.Time) (*Feedback, error) {
	if now.IsZero() {
		return nil, domain.ErrInvalidTimestamp
	}

	return &Feedback{
		id:            id,
		userID:        userID,
		transactionID: transactionID,
		kind:          kind,
		createdAt:     now,
		updatedAt:     now,
	}, nil
}

// ReconstituteFeedback rebuilds a feedback entity from storage.
func ReconstituteFeedback(id valueobjects.FeedbackID, userID valueobjects.UserID, transactionID valueobjects.TransactionID, kind valueobjects.FeedbackKind, createdAt, updatedAt time.Time) *Feedback {
	return &Feedback{
		id:            id,
		userID:        userID,
		transactionID: transactionID,
		kind:          kind,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

// RecordUpserted records the feedback upsert event.
func (f *Feedback) RecordUpserted(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.UpsertTransactionFeedbackEventType, map[string]any{
		"action":         "upsert",
		"resource":       "transaction_feedback",
		"target_id":      f.ID().String(),
		"transaction_id": f.TransactionID(),
		"kind":           f.Kind(),
	})
	if err != nil {
		return err
	}

	f.recordDomainEvent(event)
	return nil
}

// RecordDeleted records the feedback deletion event.
func (f *Feedback) RecordDeleted(now time.Time, actorUserID, actorGroupID string) error {
	event, err := events.NewAdminActionOccurred(now, actorUserID, actorGroupID, events.DeleteTransactionFeedbackEventType, map[string]any{
		"action":         "delete",
		"resource":       "transaction_feedback",
		"target_id":      f.ID().String(),
		"transaction_id": f.TransactionID(),
	})
	if err != nil {
		return err
	}

	f.recordDomainEvent(event)
	return nil
}

// ID returns the feedback identifier.
func (f *Feedback) ID() valueobjects.FeedbackID {
	return f.id
}

// UserID returns the user identifier.
func (f *Feedback) UserID() string {
	return f.userID.String()
}

// TransactionID returns the transaction identifier.
func (f *Feedback) TransactionID() string {
	return f.transactionID.String()
}

// Kind returns the feedback kind.
func (f *Feedback) Kind() string {
	return f.kind.String()
}

// CreatedAt returns the creation timestamp.
func (f *Feedback) CreatedAt() time.Time {
	return f.createdAt
}

// UpdatedAt returns the update timestamp.
func (f *Feedback) UpdatedAt() time.Time {
	return f.updatedAt
}

// ChangeKind updates the feedback kind and timestamp.
func (f *Feedback) ChangeKind(kind valueobjects.FeedbackKind, now time.Time) error {
	if now.IsZero() {
		return domain.ErrInvalidTimestamp
	}

	f.kind = kind
	f.updatedAt = now
	return nil
}

// Equal reports whether two feedbacks share the same identity.
func (f *Feedback) Equal(other *Feedback) bool {
	if f == nil || other == nil {
		return false
	}

	return f.id.Equal(other.id)
}
