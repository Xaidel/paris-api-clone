//go:build integration

package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gyud-adb/paris-api/internal/infrastructure/config"
	infraDB "github.com/gyud-adb/paris-api/internal/infrastructure/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	expectedExclusionSeedRowCount = 4
	expectedU1SeedRowCount        = 33
)

func TestRunSeedReferenceDataWithoutMigrationsReturnsMissingSchemaError(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("set INTEGRATION=true to run integration tests")
	}

	adminDatabaseURL := os.Getenv("DATABASE_URL")
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if adminDatabaseURL == "" || migrationsPath == "" {
		t.Skip("DATABASE_URL and MIGRATIONS_PATH are required")
	}

	ctx := context.Background()
	tempDatabaseURL, absMigrationsPath := setupMigrateIntegrationEnvironment(t, ctx, adminDatabaseURL, migrationsPath)
	setRunEnvironment(t, tempDatabaseURL, absMigrationsPath)

	err := run([]string{migrationActionSeedReferenceData})
	if err == nil {
		t.Fatalf("run(%q) error = nil, want missing schema error", migrationActionSeedReferenceData)
	}

	if !strings.Contains(err.Error(), "exclusion_list") {
		t.Fatalf("run(%q) error = %q, want message to mention exclusion_list", migrationActionSeedReferenceData, err.Error())
	}
}

func TestRunWithoutArgsAppliesMigrations(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("set INTEGRATION=true to run integration tests")
	}

	adminDatabaseURL := os.Getenv("DATABASE_URL")
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if adminDatabaseURL == "" || migrationsPath == "" {
		t.Skip("DATABASE_URL and MIGRATIONS_PATH are required")
	}

	ctx := context.Background()
	tempDatabaseURL, absMigrationsPath := setupMigrateIntegrationEnvironment(t, ctx, adminDatabaseURL, migrationsPath)
	setRunEnvironment(t, tempDatabaseURL, absMigrationsPath)

	err := run(nil)
	if err != nil {
		t.Fatalf("run(nil) error = %v", err)
	}

	pool, err := pgxpool.New(ctx, tempDatabaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	t.Cleanup(pool.Close)

	var exclusionListTable string
	if err := pool.QueryRow(ctx, `SELECT COALESCE(to_regclass('public.exclusion_list')::text, '')`).Scan(&exclusionListTable); err != nil {
		t.Fatalf("SELECT to_regclass('public.exclusion_list') error = %v", err)
	}
	if exclusionListTable != "exclusion_list" {
		t.Fatalf("exclusion_list table lookup = %q, want %q", exclusionListTable, "exclusion_list")
	}
}

func TestRunSeedReferenceDataSucceedsAfterExplicitMigrations(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("set INTEGRATION=true to run integration tests")
	}

	adminDatabaseURL := os.Getenv("DATABASE_URL")
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if adminDatabaseURL == "" || migrationsPath == "" {
		t.Skip("DATABASE_URL and MIGRATIONS_PATH are required")
	}

	ctx := context.Background()
	tempDatabaseURL, absMigrationsPath := setupMigrateIntegrationEnvironment(t, ctx, adminDatabaseURL, migrationsPath)
	setRunEnvironment(t, tempDatabaseURL, absMigrationsPath)

	migrator := infraDB.NewMigrator(config.DatabaseConfig{
		URL:            tempDatabaseURL,
		MigrationsPath: absMigrationsPath,
	})
	if err := migrator.Up(); err != nil {
		t.Fatalf("Migrator.Up() error = %v", err)
	}

	err := run([]string{migrationActionSeedReferenceData})
	if err != nil {
		t.Fatalf("run(%q) error = %v", migrationActionSeedReferenceData, err)
	}

	pool, err := pgxpool.New(ctx, tempDatabaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	t.Cleanup(pool.Close)

	var exclusionCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM exclusion_list`).Scan(&exclusionCount); err != nil {
		t.Fatalf("SELECT COUNT(*) FROM exclusion_list error = %v", err)
	}
	if exclusionCount != expectedExclusionSeedRowCount {
		t.Fatalf("exclusion_list row count = %d, want %d", exclusionCount, expectedExclusionSeedRowCount)
	}

	var u1Count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM u1_list`).Scan(&u1Count); err != nil {
		t.Fatalf("SELECT COUNT(*) FROM u1_list error = %v", err)
	}
	if u1Count != expectedU1SeedRowCount {
		t.Fatalf("u1_list row count = %d, want %d", u1Count, expectedU1SeedRowCount)
	}
}

