package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	// PostgreSQL reports INSERT for both insert and conflict-update upserts, so the
	// seeder returns `xmax = 0` to distinguish freshly inserted rows from updates.
	upsertExclusionListEntryQuery = `
INSERT INTO exclusion_list (id, activity_type, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE
SET activity_type = EXCLUDED.activity_type,
    updated_at = EXCLUDED.updated_at
RETURNING (xmax = 0) AS inserted
`
	upsertU1ListEntryQuery = `
INSERT INTO u1_list (id, sector, eligible_operation_type, condition_guidance, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE
SET sector = EXCLUDED.sector,
    eligible_operation_type = EXCLUDED.eligible_operation_type,
    condition_guidance = EXCLUDED.condition_guidance,
    updated_at = EXCLUDED.updated_at
RETURNING (xmax = 0) AS inserted
`
)

type referenceDataTxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// ReferenceDataSeedResult reports how many canonical reference rows took each upsert path.
type ReferenceDataSeedResult struct {
	// ExclusionInserted counts exclusion rows inserted by the seed run.
	ExclusionInserted int
	// ExclusionUpdated counts exclusion rows that hit ON CONFLICT DO UPDATE.
	// This means the row already existed; it does not imply any business field changed.
	ExclusionUpdated int
	// U1Inserted counts U1 rows inserted by the seed run.
	U1Inserted int
	// U1Updated counts U1 rows that hit ON CONFLICT DO UPDATE.
	// This means the row already existed; it does not imply any business field changed.
	U1Updated int
}

// ReferenceDataSeeder upserts the canonical exclusion and U1 reference datasets.
type ReferenceDataSeeder struct {
	beginner referenceDataTxBeginner
}

// NewReferenceDataSeeder builds a ReferenceDataSeeder backed by a transactional beginner.
func NewReferenceDataSeeder(beginner referenceDataTxBeginner) *ReferenceDataSeeder {
	return &ReferenceDataSeeder{beginner: beginner}
}

// Seed upserts the canonical exclusion and U1 reference rows in a single transaction.
func (s *ReferenceDataSeeder) Seed(ctx context.Context) (ReferenceDataSeedResult, error) {
	tx, err := s.beginner.Begin(ctx)
	if err != nil {
		return ReferenceDataSeedResult{}, fmt.Errorf("beginning reference data seed transaction: %w", err)
	}

	now := time.Now().UTC()
	result, err := seedReferenceDataTx(ctx, tx, now)
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
			return ReferenceDataSeedResult{}, fmt.Errorf("rolling back reference data seed transaction after seed failure: %w: %w", err, rollbackErr)
		}

		return ReferenceDataSeedResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return ReferenceDataSeedResult{}, fmt.Errorf("committing reference data seed transaction: %w", err)
	}

	return result, nil
}

func seedReferenceDataTx(ctx context.Context, tx pgx.Tx, now time.Time) (ReferenceDataSeedResult, error) {
	result, err := seedExclusionEntries(ctx, tx, now)
	if err != nil {
		return ReferenceDataSeedResult{}, err
	}

	u1Result, err := seedU1Entries(ctx, tx, now)
	if err != nil {
		return ReferenceDataSeedResult{}, err
	}

	result.U1Inserted = u1Result.U1Inserted
	result.U1Updated = u1Result.U1Updated

	return result, nil
}

func seedExclusionEntries(ctx context.Context, tx pgx.Tx, now time.Time) (ReferenceDataSeedResult, error) {
	result := ReferenceDataSeedResult{}

	for _, entry := range canonicalExclusionSeedEntries() {
		var inserted bool
		if err := tx.QueryRow(ctx, upsertExclusionListEntryQuery, entry.ID, entry.ActivityType, seededReferenceDataCreatedBy, now, now).Scan(&inserted); err != nil {
			return ReferenceDataSeedResult{}, fmt.Errorf("upserting exclusion list entry %q: %w", entry.ActivityType, err)
		}

		if inserted {
			result.ExclusionInserted++
		} else {
			result.ExclusionUpdated++
		}
	}

	return result, nil
}

func seedU1Entries(ctx context.Context, tx pgx.Tx, now time.Time) (ReferenceDataSeedResult, error) {
	result := ReferenceDataSeedResult{}

	for _, entry := range canonicalU1SeedEntries() {
		var inserted bool
		if err := tx.QueryRow(ctx, upsertU1ListEntryQuery, entry.ID, entry.Sector, entry.EligibleOperationType, entry.ConditionGuidance, seededReferenceDataCreatedBy, now, now).Scan(&inserted); err != nil {
			return ReferenceDataSeedResult{}, fmt.Errorf("upserting u1 list entry %q: %w", entry.EligibleOperationType, err)
		}

		if inserted {
			result.U1Inserted++
		} else {
			result.U1Updated++
		}
	}

	return result, nil
}
