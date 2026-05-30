// Package httpserver assembles the Gin-based HTTP surface for the application.
// It owns router construction, middleware attachment, and route registration for
// inbound adapters that expose the application over HTTP.
//
// The package should stay narrowly focused on server composition. Request
// parsing, response mapping, and transport-specific behavior belong in the HTTP
// adapters under internal/adapters, while business orchestration remains in the
// application layer.
//
// New contributors should read router.go together with one adapter such as
// internal/adapters/user_http_adapter.go. That pairing shows the intended split
// between route registration and request handling.
package httpserver
