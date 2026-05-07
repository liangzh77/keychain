package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/liangzh77/keychain/internal/admin"
	keydb "github.com/liangzh77/keychain/internal/db"
)

func TestRuntimeAPIRequiresBearerToken(t *testing.T) {
	handler, _ := newRuntimeTestRouter(t)

	request := httptest.NewRequest(http.MethodGet, "/api/runtime/providers", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestRuntimeAPIFlow(t *testing.T) {
	handler, fixtures := newRuntimeTestRouter(t)

	userResponse := postRuntime[runtimeUserResponse](t, handler, "/api/runtime/users", map[string]any{
		"channelId": fixtures.ChannelID,
		"name":      "Student 001",
	})
	if userResponse.ID == "" || userResponse.Name != "Student 001" {
		t.Fatalf("user response = %#v", userResponse)
	}

	permissionsRequest := httptest.NewRequest(http.MethodGet, "/api/runtime/users/"+userResponse.ID+"/permissions", nil)
	permissionsRequest.Header.Set("Authorization", "Bearer test-runtime-token")
	permissionsRecorder := httptest.NewRecorder()
	handler.ServeHTTP(permissionsRecorder, permissionsRequest)
	if permissionsRecorder.Code != http.StatusOK {
		t.Fatalf("permissions status = %d, body=%s", permissionsRecorder.Code, permissionsRecorder.Body.String())
	}
	var permissions runtimePermissionsResponse
	if err := json.Unmarshal(permissionsRecorder.Body.Bytes(), &permissions); err != nil {
		t.Fatalf("decode permissions: %v", err)
	}
	if len(permissions.Permissions) != 1 || !permissions.Permissions[0].Allowed {
		t.Fatalf("permissions = %#v, want one allowed permission", permissions.Permissions)
	}

	dispatch := postRuntime[dispatchKeyResponse](t, handler, "/api/runtime/dispatch-key", map[string]any{
		"channelId":  fixtures.ChannelID,
		"userId":     userResponse.ID,
		"providerId": fixtures.ProviderID,
		"modelId":    fixtures.ModelID,
	})
	if dispatch.Key != "sk-runtime" || dispatch.KeyAlias != "runtime-main" {
		t.Fatalf("dispatch = %#v, want runtime key", dispatch)
	}

	failure := postRuntime[keyFailureResponse](t, handler, "/api/runtime/key-failures", map[string]any{
		"dispatchLogId": dispatch.DispatchLogID,
		"errorCode":     "rate_limit",
		"errorMessage":  "provider returned 429",
	})
	if !failure.Reported || failure.IsAvailable {
		t.Fatalf("failure = %#v, want reported unavailable", failure)
	}
}

func postRuntime[T any](t *testing.T, handler http.Handler, path string, payload map[string]any) T {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer test-runtime-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("%s status = %d, body=%s", path, response.Code, response.Body.String())
	}
	var decoded T
	if err := json.Unmarshal(response.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode %s response: %v", path, err)
	}
	return decoded
}

type runtimeFixtures struct {
	ChannelID  string
	ProviderID string
	ModelID    string
}

func newRuntimeTestRouter(t *testing.T) (http.Handler, runtimeFixtures) {
	t.Helper()

	database, err := keydb.Open(context.Background(), t.TempDir()+"/runtime-api-test.db")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := database.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	store, err := admin.NewStore(admin.Options{
		DB:  database.SQL(),
		Now: func() time.Time { return time.Date(2026, 5, 6, 1, 2, 3, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	provider, err := store.CreateProvider(context.Background(), admin.CreateProviderInput{
		Name:             "OpenAI",
		Code:             "openai",
		IsEnabled:        true,
		RotationStrategy: "STICKY_FIRST_AVAILABLE",
	})
	if err != nil {
		t.Fatalf("CreateProvider() error = %v", err)
	}
	model, err := store.CreateModel(context.Background(), admin.CreateModelInput{
		ProviderID: provider.ID,
		Name:       "GPT 4.1",
		Code:       "gpt-4.1",
		IsEnabled:  true,
	})
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}
	if _, err := store.CreateAPIKey(context.Background(), admin.CreateAPIKeyInput{
		ProviderID:  provider.ID,
		Alias:       "runtime-main",
		SecretValue: "sk-runtime",
		IsEnabled:   true,
		IsAvailable: true,
		SortOrder:   1,
	}); err != nil {
		t.Fatalf("CreateAPIKey() error = %v", err)
	}
	channel, err := store.CreateChannel(context.Background(), admin.CreateChannelInput{
		Name:                  "Main",
		Code:                  "main",
		DefaultPermissionMode: "DENY",
		IsEnabled:             true,
	})
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	if err := store.SetChannelPermissionDefault(context.Background(), channel.ID, provider.ID, model.ID, true); err != nil {
		t.Fatalf("SetChannelPermissionDefault() error = %v", err)
	}

	handler := NewRouter(Options{AdminStore: store, RuntimeToken: "test-runtime-token"})
	return handler, runtimeFixtures{ChannelID: channel.ID, ProviderID: provider.ID, ModelID: model.ID}
}
