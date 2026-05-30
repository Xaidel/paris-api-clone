package db

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/gyud-adb/paris-api/internal/infrastructure/config"
)

type migrationClientStub struct {
	upErr      error
	downErr    error
	version    uint
	dirty      bool
	versionErr error
	closed     bool
}

func (s *migrationClientStub) Up() error { return s.upErr }

func (s *migrationClientStub) Down() error { return s.downErr }

func (s *migrationClientStub) Version() (uint, bool, error) { return s.version, s.dirty, s.versionErr }

func (s *migrationClientStub) Close() (error, error) {
	s.closed = true
	return nil, nil
}

// TestNewMigrator verifies the new migrator behavior and the expected outcome asserted below.
func TestNewMigrator(t *testing.T) {
	t.Parallel()

	migrator := NewMigrator(config.DatabaseConfig{URL: "postgres://example", MigrationsPath: "internal/infrastructure/db/migrations"})
	if migrator == nil || migrator.factory == nil {
		t.Fatal("expected migrator with factory")
	}
}

// TestMigratorOperations verifies the migrator operations behavior and the expected outcome asserted below.
func TestMigratorOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*migrationClientStub)
		invoke    func(*Migrator) error
		assert    func(t *testing.T, stub *migrationClientStub, err error)
	}{
		{
			name: "up ignores no change",
			configure: func(stub *migrationClientStub) {
				stub.upErr = migrate.ErrNoChange
			},
			invoke: func(migrator *Migrator) error { return migrator.Up() },
			assert: func(t *testing.T, stub *migrationClientStub, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("Up() error = %v", err)
				}
				if !stub.closed {
					t.Fatal("expected client to be closed")
				}
			},
		},
		{
			name: "down wraps errors",
			configure: func(stub *migrationClientStub) {
				stub.downErr = errors.New("boom")
			},
			invoke: func(migrator *Migrator) error { return migrator.Down() },
			assert: func(t *testing.T, _ *migrationClientStub, err error) {
				t.Helper()
				if err == nil || !strings.Contains(err.Error(), "running down migrations") {
					t.Fatalf("unexpected error = %v", err)
				}
			},
		},
		{
			name: "version returns zero when none applied",
			configure: func(stub *migrationClientStub) {
				stub.versionErr = migrate.ErrNilVersion
			},
			invoke: func(migrator *Migrator) error {
				version, dirty, err := migrator.Version()
				if err != nil {
					return err
				}
				if version != 0 || dirty {
					return errors.New("expected zero version and clean state")
				}
				return nil
			},
			assert: func(t *testing.T, stub *migrationClientStub, err error) {
				t.Helper()
				if err != nil {
					t.Fatalf("Version() error = %v", err)
				}
				if !stub.closed {
					t.Fatal("expected client to be closed")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stub := &migrationClientStub{}
			tt.configure(stub)

			migrator := &Migrator{
				databaseURL:    "postgres://example",
				migrationsPath: "internal/infrastructure/db/migrations",
				factory: func(sourceURL, databaseURL string) (migrationClient, error) {
					return stub, nil
				},
			}

			err := tt.invoke(migrator)
			tt.assert(t, stub, err)
		})
	}
}

// TestMigrationSourceURL verifies the migration source url behavior and the expected outcome asserted below.
func TestMigrationSourceURL(t *testing.T) {
	t.Parallel()

	url, err := migrationSourceURL("internal/infrastructure/db/migrations")
	if err != nil {
		t.Fatalf("migrationSourceURL() error = %v", err)
	}

	if !strings.HasPrefix(url, "file://") {
		t.Fatalf("migration source url = %q", url)
	}
}

// TestMigrationDatabaseURL verifies the migration database url behavior and the expected outcome asserted below.
func TestMigrationDatabaseURL(t *testing.T) {
	t.Parallel()

	url, err := migrationDatabaseURL("postgres://admin:admin@localhost:5432/paris?sslmode=disable")
	if err != nil {
		t.Fatalf("migrationDatabaseURL() error = %v", err)
	}

	if !strings.HasPrefix(url, "pgx5://") {
		t.Fatalf("migration database url = %q", url)
	}
}

