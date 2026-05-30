package entities

import (
	"slices"

	"github.com/gyud-adb/paris-api/internal/domain"
)

type aggregateRoot struct {
	domainEvents []domain.DomainEvent
}

func (a *aggregateRoot) recordDomainEvent(event domain.DomainEvent) {
	// Ignore nil events so entity methods can build an event conditionally
	// without forcing every call site to add its own guard.
	if event == nil {
		return
	}

	a.domainEvents = append(a.domainEvents, event)
}

// PullDomainEvents returns and clears the entity's pending domain events.
func (a *aggregateRoot) PullDomainEvents() []domain.DomainEvent {
	events := slices.Clone(a.domainEvents)
	a.domainEvents = nil
	return events
}
