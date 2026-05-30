// Package db contains infrastructure helpers for PostgreSQL connectivity and
// schema migration. It is intentionally small and operational: create the pool,
// configure connection behavior, and run migrations before the application is
// wired.
//
// The rest of the codebase should not use this package for data access logic.
// Repositories in internal/adapters own SQL and persistence mapping. This
// package only provides the shared database primitives that the composition root
// needs during startup.
//
// For startup flow, read migrate.go first and then pool.go. That sequence shows
// how the application ensures schema readiness before opening the long-lived
// connection pool used by Postgres-backed adapters.
package db
