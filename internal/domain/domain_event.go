package domain

import "time"

// DomainEvent describes an immutable event emitted by the domain layer.
type DomainEvent interface {
	// EventName returns the stable event type used by recorders and outbox
	// publishers.
	EventName() string
	// OccurredAt returns the timestamp captured when the domain fact happened.
	OccurredAt() time.Time
}
