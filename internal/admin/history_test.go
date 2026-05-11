package admin

import (
	"context"
	"testing"
)

func TestDispatchHistoryQueries(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	provider, err := store.CreateProvider(ctx, CreateProviderInput{
		Name:             "OpenAI",
		Code:             "openai",
		IsEnabled:        true,
		RotationStrategy: "ROUND_ROBIN",
	})
	if err != nil {
		t.Fatalf("CreateProvider() error = %v", err)
	}
	model, err := store.CreateModel(ctx, CreateModelInput{
		ProviderID: provider.ID,
		Name:       "GPT 4.1",
		Code:       "gpt-4.1",
		IsEnabled:  true,
	})
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}
	first, err := store.CreateAPIKey(ctx, CreateAPIKeyInput{
		ProviderID:  provider.ID,
		Alias:       "first-key",
		SecretValue: "sk-first",
		IsEnabled:   true,
		IsAvailable: true,
		SortOrder:   1,
	})
	if err != nil {
		t.Fatalf("CreateAPIKey(first) error = %v", err)
	}
	second, err := store.CreateAPIKey(ctx, CreateAPIKeyInput{
		ProviderID:  provider.ID,
		Alias:       "second-key",
		SecretValue: "sk-second",
		IsEnabled:   true,
		IsAvailable: true,
		SortOrder:   2,
	})
	if err != nil {
		t.Fatalf("CreateAPIKey(second) error = %v", err)
	}
	channel, err := store.CreateChannel(ctx, CreateChannelInput{
		Name:                  "Main",
		Code:                  "main",
		DefaultPermissionMode: "DENY",
		UserManagementMode:    "EXTERNAL_MANAGED",
		IsEnabled:             true,
	})
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	user, err := store.UpsertRuntimeExternalUser(ctx, UpsertRuntimeExternalUserInput{
		ChannelName:    channel.Name,
		ExternalUserID: "student-001",
		Name:           "Student 001",
		IsEnabled:      true,
	})
	if err != nil {
		t.Fatalf("UpsertRuntimeExternalUser() error = %v", err)
	}
	if err := store.SetChannelPermissionDefault(ctx, channel.ID, provider.ID, model.ID, true); err != nil {
		t.Fatalf("SetChannelPermissionDefault() error = %v", err)
	}
	if err := store.SetUserPermission(ctx, user.ID, provider.ID, model.ID, true); err != nil {
		t.Fatalf("SetUserPermission() error = %v", err)
	}
	if err := store.SetUserKeyPermission(ctx, user.ID, provider.ID, first.ID, true); err != nil {
		t.Fatalf("SetUserKeyPermission(first) error = %v", err)
	}
	if err := store.SetUserKeyPermission(ctx, user.ID, provider.ID, second.ID, true); err != nil {
		t.Fatalf("SetUserKeyPermission(second) error = %v", err)
	}

	failedDispatch, err := store.DispatchRuntimeKey(ctx, DispatchKeyInput{
		ChannelName: channel.Name,
		UserID:      user.ID,
		ProviderID:  provider.ID,
		ModelID:     model.ID,
	})
	if err != nil {
		t.Fatalf("DispatchRuntimeKey(first) error = %v", err)
	}
	successDispatch, err := store.DispatchRuntimeKey(ctx, DispatchKeyInput{
		ChannelName: channel.Name,
		UserID:      user.ID,
		ProviderID:  provider.ID,
		ModelID:     model.ID,
	})
	if err != nil {
		t.Fatalf("DispatchRuntimeKey(second) error = %v", err)
	}
	if successDispatch.KeyID != second.ID {
		t.Fatalf("second dispatch key = %s, want %s", successDispatch.KeyID, second.ID)
	}
	if _, err := store.ReportRuntimeKeyFailure(ctx, failedDispatch.DispatchLogID, "rate_limit", "provider returned 429"); err != nil {
		t.Fatalf("ReportRuntimeKeyFailure() error = %v", err)
	}
	userOptions, err := store.ListDispatchHistoryUserOptions(ctx, "")
	if err != nil {
		t.Fatalf("ListDispatchHistoryUserOptions() error = %v", err)
	}
	if len(userOptions) != 1 || userOptions[0].ID != user.ID || userOptions[0].ChannelName != channel.Name {
		t.Fatalf("user options = %#v", userOptions)
	}

	filter := DispatchHistoryFilter{ChannelID: channel.ID, UserID: user.ID, ProviderID: provider.ID, ModelID: model.ID, PageSize: 10}
	rows, pagination, err := store.ListDispatchHistory(ctx, filter)
	if err != nil {
		t.Fatalf("ListDispatchHistory() error = %v", err)
	}
	if len(rows) != 2 || pagination.TotalItems != 2 {
		t.Fatalf("history rows = %d total = %d, want 2", len(rows), pagination.TotalItems)
	}
	stats, err := store.GetDispatchHistoryStats(ctx, filter)
	if err != nil {
		t.Fatalf("GetDispatchHistoryStats() error = %v", err)
	}
	if stats.TotalCount != 2 || stats.FailedCount != 1 || stats.SuccessCount != 1 || stats.UniqueUserCount != 1 || stats.UniqueKeyCount != 2 {
		t.Fatalf("stats = %#v", stats)
	}
	failedRows, _, err := store.ListDispatchHistory(ctx, DispatchHistoryFilter{Status: "FAILED", KeyID: first.ID, PageSize: 10})
	if err != nil {
		t.Fatalf("ListDispatchHistory(failed) error = %v", err)
	}
	if len(failedRows) != 1 || failedRows[0].KeyAlias != "first-key" || failedRows[0].FailureErrorCode == nil {
		t.Fatalf("failed rows = %#v", failedRows)
	}
	points, err := store.GetDispatchHistorySeries(ctx, filter, "day")
	if err != nil {
		t.Fatalf("GetDispatchHistorySeries() error = %v", err)
	}
	if len(points) != 1 || points[0].TotalCount != 2 || points[0].FailedCount != 1 {
		t.Fatalf("series = %#v", points)
	}
}
