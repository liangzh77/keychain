package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/liangzh77/keychain/internal/auth"
	keydb "github.com/liangzh77/keychain/internal/db"
)

func TestAuthAPIFlow(t *testing.T) {
	handler := newAuthTestRouter(t)

	loginBody := bytes.NewBufferString(`{"username":"admin","password":"password"}`)
	loginRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	loginResponse := httptest.NewRecorder()
	handler.ServeHTTP(loginResponse, loginRequest)

	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d, body=%s", loginResponse.Code, http.StatusOK, loginResponse.Body.String())
	}
	cookies := loginResponse.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != auth.CookieName {
		t.Fatalf("login cookies = %#v, want %s", cookies, auth.CookieName)
	}

	meRequest := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meRequest.AddCookie(cookies[0])
	meRecorder := httptest.NewRecorder()
	handler.ServeHTTP(meRecorder, meRequest)

	if meRecorder.Code != http.StatusOK {
		t.Fatalf("me status = %d, want %d", meRecorder.Code, http.StatusOK)
	}
	var body meResponse
	if err := json.Unmarshal(meRecorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode me response: %v", err)
	}
	if body.Username != "admin" {
		t.Fatalf("username = %q, want admin", body.Username)
	}

	logoutRequest := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutRequest.AddCookie(cookies[0])
	logoutResponse := httptest.NewRecorder()
	handler.ServeHTTP(logoutResponse, logoutRequest)
	if logoutResponse.Code != http.StatusOK {
		t.Fatalf("logout status = %d, want %d", logoutResponse.Code, http.StatusOK)
	}
}

func TestAuthAPIRejectsInvalidCredentials(t *testing.T) {
	handler := newAuthTestRouter(t)

	loginBody := bytes.NewBufferString(`{"username":"admin","password":"wrong"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", loginBody)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if strings.Contains(response.Body.String(), "password") {
		t.Fatalf("response body leaks password context: %s", response.Body.String())
	}
}

func TestMeRequiresLogin(t *testing.T) {
	handler := newAuthTestRouter(t)

	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestAdminPageRedirectsUntilLoggedIn(t *testing.T) {
	handler := newAuthTestRouter(t)

	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusFound)
	}
	if location := response.Header().Get("Location"); location != "/login" {
		t.Fatalf("Location = %q, want /login", location)
	}
}

func newAuthTestRouter(t *testing.T) http.Handler {
	t.Helper()

	database, err := keydb.Open(context.Background(), t.TempDir()+"/server-auth-test.db")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := database.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	authService, err := auth.NewService(auth.Options{
		DB:            database.SQL(),
		AdminUsername: "admin",
		AdminPassword: "password",
		SessionSecret: "test-session-secret",
		Now:           func() time.Time { return time.Date(2026, 5, 6, 1, 2, 3, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return NewRouter(Options{Auth: authService})
}
