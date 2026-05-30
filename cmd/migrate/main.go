package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/gyud-adb/paris-api/internal/infrastructure/config"
	infraDB "github.com/gyud-adb/paris-api/internal/infrastructure/db"
	"github.com/gyud-adb/paris-api/internal/infrastructure/observability"
	"go.uber.org/zap"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		panic(err)
	}
}

// run executes the migration command.
func run(args []string) error {
	action, err := parseAction(args)
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger, cleanupLogger, err := observability.NewLogger(cfg.Log)
	if err != nil {
		return err
	}
	defer func() {
		_ = cleanupLogger()
	}()

	switch action {
	case migrationActionSeedReferenceData:
		pool, err := infraDB.NewPool(context.Background(), cfg.Database)
		if err != nil {
			return err
		}
		defer pool.Close()

		result, err := infraDB.NewReferenceDataSeeder(pool).Seed(context.Background())
		if err != nil {
			return err
		}

		logger.Info(
			"reference data seeded",
			zap.Int("exclusion_inserted", result.ExclusionInserted),
			zap.Int("exclusion_updated", result.ExclusionUpdated),
			zap.Int("u1_inserted", result.U1Inserted),
			zap.Int("u1_updated", result.U1Updated),
		)

		return nil
	case migrationActionUp:
		migrator := infraDB.NewMigrator(cfg.Database)
		if err := migrator.Up(); err != nil {
			return err
		}
		logger.Info("migrations applied")
	case migrationActionDown:
		migrator := infraDB.NewMigrator(cfg.Database)
		if err := migrator.Down(); err != nil {
			return err
		}
		logger.Info("migrations rolled back")
	case migrationActionVersion:
		migrator := infraDB.NewMigrator(cfg.Database)
		version, dirty, err := migrator.Version()
		if err != nil {
			return err
		}
		logger.Info("migration version", zap.Uint("version", version), zap.Bool("dirty", dirty))
	default:
		return errors.New("unsupported migration action")
	}

	return nil
}

const (
	migrationActionUp                = "up"
	migrationActionDown              = "down"
	migrationActionSeedReferenceData = "seed-reference-data"
	migrationActionVersion           = "version"
)

// parseAction parses the migration action argument.
func parseAction(args []string) (string, error) {
	if len(args) == 0 {
		return migrationActionUp, nil
	}

	action := strings.ToLower(strings.TrimSpace(args[0]))
	if action == migrationActionUp || action == migrationActionDown || action == migrationActionSeedReferenceData || action == migrationActionVersion {
		return action, nil
	}

	return "", errors.New("migration action must be one of: up, down, seed-reference-data, version")
}
