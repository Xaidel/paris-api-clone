// Package domain defines the shared language of the system. It holds errors,
// validation primitives, shared contracts such as DomainEvent, and other core
// concepts that must remain independent from adapters, transport protocols, and
// infrastructure details.
//
// This package works with the domain subpackages rather than replacing them:
//   - entities contains aggregates and records with lifecycle behavior.
//   - valueobjects contains immutable typed values and structured pipeline data.
//   - services contains business algorithms that do not naturally belong to one
//     entity.
//   - events contains concrete domain event implementations.
//
// Files at the root of package domain generally define shared validation and
// error types used across the rest of the model. When adding new core concepts,
// prefer the narrowest package that matches the concept. Keep I/O, logging,
// database knowledge, and framework types out of the domain entirely.
//
// New developers should read this package alongside the subpackage doc.go files.
// Together they explain the split between reusable domain primitives, aggregate
// behavior, immutable value types, and event publication.
package domain
