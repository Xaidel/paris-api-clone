// Package observability contains runtime logging and metrics setup used by the
// composition root and long-running workers. It centralizes cross-cutting
// telemetry concerns so the rest of the code can depend on injected loggers and
// metrics objects instead of configuring them ad hoc.
//
// Typical responsibilities include logger construction, cleanup hooks, and
// metrics collectors that track classification throughput, latency, or failures.
// Keep exporter setup and instrumentation primitives here rather than scattering
// operational details across adapters and services.
//
// When extending telemetry, prefer adding a focused helper in this package and
// injecting it through constructors. That keeps instrumentation consistent and
// makes startup behavior visible from internal/infrastructure/di.
package observability
