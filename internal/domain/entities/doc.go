// Package entities contains the mutable domain records and aggregates that carry
// transaction, user, upload, list-entry, and feedback lifecycle behavior. These
// types enforce invariants for state transitions and emit domain events through
// the shared aggregate root support when important actions occur.
//
// The package is organized around persisted records rather than transport DTOs.
// A few useful reading paths are:
//   - transaction.go for the most feature-rich aggregate and classification
//     state transitions.
//   - user.go and group.go for identity and access-related records.
//   - transaction_upload.go plus related list-entry files for import and lookup
//     workflows.
//   - aggregate_root.go for the event recording pattern used by entities that
//     publish administrative or domain events.
//
// Constructors in this package create valid entities or return explicit domain
// errors. Reconstitution helpers exist where storage needs to rebuild state
// without replaying creation-time side effects. Code outside the domain should
// prefer these constructors and methods rather than mutating struct fields
// directly.
package entities
