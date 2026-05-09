package admin

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type Channel struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Code                  string `json:"code"`
	DefaultPermissionMode string `json:"defaultPermissionMode"`
	UserManagementMode    string `json:"userManagementMode"`
	IsEnabled             bool   `json:"isEnabled"`
}

type User struct {
	ID             string `json:"id"`
	ChannelID      string `json:"channelId"`
	ExternalUserID string `json:"externalUserId"`
	DisplayName    string `json:"displayName"`
	IsEnabled      bool   `json:"isEnabled"`
}

type ProviderModelOption struct {
	ProviderID   string
	ProviderName string
	ProviderCode string
	ModelID      string
	ModelName    string
	ModelCode    string
}

type ChannelPermissionRow struct {
	ProviderModelOption
	DefaultAllowed bool
	HasDefault     bool
}

type UserPermissionRow struct {
	ProviderModelOption
	Allowed     bool
	HasExplicit bool
}

type UserKeyPermissionRow struct {
	ProviderID   string
	ProviderName string
	KeyID        string
	KeyAlias     string
	Allowed      bool
	HasExplicit  bool
}

type CreateChannelInput struct {
	Name                  string
	Code                  string
	DefaultPermissionMode string
	UserManagementMode    string
	IsEnabled             bool
}

type UpdateChannelInput = CreateChannelInput

type CreateUserInput struct {
	ChannelID      string
	ExternalUserID string
	DisplayName    string
	IsEnabled      bool
}

type UpdateUserInput struct {
	ExternalUserID string
	DisplayName    string
	IsEnabled      bool
}

func (store *Store) ListChannels(ctx context.Context) ([]Channel, error) {
	rows, err := store.db.QueryContext(ctx, `
SELECT id, name, code, default_permission_mode, user_management_mode, is_enabled
FROM channels
ORDER BY created_at DESC, name ASC;
`)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		var channel Channel
		var isEnabled int
		if err := rows.Scan(&channel.ID, &channel.Name, &channel.Code, &channel.DefaultPermissionMode, &channel.UserManagementMode, &isEnabled); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		channel.IsEnabled = isEnabled == 1
		channels = append(channels, channel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate channels: %w", err)
	}
	return channels, nil
}

func (store *Store) CreateChannel(ctx context.Context, input CreateChannelInput) (Channel, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Code = strings.TrimSpace(input.Code)
	if input.Code == "" {
		input.Code = input.Name
	}
	if input.DefaultPermissionMode == "" {
		input.DefaultPermissionMode = "DENY"
	}
	if input.UserManagementMode == "" {
		input.UserManagementMode = "EXTERNAL_MANAGED"
	}
	if input.Name == "" {
		return Channel{}, fmt.Errorf("channel name is required")
	}
	if input.DefaultPermissionMode != "ALLOW" && input.DefaultPermissionMode != "DENY" {
		return Channel{}, fmt.Errorf("invalid channel default permission mode")
	}
	if input.UserManagementMode != "EXTERNAL_MANAGED" && input.UserManagementMode != "KEYCHAIN_HOSTED" {
		return Channel{}, fmt.Errorf("invalid channel user management mode")
	}
	now := formatTime(store.now())
	channel := Channel{
		ID:                    newID("channel"),
		Name:                  input.Name,
		Code:                  input.Code,
		DefaultPermissionMode: input.DefaultPermissionMode,
		UserManagementMode:    input.UserManagementMode,
		IsEnabled:             input.IsEnabled,
	}
	if _, err := store.db.ExecContext(ctx, `
INSERT INTO channels (id, name, code, default_permission_mode, user_management_mode, is_enabled, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);
`, channel.ID, channel.Name, channel.Code, channel.DefaultPermissionMode, channel.UserManagementMode, boolToInt(channel.IsEnabled), now, now); err != nil {
		return Channel{}, fmt.Errorf("create channel: %w", err)
	}
	return channel, nil
}

