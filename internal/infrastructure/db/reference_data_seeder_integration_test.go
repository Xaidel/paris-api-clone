//go:build integration

package db

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
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestReferenceDataSeederSeedUsesPostgresXmaxContract(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("set INTEGRATION=true to run integration tests")
	}

	adminDatabaseURL := os.Getenv("DATABASE_URL")
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if adminDatabaseURL == "" || migrationsPath == "" {
		t.Skip("DATABASE_URL and MIGRATIONS_PATH are required")
	}

	absMigrationsPath, err := resolveIntegrationMigrationsPath(migrationsPath)
	if err != nil {
		t.Fatalf("resolveIntegrationMigrationsPath(%q) error = %v", migrationsPath, err)
	}

	ctx := context.Background()
	tempDatabaseName := uniqueIntegrationDatabaseName()
	tempDatabaseURL, err := databaseURLWithName(adminDatabaseURL, tempDatabaseName)
	if err != nil {
		t.Fatalf("databaseURLWithName() error = %v", err)
	}

	createTemporaryIntegrationDatabase(t, ctx, adminDatabaseURL, tempDatabaseName)

	migrator := NewMigrator(config.DatabaseConfig{
		URL:            tempDatabaseURL,
		MigrationsPath: absMigrationsPath,
	})
	if err := migrator.Up(); err != nil {
		t.Fatalf("Migrator.Up() error = %v", err)
	}

	pool, err := NewPool(ctx, config.DatabaseConfig{
		URL:            tempDatabaseURL,
		MigrationsPath: absMigrationsPath,
		PingTimeout:    10 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		dropTemporaryIntegrationDatabase(t, ctx, adminDatabaseURL, tempDatabaseName)
	})

	if _, err := pool.Exec(ctx, "DELETE FROM u1_list"); err != nil {
		t.Fatalf("DELETE FROM u1_list error = %v", err)
	}

	if _, err := pool.Exec(ctx, "DELETE FROM exclusion_list"); err != nil {
		t.Fatalf("DELETE FROM exclusion_list error = %v", err)
	}

	seeder := NewReferenceDataSeeder(pool)

	firstResult, err := seeder.Seed(ctx)
	if err != nil {
		t.Fatalf("first Seed() error = %v", err)
	}

	if firstResult.ExclusionInserted != len(canonicalExclusionSeedEntries()) {
		t.Fatalf("first Seed() ExclusionInserted = %d, want %d", firstResult.ExclusionInserted, len(canonicalExclusionSeedEntries()))
	}

	if firstResult.ExclusionUpdated != 0 {
		t.Fatalf("first Seed() ExclusionUpdated = %d, want 0", firstResult.ExclusionUpdated)
	}

	if firstResult.U1Inserted != len(canonicalU1SeedEntries()) {
		t.Fatalf("first Seed() U1Inserted = %d, want %d", firstResult.U1Inserted, len(canonicalU1SeedEntries()))
	}

	if firstResult.U1Updated != 0 {
		t.Fatalf("first Seed() U1Updated = %d, want 0", firstResult.U1Updated)
	}

	exclusionEntry := canonicalExclusionSeedEntries()[0]
	u1Entry := canonicalU1SeedEntries()[0]

	if _, err := pool.Exec(ctx, `UPDATE exclusion_list SET activity_type = $2 WHERE id = $1`, exclusionEntry.ID, "Drifted exclusion activity type"); err != nil {
		t.Fatalf("UPDATE exclusion_list drift row error = %v", err)
	}

	if _, err := pool.Exec(
		ctx,
		`UPDATE u1_list
SET sector = $2,
    eligible_operation_type = $3,
    condition_guidance = $4
WHERE id = $1`,
		u1Entry.ID,
		"Drifted sector",
		"Drifted eligible operation type",
		"Drifted condition guidance",
	); err != nil {
		t.Fatalf("UPDATE u1_list drift row error = %v", err)
	}

	secondResult, err := seeder.Seed(ctx)
	if err != nil {
		t.Fatalf("second Seed() error = %v", err)
	}

	if secondResult.ExclusionInserted != 0 {
		t.Fatalf("second Seed() ExclusionInserted = %d, want 0", secondResult.ExclusionInserted)
	}

	if secondResult.ExclusionUpdated != len(canonicalExclusionSeedEntries()) {
		t.Fatalf("second Seed() ExclusionUpdated = %d, want %d", secondResult.ExclusionUpdated, len(canonicalExclusionSeedEntries()))
	}

	if secondResult.U1Inserted != 0 {
		t.Fatalf("second Seed() U1Inserted = %d, want 0", secondResult.U1Inserted)
	}

	if secondResult.U1Updated != len(canonicalU1SeedEntries()) {
		t.Fatalf("second Seed() U1Updated = %d, want %d", secondResult.U1Updated, len(canonicalU1SeedEntries()))
	}

	var restoredExclusionActivityType string
	if err := pool.QueryRow(ctx, `SELECT activity_type FROM exclusion_list WHERE id = $1`, exclusionEntry.ID).Scan(&restoredExclusionActivityType); err != nil {
		t.Fatalf("SELECT exclusion_list restored row error = %v", err)
	}

	if restoredExclusionActivityType != exclusionEntry.ActivityType {
		t.Fatalf("restored exclusion activity_type = %q, want %q", restoredExclusionActivityType, exclusionEntry.ActivityType)
	}

	var restoredU1Sector string
	var restoredU1EligibleOperationType string
	var restoredU1ConditionGuidance string
	if err := pool.QueryRow(
		ctx,
		`SELECT sector, eligible_operation_type, condition_guidance FROM u1_list WHERE id = $1`,
		u1Entry.ID,
	).Scan(&restoredU1Sector, &restoredU1EligibleOperationType, &restoredU1ConditionGuidance); err != nil {
		t.Fatalf("SELECT u1_list restored row error = %v", err)
	}

	if restoredU1Sector != u1Entry.Sector {
		t.Fatalf("restored u1 sector = %q, want %q", restoredU1Sector, u1Entry.Sector)
	}

	if restoredU1EligibleOperationType != u1Entry.EligibleOperationType {
		t.Fatalf("restored u1 eligible_operation_type = %q, want %q", restoredU1EligibleOperationType, u1Entry.EligibleOperationType)
	}

	if restoredU1ConditionGuidance != u1Entry.ConditionGuidance {
		t.Fatalf("restored u1 condition_guidance = %q, want %q", restoredU1ConditionGuidance, u1Entry.ConditionGuidance)
	}

	var exclusionCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM exclusion_list`).Scan(&exclusionCount); err != nil {
		t.Fatalf("SELECT COUNT(*) FROM exclusion_list error = %v", err)
	}

	if exclusionCount != len(canonicalExclusionSeedEntries()) {
		t.Fatalf("exclusion_list row count = %d, want %d", exclusionCount, len(canonicalExclusionSeedEntries()))
	}

	var u1Count int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM u1_list`).Scan(&u1Count); err != nil {
		t.Fatalf("SELECT COUNT(*) FROM u1_list error = %v", err)
	}

	if u1Count != len(canonicalU1SeedEntries()) {
		t.Fatalf("u1_list row count = %d, want %d", u1Count, len(canonicalU1SeedEntries()))
	}
}

