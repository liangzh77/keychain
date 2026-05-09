package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UpsertRuntimeExternalUserInput struct {
	ChannelName    string
	ExternalUserID string
	Name           string
	IsEnabled      bool
}

type RegisterRuntimeHostedUserInput struct {
	ChannelName string
	Username    string
	Name        string
	Password    string
}

type LoginRuntimeHostedUserInput struct {
	ChannelName string
	Username    string
	Password    string
}

type ResetRuntimeHostedUserPasswordInput struct {
	ChannelName string
	UserID      string
	Password    string
}

type RuntimeProvider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RuntimeModel struct {
	ID         string `json:"id"`
	ProviderID string `json:"providerId"`
	Name       string `json:"name"`
}

type EffectivePermission struct {
	ProviderID   string `json:"providerId"`
	ProviderName string `json:"providerName"`
	ModelID      string `json:"modelId"`
	ModelName    string `json:"modelName"`
	Allowed      bool   `json:"allowed"`
}

type DispatchKeyInput struct {
	ChannelName string
	UserID      string
	ProviderID  string
	ModelID     string
}

type DispatchKeyResult struct {
	DispatchLogID string
	ProviderName  string
	ModelName     string
	KeyID         string
	KeyAlias      string
	Key           string
}

type KeyFailureResult struct {
	KeyID       string
	KeyAlias    string
	IsAvailable bool
}