// TestPhase6EmbeddingCleanupMigrationFilesExist verifies the phase 6 embedding cleanup migration files exist and include the expected statements.
func TestPhase6EmbeddingCleanupMigrationFilesExist(t *testing.T) {
	t.Parallel()

	basePath := filepath.Join("internal", "infrastructure", "db", "migrations")
	if _, err := os.Stat(basePath); err != nil {
		basePath = filepath.Join("..", "..", "..", basePath)
	}

	upPath := filepath.Join(basePath, "000206_drop_unused_embedding_artifacts.up.sql")
	downPath := filepath.Join(basePath, "000206_drop_unused_embedding_artifacts.down.sql")

	upContent, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", upPath, err)
	}

	downContent, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", downPath, err)
	}

	upText := strings.ReplaceAll(string(upContent), "\r\n", "\n")
	downText := strings.ReplaceAll(string(downContent), "\r\n", "\n")

	upExpected := []string{
		"DROP INDEX IF EXISTS idx_transaction_description_embeddings_embedding_hnsw;",
		"DROP INDEX IF EXISTS idx_transaction_description_embeddings_exact_lookup;",
		"DROP TABLE IF EXISTS transaction_description_embeddings;",
		"DROP INDEX IF EXISTS idx_classification_entry_embedding_sector_hnsw;",
		"DROP INDEX IF EXISTS idx_classification_entry_embedding_u2_hnsw;",
		"DROP INDEX IF EXISTS idx_classification_entry_embedding_u1_hnsw;",
		"DROP INDEX IF EXISTS idx_classification_entry_embedding_model_active;",
		"DROP TABLE IF EXISTS classification_entry_embedding;",
		"DROP INDEX IF EXISTS idx_classification_entry_list_type_active;",
		"DROP TABLE IF EXISTS classification_entry;",
		"DROP INDEX IF EXISTS idx_list_entry_embeddings_embedding_hnsw;",
		"DROP TABLE IF EXISTS list_entry_embeddings;",
	}

	for _, want := range upExpected {
		if !strings.Contains(upText, want) {
			t.Errorf("up migration missing %q", want)
		}
	}

	downExpected := []string{
		"SELECT 1 FROM pg_available_extensions WHERE name = 'vector'",
		"CREATE EXTENSION IF NOT EXISTS vector",
		"CREATE TABLE IF NOT EXISTS list_entry_embeddings",
		"CREATE INDEX IF NOT EXISTS idx_list_entry_embeddings_embedding_hnsw",
		"CREATE TABLE IF NOT EXISTS classification_entry_embedding",
		"CREATE TABLE IF NOT EXISTS classification_entry",
		"CREATE INDEX IF NOT EXISTS idx_classification_entry_list_type_active",
		"CREATE INDEX IF NOT EXISTS idx_classification_entry_embedding_model_active",
		"CREATE INDEX IF NOT EXISTS idx_classification_entry_embedding_u1_hnsw",
		"CREATE INDEX IF NOT EXISTS idx_classification_entry_embedding_u2_hnsw",
		"CREATE INDEX IF NOT EXISTS idx_classification_entry_embedding_sector_hnsw",
		"CREATE TABLE IF NOT EXISTS transaction_description_embeddings",
		"CREATE INDEX IF NOT EXISTS idx_transaction_description_embeddings_exact_lookup",
		"CREATE INDEX IF NOT EXISTS idx_transaction_description_embeddings_embedding_hnsw",
	}

	for _, want := range downExpected {
		if !strings.Contains(downText, want) {
			t.Errorf("down migration missing %q", want)
		}
	}

	listEntryBaseIndex := strings.Index(downText, "CREATE TABLE IF NOT EXISTS list_entry_embeddings (\n    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),")
	if listEntryBaseIndex == -1 {
		t.Fatalf("down migration missing base list_entry_embeddings restore block")
	}

	guardIndex := strings.Index(downText, "IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'vector') THEN")
	if guardIndex == -1 {
		t.Fatalf("down migration missing vector-availability guard")
	}

	if listEntryBaseIndex > guardIndex {
		t.Errorf("down migration should restore base list_entry_embeddings table before the vector guard")
	}

	guardedFragments := []string{
		"CREATE EXTENSION IF NOT EXISTS vector",
		"ALTER TABLE list_entry_embeddings\n    ADD COLUMN IF NOT EXISTS embedding vector(1536)",
		"CREATE INDEX IF NOT EXISTS idx_list_entry_embeddings_embedding_hnsw\n    ON list_entry_embeddings USING hnsw (embedding vector_cosine_ops)",
		"CREATE TABLE IF NOT EXISTS classification_entry (\n    entry_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),",
		"CREATE TABLE IF NOT EXISTS classification_entry_embedding (\n    embedding_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),",
		"CREATE TABLE IF NOT EXISTS transaction_description_embeddings (\n    transaction_id UUID NOT NULL REFERENCES transactions (id) ON DELETE CASCADE,",
	}

	previousIndex := guardIndex
	for _, fragment := range guardedFragments {
		currentIndex := strings.Index(downText, fragment)
		if currentIndex == -1 {
			t.Fatalf("down migration missing guarded fragment %q", fragment)
		}
		if currentIndex < guardIndex {
			t.Errorf("down migration fragment %q should be inside the vector guard", fragment)
		}
		if currentIndex < previousIndex {
			t.Errorf("down migration fragment %q appears before a required predecessor", fragment)
		}
		previousIndex = currentIndex
	}

	if strings.Contains(downText, "SELECT 1 FROM pg_extension WHERE extname = 'vector'") {
		t.Errorf("down migration should check pg_available_extensions and install vector when available")
	}
}