func TestRunVersionSucceedsAfterExplicitMigrations(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("set INTEGRATION=true to run integration tests")
	}

	adminDatabaseURL := os.Getenv("DATABASE_URL")
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if adminDatabaseURL == "" || migrationsPath == "" {
		t.Skip("DATABASE_URL and MIGRATIONS_PATH are required")
	}

	ctx := context.Background()
	tempDatabaseURL, absMigrationsPath := setupMigrateIntegrationEnvironment(t, ctx, adminDatabaseURL, migrationsPath)
	setRunEnvironment(t, tempDatabaseURL, absMigrationsPath)

	migrator := newIntegrationTestMigrator(tempDatabaseURL, absMigrationsPath)
	if err := migrator.Up(); err != nil {
		t.Fatalf("Migrator.Up() error = %v", err)
	}

	beforeVersion, beforeDirty, err := migrator.Version()
	if err != nil {
		t.Fatalf("Migrator.Version() before run error = %v", err)
	}

	err = run([]string{migrationActionVersion})
	if err != nil {
		t.Fatalf("run(%q) error = %v", migrationActionVersion, err)
	}

	afterVersion, afterDirty, err := migrator.Version()
	if err != nil {
		t.Fatalf("Migrator.Version() after run error = %v", err)
	}
	if afterVersion != beforeVersion || afterDirty != beforeDirty {
		t.Fatalf("migration version after run = (%d, %t), want (%d, %t)", afterVersion, afterDirty, beforeVersion, beforeDirty)
	}
}

func TestRunDownSucceedsAfterExplicitMigrations(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("set INTEGRATION=true to run integration tests")
	}

	adminDatabaseURL := os.Getenv("DATABASE_URL")
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if adminDatabaseURL == "" || migrationsPath == "" {
		t.Skip("DATABASE_URL and MIGRATIONS_PATH are required")
	}

	ctx := context.Background()
	tempDatabaseURL, absMigrationsPath := setupMigrateIntegrationEnvironment(t, ctx, adminDatabaseURL, migrationsPath)
	setRunEnvironment(t, tempDatabaseURL, absMigrationsPath)

	migrator := newIntegrationTestMigrator(tempDatabaseURL, absMigrationsPath)
	if err := migrator.Up(); err != nil {
		t.Fatalf("Migrator.Up() error = %v", err)
	}

	err := run([]string{migrationActionDown})
	if err != nil {
		t.Fatalf("run(%q) error = %v", migrationActionDown, err)
	}

	version, dirty, err := migrator.Version()
	if err != nil {
		t.Fatalf("Migrator.Version() error = %v", err)
	}
	if version != 0 || dirty {
		t.Fatalf("migration version after down = (%d, %t), want (0, false)", version, dirty)
	}
}

func setupMigrateIntegrationEnvironment(t *testing.T, ctx context.Context, adminDatabaseURL string, migrationsPath string) (string, string) {
	t.Helper()

	absMigrationsPath, err := resolveMigrateIntegrationMigrationsPath(migrationsPath)
	if err != nil {
		t.Fatalf("resolveMigrateIntegrationMigrationsPath(%q) error = %v", migrationsPath, err)
	}

	tempDatabaseName := uniqueMigrateIntegrationDatabaseName()
	tempDatabaseURL, err := databaseURLWithName(adminDatabaseURL, tempDatabaseName)
	if err != nil {
		t.Fatalf("databaseURLWithName() error = %v", err)
	}

	createTemporaryIntegrationDatabase(t, ctx, adminDatabaseURL, tempDatabaseName)
	t.Cleanup(func() {
		dropTemporaryIntegrationDatabase(t, ctx, adminDatabaseURL, tempDatabaseName)
	})

	return tempDatabaseURL, absMigrationsPath
}

