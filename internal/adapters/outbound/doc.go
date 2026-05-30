// Package adapters contains the outbound boundary implementations for the
// application. It is where Postgres repositories, storage backends, password
// hashing, queue persistence, and AI gateways translate between external
// systems and the port contracts defined in package ports.
//
// The files in this package fall into a few stable groups:
//   - Postgres repositories such as transaction_postgres_repository.go and
//     group_postgres_repository.go implement persistence ports.
//   - Queue and transaction helpers such as classification_job_queue_postgres.go
//     and transaction_manager_pgx.go coordinate database-backed workflows.
//   - External service adapters such as
//     transaction_classification_react_gateway.go bridge to AI providers.
//   - Storage adapters such as raw_file_local_store.go and
//     raw_file_azure_blob_store.go persist uploaded raw assets.
//
// When developing in this package, keep the layer contract in mind: adapters do
// not own business rules. SQL mapping, retry behavior, and response
// translation from external systems belong here. Domain invariants, workflow
// policy, and cross-entity business decisions do not.
//
// A good entry path for new contributors is:
//  1. Read the matching outbound port in internal/ports/outbound.
//  2. Read the repository or gateway implementation that satisfies the port.
//  3. Read the matching use case in internal/application/use-cases.
//  4. Finish with internal/infrastructure/di/wire.go to see runtime wiring.
package adapters
