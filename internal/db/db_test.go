package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestOpenConfiguresSQLite(t *testing.T) {
	database := openTestDB(t)
	defer database.Close()

	assertPragma(t, database.SQL(), "foreign_keys", "1")
	assertPragma(t, database.SQL(), "busy_timeout", "5000")

	var journalMode string
	if err := database.SQL().QueryRowContext(context.Background(), "PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Fatalf("journal_mode = %q, want wal", journalMode)
	}
}

func TestMigrateCreatesCoreSchemaAndIsIdempotent(t *testing.T) {
	database := openTestDB(t)
	defer database.Close()

	if err := database.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate() first run error = %v", err)
	}
	if err := database.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate() second run error = %v", err)
	}

	expectedTables := []string{
		"schema_migrations",
		"providers",
		"models",
		"api_keys",
		"channels",
		"users",
		"channel_permission_defaults",
		"user_permissions",
		"dispatch_logs",
		"failure_reports",
		"admin_sessions",
	}
	for _, table := range expectedTables {
		if !tableExists(t, database.SQL(), table) {
			t.Fatalf("table %s does not exist", table)
		}
	}

	var appliedCount int
	if err := database.SQL().QueryRowContext(context.Background(), "SELECT COUNT(*) FROM schema_migrations").Scan(&appliedCount); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if appliedCount != 1 {
		t.Fatalf("applied migrations = %d, want 1", appliedCount)
	}
}

func TestMigrateEnforcesForeignKeys(t *testing.T) {
	database := openTestDB(t)
	defer database.Close()

	if err := database.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	_, err := database.SQL().ExecContext(context.Background(), `
INSERT INTO models (id, provider_id, name, code, created_at, updated_at)
VALUES ('model_1', 'missing_provider', 'Model 1', 'model-1', '2026-05-06T00:00:00Z', '2026-05-06T00:00:00Z');
`)
	if err == nil {
		t.Fatal("insert model with missing provider succeeded, want foreign key error")
	}
}

func openTestDB(t *testing.T) *DB {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.db")
	database, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return database
}

func assertPragma(t *testing.T, sqlDB *sql.DB, name string, want string) {
	t.Helper()

	var got string
	if err := sqlDB.QueryRowContext(context.Background(), "PRAGMA "+name).Scan(&got); err != nil {
		t.Fatalf("query pragma %s: %v", name, err)
	}
	if got != want {
		t.Fatalf("pragma %s = %q, want %q", name, got, want)
	}
}

func tableExists(t *testing.T, sqlDB *sql.DB, table string) bool {
	t.Helper()

	var name string
	err := sqlDB.QueryRowContext(context.Background(), `
SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?;
`, table).Scan(&name)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		t.Fatalf("check table %s: %v", table, err)
	}
	return name == table
}
