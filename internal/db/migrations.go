package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Migration struct {
	Version int
	Name    string
	SQL     string
}

var migrations = []Migration{
	{
		Version: 1,
		Name:    "create_core_schema",
		SQL: `
CREATE TABLE providers (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  code TEXT NOT NULL UNIQUE,
  is_enabled INTEGER NOT NULL DEFAULT 1,
  rotation_strategy TEXT NOT NULL DEFAULT 'ROUND_ROBIN',
  round_robin_cursor INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE models (
  id TEXT PRIMARY KEY,
  provider_id TEXT NOT NULL,
  name TEXT NOT NULL,
  code TEXT NOT NULL,
  is_enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE CASCADE,
  UNIQUE (provider_id, code)
);

CREATE TABLE api_keys (
  id TEXT PRIMARY KEY,
  provider_id TEXT NOT NULL,
  alias TEXT NOT NULL,
  secret_value TEXT NOT NULL,
  is_enabled INTEGER NOT NULL DEFAULT 1,
  is_available INTEGER NOT NULL DEFAULT 1,
  sort_order INTEGER NOT NULL DEFAULT 0,
  failure_count INTEGER NOT NULL DEFAULT 0,
  last_failed_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE CASCADE,
  UNIQUE (provider_id, alias)
);

CREATE TABLE channels (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  code TEXT NOT NULL UNIQUE,
  default_permission_mode TEXT NOT NULL DEFAULT 'DENY',
  user_management_mode TEXT NOT NULL DEFAULT 'EXTERNAL_MANAGED',
  is_enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE users (
  id TEXT PRIMARY KEY,
  channel_id TEXT NOT NULL,
  external_user_id TEXT NOT NULL,
  display_name TEXT NOT NULL,
  is_enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
  UNIQUE (channel_id, external_user_id)
);

CREATE TABLE hosted_user_credentials (
  user_id TEXT PRIMARY KEY,
  password_hash TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE channel_permission_defaults (
  id TEXT PRIMARY KEY,
  channel_id TEXT NOT NULL,
  provider_id TEXT NOT NULL,
  model_id TEXT NOT NULL,
  default_allowed INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
  FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE CASCADE,
  FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
  UNIQUE (channel_id, provider_id, model_id)
);

CREATE TABLE user_permissions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  provider_id TEXT NOT NULL,
  model_id TEXT NOT NULL,
  allowed INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE CASCADE,
  FOREIGN KEY (model_id) REFERENCES models(id) ON DELETE CASCADE,
  UNIQUE (user_id, provider_id, model_id)
);

CREATE TABLE user_key_permissions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  provider_id TEXT NOT NULL,
  key_id TEXT NOT NULL,
  allowed INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE CASCADE,
  FOREIGN KEY (key_id) REFERENCES api_keys(id) ON DELETE CASCADE,
  UNIQUE (user_id, provider_id, key_id)
);

CREATE TABLE dispatch_logs (
  id TEXT PRIMARY KEY,
  channel_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  provider_id TEXT NOT NULL,
  model_id TEXT NOT NULL,
  key_id TEXT NOT NULL,
  key_alias_snapshot TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'DISPATCHED',
  created_at TEXT NOT NULL,
  FOREIGN KEY (channel_id) REFERENCES channels(id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (provider_id) REFERENCES providers(id),
  FOREIGN KEY (model_id) REFERENCES models(id),
  FOREIGN KEY (key_id) REFERENCES api_keys(id)
);

CREATE TABLE failure_reports (
  id TEXT PRIMARY KEY,
  dispatch_log_id TEXT NOT NULL,
  channel_id TEXT NOT NULL,
  user_id TEXT NOT NULL,
  provider_id TEXT NOT NULL,
  model_id TEXT NOT NULL,
  key_id TEXT NOT NULL,
  key_alias_snapshot TEXT NOT NULL,
  error_code TEXT,
  error_message TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY (dispatch_log_id) REFERENCES dispatch_logs(id),
  FOREIGN KEY (channel_id) REFERENCES channels(id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (provider_id) REFERENCES providers(id),
  FOREIGN KEY (model_id) REFERENCES models(id),
  FOREIGN KEY (key_id) REFERENCES api_keys(id)
);

CREATE TABLE admin_sessions (
  id TEXT PRIMARY KEY,
  session_token_hash TEXT NOT NULL UNIQUE,
  expires_at TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE INDEX idx_models_provider_id ON models(provider_id);
CREATE INDEX idx_api_keys_provider_id ON api_keys(provider_id);
CREATE INDEX idx_api_keys_provider_available ON api_keys(provider_id, is_enabled, is_available, sort_order);
CREATE UNIQUE INDEX idx_channels_name_unique ON channels(name);
CREATE INDEX idx_users_channel_id ON users(channel_id);
CREATE INDEX idx_users_channel_external ON users(channel_id, external_user_id);
CREATE INDEX idx_hosted_user_credentials_user ON hosted_user_credentials(user_id);
CREATE INDEX idx_user_permissions_user ON user_permissions(user_id);
CREATE INDEX idx_user_key_permissions_user_provider ON user_key_permissions(user_id, provider_id);
CREATE INDEX idx_channel_permission_defaults_channel ON channel_permission_defaults(channel_id);
CREATE INDEX idx_dispatch_logs_created_at ON dispatch_logs(created_at);
CREATE INDEX idx_dispatch_logs_user ON dispatch_logs(user_id);
CREATE INDEX idx_dispatch_logs_channel ON dispatch_logs(channel_id);
CREATE INDEX idx_dispatch_logs_provider_model ON dispatch_logs(provider_id, model_id);
CREATE INDEX idx_dispatch_logs_key ON dispatch_logs(key_id);
CREATE INDEX idx_failure_reports_dispatch_log ON failure_reports(dispatch_log_id);
CREATE INDEX idx_admin_sessions_expires_at ON admin_sessions(expires_at);
`,
	},
	{
		Version: 2,
		Name:    "add_user_key_permissions",
		SQL: `
CREATE TABLE IF NOT EXISTS user_key_permissions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  provider_id TEXT NOT NULL,
  key_id TEXT NOT NULL,
  allowed INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE CASCADE,
  FOREIGN KEY (key_id) REFERENCES api_keys(id) ON DELETE CASCADE,
  UNIQUE (user_id, provider_id, key_id)
);

CREATE INDEX IF NOT EXISTS idx_user_key_permissions_user_provider ON user_key_permissions(user_id, provider_id);
`,
	},
	{
		Version: 3,
		Name:    "add_hosted_user_credentials",
		SQL: `
CREATE TABLE IF NOT EXISTS hosted_user_credentials (
  user_id TEXT PRIMARY KEY,
  password_hash TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_hosted_user_credentials_user ON hosted_user_credentials(user_id);
`,
	},
	{
		Version: 4,
		Name:    "add_unique_channel_names",
		SQL: `
CREATE UNIQUE INDEX IF NOT EXISTS idx_channels_name_unique ON channels(name);
`,
	},
}

func (database *DB) Migrate(ctx context.Context) error {
	tx, err := database.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, migration := range migrations {
		applied, err := migrationApplied(ctx, tx, migration.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			return fmt.Errorf("apply migration %d %s: %w", migration.Version, migration.Name, err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO schema_migrations (version, name) VALUES (?, ?);
`, migration.Version, migration.Name); err != nil {
			return fmt.Errorf("record migration %d %s: %w", migration.Version, migration.Name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}
	return nil
}

func migrationApplied(ctx context.Context, tx *sql.Tx, version int) (bool, error) {
	var exists int
	if err := tx.QueryRowContext(ctx, `
SELECT 1 FROM schema_migrations WHERE version = ?;
`, version).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("check migration %d: %w", version, err)
	}
	return exists == 1, nil
}
