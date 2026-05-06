package auth

import (
	"context"
	"testing"
	"time"

	keydb "github.com/liangzh77/keychain/internal/db"
)

func TestAuthenticateUsesConfiguredAdmin(t *testing.T) {
	service := newTestService(t, time.Now)

	if !service.Authenticate("admin", "password") {
		t.Fatal("Authenticate() = false, want true")
	}
	if service.Authenticate("admin", "wrong") {
		t.Fatal("Authenticate() with wrong password = true, want false")
	}
}

func TestSessionLifecycle(t *testing.T) {
	now := time.Date(2026, 5, 6, 1, 2, 3, 0, time.UTC)
	service := newTestService(t, func() time.Time { return now })

	token, created, err := service.CreateSession(context.Background())
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if token == "" {
		t.Fatal("CreateSession() returned empty token")
	}
	if !created.ExpiresAt.Equal(now.Add(24 * time.Hour)) {
		t.Fatalf("ExpiresAt = %v, want 24h later", created.ExpiresAt)
	}

	session, ok, err := service.GetSession(context.Background(), token)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if !ok {
		t.Fatal("GetSession() ok = false, want true")
	}
	if session.Username != "admin" {
		t.Fatalf("Username = %q, want admin", session.Username)
	}

	if err := service.DeleteSession(context.Background(), token); err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	_, ok, err = service.GetSession(context.Background(), token)
	if err != nil {
		t.Fatalf("GetSession() after delete error = %v", err)
	}
	if ok {
		t.Fatal("GetSession() after delete ok = true, want false")
	}
}

func TestExpiredSessionIsRejected(t *testing.T) {
	now := time.Date(2026, 5, 6, 1, 2, 3, 0, time.UTC)
	currentTime := now
	service := newTestService(t, func() time.Time { return currentTime })

	token, _, err := service.CreateSession(context.Background())
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	currentTime = now.Add(25 * time.Hour)
	_, ok, err := service.GetSession(context.Background(), token)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if ok {
		t.Fatal("GetSession() for expired session ok = true, want false")
	}
}

func newTestService(t *testing.T, now func() time.Time) *Service {
	t.Helper()

	database, err := keydb.Open(context.Background(), t.TempDir()+"/auth-test.db")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { database.Close() })
	if err := database.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	service, err := NewService(Options{
		DB:            database.SQL(),
		AdminUsername: "admin",
		AdminPassword: "password",
		SessionSecret: "test-session-secret",
		Now:           now,
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return service
}
