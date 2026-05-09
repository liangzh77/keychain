package admin

import (
	"context"
	"strings"
	"testing"
	"time"

	keydb "github.com/liangzh77/keychain/internal/db"
)

func TestProviderModelKeyFlow(t *testing.T) {
	store := newTestStore(t)

	provider, err := store.CreateProvider(context.Background(), CreateProviderInput{
		Name:             "OpenAI",
		Code:             "openai",
		IsEnabled:        true,
		RotationStrategy: "ROUND_ROBIN",
	})
	if err != nil {
		t.Fatalf("CreateProvider() error = %v", err)
	}

	model, err := store.CreateModel(context.Background(), CreateModelInput{
		ProviderID: provider.ID,
		Name:       "GPT 4.1",
		Code:       "gpt-4.1",
		IsEnabled:  true,
	})
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	key, err := store.CreateAPIKey(context.Background(), CreateAPIKeyInput{
		ProviderID:  provider.ID,
		Alias:       "openai-main",
		SecretValue: "sk-test-secret-1234",
		IsEnabled:   true,
		IsAvailable: true,
		SortOrder:   10,
	})
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}
	if strings.Contains(key.MaskedValue, "secret") {
		t.Fatalf("masked value leaks secret: %s", key.MaskedValue)
	}

	providers, err := store.ListProviders(context.Background())
	if err != nil {
		t.Fatalf("ListProviders() error = %v", err)
	}
	if len(providers) != 1 || providers[0].ID != provider.ID {
		t.Fatalf("providers = %#v, want created provider", providers)
	}

	models, err := store.ListModels(context.Background(), provider.ID, "")
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}
	if len(models) != 1 || models[0].ID != model.ID {
		t.Fatalf("models = %#v, want created model", models)
	}

	keys, err := store.ListAPIKeys(context.Background(), provider.ID)
	if err != nil {
		t.Fatalf("ListAPIKeys() error = %v", err)
	}
	if len(keys) != 1 || keys[0].ID != key.ID {
		t.Fatalf("keys = %#v, want created key", keys)
	}
	if strings.Contains(keys[0].MaskedValue, "secret") {
		t.Fatalf("listed masked value leaks secret: %s", keys[0].MaskedValue)
	}
}

func TestListModelsRequiresProviderFilter(t *testing.T) {
	store := newTestStore(t)

	_, err := store.ListModels(context.Background(), "", "")
	if err == nil {
		t.Fatal("ListModels() error = nil, want provider filter error")
	}
}

func TestUpdateProviderModelAndKey(t *testing.T) {
	store := newTestStore(t)

	provider, err := store.CreateProvider(context.Background(), CreateProviderInput{
		Name:             "OpenAI",
		Code:             "openai",
		IsEnabled:        true,
		RotationStrategy: "ROUND_ROBIN",
	})
	if err != nil {
		t.Fatalf("CreateProvider() error = %v", err)
	}
	model, err := store.CreateModel(context.Background(), CreateModelInput{
		ProviderID: provider.ID,
		Name:       "GPT 4.1",
		Code:       "gpt-4.1",
		IsEnabled:  true,
	})
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}
	key, err := store.CreateAPIKey(context.Background(), CreateAPIKeyInput{
		ProviderID:  provider.ID,
		Alias:       "main",
		SecretValue: "sk-original-1234",
		IsEnabled:   true,
		IsAvailable: true,
		SortOrder:   1,
	})
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}

	if _, err := store.UpdateProvider(context.Background(), provider.ID, UpdateProviderInput{
		Name:             "OpenAI Updated",
		Code:             "openai-updated",
		IsEnabled:        false,
		RotationStrategy: "STICKY_FIRST_AVAILABLE",
	}); err != nil {
		t.Fatalf("UpdateProvider() error = %v", err)
	}
	if _, err := store.UpdateModel(context.Background(), model.ID, UpdateModelInput{
		Name:      "GPT 4.1 Mini",
		Code:      "gpt-4.1-mini",
		IsEnabled: false,
	}); err != nil {
		t.Fatalf("UpdateModel() error = %v", err)
	}
	if _, err := store.UpdateAPIKey(context.Background(), key.ID, UpdateAPIKeyInput{
		Alias:       "main-updated",
		SecretValue: "",
		IsEnabled:   false,
		IsAvailable: false,
		SortOrder:   7,
	}); err != nil {
		t.Fatalf("UpdateAPIKey() error = %v", err)
	}

	providers, err := store.ListProviders(context.Background())
	if err != nil {
		t.Fatalf("ListProviders() error = %v", err)
	}
	if providers[0].Name != "OpenAI Updated" || providers[0].IsEnabled {
		t.Fatalf("updated provider = %#v", providers[0])
	}
	models, err := store.ListModels(context.Background(), provider.ID, "")
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}
	if models[0].Code != "gpt-4.1-mini" || models[0].IsEnabled {
		t.Fatalf("updated model = %#v", models[0])
	}
	keys, err := store.ListAPIKeys(context.Background(), provider.ID)
	if err != nil {
		t.Fatalf("ListAPIKeys() error = %v", err)
	}
	if keys[0].Alias != "main-updated" || keys[0].IsEnabled || keys[0].IsAvailable || keys[0].SortOrder != 7 {
		t.Fatalf("updated key = %#v", keys[0])
	}
}