func setRunEnvironment(t *testing.T, databaseURL string, migrationsPath string) {
	t.Helper()

	logPath := filepath.Join(t.TempDir(), "migrate.log")
	t.Setenv("ENV", "test")
	t.Setenv("SERVICE_NAME", "paris-api")
	t.Setenv("HTTP_PORT", "9000")
	t.Setenv("HTTP_READ_TIMEOUT", "5s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "10s")
	t.Setenv("HTTP_IDLE_TIMEOUT", "60s")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "10s")
	t.Setenv("DATABASE_URL", databaseURL)
	t.Setenv("DATABASE_PING_TIMEOUT", "5s")
	t.Setenv("MIGRATIONS_PATH", migrationsPath)
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("LOG_FILE_PATH", logPath)
	t.Setenv("STORAGE_PROVIDER", "local")
	t.Setenv("LOCAL_TRANSACTION_FILE_PATH", ".tx-files")
}

func newIntegrationTestMigrator(databaseURL string, migrationsPath string) *infraDB.Migrator {
	return infraDB.NewMigrator(config.DatabaseConfig{
		URL:            databaseURL,
		MigrationsPath: migrationsPath,
	})
}

func resolveMigrateIntegrationMigrationsPath(path string) (string, error) {
	candidates := []string{path}
	if !filepath.IsAbs(path) {
		candidates = append(candidates, filepath.Join("..", "..", path))
	}

	for _, candidate := range candidates {
		absCandidate, err := filepath.Abs(candidate)
		if err != nil {
			return "", err
		}

		if _, err := os.Stat(absCandidate); err == nil {
			return absCandidate, nil
		}
	}

	return "", os.ErrNotExist
}

func uniqueMigrateIntegrationDatabaseName() string {
	return fmt.Sprintf("cmd_migrate_test_%d", time.Now().UTC().UnixNano())
}

func databaseURLWithName(databaseURL string, databaseName string) (string, error) {
	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("parsing database url: %w", err)
	}

	if parsedURL.Scheme == "" {
		return "", fmt.Errorf("database url scheme is required")
	}

	parsedURL.Path = "/" + databaseName
	return parsedURL.String(), nil
}

func createTemporaryIntegrationDatabase(t *testing.T, ctx context.Context, adminDatabaseURL string, databaseName string) {
	t.Helper()

	adminConn, err := pgx.Connect(ctx, adminDatabaseURL)
	if err != nil {
		t.Skipf("admin database connection unavailable for temp database setup: %v", err)
	}
	defer adminConn.Close(ctx)

	if _, err := adminConn.Exec(ctx, "CREATE DATABASE \""+strings.ReplaceAll(databaseName, "\"", "\"\"")+"\""); err != nil {
		if isInsufficientPrivilegeError(err) {
			t.Skipf("creating temporary integration database %q requires CREATEDB privileges: %v", databaseName, err)
		}

		t.Fatalf("CREATE DATABASE %q error = %v", databaseName, err)
	}
}

func dropTemporaryIntegrationDatabase(t *testing.T, ctx context.Context, adminDatabaseURL string, databaseName string) {
	t.Helper()

	adminConn, err := pgx.Connect(ctx, adminDatabaseURL)
	if err != nil {
		t.Fatalf("admin database connection for cleanup error = %v", err)
	}
	defer adminConn.Close(ctx)

	quotedDatabaseName := "\"" + strings.ReplaceAll(databaseName, "\"", "\"\"") + "\""
	if _, err := adminConn.Exec(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, databaseName); err != nil {
		t.Fatalf("terminating connections to temporary database %q error = %v", databaseName, err)
	}

	if _, err := adminConn.Exec(ctx, "DROP DATABASE IF EXISTS "+quotedDatabaseName); err != nil {
		t.Fatalf("DROP DATABASE %q error = %v", databaseName, err)
	}
}

func isInsufficientPrivilegeError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == "42501"
}
