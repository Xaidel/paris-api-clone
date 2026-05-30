// Package di is the composition root for the service. It wires configuration,
// observability, database connectivity, concrete adapters, application services,
// use cases, and HTTP routing into one runtime Application graph.
//
// Bootstrap is the main entry point. It performs startup work in roughly this
// order:
//  1. Load and validate configuration.
//  2. Build the logger and metrics.
//  3. Run database migrations and open the Postgres pool.
//  4. Construct adapters, services, and use cases.
//  5. Register HTTP handlers and return the assembled Application.
//
// This package is the only place in the codebase that should know about the full
// dependency graph. If a constructor needs a new collaborator, update the port
// or package that owns the behavior first, then wire the concrete implementation
// here. Avoid letting business decisions accumulate in this package; keep it as
// orchestration and assembly only.
package di
