// Package usecases contains the application's named actions. Each exported use
// case coordinates one business operation by validating actor context, loading
// domain state through ports, delegating invariant checks to domain objects,
// persisting the result, and recording any resulting events.
//
// The package also holds a small set of result mappers and helper types that
// keep transport-facing DTOs separate from domain entities. Common patterns in
// this package include:
//   - CRUD-oriented use cases for users, groups, sectors, exclusions, bug
//     reports, transactions, uploads, and feedback.
//   - Query-oriented use cases that list records or fetch a single aggregate.
//   - Cross-cutting helpers such as event_publisher.go, mapper files, and
//     shared actor validation helpers.
//
// A typical use case constructor accepts only ports and pure collaborators.
// Concrete adapters are wired in internal/infrastructure/di and are never
// referenced directly from here. If a new workflow needs an external capability,
// define or extend a port first and keep the orchestration logic here focused on
// sequencing, error context, and transactional boundaries.
//
// New contributors usually get the fastest picture by reading one create use
// case, one get/list use case, and one mapper file together. That shows the
// package's main responsibility: expose a stable application contract without
// leaking adapter concerns outward or domain internals across boundaries.
package usecases
