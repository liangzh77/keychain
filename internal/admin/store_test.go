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
