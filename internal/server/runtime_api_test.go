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

	userPath := "/api/runtime/channels/" + fixtures.ChannelID + "/external-users/student-001"
	userResponse := runtimeRequest[runtimeUserResponse](t, handler, http.MethodPut, userPath, map[string]any{
		"name": "Student 001",
	})
	if userResponse.ID == "" || userResponse.ExternalUserID != "student-001" || userResponse.Name != "Student 001" {
		t.Fatalf("user response = %#v", userResponse)
	}
	updatedUser := runtimeRequest[runtimeUserResponse](t, handler, http.MethodPut, userPath, map[string]any{
		"name":      "Student 001",
		"isEnabled": false,
	})
	if updatedUser.ID != userResponse.ID || updatedUser.IsEnabled {
		t.Fatalf("updated user = %#v, want same disabled user %s", updatedUser, userResponse.ID)
	}
	userResponse = runtimeRequest[runtimeUserResponse](t, handler, http.MethodPut, userPath, map[string]any{
		"name": "Student 001",
	})

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

	dispatch := postRuntime[dispatchKeyResponse](t, handler, "/api/runtime/dispatches", map[string]any{
		"channelId":  fixtures.ChannelID,
		"userId":     userResponse.ID,
		"providerId": fixtures.ProviderID,
		"modelId":    fixtures.ModelID,
	})
	if dispatch.Key != "sk-runtime" || dispatch.KeyAlias != "runtime-main" {
		t.Fatalf("dispatch = %#v, want runtime key", dispatch)
	}

	failure := postRuntime[keyFailureResponse](t, handler, "/api/runtime/dispatches/"+dispatch.DispatchLogID+"/failure", map[string]any{
		"errorCode":    "rate_limit",
		"errorMessage": "provider returned 429",
	})
	if !failure.Reported || failure.IsAvailable {
		t.Fatalf("failure = %#v, want reported unavailable", failure)
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, userPath, nil)
	deleteRequest.Header.Set("Authorization", "Bearer test-runtime-token")
	deleteRecorder := httptest.NewRecorder()
	handler.ServeHTTP(deleteRecorder, deleteRequest)
	if deleteRecorder.Code != http.StatusOK {
		t.Fatalf("delete status = %d, body=%s", deleteRecorder.Code, deleteRecorder.Body.String())
	}
}

func postRuntime[T any](t *testing.T, handler http.Handler, path string, payload map[string]any) T {
	t.Helper()
	return runtimeRequest[T](t, handler, http.MethodPost, path, payload)
}

func runtimeRequest[T any](t *testing.T, handler http.Handler, method string, path string, payload map[string]any) T {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
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

func TestRuntimeExternalUsersRejectHostedChannels(t *testing.T) {
	handler, fixtures := newRuntimeTestRouter(t)

	request := httptest.NewRequest(http.MethodPut, "/api/runtime/channels/"+fixtures.HostedChannelID+"/external-users/student-001", bytes.NewReader([]byte(`{"name":"Student 001"}`)))
	request.Header.Set("Authorization", "Bearer test-runtime-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body=%s", response.Code, response.Body.String())
	}
}