func TestReferenceDataSeederSeedLeavesUnrelatedRowsUntouched(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("set INTEGRATION=true to run integration tests")
	}

	adminDatabaseURL := os.Getenv("DATABASE_URL")
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if adminDatabaseURL == "" || migrationsPath == "" {
		t.Skip("DATABASE_URL and MIGRATIONS_PATH are required")
	}

	absMigrationsPath, err := resolveIntegrationMigrationsPath(migrationsPath)
	if err != nil {
		t.Fatalf("resolveIntegrationMigrationsPath(%q) error = %v", migrationsPath, err)
	}

	ctx := context.Background()
	tempDatabaseName := uniqueIntegrationDatabaseName()
	tempDatabaseURL, err := databaseURLWithName(adminDatabaseURL, tempDatabaseName)
	if err != nil {
		t.Fatalf("databaseURLWithName() error = %v", err)
	}

	createTemporaryIntegrationDatabase(t, ctx, adminDatabaseURL, tempDatabaseName)

	migrator := NewMigrator(config.DatabaseConfig{
		URL:            tempDatabaseURL,
		MigrationsPath: absMigrationsPath,
	})
	if err := migrator.Up(); err != nil {
		t.Fatalf("Migrator.Up() error = %v", err)
	}

	pool, err := NewPool(ctx, config.DatabaseConfig{
		URL:            tempDatabaseURL,
		MigrationsPath: absMigrationsPath,
		PingTimeout:    10 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		dropTemporaryIntegrationDatabase(t, ctx, adminDatabaseURL, tempDatabaseName)
	})

	seeder := NewReferenceDataSeeder(pool)
	if _, err := seeder.Seed(ctx); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	const unrelatedExclusionID = "01962b8f-aeb2-7e03-a8ff-1edce1302999"
	const unrelatedExclusionActivityType = "Unrelated non-canonical activity"

	if _, err := pool.Exec(
		ctx,
		`INSERT INTO exclusion_list (id, activity_type, created_by, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())`,
		unrelatedExclusionID,
		unrelatedExclusionActivityType,
		seededReferenceDataCreatedBy,
	); err != nil {
		t.Fatalf("INSERT unrelated exclusion_list row error = %v", err)
	}

	if _, err := seeder.Seed(ctx); err != nil {
		t.Fatalf("second Seed() error = %v", err)
	}

	var activityType string
	if err := pool.QueryRow(ctx, `SELECT activity_type FROM exclusion_list WHERE id = $1`, unrelatedExclusionID).Scan(&activityType); err != nil {
		t.Fatalf("SELECT unrelated exclusion_list row error = %v", err)
	}

	if activityType != unrelatedExclusionActivityType {
		t.Fatalf("unrelated exclusion activity_type = %q, want %q", activityType, unrelatedExclusionActivityType)
	}

	var exclusionCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM exclusion_list`).Scan(&exclusionCount); err != nil {
		t.Fatalf("SELECT COUNT(*) FROM exclusion_list error = %v", err)
	}

	if exclusionCount != len(canonicalExclusionSeedEntries())+1 {
		t.Fatalf("exclusion_list row count = %d, want %d", exclusionCount, len(canonicalExclusionSeedEntries())+1)
	}
}

func resolveIntegrationMigrationsPath(path string) (string, error) {
	candidates := []string{path}
	if !filepath.IsAbs(path) {
		candidates = append(candidates, filepath.Join("..", "..", "..", path))
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

func uniqueIntegrationDatabaseName() string {
	return fmt.Sprintf("reference_data_seeder_test_%d", time.Now().UTC().UnixNano())
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