func (store *Store) UpdateChannel(ctx context.Context, id string, input UpdateChannelInput) error {
	id = strings.TrimSpace(id)
	input.Name = strings.TrimSpace(input.Name)
	input.Code = strings.TrimSpace(input.Code)
	if input.Code == "" {
		input.Code = input.Name
	}
	if id == "" || input.Name == "" {
		return fmt.Errorf("channel id and name are required")
	}
	if input.DefaultPermissionMode != "ALLOW" && input.DefaultPermissionMode != "DENY" {
		return fmt.Errorf("invalid channel default permission mode")
	}
	if input.UserManagementMode == "" {
		input.UserManagementMode = "EXTERNAL_MANAGED"
	}
	if input.UserManagementMode != "EXTERNAL_MANAGED" && input.UserManagementMode != "KEYCHAIN_HOSTED" {
		return fmt.Errorf("invalid channel user management mode")
	}
	if _, err := store.db.ExecContext(ctx, `
UPDATE channels
SET name = ?, code = ?, default_permission_mode = ?, user_management_mode = ?, is_enabled = ?, updated_at = ?
WHERE id = ?;
`, input.Name, input.Code, input.DefaultPermissionMode, input.UserManagementMode, boolToInt(input.IsEnabled), formatTime(store.now()), id); err != nil {
		return fmt.Errorf("update channel: %w", err)
	}
	return nil
}

func (store *Store) DeleteChannel(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("channel id is required")
	}
	if _, err := store.db.ExecContext(ctx, `DELETE FROM channels WHERE id = ?;`, id); err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}
	return nil
}

func (store *Store) ListUsers(ctx context.Context, channelID string) ([]User, error) {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return nil, fmt.Errorf("channel id is required")
	}
	rows, err := store.db.QueryContext(ctx, `
SELECT id, channel_id, external_user_id, display_name, is_enabled
FROM users
WHERE channel_id = ?
ORDER BY created_at DESC, display_name ASC;
`, channelID)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var isEnabled int
		if err := rows.Scan(&user.ID, &user.ChannelID, &user.ExternalUserID, &user.DisplayName, &isEnabled); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		user.IsEnabled = isEnabled == 1
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

func (store *Store) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	input.ChannelID = strings.TrimSpace(input.ChannelID)
	input.ExternalUserID = strings.TrimSpace(input.ExternalUserID)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if input.ExternalUserID == "" {
		input.ExternalUserID = input.DisplayName
	}
	if input.DisplayName == "" {
		input.DisplayName = input.ExternalUserID
	}
	if input.ChannelID == "" || input.DisplayName == "" {
		return User{}, fmt.Errorf("channel id and user name are required")
	}
	now := formatTime(store.now())
	user := User{
		ID:             newID("user"),
		ChannelID:      input.ChannelID,
		ExternalUserID: input.ExternalUserID,
		DisplayName:    input.DisplayName,
		IsEnabled:      input.IsEnabled,
	}
	if _, err := store.db.ExecContext(ctx, `
INSERT INTO users (id, channel_id, external_user_id, display_name, is_enabled, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?);
`, user.ID, user.ChannelID, user.ExternalUserID, user.DisplayName, boolToInt(user.IsEnabled), now, now); err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (store *Store) UpdateUser(ctx context.Context, id string, input UpdateUserInput) error {
	id = strings.TrimSpace(id)
	input.ExternalUserID = strings.TrimSpace(input.ExternalUserID)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if input.ExternalUserID == "" {
		input.ExternalUserID = input.DisplayName
	}
	if input.DisplayName == "" {
		input.DisplayName = input.ExternalUserID
	}
	if id == "" || input.DisplayName == "" {
		return fmt.Errorf("user id and user name are required")
	}
	if _, err := store.db.ExecContext(ctx, `
UPDATE users
SET external_user_id = ?, display_name = ?, is_enabled = ?, updated_at = ?
WHERE id = ?;
`, input.ExternalUserID, input.DisplayName, boolToInt(input.IsEnabled), formatTime(store.now()), id); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (store *Store) DeleteUser(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("user id is required")
	}
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete user: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM failure_reports WHERE user_id = ?;`, id); err != nil {
		return fmt.Errorf("delete user failure reports: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM dispatch_logs WHERE user_id = ?;`, id); err != nil {
		return fmt.Errorf("delete user dispatch logs: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_key_permissions WHERE user_id = ?;`, id); err != nil {
		return fmt.Errorf("delete user key permissions: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM hosted_user_credentials WHERE user_id = ?;`, id); err != nil {
		return fmt.Errorf("delete hosted user credentials: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM users WHERE id = ?;`, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete user: %w", err)
	}
	return nil
}

func (store *Store) ListChannelPermissionRows(ctx context.Context, channelID string) ([]ChannelPermissionRow, error) {
	options, err := store.ListProviderModelOptions(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]ChannelPermissionRow, 0, len(options))
	for _, option := range options {
		var allowed int
		err := store.db.QueryRowContext(ctx, `
SELECT default_allowed
FROM channel_permission_defaults
WHERE channel_id = ? AND provider_id = ? AND model_id = ?;
`, channelID, option.ProviderID, option.ModelID).Scan(&allowed)
		item := ChannelPermissionRow{ProviderModelOption: option}
		if err == nil {
			item.DefaultAllowed = allowed == 1
			item.HasDefault = true
		} else if err != sql.ErrNoRows {
			return nil, fmt.Errorf("read channel permission default: %w", err)
		}
		rows = append(rows, item)
	}
	return rows, nil
}

