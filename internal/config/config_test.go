package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReadsEnvFile(t *testing.T) {
	t.Setenv("ADMIN_USERNAME", "")
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("SESSION_SECRET", "")
	t.Setenv("RUNTIME_API_TOKEN", "")
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("HTTP_ADDR", "")

	envPath := filepath.Join(t.TempDir(), ".env")
	content := strings.Join([]string{
		"ADMIN_USERNAME=admin",
		"ADMIN_PASSWORD=secret",
		"SESSION_SECRET=session-secret",
		"RUNTIME_API_TOKEN=runtime-token",
		"DATABASE_PATH=./keychain.db",
	}, "\n")
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := Load(envPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AdminUsername != "admin" {
		t.Fatalf("AdminUsername = %q, want admin", cfg.AdminUsername)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
}

func TestLoadLetsEnvironmentOverrideEnvFile(t *testing.T) {
	t.Setenv("ADMIN_USERNAME", "admin-from-env")
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("SESSION_SECRET", "")
	t.Setenv("RUNTIME_API_TOKEN", "")
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("HTTP_ADDR", "127.0.0.1:9090")

	envPath := filepath.Join(t.TempDir(), ".env")
	content := strings.Join([]string{
		"ADMIN_USERNAME=admin-from-file",
		"ADMIN_PASSWORD=secret",
		"SESSION_SECRET=session-secret",
		"RUNTIME_API_TOKEN=runtime-token",
		"DATABASE_PATH=./keychain.db",
	}, "\n")
	if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := Load(envPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AdminUsername != "admin-from-env" {
		t.Fatalf("AdminUsername = %q, want admin-from-env", cfg.AdminUsername)
	}
	if cfg.HTTPAddr != "127.0.0.1:9090" {
		t.Fatalf("HTTPAddr = %q, want 127.0.0.1:9090", cfg.HTTPAddr)
	}
}

func TestLoadRejectsMissingRequiredConfig(t *testing.T) {
	t.Setenv("ADMIN_USERNAME", "")
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("SESSION_SECRET", "")
	t.Setenv("RUNTIME_API_TOKEN", "")
	t.Setenv("DATABASE_PATH", "")
	t.Setenv("HTTP_ADDR", "")

	_, err := Load(filepath.Join(t.TempDir(), ".env"))
	if err == nil {
		t.Fatal("Load() error = nil, want missing config error")
	}
	if !strings.Contains(err.Error(), "ADMIN_USERNAME") {
		t.Fatalf("Load() error = %q, want missing ADMIN_USERNAME", err)
	}
}