func TestReorderAPIKeys(t *testing.T) {
	store := newTestStore(t)

	provider, err := store.CreateProvider(context.Background(), CreateProviderInput{
		Name:             "OpenAI",
		Code:             "openai",
		IsEnabled:        true,
		RotationStrategy: "ROUND_ROBIN",
	})
	if err != nil {
		t.Fatalf("CreateProvider() error = %v", err)
	}
	first, err := store.CreateAPIKey(context.Background(), CreateAPIKeyInput{
		ProviderID:  provider.ID,
		Alias:       "first",
		SecretValue: "sk-first",
		IsEnabled:   true,
		IsAvailable: true,
		SortOrder:   1,
	})
	if err != nil {
		t.Fatalf("CreateAPIKey(first) error = %v", err)
	}
	second, err := store.CreateAPIKey(context.Background(), CreateAPIKeyInput{
		ProviderID:  provider.ID,
		Alias:       "second",
		SecretValue: "sk-second",
		IsEnabled:   true,
		IsAvailable: true,
		SortOrder:   2,
	})
	if err != nil {
		t.Fatalf("CreateAPIKey(second) error = %v", err)
	}
	third, err := store.CreateAPIKey(context.Background(), CreateAPIKeyInput{
		ProviderID:  provider.ID,
		Alias:       "third",
		SecretValue: "sk-third",
		IsEnabled:   true,
		IsAvailable: true,
		SortOrder:   3,
	})
	if err != nil {
		t.Fatalf("CreateAPIKey(third) error = %v", err)
	}

	if err := store.ReorderAPIKeys(context.Background(), provider.ID, []string{third.ID, first.ID, second.ID}); err != nil {
		t.Fatalf("ReorderAPIKeys() error = %v", err)
	}

	keys, err := store.ListAPIKeys(context.Background(), provider.ID)
	if err != nil {
		t.Fatalf("ListAPIKeys() error = %v", err)
	}
	got := []string{keys[0].Alias, keys[1].Alias, keys[2].Alias}
	want := []string{"third", "first", "second"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("key order = %v, want %v", got, want)
		}
	}
}

