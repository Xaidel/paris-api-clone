// Package events defines concrete domain events and event naming constants used
// across the application. These events capture business-significant facts in the
// past tense so adapters and application services can record, persist, or publish
// them without re-deriving intent from mutable entities.
//
// In this codebase, administrative actions are a major event source. Files in
// this package describe the event payload and canonical event type names that are
// later consumed by event recorders and outbox repositories.
//
// Keep event types descriptive and stable. They should say what happened, not
// how a caller reached that state. Any transport formatting, database schema, or
// delivery mechanics belong outside this package.
package events
