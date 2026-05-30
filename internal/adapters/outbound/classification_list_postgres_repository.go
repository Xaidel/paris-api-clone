package adapters

import (
	"context"
	"fmt"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

const (
	listClassificationEntriesU1Query = `
SELECT id::text AS source_row_id,
       CONCAT('sector: ', sector, '; eligible_operation_type: ', eligible_operation_type, '; condition_guidance: ', condition_guidance) AS entry_text
FROM u1_list
ORDER BY sector ASC, eligible_operation_type ASC, id ASC
`
	listClassificationEntriesU2Query = `
SELECT id::text AS source_row_id,
       CONCAT('activity_type: ', activity_type) AS entry_text
FROM exclusion_list
ORDER BY activity_type ASC, id ASC
`
	listClassificationEntriesSectorQuery = `
SELECT id::text AS source_row_id,
       CONCAT('type: ', type, '; name: ', name, '; description: ', description) AS entry_text
FROM sector
ORDER BY type ASC, name ASC, id ASC
`
)

// PostgresClassificationListRepository loads classification list entries from PostgreSQL.
type PostgresClassificationListRepository struct {
	pool pgxQuerier
}

// NewPostgresClassificationListRepository builds a PostgresClassificationListRepository.
func NewPostgresClassificationListRepository(pool pgxQuerier) *PostgresClassificationListRepository {
	return &PostgresClassificationListRepository{pool: pool}
}

// GetEntries returns entry texts for the requested list type.
func (r *PostgresClassificationListRepository) GetEntries(ctx context.Context, listType valueobjects.ListType) ([]string, error) {
	entryDocuments, err := r.GetEntryDocuments(ctx, listType)
	if err != nil {
		return nil, err
	}

	entries := make([]string, 0, len(entryDocuments))
	for _, entryDocument := range entryDocuments {
		entries = append(entries, entryDocument.EntryText().String())
	}

	return entries, nil
}

// GetEntryDocuments returns canonical classification entry documents for the requested list type.
func (r *PostgresClassificationListRepository) GetEntryDocuments(ctx context.Context, listType valueobjects.ListType) ([]valueobjects.ClassificationListEntryDocument, error) {
	querier := txQuerierFromContext(ctx, r.pool)
	query, err := classificationListQueryForType(listType)
	if err != nil {
		return nil, fmt.Errorf("selecting %q classification list query: %w", listType.String(), err)
	}

	rows, err := querier.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying %s classification list entries: %w", listType.String(), err)
	}
	defer rows.Close()

	entryDocuments := make([]valueobjects.ClassificationListEntryDocument, 0)
	for rows.Next() {
		var (
			sourceRowID string
			entryText   string
		)
		if err := rows.Scan(&sourceRowID, &entryText); err != nil {
			return nil, fmt.Errorf("scanning %s classification list entry: %w", listType.String(), err)
		}

		canonicalEntryText, err := valueobjects.ClassificationListEntryTextFromString(entryText)
		if err != nil {
			return nil, fmt.Errorf("parsing %s classification list entry text: %w", listType.String(), err)
		}

		entryDocuments = append(entryDocuments, valueobjects.NewClassificationListEntryDocument(sourceRowID, canonicalEntryText))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating %s classification list entries: %w", listType.String(), err)
	}

	return entryDocuments, nil
}

func classificationListQueryForType(listType valueobjects.ListType) (string, error) {
	switch listType.String() {
	case valueobjects.U1ListType().String():
		return listClassificationEntriesU1Query, nil
	case valueobjects.U2ListType().String():
		return listClassificationEntriesU2Query, nil
	case valueobjects.SectorListType().String():
		return listClassificationEntriesSectorQuery, nil
	default:
		return "", fmt.Errorf("unsupported classification list type %q", listType.String())
	}
}