// TestTransactionUploadStatusMigrationFilesExist verifies the transaction upload status migration files exist and include the expected statements.
func TestTransactionUploadStatusMigrationFilesExist(t *testing.T) {
	t.Parallel()

	basePath := filepath.Join("internal", "infrastructure", "db", "migrations")
	if _, err := os.Stat(basePath); err != nil {
		basePath = filepath.Join("..", "..", "..", basePath)
	}

	upPath := filepath.Join(basePath, "000207_add_transaction_upload_status.up.sql")
	downPath := filepath.Join(basePath, "000207_add_transaction_upload_status.down.sql")

	upContent, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", upPath, err)
	}

	downContent, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", downPath, err)
	}

	upText := strings.ReplaceAll(string(upContent), "\r\n", "\n")
	downText := strings.ReplaceAll(string(downContent), "\r\n", "\n")

	upExpected := []string{
		"ALTER TABLE transaction_upload ADD COLUMN status TEXT",
		"UPDATE transaction_upload SET status = 'uploaded' WHERE status IS NULL",
		"ALTER TABLE transaction_upload ALTER COLUMN status SET NOT NULL",
		"ADD CONSTRAINT chk_transaction_upload_status",
		"CHECK (status IN ('uploaded', 'failed'))",
	}

	for _, want := range upExpected {
		if !strings.Contains(upText, want) {
			t.Errorf("up migration missing %q", want)
		}
	}

	upOrdered := []string{
		"ALTER TABLE transaction_upload ADD COLUMN status TEXT",
		"UPDATE transaction_upload SET status = 'uploaded' WHERE status IS NULL",
		"ALTER TABLE transaction_upload ALTER COLUMN status SET NOT NULL",
		"ADD CONSTRAINT chk_transaction_upload_status",
	}

	previousIndex := -1
	for _, fragment := range upOrdered {
		currentIndex := strings.Index(upText, fragment)
		if currentIndex == -1 {
			t.Fatalf("up migration missing ordered fragment %q", fragment)
		}
		if currentIndex < previousIndex {
			t.Errorf("up migration fragment %q appears before a required predecessor", fragment)
		}
		previousIndex = currentIndex
	}

	downExpected := []string{
		"ALTER TABLE transaction_upload DROP CONSTRAINT IF EXISTS chk_transaction_upload_status",
		"ALTER TABLE transaction_upload DROP COLUMN IF EXISTS status",
	}

	for _, want := range downExpected {
		if !strings.Contains(downText, want) {
			t.Errorf("down migration missing %q", want)
		}
	}
}

func TestTransactionUploadPreviewAndGroupIDMigrationFilesExist(t *testing.T) {
	t.Parallel()

	basePath := filepath.Join("internal", "infrastructure", "db", "migrations")
	if _, err := os.Stat(basePath); err != nil {
		basePath = filepath.Join("..", "..", "..", basePath)
	}

	upPath := filepath.Join(basePath, "000208_create_transaction_upload_preview.up.sql")
	downPath := filepath.Join(basePath, "000208_create_transaction_upload_preview.down.sql")

	upContent, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", upPath, err)
	}

	downContent, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", downPath, err)
	}

	upText := strings.ReplaceAll(string(upContent), "\r\n", "\n")
	downText := strings.ReplaceAll(string(downContent), "\r\n", "\n")

	upExpected := []string{
		"ALTER TABLE transaction_upload ADD COLUMN group_id TEXT",
		"UPDATE transaction_upload SET group_id = (SELECT id FROM user_group WHERE name = 'superadmin') WHERE group_id IS NULL",
		"ALTER TABLE transaction_upload ALTER COLUMN group_id SET NOT NULL",
		"REFERENCES user_group (id)",
		"CREATE INDEX IF NOT EXISTS idx_transaction_upload_group_id ON transaction_upload (group_id)",
		"CREATE TABLE IF NOT EXISTS transaction_upload_preview",
	}

	for _, want := range upExpected {
		if !strings.Contains(upText, want) {
			t.Errorf("up migration missing %q", want)
		}
	}

	upOrdered := []string{
		"ALTER TABLE transaction_upload ADD COLUMN group_id TEXT",
		"UPDATE transaction_upload SET group_id = (SELECT id FROM user_group WHERE name = 'superadmin') WHERE group_id IS NULL",
		"ALTER TABLE transaction_upload ALTER COLUMN group_id SET NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_transaction_upload_group_id ON transaction_upload (group_id)",
		"CREATE TABLE IF NOT EXISTS transaction_upload_preview",
	}

	previousIndex := -1
	for _, fragment := range upOrdered {
		currentIndex := strings.Index(upText, fragment)
		if currentIndex == -1 {
			t.Fatalf("up migration missing ordered fragment %q", fragment)
		}
		if currentIndex < previousIndex {
			t.Errorf("up migration fragment %q appears before a required predecessor", fragment)
		}
		previousIndex = currentIndex
	}

	downExpected := []string{
		"DROP TABLE IF EXISTS transaction_upload_preview",
		"DROP INDEX IF EXISTS idx_transaction_upload_group_id",
		"ALTER TABLE transaction_upload DROP COLUMN IF EXISTS group_id",
	}

	for _, want := range downExpected {
		if !strings.Contains(downText, want) {
			t.Errorf("down migration missing %q", want)
		}
	}
}