func (store *Store) SetChannelPermissionDefault(ctx context.Context, channelID string, providerID string, modelID string, allowed bool) error {
	if channelID == "" || providerID == "" || modelID == "" {
		return fmt.Errorf("channel id, provider id and model id are required")
	}
	now := formatTime(store.now())
	if _, err := store.db.ExecContext(ctx, `
INSERT INTO channel_permission_defaults (id, channel_id, provider_id, model_id, default_allowed, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(channel_id, provider_id, model_id)
DO UPDATE SET default_allowed = excluded.default_allowed, updated_at = excluded.updated_at;
`, newID("channel_permission"), channelID, providerID, modelID, boolToInt(allowed), now, now); err != nil {
		return fmt.Errorf("set channel permission default: %w", err)
	}
	return nil
}

func (store *Store) ListUserPermissionRows(ctx context.Context, userID string) ([]UserPermissionRow, error) {
	options, err := store.ListProviderModelOptions(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]UserPermissionRow, 0, len(options))
	for _, option := range options {
		var allowed int
		err := store.db.QueryRowContext(ctx, `
SELECT allowed
FROM user_permissions
WHERE user_id = ? AND provider_id = ? AND model_id = ?;
`, userID, option.ProviderID, option.ModelID).Scan(&allowed)
		item := UserPermissionRow{ProviderModelOption: option}
		if err == nil {
			item.Allowed = allowed == 1
			item.HasExplicit = true
		} else if err != sql.ErrNoRows {
			return nil, fmt.Errorf("read user permission: %w", err)
		}
		rows = append(rows, item)
	}
	return rows, nil
}

func (store *Store) SetUserPermission(ctx context.Context, userID string, providerID string, modelID string, allowed bool) error {
	if userID == "" || providerID == "" || modelID == "" {
		return fmt.Errorf("user id, provider id and model id are required")
	}
	now := formatTime(store.now())
	if _, err := store.db.ExecContext(ctx, `
INSERT INTO user_permissions (id, user_id, provider_id, model_id, allowed, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(user_id, provider_id, model_id)
DO UPDATE SET allowed = excluded.allowed, updated_at = excluded.updated_at;
`, newID("user_permission"), userID, providerID, modelID, boolToInt(allowed), now, now); err != nil {
		return fmt.Errorf("set user permission: %w", err)
	}
	return nil
}

func (store *Store) ListUserKeyPermissionRows(ctx context.Context, userID string, providerID string) ([]UserKeyPermissionRow, error) {
	userID = strings.TrimSpace(userID)
	providerID = strings.TrimSpace(providerID)
	if userID == "" || providerID == "" {
		return nil, fmt.Errorf("user id and provider id are required")
	}
	rows, err := store.db.QueryContext(ctx, `
SELECT providers.id, providers.name, api_keys.id, api_keys.alias,
  COALESCE(user_key_permissions.allowed, 1) AS allowed,
  CASE WHEN user_key_permissions.id IS NULL THEN 0 ELSE 1 END AS has_explicit
FROM api_keys
JOIN providers ON providers.id = api_keys.provider_id
LEFT JOIN user_key_permissions ON user_key_permissions.user_id = ? AND user_key_permissions.provider_id = providers.id AND user_key_permissions.key_id = api_keys.id
WHERE api_keys.provider_id = ?
ORDER BY api_keys.sort_order ASC, api_keys.created_at DESC, api_keys.alias ASC;
`, userID, providerID)
	if err != nil {
		return nil, fmt.Errorf("list user key permission rows: %w", err)
	}
	defer rows.Close()

	var permissions []UserKeyPermissionRow
	for rows.Next() {
		var permission UserKeyPermissionRow
		var allowed, hasExplicit int
		if err := rows.Scan(&permission.ProviderID, &permission.ProviderName, &permission.KeyID, &permission.KeyAlias, &allowed, &hasExplicit); err != nil {
			return nil, fmt.Errorf("scan user key permission row: %w", err)
		}
		permission.Allowed = allowed == 1
		permission.HasExplicit = hasExplicit == 1
		permissions = append(permissions, permission)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user key permission rows: %w", err)
	}
	return permissions, nil
}