func (store *Store) UpsertRuntimeExternalUser(ctx context.Context, input UpsertRuntimeExternalUserInput) (User, error) {
	input.ChannelName = strings.TrimSpace(input.ChannelName)
	input.ExternalUserID = strings.TrimSpace(input.ExternalUserID)
	input.Name = strings.TrimSpace(input.Name)
	if input.ExternalUserID == "" {
		return User{}, fmt.Errorf("external user id is required")
	}
	if input.Name == "" {
		input.Name = input.ExternalUserID
	}
	channel, err := store.lookupEnabledChannel(ctx, input.ChannelName)
	if err != nil {
		return User{}, err
	}
	if channel.UserManagementMode != "EXTERNAL_MANAGED" {
		return User{}, fmt.Errorf("channel does not accept external users")
	}

	now := formatTime(store.now())
	id := newID("user")
	if _, err := store.db.ExecContext(ctx, `
INSERT INTO users (id, channel_id, external_user_id, display_name, is_enabled, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(channel_id, external_user_id)
DO UPDATE SET display_name = excluded.display_name, is_enabled = excluded.is_enabled, updated_at = excluded.updated_at;
`, id, channel.ID, input.ExternalUserID, input.Name, boolToInt(input.IsEnabled), now, now); err != nil {
		return User{}, fmt.Errorf("upsert runtime external user: %w", err)
	}
	user, err := store.lookupUserByExternalID(ctx, channel.ID, input.ExternalUserID)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (store *Store) DeleteRuntimeExternalUser(ctx context.Context, channelName string, externalUserID string) error {
	channel, err := store.lookupEnabledChannel(ctx, channelName)
	if err != nil {
		return err
	}
	if channel.UserManagementMode != "EXTERNAL_MANAGED" {
		return fmt.Errorf("channel does not accept external users")
	}
	user, err := store.lookupUserByExternalID(ctx, channel.ID, strings.TrimSpace(externalUserID))
	if err != nil {
		return err
	}
	return store.DeleteUser(ctx, user.ID)
}

func (store *Store) RegisterRuntimeHostedUser(ctx context.Context, input RegisterRuntimeHostedUserInput) (User, error) {
	input.ChannelName = strings.TrimSpace(input.ChannelName)
	input.Username = strings.TrimSpace(input.Username)
	input.Name = strings.TrimSpace(input.Name)
	if input.Username == "" {
		return User{}, fmt.Errorf("username is required")
	}
	if input.Name == "" {
		input.Name = input.Username
	}
	if strings.TrimSpace(input.Password) == "" {
		return User{}, fmt.Errorf("password is required")
	}
	channel, err := store.lookupEnabledChannel(ctx, input.ChannelName)
	if err != nil {
		return User{}, err
	}
	if channel.UserManagementMode != "KEYCHAIN_HOSTED" {
		return User{}, fmt.Errorf("channel does not accept hosted users")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hash hosted user password: %w", err)
	}

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, fmt.Errorf("begin register hosted user: %w", err)
	}
	defer tx.Rollback()

	now := formatTime(store.now())
	user := User{
		ID:             newID("user"),
		ChannelID:      channel.ID,
		ExternalUserID: input.Username,
		DisplayName:    input.Name,
		IsEnabled:      true,
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO users (id, channel_id, external_user_id, display_name, is_enabled, created_at, updated_at)
VALUES (?, ?, ?, ?, 1, ?, ?);
`, user.ID, user.ChannelID, user.ExternalUserID, user.DisplayName, now, now); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return User{}, fmt.Errorf("hosted user already exists")
		}
		return User{}, fmt.Errorf("create hosted user: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO hosted_user_credentials (user_id, password_hash, created_at, updated_at)
VALUES (?, ?, ?, ?);
`, user.ID, string(passwordHash), now, now); err != nil {
		return User{}, fmt.Errorf("create hosted user credentials: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return User{}, fmt.Errorf("commit register hosted user: %w", err)
	}
	return user, nil
}

func (store *Store) LoginRuntimeHostedUser(ctx context.Context, input LoginRuntimeHostedUserInput) (User, error) {
	input.ChannelName = strings.TrimSpace(input.ChannelName)
	input.Username = strings.TrimSpace(input.Username)
	if input.Username == "" || strings.TrimSpace(input.Password) == "" {
		return User{}, fmt.Errorf("username and password are required")
	}
	channel, err := store.lookupEnabledChannel(ctx, input.ChannelName)
	if err != nil {
		return User{}, err
	}
	if channel.UserManagementMode != "KEYCHAIN_HOSTED" {
		return User{}, fmt.Errorf("channel does not accept hosted users")
	}

	var user User
	var isEnabled int
	var passwordHash string
	if err := store.db.QueryRowContext(ctx, `
SELECT users.id, users.channel_id, users.external_user_id, users.display_name, users.is_enabled, hosted_user_credentials.password_hash
FROM users
JOIN hosted_user_credentials ON hosted_user_credentials.user_id = users.id
WHERE users.channel_id = ? AND users.external_user_id = ?;
`, channel.ID, input.Username).Scan(&user.ID, &user.ChannelID, &user.ExternalUserID, &user.DisplayName, &isEnabled, &passwordHash); err != nil {
		if err == sql.ErrNoRows {
			return User{}, fmt.Errorf("invalid credentials")
		}
		return User{}, fmt.Errorf("read hosted user credentials: %w", err)
	}
	if isEnabled != 1 {
		return User{}, fmt.Errorf("user is disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(input.Password)); err != nil {
		return User{}, fmt.Errorf("invalid credentials")
	}
	user.IsEnabled = true
	return user, nil
}

func (store *Store) ResetRuntimeHostedUserPassword(ctx context.Context, input ResetRuntimeHostedUserPasswordInput) (User, error) {
	input.ChannelName = strings.TrimSpace(input.ChannelName)
	input.UserID = strings.TrimSpace(input.UserID)
	if input.UserID == "" {
		return User{}, fmt.Errorf("user id is required")
	}
	if strings.TrimSpace(input.Password) == "" {
		return User{}, fmt.Errorf("password is required")
	}
	channel, err := store.lookupEnabledChannel(ctx, input.ChannelName)
	if err != nil {
		return User{}, err
	}
	if channel.UserManagementMode != "KEYCHAIN_HOSTED" {
		return User{}, fmt.Errorf("channel does not accept hosted users")
	}
	user, err := store.lookupUserByID(ctx, input.UserID)
	if err != nil {
		return User{}, err
	}
	if user.ChannelID != channel.ID {
		return User{}, fmt.Errorf("user does not belong to channel")
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hash hosted user password: %w", err)
	}
	result, err := store.db.ExecContext(ctx, `
UPDATE hosted_user_credentials
SET password_hash = ?, updated_at = ?
WHERE user_id = ?;
`, string(passwordHash), formatTime(store.now()), user.ID)
	if err != nil {
		return User{}, fmt.Errorf("reset hosted user password: %w", err)
	}
	if rowsAffected, err := result.RowsAffected(); err == nil && rowsAffected == 0 {
		return User{}, fmt.Errorf("hosted user credentials not found")
	}
	return user, nil
}

func (store *Store) DeleteRuntimeHostedUser(ctx context.Context, channelName string, userID string) error {
	channel, err := store.lookupEnabledChannel(ctx, channelName)
	if err != nil {
		return err
	}
	if channel.UserManagementMode != "KEYCHAIN_HOSTED" {
		return fmt.Errorf("channel does not accept hosted users")
	}
	user, err := store.lookupUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.ChannelID != channel.ID {
		return fmt.Errorf("user does not belong to channel")
	}
	return store.DeleteUser(ctx, user.ID)
}

func (store *Store) ListRuntimeProviders(ctx context.Context) ([]RuntimeProvider, error) {
	rows, err := store.db.QueryContext(ctx, `
SELECT id, name
FROM providers
WHERE is_enabled = 1
ORDER BY name ASC;
`)
	if err != nil {
		return nil, fmt.Errorf("list runtime providers: %w", err)
	}
	defer rows.Close()

	var providers []RuntimeProvider
	for rows.Next() {
		var provider RuntimeProvider
		if err := rows.Scan(&provider.ID, &provider.Name); err != nil {
			return nil, fmt.Errorf("scan runtime provider: %w", err)
		}
		providers = append(providers, provider)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runtime providers: %w", err)
	}
	return providers, nil
}

func (store *Store) ListRuntimeModels(ctx context.Context, providerID string) ([]RuntimeModel, error) {
	provider, err := store.lookupEnabledProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}
	rows, err := store.db.QueryContext(ctx, `
SELECT id, provider_id, name
FROM models
WHERE provider_id = ? AND is_enabled = 1
ORDER BY name ASC;
`, provider.ID)
	if err != nil {
		return nil, fmt.Errorf("list runtime models: %w", err)
	}
	defer rows.Close()

	var models []RuntimeModel
	for rows.Next() {
		var model RuntimeModel
		if err := rows.Scan(&model.ID, &model.ProviderID, &model.Name); err != nil {
			return nil, fmt.Errorf("scan runtime model: %w", err)
		}
		models = append(models, model)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate runtime models: %w", err)
	}
	return models, nil
}

func (store *Store) ListEffectiveUserPermissions(ctx context.Context, userID string) ([]EffectivePermission, error) {
	user, channel, err := store.lookupEnabledUserAndChannel(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, err := store.db.QueryContext(ctx, `
SELECT providers.id, providers.name, models.id, models.name,
  COALESCE(user_permissions.allowed, channel_permission_defaults.default_allowed,
    CASE WHEN channels.default_permission_mode = 'ALLOW' THEN 1 ELSE 0 END) AS allowed
FROM providers
JOIN models ON models.provider_id = providers.id
JOIN channels ON channels.id = ?
LEFT JOIN user_permissions ON user_permissions.user_id = ? AND user_permissions.provider_id = providers.id AND user_permissions.model_id = models.id
LEFT JOIN channel_permission_defaults ON channel_permission_defaults.channel_id = channels.id AND channel_permission_defaults.provider_id = providers.id AND channel_permission_defaults.model_id = models.id
WHERE providers.is_enabled = 1 AND models.is_enabled = 1
ORDER BY providers.name ASC, models.name ASC;
`, channel.ID, user.ID)
	if err != nil {
		return nil, fmt.Errorf("list effective user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []EffectivePermission
	for rows.Next() {
		var permission EffectivePermission
		var allowed int
		if err := rows.Scan(&permission.ProviderID, &permission.ProviderName, &permission.ModelID, &permission.ModelName, &allowed); err != nil {
			return nil, fmt.Errorf("scan effective permission: %w", err)
		}
		permission.Allowed = allowed == 1
		permissions = append(permissions, permission)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate effective permissions: %w", err)
	}
	return permissions, nil
}

func (store *Store) DispatchRuntimeKey(ctx context.Context, input DispatchKeyInput) (DispatchKeyResult, error) {
	input.ChannelName = strings.TrimSpace(input.ChannelName)
	input.UserID = strings.TrimSpace(input.UserID)
	input.ProviderID = strings.TrimSpace(input.ProviderID)
	input.ModelID = strings.TrimSpace(input.ModelID)
	if input.ChannelName == "" || input.UserID == "" || input.ProviderID == "" || input.ModelID == "" {
		return DispatchKeyResult{}, fmt.Errorf("channel name, user id, provider id and model id are required")
	}

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return DispatchKeyResult{}, fmt.Errorf("begin dispatch key: %w", err)
	}
	defer tx.Rollback()

	var channelID, channelName, defaultMode string
	var channelEnabled int
	if err := tx.QueryRowContext(ctx, `
SELECT id, name, default_permission_mode, is_enabled
FROM channels
WHERE name = ?;
`, input.ChannelName).Scan(&channelID, &channelName, &defaultMode, &channelEnabled); err != nil {
		return DispatchKeyResult{}, wrapNotFound(err, "channel not found")
	}
	if channelEnabled != 1 {
		return DispatchKeyResult{}, fmt.Errorf("channel is disabled")
	}

	var userChannelID string
	var userEnabled int
	if err := tx.QueryRowContext(ctx, `
SELECT channel_id, is_enabled
FROM users
WHERE id = ?;
`, input.UserID).Scan(&userChannelID, &userEnabled); err != nil {
		return DispatchKeyResult{}, wrapNotFound(err, "user not found")
	}
	if userChannelID != channelID {
		return DispatchKeyResult{}, fmt.Errorf("user does not belong to channel")
	}
	if userEnabled != 1 {
		return DispatchKeyResult{}, fmt.Errorf("user is disabled")
	}

	var providerName, rotationStrategy string
	var providerEnabled int
	var roundRobinCursor int
	if err := tx.QueryRowContext(ctx, `
SELECT name, is_enabled, rotation_strategy, round_robin_cursor
FROM providers
WHERE id = ?;
`, input.ProviderID).Scan(&providerName, &providerEnabled, &rotationStrategy, &roundRobinCursor); err != nil {
		return DispatchKeyResult{}, wrapNotFound(err, "provider not found")
	}
	if providerEnabled != 1 {
		return DispatchKeyResult{}, fmt.Errorf("provider is disabled")
	}

	var modelName, modelProviderID string
	var modelEnabled int
	if err := tx.QueryRowContext(ctx, `
SELECT name, provider_id, is_enabled
FROM models
WHERE id = ?;
`, input.ModelID).Scan(&modelName, &modelProviderID, &modelEnabled); err != nil {
		return DispatchKeyResult{}, wrapNotFound(err, "model not found")
	}
	if modelProviderID != input.ProviderID {
		return DispatchKeyResult{}, fmt.Errorf("model does not belong to provider")
	}
	if modelEnabled != 1 {
		return DispatchKeyResult{}, fmt.Errorf("model is disabled")
	}

	allowed, err := effectivePermissionInTx(ctx, tx, channelID, input.UserID, input.ProviderID, input.ModelID, defaultMode)
	if err != nil {
		return DispatchKeyResult{}, err
	}
	if !allowed {
		return DispatchKeyResult{}, fmt.Errorf("permission denied")
	}

	keys, err := availableKeysInTx(ctx, tx, input.ProviderID, input.UserID)
	if err != nil {
		return DispatchKeyResult{}, err
	}
	if len(keys) == 0 {
		return DispatchKeyResult{}, fmt.Errorf("no available key")
	}
	selected := keys[0]
	if rotationStrategy == "ROUND_ROBIN" {
		selected = keys[roundRobinCursor%len(keys)]
		if _, err := tx.ExecContext(ctx, `
UPDATE providers
SET round_robin_cursor = ?, updated_at = ?
WHERE id = ?;
`, (roundRobinCursor+1)%len(keys), formatTime(store.now()), input.ProviderID); err != nil {
			return DispatchKeyResult{}, fmt.Errorf("update round robin cursor: %w", err)
		}
	}

	now := formatTime(store.now())
	dispatchLogID := newID("dispatch")
	if _, err := tx.ExecContext(ctx, `
INSERT INTO dispatch_logs (id, channel_id, user_id, provider_id, model_id, key_id, key_alias_snapshot, status, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, 'DISPATCHED', ?);
`, dispatchLogID, channelID, input.UserID, input.ProviderID, input.ModelID, selected.ID, selected.Alias, now); err != nil {
		return DispatchKeyResult{}, fmt.Errorf("insert dispatch log: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return DispatchKeyResult{}, fmt.Errorf("commit dispatch key: %w", err)
	}

	return DispatchKeyResult{
		DispatchLogID: dispatchLogID,
		ProviderName:  providerName,
		ModelName:     modelName,
		KeyID:         selected.ID,
		KeyAlias:      selected.Alias,
		Key:           selected.SecretValue,
	}, nil
}

func (store *Store) ReportRuntimeKeyFailure(ctx context.Context, dispatchLogID string, errorCode string, errorMessage string) (KeyFailureResult, error) {
	dispatchLogID = strings.TrimSpace(dispatchLogID)
	if dispatchLogID == "" {
		return KeyFailureResult{}, fmt.Errorf("dispatch log id is required")
	}
	errorCode = strings.TrimSpace(errorCode)
	errorMessage = strings.TrimSpace(errorMessage)

	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return KeyFailureResult{}, fmt.Errorf("begin key failure report: %w", err)
	}
	defer tx.Rollback()

	var channelID, userID, providerID, modelID, keyID, keyAlias string
	if err := tx.QueryRowContext(ctx, `
SELECT channel_id, user_id, provider_id, model_id, key_id, key_alias_snapshot
FROM dispatch_logs
WHERE id = ?;
`, dispatchLogID).Scan(&channelID, &userID, &providerID, &modelID, &keyID, &keyAlias); err != nil {
		return KeyFailureResult{}, wrapNotFound(err, "dispatch log not found")
	}

	now := formatTime(store.now())
	if _, err := tx.ExecContext(ctx, `
INSERT INTO failure_reports (id, dispatch_log_id, channel_id, user_id, provider_id, model_id, key_id, key_alias_snapshot, error_code, error_message, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
`, newID("failure"), dispatchLogID, channelID, userID, providerID, modelID, keyID, keyAlias, errorCode, errorMessage, now); err != nil {
		return KeyFailureResult{}, fmt.Errorf("insert key failure report: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE api_keys
SET is_available = 0, failure_count = failure_count + 1, last_failed_at = ?, updated_at = ?
WHERE id = ?;
`, now, now, keyID); err != nil {
		return KeyFailureResult{}, fmt.Errorf("mark key unavailable: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE dispatch_logs
SET status = 'FAILED'
WHERE id = ?;
`, dispatchLogID); err != nil {
		return KeyFailureResult{}, fmt.Errorf("update dispatch log status: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return KeyFailureResult{}, fmt.Errorf("commit key failure report: %w", err)
	}
	return KeyFailureResult{KeyID: keyID, KeyAlias: keyAlias, IsAvailable: false}, nil
}

type runtimeKeyCandidate struct {
	ID          string
	Alias       string
	SecretValue string
}

func availableKeysInTx(ctx context.Context, tx *sql.Tx, providerID string, userID string) ([]runtimeKeyCandidate, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT api_keys.id, api_keys.alias, api_keys.secret_value
FROM api_keys
LEFT JOIN user_key_permissions ON user_key_permissions.user_id = ? AND user_key_permissions.provider_id = api_keys.provider_id AND user_key_permissions.key_id = api_keys.id
WHERE api_keys.provider_id = ? AND api_keys.is_enabled = 1 AND api_keys.is_available = 1
  AND COALESCE(user_key_permissions.allowed, 1) = 1
ORDER BY api_keys.sort_order ASC, api_keys.created_at DESC, api_keys.alias ASC;
`, userID, providerID)
	if err != nil {
		return nil, fmt.Errorf("list available keys: %w", err)
	}
	defer rows.Close()

	var keys []runtimeKeyCandidate
	for rows.Next() {
		var key runtimeKeyCandidate
		if err := rows.Scan(&key.ID, &key.Alias, &key.SecretValue); err != nil {
			return nil, fmt.Errorf("scan available key: %w", err)
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate available keys: %w", err)
	}
	return keys, nil
}

func effectivePermissionInTx(ctx context.Context, tx *sql.Tx, channelID string, userID string, providerID string, modelID string, defaultMode string) (bool, error) {
	var allowed int
	err := tx.QueryRowContext(ctx, `
SELECT allowed
FROM user_permissions
WHERE user_id = ? AND provider_id = ? AND model_id = ?;
`, userID, providerID, modelID).Scan(&allowed)
	if err == nil {
		return allowed == 1, nil
	}
	if err != sql.ErrNoRows {
		return false, fmt.Errorf("read user permission: %w", err)
	}

	err = tx.QueryRowContext(ctx, `
SELECT default_allowed
FROM channel_permission_defaults
WHERE channel_id = ? AND provider_id = ? AND model_id = ?;
`, channelID, providerID, modelID).Scan(&allowed)
	if err == nil {
		return allowed == 1, nil
	}
	if err != sql.ErrNoRows {
		return false, fmt.Errorf("read channel permission default: %w", err)
	}
	return defaultMode == "ALLOW", nil
}

func (store *Store) lookupEnabledChannel(ctx context.Context, channelName string) (Channel, error) {
	channelName = strings.TrimSpace(channelName)
	if channelName == "" {
		return Channel{}, fmt.Errorf("channel name is required")
	}
	row := store.db.QueryRowContext(ctx, `
SELECT id, name, code, default_permission_mode, user_management_mode, is_enabled
FROM channels
WHERE name = ?;
`, channelName)
	var channel Channel
	var isEnabled int
	if err := row.Scan(&channel.ID, &channel.Name, &channel.Code, &channel.DefaultPermissionMode, &channel.UserManagementMode, &isEnabled); err != nil {
		return Channel{}, wrapNotFound(err, "channel not found")
	}
	if isEnabled != 1 {
		return Channel{}, fmt.Errorf("channel is disabled")
	}
	channel.IsEnabled = true
	return channel, nil
}

func (store *Store) lookupEnabledProvider(ctx context.Context, providerID string) (Provider, error) {
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		return Provider{}, fmt.Errorf("provider id is required")
	}
	row := store.db.QueryRowContext(ctx, `
SELECT id, name, code, is_enabled, rotation_strategy, created_at, updated_at
FROM providers
WHERE id = ?;
`, providerID)
	var provider Provider
	var isEnabled int
	if err := row.Scan(&provider.ID, &provider.Name, &provider.Code, &isEnabled, &provider.RotationStrategy, &provider.CreatedAt, &provider.UpdatedAt); err != nil {
		return Provider{}, wrapNotFound(err, "provider not found")
	}
	if isEnabled != 1 {
		return Provider{}, fmt.Errorf("provider is disabled")
	}
	provider.IsEnabled = true
	return provider, nil
}

func (store *Store) lookupUserByExternalID(ctx context.Context, channelID string, externalUserID string) (User, error) {
	var user User
	var isEnabled int
	if err := store.db.QueryRowContext(ctx, `
SELECT id, channel_id, external_user_id, display_name, is_enabled
FROM users
WHERE channel_id = ? AND external_user_id = ?;
`, channelID, strings.TrimSpace(externalUserID)).Scan(&user.ID, &user.ChannelID, &user.ExternalUserID, &user.DisplayName, &isEnabled); err != nil {
		return User{}, wrapNotFound(err, "user not found")
	}
	user.IsEnabled = isEnabled == 1
	return user, nil
}

func (store *Store) lookupUserByID(ctx context.Context, userID string) (User, error) {
	var user User
	var isEnabled int
	if err := store.db.QueryRowContext(ctx, `
SELECT id, channel_id, external_user_id, display_name, is_enabled
FROM users
WHERE id = ?;
`, strings.TrimSpace(userID)).Scan(&user.ID, &user.ChannelID, &user.ExternalUserID, &user.DisplayName, &isEnabled); err != nil {
		return User{}, wrapNotFound(err, "user not found")
	}
	user.IsEnabled = isEnabled == 1
	return user, nil
}

func (store *Store) lookupEnabledUserAndChannel(ctx context.Context, userID string) (User, Channel, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return User{}, Channel{}, fmt.Errorf("user id is required")
	}
	var user User
	var channel Channel
	var userEnabled, channelEnabled int
	if err := store.db.QueryRowContext(ctx, `
SELECT users.id, users.channel_id, users.external_user_id, users.display_name, users.is_enabled,
  channels.id, channels.name, channels.code, channels.default_permission_mode, channels.user_management_mode, channels.is_enabled
FROM users
JOIN channels ON channels.id = users.channel_id
WHERE users.id = ?;
`, userID).Scan(&user.ID, &user.ChannelID, &user.ExternalUserID, &user.DisplayName, &userEnabled, &channel.ID, &channel.Name, &channel.Code, &channel.DefaultPermissionMode, &channel.UserManagementMode, &channelEnabled); err != nil {
		return User{}, Channel{}, wrapNotFound(err, "user not found")
	}
	if userEnabled != 1 {
		return User{}, Channel{}, fmt.Errorf("user is disabled")
	}
	if channelEnabled != 1 {
		return User{}, Channel{}, fmt.Errorf("channel is disabled")
	}
	user.IsEnabled = true
	channel.IsEnabled = true
	return user, channel, nil
}

func wrapNotFound(err error, message string) error {
	if err == sql.ErrNoRows {
		return errors.New(message)
	}
	return err
}