func TestAccessDataAndPermissions(t *testing.T) {
	store := newTestStore(t)

	provider, err := store.CreateProvider(context.Background(), CreateProviderInput{
		Name:             "OpenAI",
		Code:             "openai",
		IsEnabled:        true,
		RotationStrategy: "ROUND_ROBIN",
	})
	if err != nil {
		t.Fatalf("CreateProvider() error = %v", err)
	}
	model, err := store.CreateModel(context.Background(), CreateModelInput{
		ProviderID: provider.ID,
		Name:       "GPT 4.1",
		Code:       "gpt-4.1",
		IsEnabled:  true,
	})
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}

	if err := store.SeedDemoAccessData(context.Background()); err != nil {
		t.Fatalf("SeedDemoAccessData() error = %v", err)
	}
	channels, err := store.ListChannels(context.Background())
	if err != nil {
		t.Fatalf("ListChannels() error = %v", err)
	}
	if len(channels) != 2 {
		t.Fatalf("channels length = %d, want 2", len(channels))
	}

	users, err := store.ListUsers(context.Background(), channels[0].ID)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) == 0 {
		t.Fatal("expected demo users")
	}

	if err := store.SetChannelPermissionDefault(context.Background(), channels[0].ID, provider.ID, model.ID, true); err != nil {
		t.Fatalf("SetChannelPermissionDefault() error = %v", err)
	}
	channelRows, err := store.ListChannelPermissionRows(context.Background(), channels[0].ID)
	if err != nil {
		t.Fatalf("ListChannelPermissionRows() error = %v", err)
	}
	if len(channelRows) != 1 || !channelRows[0].DefaultAllowed || !channelRows[0].HasDefault {
		t.Fatalf("channel permission rows = %#v", channelRows)
	}

	if err := store.SetUserPermission(context.Background(), users[0].ID, provider.ID, model.ID, true); err != nil {
		t.Fatalf("SetUserPermission() error = %v", err)
	}
	userRows, err := store.ListUserPermissionRows(context.Background(), users[0].ID)
	if err != nil {
		t.Fatalf("ListUserPermissionRows() error = %v", err)
	}
	if len(userRows) != 1 || !userRows[0].Allowed || !userRows[0].HasExplicit {
		t.Fatalf("user permission rows = %#v", userRows)
	}
	apiKey, err := store.CreateAPIKey(context.Background(), CreateAPIKeyInput{
		ProviderID:  provider.ID,
		Alias:       "main-key",
		SecretValue: "sk-main",
		IsEnabled:   true,
		IsAvailable: true,
		SortOrder:   1,
	})
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}
	keyRows, err := store.ListUserKeyPermissionRows(context.Background(), users[0].ID, provider.ID)
	if err != nil {
		t.Fatalf("ListUserKeyPermissionRows() error = %v", err)
	}
	if len(keyRows) != 1 || keyRows[0].KeyID != apiKey.ID || !keyRows[0].Allowed || keyRows[0].HasExplicit {
		t.Fatalf("default user key permission rows = %#v", keyRows)
	}
	if err := store.SetUserKeyPermission(context.Background(), users[0].ID, provider.ID, apiKey.ID, false); err != nil {
		t.Fatalf("SetUserKeyPermission() error = %v", err)
	}
	keyRows, err = store.ListUserKeyPermissionRows(context.Background(), users[0].ID, provider.ID)
	if err != nil {
		t.Fatalf("ListUserKeyPermissionRows() after set error = %v", err)
	}
	if len(keyRows) != 1 || keyRows[0].Allowed || !keyRows[0].HasExplicit {
		t.Fatalf("explicit user key permission rows = %#v", keyRows)
	}
}