func TestRuntimeHostedUserFlow(t *testing.T) {
	handler, fixtures := newRuntimeTestRouter(t)

	registerPath := "/api/runtime/channels/" + fixtures.HostedChannelID + "/hosted-users/register"
	user := runtimeRequest[runtimeUserResponse](t, handler, http.MethodPost, registerPath, map[string]any{
		"username": "hosted-student-001",
		"name":     "托管学生 001",
		"password": "first-password",
	})
	if user.ID == "" || user.ExternalUserID != "hosted-student-001" || user.Name != "托管学生 001" || !user.IsEnabled {
		t.Fatalf("hosted register response = %#v", user)
	}

	conflictRequest := httptest.NewRequest(http.MethodPost, registerPath, bytes.NewReader([]byte(`{"username":"hosted-student-001","password":"first-password"}`)))
	conflictRequest.Header.Set("Authorization", "Bearer test-runtime-token")
	conflictResponse := httptest.NewRecorder()
	handler.ServeHTTP(conflictResponse, conflictRequest)
	if conflictResponse.Code != http.StatusConflict {
		t.Fatalf("duplicate register status = %d, body=%s", conflictResponse.Code, conflictResponse.Body.String())
	}

	loginPath := "/api/runtime/channels/" + fixtures.HostedChannelID + "/hosted-users/login"
	login := runtimeRequest[runtimeUserResponse](t, handler, http.MethodPost, loginPath, map[string]any{
		"username": "hosted-student-001",
		"password": "first-password",
	})
	if login.ID != user.ID {
		t.Fatalf("login user id = %s, want %s", login.ID, user.ID)
	}

	badLoginRequest := httptest.NewRequest(http.MethodPost, loginPath, bytes.NewReader([]byte(`{"username":"hosted-student-001","password":"bad-password"}`)))
	badLoginRequest.Header.Set("Authorization", "Bearer test-runtime-token")
	badLoginResponse := httptest.NewRecorder()
	handler.ServeHTTP(badLoginResponse, badLoginRequest)
	if badLoginResponse.Code != http.StatusUnauthorized {
		t.Fatalf("bad login status = %d, body=%s", badLoginResponse.Code, badLoginResponse.Body.String())
	}

	resetPath := "/api/runtime/channels/" + fixtures.HostedChannelID + "/hosted-users/" + user.ID + "/reset-password"
	reset := runtimeRequest[runtimeUserResponse](t, handler, http.MethodPost, resetPath, map[string]any{
		"password": "second-password",
	})
	if reset.ID != user.ID {
		t.Fatalf("reset user id = %s, want %s", reset.ID, user.ID)
	}

	newLogin := runtimeRequest[runtimeUserResponse](t, handler, http.MethodPost, loginPath, map[string]any{
		"username": "hosted-student-001",
		"password": "second-password",
	})
	if newLogin.ID != user.ID {
		t.Fatalf("new login user id = %s, want %s", newLogin.ID, user.ID)
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/runtime/channels/"+fixtures.HostedChannelID+"/hosted-users/"+user.ID, nil)
	deleteRequest.Header.Set("Authorization", "Bearer test-runtime-token")
	deleteResponse := httptest.NewRecorder()
	handler.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("delete hosted user status = %d, body=%s", deleteResponse.Code, deleteResponse.Body.String())
	}
}

func TestRuntimeHostedUsersRejectExternalManagedChannels(t *testing.T) {
	handler, fixtures := newRuntimeTestRouter(t)

	request := httptest.NewRequest(http.MethodPost, "/api/runtime/channels/"+fixtures.ChannelID+"/hosted-users/register", bytes.NewReader([]byte(`{"username":"student-001","password":"password"}`)))
	request.Header.Set("Authorization", "Bearer test-runtime-token")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body=%s", response.Code, response.Body.String())
	}
}

type runtimeFixtures struct {
	ChannelID       string
	HostedChannelID string
	ProviderID      string
	ModelID         string
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
		UserManagementMode:    "EXTERNAL_MANAGED",
		IsEnabled:             true,
	})
	if err != nil {
		t.Fatalf("CreateChannel() error = %v", err)
	}
	if err := store.SetChannelPermissionDefault(context.Background(), channel.ID, provider.ID, model.ID, true); err != nil {
		t.Fatalf("SetChannelPermissionDefault() error = %v", err)
	}
	hostedChannel, err := store.CreateChannel(context.Background(), admin.CreateChannelInput{
		Name:                  "Hosted",
		Code:                  "hosted",
		DefaultPermissionMode: "DENY",
		UserManagementMode:    "KEYCHAIN_HOSTED",
		IsEnabled:             true,
	})
	if err != nil {
		t.Fatalf("CreateChannel() hosted error = %v", err)
	}

	handler := NewRouter(Options{AdminStore: store, RuntimeToken: "test-runtime-token"})
	return handler, runtimeFixtures{ChannelID: channel.ID, HostedChannelID: hostedChannel.ID, ProviderID: provider.ID, ModelID: model.ID}
}
