package admin

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"
)

type Store struct {
	db  *sql.DB
	now func() time.Time
}

type Options struct {
	DB  *sql.DB
	Now func() time.Time
}

type Provider struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Code             string `json:"code"`
	IsEnabled        bool   `json:"isEnabled"`
	RotationStrategy string `json:"rotationStrategy"`
	CreatedAt        string `json:"createdAt"`
	UpdatedAt        string `json:"updatedAt"`
}

type Model struct {
	ID         string `json:"id"`
	ProviderID string `json:"providerId"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	IsEnabled  bool   `json:"isEnabled"`
}

type APIKey struct {
	ID           string  `json:"id"`
	ProviderID   string  `json:"providerId"`
	Alias        string  `json:"alias"`
	MaskedValue  string  `json:"maskedValue"`
	IsEnabled    bool    `json:"isEnabled"`
	IsAvailable  bool    `json:"isAvailable"`
	SortOrder    int     `json:"sortOrder"`
	FailureCount int     `json:"failureCount"`
	LastFailedAt *string `json:"lastFailedAt"`
}

type CreateProviderInput struct {
	Name             string
	Code             string
	IsEnabled        bool
	RotationStrategy string
}

type CreateModelInput struct {
	ProviderID string
	Name       string
	Code       string
	IsEnabled  bool
}

type CreateAPIKeyInput struct {
	ProviderID  string
	Alias       string
	SecretValue string
	IsEnabled   bool
	IsAvailable bool
	SortOrder   int
}

type UpdateProviderInput struct {
	Name             string
	Code             string
	IsEnabled        bool
	RotationStrategy string
}

type UpdateModelInput struct {
	Name      string
	Code      string
	IsEnabled bool
}

type UpdateAPIKeyInput struct {
	Alias       string
	SecretValue string
	IsEnabled   bool
	IsAvailable bool
	SortOrder   int
}

func NewStore(options Options) (*Store, error) {
	if options.DB == nil {
		return nil, fmt.Errorf("admin database is required")
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Store{db: options.DB, now: options.Now}, nil
}

func (store *Store) ListProviders(ctx context.Context) ([]Provider, error) {
	rows, err := store.db.QueryContext(ctx, `
SELECT id, name, code, is_enabled, rotation_strategy, created_at, updated_at
FROM providers
ORDER BY created_at DESC, name ASC;
`)
	if err != nil {
		return nil, fmt.Errorf("list providers: %w", err)
	}
	defer rows.Close()

	var providers []Provider
	for rows.Next() {
		var provider Provider
		var isEnabled int
		if err := rows.Scan(&provider.ID, &provider.Name, &provider.Code, &isEnabled, &provider.RotationStrategy, &provider.CreatedAt, &provider.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan provider: %w", err)
		}
		provider.IsEnabled = isEnabled == 1
		providers = append(providers, provider)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate providers: %w", err)
	}
	return providers, nil
}

func (store *Store) CreateProvider(ctx context.Context, input CreateProviderInput) (Provider, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Code = strings.TrimSpace(input.Code)
	if input.Code == "" {
		input.Code = input.Name
	}
	if input.Name == "" {
		return Provider{}, fmt.Errorf("provider name is required")
	}
	if input.RotationStrategy == "" {
		input.RotationStrategy = "ROUND_ROBIN"
	}
	if input.RotationStrategy != "ROUND_ROBIN" && input.RotationStrategy != "STICKY_FIRST_AVAILABLE" {
		return Provider{}, fmt.Errorf("invalid rotation strategy")
	}

	now := formatTime(store.now())
	provider := Provider{
		ID:               newID("provider"),
		Name:             input.Name,
		Code:             input.Code,
		IsEnabled:        input.IsEnabled,
		RotationStrategy: input.RotationStrategy,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if _, err := store.db.ExecContext(ctx, `
INSERT INTO providers (id, name, code, is_enabled, rotation_strategy, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);
`, provider.ID, provider.Name, provider.Code, boolToInt(provider.IsEnabled), provider.RotationStrategy, provider.CreatedAt, provider.UpdatedAt); err != nil {
		return Provider{}, fmt.Errorf("create provider: %w", err)
	}
	return provider, nil
}

func (store *Store) UpdateProvider(ctx context.Context, id string, input UpdateProviderInput) (Provider, error) {
	id = strings.TrimSpace(id)
	input.Name = strings.TrimSpace(input.Name)
	input.Code = strings.TrimSpace(input.Code)
	if input.Code == "" {
		input.Code = input.Name
	}
	if id == "" || input.Name == "" {
		return Provider{}, fmt.Errorf("provider id and name are required")
	}
	if input.RotationStrategy != "ROUND_ROBIN" && input.RotationStrategy != "STICKY_FIRST_AVAILABLE" {
		return Provider{}, fmt.Errorf("invalid rotation strategy")
	}

	updatedAt := formatTime(store.now())
	result, err := store.db.ExecContext(ctx, `
UPDATE providers
SET name = ?, code = ?, is_enabled = ?, rotation_strategy = ?, updated_at = ?
WHERE id = ?;
`, input.Name, input.Code, boolToInt(input.IsEnabled), input.RotationStrategy, updatedAt, id)
	if err != nil {
		return Provider{}, fmt.Errorf("update provider: %w", err)
	}
	if rows, err := result.RowsAffected(); err != nil || rows == 0 {
		if err != nil {
			return Provider{}, fmt.Errorf("update provider rows affected: %w", err)
		}
		return Provider{}, fmt.Errorf("provider not found")
	}

	return Provider{
		ID:               id,
		Name:             input.Name,
		Code:             input.Code,
		IsEnabled:        input.IsEnabled,
		RotationStrategy: input.RotationStrategy,
		UpdatedAt:        updatedAt,
	}, nil
}

func (store *Store) DeleteProvider(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("provider id is required")
	}
	if _, err := store.db.ExecContext(ctx, `DELETE FROM providers WHERE id = ?;`, id); err != nil {
		return fmt.Errorf("delete provider: %w", err)
	}
	return nil
}

func (store *Store) ListModels(ctx context.Context, providerID string, providerCode string) ([]Model, error) {
	providerID = strings.TrimSpace(providerID)
	providerCode = strings.TrimSpace(providerCode)
	if providerID == "" && providerCode == "" {
		return nil, fmt.Errorf("providerId or providerCode is required")
	}

	query := `
SELECT models.id, models.provider_id, models.name, models.code, models.is_enabled
FROM models
JOIN providers ON providers.id = models.provider_id
WHERE models.provider_id = ? OR providers.code = ?
ORDER BY models.created_at DESC, models.name ASC;
`
	rows, err := store.db.QueryContext(ctx, query, providerID, providerCode)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		var model Model
		var isEnabled int
		if err := rows.Scan(&model.ID, &model.ProviderID, &model.Name, &model.Code, &isEnabled); err != nil {
			return nil, fmt.Errorf("scan model: %w", err)
		}
		model.IsEnabled = isEnabled == 1
		models = append(models, model)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate models: %w", err)
	}
	return models, nil
}

func (store *Store) CreateModel(ctx context.Context, input CreateModelInput) (Model, error) {
	input.ProviderID = strings.TrimSpace(input.ProviderID)
	input.Name = strings.TrimSpace(input.Name)
	input.Code = strings.TrimSpace(input.Code)
	if input.Code == "" {
		input.Code = input.Name
	}
	if input.ProviderID == "" || input.Name == "" {
		return Model{}, fmt.Errorf("provider id and model name are required")
	}

	now := formatTime(store.now())
	model := Model{
		ID:         newID("model"),
		ProviderID: input.ProviderID,
		Name:       input.Name,
		Code:       input.Code,
		IsEnabled:  input.IsEnabled,
	}
	if _, err := store.db.ExecContext(ctx, `
INSERT INTO models (id, provider_id, name, code, is_enabled, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);
`, model.ID, model.ProviderID, model.Name, model.Code, boolToInt(model.IsEnabled), now, now); err != nil {
		return Model{}, fmt.Errorf("create model: %w", err)
	}
	return model, nil
}

func (store *Store) UpdateModel(ctx context.Context, id string, input UpdateModelInput) (Model, error) {
	id = strings.TrimSpace(id)
	input.Name = strings.TrimSpace(input.Name)
	input.Code = strings.TrimSpace(input.Code)
	if input.Code == "" {
		input.Code = input.Name
	}
	if id == "" || input.Name == "" {
		return Model{}, fmt.Errorf("model id and name are required")
	}

	updatedAt := formatTime(store.now())
	result, err := store.db.ExecContext(ctx, `
UPDATE models
SET name = ?, code = ?, is_enabled = ?, updated_at = ?
WHERE id = ?;
`, input.Name, input.Code, boolToInt(input.IsEnabled), updatedAt, id)
	if err != nil {
		return Model{}, fmt.Errorf("update model: %w", err)
	}
	if rows, err := result.RowsAffected(); err != nil || rows == 0 {
		if err != nil {
			return Model{}, fmt.Errorf("update model rows affected: %w", err)
		}
		return Model{}, fmt.Errorf("model not found")
	}
	return Model{ID: id, Name: input.Name, Code: input.Code, IsEnabled: input.IsEnabled}, nil
}

func (store *Store) DeleteModel(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("model id is required")
	}
	if _, err := store.db.ExecContext(ctx, `DELETE FROM models WHERE id = ?;`, id); err != nil {
		return fmt.Errorf("delete model: %w", err)
	}
	return nil
}

func (store *Store) ListAPIKeys(ctx context.Context, providerID string) ([]APIKey, error) {
	providerID = strings.TrimSpace(providerID)
	query := `
SELECT id, provider_id, alias, secret_value, is_enabled, is_available, sort_order, failure_count, last_failed_at
FROM api_keys
`
	var args []any
	if providerID != "" {
		query += "WHERE provider_id = ? "
		args = append(args, providerID)
	}
	query += "ORDER BY sort_order ASC, created_at DESC, alias ASC;"

	rows, err := store.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var key APIKey
		var secretValue string
		var isEnabled int
		var isAvailable int
		var lastFailedAt sql.NullString
		if err := rows.Scan(&key.ID, &key.ProviderID, &key.Alias, &secretValue, &isEnabled, &isAvailable, &key.SortOrder, &key.FailureCount, &lastFailedAt); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		key.MaskedValue = MaskSecret(secretValue)
		key.IsEnabled = isEnabled == 1
		key.IsAvailable = isAvailable == 1
		if lastFailedAt.Valid {
			key.LastFailedAt = &lastFailedAt.String
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate api keys: %w", err)
	}
	return keys, nil
}

func (store *Store) CreateAPIKey(ctx context.Context, input CreateAPIKeyInput) (APIKey, error) {
	input.ProviderID = strings.TrimSpace(input.ProviderID)
	input.Alias = strings.TrimSpace(input.Alias)
	input.SecretValue = strings.TrimSpace(input.SecretValue)
	if input.ProviderID == "" || input.Alias == "" || input.SecretValue == "" {
		return APIKey{}, fmt.Errorf("provider id, key alias and secret value are required")
	}

	now := formatTime(store.now())
	key := APIKey{
		ID:          newID("key"),
		ProviderID:  input.ProviderID,
		Alias:       input.Alias,
		MaskedValue: MaskSecret(input.SecretValue),
		IsEnabled:   input.IsEnabled,
		IsAvailable: input.IsAvailable,
		SortOrder:   input.SortOrder,
	}
	if _, err := store.db.ExecContext(ctx, `
INSERT INTO api_keys (id, provider_id, alias, secret_value, is_enabled, is_available, sort_order, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);
`, key.ID, key.ProviderID, key.Alias, input.SecretValue, boolToInt(key.IsEnabled), boolToInt(key.IsAvailable), key.SortOrder, now, now); err != nil {
		return APIKey{}, fmt.Errorf("create api key: %w", err)
	}
	return key, nil
}

func (store *Store) UpdateAPIKey(ctx context.Context, id string, input UpdateAPIKeyInput) (APIKey, error) {
	id = strings.TrimSpace(id)
	input.Alias = strings.TrimSpace(input.Alias)
	input.SecretValue = strings.TrimSpace(input.SecretValue)
	if id == "" || input.Alias == "" {
		return APIKey{}, fmt.Errorf("key id and alias are required")
	}

	updatedAt := formatTime(store.now())
	var result sql.Result
	var err error
	if input.SecretValue == "" {
		result, err = store.db.ExecContext(ctx, `
UPDATE api_keys
SET alias = ?, is_enabled = ?, is_available = ?, sort_order = ?, updated_at = ?
WHERE id = ?;
`, input.Alias, boolToInt(input.IsEnabled), boolToInt(input.IsAvailable), input.SortOrder, updatedAt, id)
	} else {
		result, err = store.db.ExecContext(ctx, `
UPDATE api_keys
SET alias = ?, secret_value = ?, is_enabled = ?, is_available = ?, sort_order = ?, updated_at = ?
WHERE id = ?;
`, input.Alias, input.SecretValue, boolToInt(input.IsEnabled), boolToInt(input.IsAvailable), input.SortOrder, updatedAt, id)
	}
	if err != nil {
		return APIKey{}, fmt.Errorf("update api key: %w", err)
	}
	if rows, err := result.RowsAffected(); err != nil || rows == 0 {
		if err != nil {
			return APIKey{}, fmt.Errorf("update api key rows affected: %w", err)
		}
		return APIKey{}, fmt.Errorf("key not found")
	}

	key := APIKey{
		ID:          id,
		Alias:       input.Alias,
		IsEnabled:   input.IsEnabled,
		IsAvailable: input.IsAvailable,
		SortOrder:   input.SortOrder,
	}
	if input.SecretValue != "" {
		key.MaskedValue = MaskSecret(input.SecretValue)
	}
	return key, nil
}

func (store *Store) ReorderAPIKeys(ctx context.Context, providerID string, keyIDs []string) error {
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		return fmt.Errorf("provider id is required")
	}
	if len(keyIDs) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(keyIDs))
	for i, id := range keyIDs {
		keyIDs[i] = strings.TrimSpace(id)
		if keyIDs[i] == "" {
			return fmt.Errorf("key id is required")
		}
		if _, ok := seen[keyIDs[i]]; ok {
			return fmt.Errorf("duplicate key id: %s", keyIDs[i])
		}
		seen[keyIDs[i]] = struct{}{}
	}

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reorder api keys: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
UPDATE api_keys
SET sort_order = ?, updated_at = ?
WHERE id = ? AND provider_id = ?;
`)
	if err != nil {
		return fmt.Errorf("prepare reorder api keys: %w", err)
	}
	defer stmt.Close()

	updatedAt := formatTime(store.now())
	for i, id := range keyIDs {
		result, err := stmt.ExecContext(ctx, i+1, updatedAt, id, providerID)
		if err != nil {
			return fmt.Errorf("reorder api key: %w", err)
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("reorder api key rows affected: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("key not found for provider: %s", id)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reorder api keys: %w", err)
	}
	return nil
}

func (store *Store) DeleteAPIKey(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("key id is required")
	}
	if _, err := store.db.ExecContext(ctx, `DELETE FROM api_keys WHERE id = ?;`, id); err != nil {
		return fmt.Errorf("delete api key: %w", err)
	}
	return nil
}

func MaskSecret(secret string) string {
	secret = strings.TrimSpace(secret)
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:3] + "****" + secret[len(secret)-4:]
}

func newID(prefix string) string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		panic(fmt.Sprintf("generate id: %v", err))
	}
	return prefix + "_" + base64.RawURLEncoding.EncodeToString(bytes[:])
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}