func TestRuntimeDispatchAndFailureReport(t *testing.T) {
	store := newTestStore(t)

	provider, err := store.CreateProvider(context.Background(), CreateProviderInput{
		Name:             "OpenAI",
		Code:             "openai",
		IsEnabled:        true,
		RotationStrategy: "ROUND_ROBIN",
	})
	if err != nil {
		t.Fatalf("CreateProvider() error = %v", err)
	}
	model, err := store.CreateModel(context.Background(), CreateModelInput{
		ProviderID: provider.ID,
		Name:       "GPT 4.1",
		Code:       "gpt-4.1",
		IsEnabled:  true,
	})
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}
	first, err := store.CreateAPIKey(context.Background(), CreateAPIKeyInput{
		ProviderID:  provider.ID,
		Alias:       "first",
		SecretValue: "sk-first",
		IsEnabled:   true,
		IsAvailable: true,
		SortOrder:   1,
	})
	if err != nil {
		t.Fatalf("CreateAPIKey(first) error = %v", err)
	}
	second, err := store.CreateAPIKey(context.Background(), CreateAPIKeyInput{
		ProviderID:  provider.ID,
		Alias:       "second",
		SecretValue: "sk-second",
		IsEnabled:   true,
		IsAvailable: true,
		SortOrder:   2,
	})
	if err != nil {
		t.Fatalf("CreateAPIKey(second) error = %v", err)
	}
	channel, err := store.CreateChannel(context.Background(), CreateChannelInput{
		Name:                  "Main",
		Code:                  "main",
		DefaultPermissionMode: "DENY",
		UserManagementMode:    "EXTERNAL_MANAGED",
		IsEnabled:             true,
	})
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	user, err := store.UpsertRuntimeExternalUser(context.Background(), UpsertRuntimeExternalUserInput{
		ChannelName:    channel.Name,
		ExternalUserID: "student-001",
		Name:           "Student 001",
		IsEnabled:      true,
	})
	if err != nil {
		t.Fatalf("UpsertRuntimeExternalUser() error = %v", err)
	}
	if err := store.SetChannelPermissionDefault(context.Background(), channel.ID, provider.ID, model.ID, true); err != nil {
		t.Fatalf("SetChannelPermissionDefault() error = %v", err)
	}

	dispatch, err := store.DispatchRuntimeKey(context.Background(), DispatchKeyInput{
		ChannelName: channel.Name,
		UserID:      user.ID,
		ProviderID:  provider.ID,
		ModelID:     model.ID,
	})
	if err != nil {
		t.Fatalf("DispatchRuntimeKey() error = %v", err)
	}
	if dispatch.KeyID != first.ID || dispatch.Key != "sk-first" {
		t.Fatalf("first dispatch = %#v, want first key", dispatch)
	}
	secondDispatch, err := store.DispatchRuntimeKey(context.Background(), DispatchKeyInput{
		ChannelName: channel.Name,
		UserID:      user.ID,
		ProviderID:  provider.ID,
		ModelID:     model.ID,
	})
	if err != nil {
		t.Fatalf("DispatchRuntimeKey() second error = %v", err)
	}
	if secondDispatch.KeyID != second.ID {
		t.Fatalf("second dispatch key id = %s, want %s", secondDispatch.KeyID, second.ID)
	}
	if err := store.SetUserKeyPermission(context.Background(), user.ID, provider.ID, first.ID, false); err != nil {
		t.Fatalf("SetUserKeyPermission(first) error = %v", err)
	}
	if err := store.SetUserKeyPermission(context.Background(), user.ID, provider.ID, second.ID, true); err != nil {
		t.Fatalf("SetUserKeyPermission(second) error = %v", err)
	}
	filteredDispatch, err := store.DispatchRuntimeKey(context.Background(), DispatchKeyInput{
		ChannelName: channel.Name,
		UserID:      user.ID,
		ProviderID:  provider.ID,
		ModelID:     model.ID,
	})
	if err != nil {
		t.Fatalf("DispatchRuntimeKey() filtered error = %v", err)
	}
	if filteredDispatch.KeyID != second.ID {
		t.Fatalf("filtered dispatch key id = %s, want %s", filteredDispatch.KeyID, second.ID)
	}

	report, err := store.ReportRuntimeKeyFailure(context.Background(), dispatch.DispatchLogID, "rate_limit", "provider returned 429")
	if err != nil {
		t.Fatalf("ReportRuntimeKeyFailure() error = %v", err)
	}
	if report.KeyID != first.ID || report.IsAvailable {
		t.Fatalf("failure report = %#v, want first unavailable", report)
	}
	keys, err := store.ListAPIKeys(context.Background(), provider.ID)
	if err != nil {
		t.Fatalf("ListAPIKeys() error = %v", err)
	}
	for _, key := range keys {
		if key.ID == first.ID && key.IsAvailable {
			t.Fatalf("failed key is still available: %#v", key)
		}
	}
	updatedUser, err := store.UpsertRuntimeExternalUser(context.Background(), UpsertRuntimeExternalUserInput{
		ChannelName:    channel.Name,
		ExternalUserID: "student-001",
		Name:           "Student 001",
		IsEnabled:      false,
	})
	if err != nil {
		t.Fatalf("UpsertRuntimeExternalUser() update error = %v", err)
	}
	if updatedUser.ID != user.ID || updatedUser.IsEnabled {
		t.Fatalf("updated user = %#v, want same disabled user %s", updatedUser, user.ID)
	}
	if err := store.DeleteUser(context.Background(), user.ID); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
}

