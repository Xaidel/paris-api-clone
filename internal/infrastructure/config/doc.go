// Package config loads runtime configuration from the environment and turns it
// into strongly typed application settings. The resulting AppConfig is consumed
// by the composition root and should be the single source of truth for runtime
// feature flags, database settings, storage configuration, and AI provider
// choices.
//
// This package is part of infrastructure, so it may depend on environment and
// operational concerns that the core layers must never import directly. Keep the
// parsing, defaults, and validation here, and pass the resulting config values
// into adapters or services through constructors.
//
// Read config.go before adding new environment variables. The package is the
// right place to document default behavior, required settings, and validation
// failures that should stop application startup.
package config
