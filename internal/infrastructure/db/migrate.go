package db

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gyud-adb/paris-api/internal/infrastructure/config"
)

type migrationClient interface {
	Up() error
	Down() error
	Version() (uint, bool, error)
	Close() (source error, database error)
}

type migrationFactory func(sourceURL, databaseURL string) (migrationClient, error)

// Migrator creates short-lived migrate clients to run schema operations.
type Migrator struct {
	databaseURL    string
	migrationsPath string
	factory        migrationFactory
}

// NewMigrator builds a migration runner.
func NewMigrator(cfg config.DatabaseConfig) *Migrator {
	return &Migrator{
		databaseURL:    cfg.URL,
		migrationsPath: cfg.MigrationsPath,
		factory: func(sourceURL, databaseURL string) (migrationClient, error) {
			migrator, err := migrate.New(sourceURL, databaseURL)
			if err != nil {
				return nil, err
			}

			return migrator, nil
		},
	}
}

// Up applies all pending migrations.
func (m *Migrator) Up() error {
	return m.run(func(client migrationClient) error {
		if err := client.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("running up migrations: %w", err)
		}

		return nil
	})
}

// Down rolls back migrations.
func (m *Migrator) Down() error {
	return m.run(func(client migrationClient) error {
		if err := client.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("running down migrations: %w", err)
		}

		return nil
	})
}

// Version returns the current migration version.
func (m *Migrator) Version() (uint, bool, error) {
	var (
		version uint
		dirty   bool
	)

	err := m.run(func(client migrationClient) error {
		currentVersion, isDirty, err := client.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				version = 0
				dirty = false
				return nil
			}

			return fmt.Errorf("checking migration version: %w", err)
		}

		version = currentVersion
		dirty = isDirty
		return nil
	})
	if err != nil {
		return 0, false, err
	}

	return version, dirty, nil
}

func (m *Migrator) run(action func(client migrationClient) error) error {
	client, err := m.newClient()
	if err != nil {
		return err
	}
	defer closeMigrationClient(client)

	return action(client)
}

func (m *Migrator) newClient() (migrationClient, error) {
	sourceURL, err := migrationSourceURL(m.migrationsPath)
	if err != nil {
		return nil, err
	}

	databaseURL, err := migrationDatabaseURL(m.databaseURL)
	if err != nil {
		return nil, err
	}

	client, err := m.factory(sourceURL, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("creating migrator: %w", err)
	}

	return client, nil
}

func migrationSourceURL(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolving migrations path: %w", err)
	}

	return "file://" + filepath.ToSlash(absPath), nil
}

func migrationDatabaseURL(databaseURL string) (string, error) {
	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("parsing database url: %w", err)
	}

	if parsedURL.Scheme == "" {
		return "", fmt.Errorf("database url scheme is required")
	}

	parsedURL.Scheme = "pgx5"

	return parsedURL.String(), nil
}

func closeMigrationClient(client migrationClient) {
	_, _ = client.Close()
}