func TestQueueFullFailureDoesNotDisableKey(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	provider, err := store.CreateProvider(ctx, CreateProviderInput{
		Name:             "RunningHub",
		Code:             "runninghub",
		IsEnabled:        true,
		RotationStrategy: "ROUND_ROBIN",
	})
	if err != nil {
		t.Fatalf("CreateProvider() error = %v", err)
	}
	model, err := store.CreateModel(ctx, CreateModelInput{
		ProviderID: provider.ID,
		Name:       "InfiniteTalk",
		Code:       "infinitetalk",
		IsEnabled:  true,
	})
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}
	key, err := store.CreateAPIKey(ctx, CreateAPIKeyInput{
		ProviderID:  provider.ID,
		Alias:       "18501079950",
		SecretValue: "sk-runninghub",
		IsEnabled:   true,
		IsAvailable: true,
		SortOrder:   1,
	})
	if err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}
	channel, err := store.CreateChannel(ctx, CreateChannelInput{
		Name:                  "ai_video_maker",
		Code:                  "ai-video-maker",
		DefaultPermissionMode: "DENY",
		UserManagementMode:    "EXTERNAL_MANAGED",
		IsEnabled:             true,
	})
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	user, err := store.UpsertRuntimeExternalUser(ctx, UpsertRuntimeExternalUserInput{
		ChannelName:    channel.Name,
		ExternalUserID: "luxi",
		Name:           "luxi",
		IsEnabled:      true,
	})
	if err != nil {
		t.Fatalf("UpsertRuntimeExternalUser() error = %v", err)
	}
	if err := store.SetChannelPermissionDefault(ctx, channel.ID, provider.ID, model.ID, true); err != nil {
		t.Fatalf("SetChannelPermissionDefault() error = %v", err)
	}

	dispatch, err := store.DispatchRuntimeKey(ctx, DispatchKeyInput{
		ChannelName: channel.Name,
		UserID:      user.ID,
		ProviderID:  provider.ID,
		ModelID:     model.ID,
	})
	if err != nil {
		t.Fatalf("DispatchRuntimeKey() error = %v", err)
	}
	report, err := store.ReportRuntimeKeyFailure(ctx, dispatch.DispatchLogID, "provider_error", "Error: [portrait_animator] 任务队列已满，请稍后重试")
	if err != nil {
		t.Fatalf("ReportRuntimeKeyFailure() error = %v", err)
	}
	if !report.IsAvailable {
		t.Fatalf("queue full report disabled key: %#v", report)
	}
	keys, err := store.ListAPIKeys(ctx, provider.ID)
	if err != nil {
		t.Fatalf("ListAPIKeys() error = %v", err)
	}
	if len(keys) != 1 || keys[0].ID != key.ID || !keys[0].IsAvailable || keys[0].FailureCount != 0 {
		t.Fatalf("key after queue full report = %#v", keys)
	}
	history, _, err := store.ListDispatchHistory(ctx, DispatchHistoryFilter{Status: "FAILED", KeyID: key.ID, PageSize: 10})
	if err != nil {
		t.Fatalf("ListDispatchHistory() error = %v", err)
	}
	if len(history) != 1 || history[0].FailureErrorMessage == nil {
		t.Fatalf("history after queue full report = %#v", history)
	}
}

func newTestStore(t *testing.T) *Store {
	t.Helper()

	database, err := keydb.Open(context.Background(), t.TempDir()+"/admin-test.db")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := database.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	store, err := NewStore(Options{
		DB:  database.SQL(),
		Now: func() time.Time { return time.Date(2026, 5, 6, 1, 2, 3, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	return store
}