func (store *Store) SetUserKeyPermission(ctx context.Context, userID string, providerID string, keyID string, allowed bool) error {
	userID = strings.TrimSpace(userID)
	providerID = strings.TrimSpace(providerID)
	keyID = strings.TrimSpace(keyID)
	if userID == "" || providerID == "" || keyID == "" {
		return fmt.Errorf("user id, provider id and key id are required")
	}
	var keyProviderID string
	if err := store.db.QueryRowContext(ctx, `SELECT provider_id FROM api_keys WHERE id = ?;`, keyID).Scan(&keyProviderID); err != nil {
		return fmt.Errorf("read key provider: %w", wrapNotFound(err, "key not found"))
	}
	if keyProviderID != providerID {
		return fmt.Errorf("key does not belong to provider")
	}
	now := formatTime(store.now())
	if _, err := store.db.ExecContext(ctx, `
INSERT INTO user_key_permissions (id, user_id, provider_id, key_id, allowed, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(user_id, provider_id, key_id)
DO UPDATE SET allowed = excluded.allowed, updated_at = excluded.updated_at;
`, newID("user_key_permission"), userID, providerID, keyID, boolToInt(allowed), now, now); err != nil {
		return fmt.Errorf("set user key permission: %w", err)
	}
	return nil
}

func (store *Store) ListProviderModelOptions(ctx context.Context) ([]ProviderModelOption, error) {
	rows, err := store.db.QueryContext(ctx, `
SELECT providers.id, providers.name, providers.code, models.id, models.name, models.code
FROM providers
JOIN models ON models.provider_id = providers.id
ORDER BY providers.name ASC, models.name ASC;
`)
	if err != nil {
		return nil, fmt.Errorf("list provider model options: %w", err)
	}
	defer rows.Close()

	var options []ProviderModelOption
	for rows.Next() {
		var option ProviderModelOption
		if err := rows.Scan(&option.ProviderID, &option.ProviderName, &option.ProviderCode, &option.ModelID, &option.ModelName, &option.ModelCode); err != nil {
			return nil, fmt.Errorf("scan provider model option: %w", err)
		}
		options = append(options, option)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate provider model options: %w", err)
	}
	return options, nil
}

func (store *Store) SeedDemoAccessData(ctx context.Context) error {
	demoChannels := []CreateChannelInput{
		{Name: "本校默认渠道", Code: "school-default", DefaultPermissionMode: "DENY", IsEnabled: true},
		{Name: "科研试用渠道", Code: "research-trial", DefaultPermissionMode: "ALLOW", IsEnabled: true},
	}
	for _, channel := range demoChannels {
		if _, err := store.CreateChannel(ctx, channel); err != nil && !strings.Contains(err.Error(), "UNIQUE") {
			return err
		}
	}

	channels, err := store.ListChannels(ctx)
	if err != nil {
		return err
	}
	channelByCode := map[string]Channel{}
	for _, channel := range channels {
		channelByCode[channel.Code] = channel
	}
	demoUsers := []CreateUserInput{
		{ChannelID: channelByCode["school-default"].ID, ExternalUserID: "stu-2026-001", DisplayName: "教学演示用户 001", IsEnabled: true},
		{ChannelID: channelByCode["school-default"].ID, ExternalUserID: "stu-2026-002", DisplayName: "教学演示用户 002", IsEnabled: true},
		{ChannelID: channelByCode["research-trial"].ID, ExternalUserID: "lab-user-001", DisplayName: "科研试用用户 001", IsEnabled: true},
	}
	for _, user := range demoUsers {
		if _, err := store.CreateUser(ctx, user); err != nil && !strings.Contains(err.Error(), "UNIQUE") {
			return err
		}
	}
	return nil
}