// TestTransactionUploadGroupMigrationFilesExist verifies the follow-up transaction upload group migration files exist and include the expected statements.
func TestTransactionUploadGroupMigrationFilesExist(t *testing.T) {
	t.Parallel()

	basePath := filepath.Join("internal", "infrastructure", "db", "migrations")
	if _, err := os.Stat(basePath); err != nil {
		basePath = filepath.Join("..", "..", "..", basePath)
	}

	upPath := filepath.Join(basePath, "000209_add_transaction_upload_group_and_download_support.up.sql")
	downPath := filepath.Join(basePath, "000209_add_transaction_upload_group_and_download_support.down.sql")

	upContent, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", upPath, err)
	}

	downContent, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", downPath, err)
	}

	upText := strings.ReplaceAll(string(upContent), "\r\n", "\n")
	downText := strings.ReplaceAll(string(downContent), "\r\n", "\n")

	upExpected := []string{
		"DROP CONSTRAINT IF EXISTS transaction_upload_content_md5_key",
		"-- Legacy uploads that still cannot infer ownership remain assigned",
		"-- to the seeded superadmin group set by 000208.",
		"UPDATE transaction_upload tu",
		"SET group_id = up.group_id",
		"WHERE tu.id = up.upload_id",
		"AND tu.group_id = '01962b8f-aeb2-7e03-a8ff-1edce1300001'",
		"ADD CONSTRAINT uq_transaction_upload_group_content_md5",
		"UNIQUE (group_id, content_md5)",
	}

	for _, want := range upExpected {
		if !strings.Contains(upText, want) {
			t.Errorf("up migration missing %q", want)
		}
	}

	upOrdered := []string{
		"DROP CONSTRAINT IF EXISTS transaction_upload_content_md5_key",
		"UPDATE transaction_upload tu",
		"AND tu.group_id = '01962b8f-aeb2-7e03-a8ff-1edce1300001'",
		"ADD CONSTRAINT uq_transaction_upload_group_content_md5",
	}

	previousIndex := -1
	for _, fragment := range upOrdered {
		currentIndex := strings.Index(upText, fragment)
		if currentIndex == -1 {
			t.Fatalf("up migration missing ordered fragment %q", fragment)
		}
		if currentIndex < previousIndex {
			t.Errorf("up migration fragment %q appears before a required predecessor", fragment)
		}
		previousIndex = currentIndex
	}

	fallbackCommentIndex := strings.Index(upText, "-- Legacy uploads that still cannot infer ownership remain assigned")
	if fallbackCommentIndex == -1 {
		t.Fatal("up migration missing fallback audit comment")
	}

	downExpected := []string{
		"DO $$",
		"HAVING COUNT(*) > 1",
		"RAISE EXCEPTION 'cannot rollback 000209: duplicate transaction_upload.content_md5 values exist'",
		"ALTER TABLE transaction_upload DROP CONSTRAINT IF EXISTS uq_transaction_upload_group_content_md5",
		"ADD CONSTRAINT transaction_upload_content_md5_key UNIQUE (content_md5)",
	}

	for _, want := range downExpected {
		if !strings.Contains(downText, want) {
			t.Errorf("down migration missing %q", want)
		}
	}

	downOrdered := []string{
		"DO $$",
		"RAISE EXCEPTION 'cannot rollback 000209: duplicate transaction_upload.content_md5 values exist'",
		"ALTER TABLE transaction_upload DROP CONSTRAINT IF EXISTS uq_transaction_upload_group_content_md5",
		"ADD CONSTRAINT transaction_upload_content_md5_key UNIQUE (content_md5)",
	}

	previousIndex = -1
	for _, fragment := range downOrdered {
		currentIndex := strings.Index(downText, fragment)
		if currentIndex == -1 {
			t.Fatalf("down migration missing ordered fragment %q", fragment)
		}
		if currentIndex < previousIndex {
			t.Errorf("down migration fragment %q appears before a required predecessor", fragment)
		}
		previousIndex = currentIndex
	}
}
