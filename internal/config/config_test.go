package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReadsEnvFile(t *testing.T) {
	clearConfigEnv(t)

	envPath := filepath.Join(t.TempDir(), ".env")
	content := strings.Join([]string{
		"KEYCHAIN_ADMIN_USERNAME=admin",
		"KEYCHAIN_ADMIN_PASSWORD=secret",
		"KEYCHAIN_SESSION_SECRET=session-secret",
		"KEYCHAIN_RUNTIME_API_TOKEN=runtime-token",
		"KEYCHAIN_DB_PATH=./keychain.db",
		"KEYCHAIN_ADDR=127.0.0.1:8090",
		"KEYCHAIN_DATA_DIR=runtime-data",
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
	if cfg.HTTPAddr != "127.0.0.1:8090" {
		t.Fatalf("HTTPAddr = %q, want 127.0.0.1:8090", cfg.HTTPAddr)
	}
	if cfg.DataDir != "runtime-data" {
		t.Fatalf("DataDir = %q, want runtime-data", cfg.DataDir)
	}
}

func TestLoadLetsEnvironmentOverrideEnvFile(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("KEYCHAIN_ADMIN_USERNAME", "admin-from-env")
	t.Setenv("KEYCHAIN_ADDR", "127.0.0.1:9090")

	envPath := filepath.Join(t.TempDir(), ".env")
	content := strings.Join([]string{
		"KEYCHAIN_ADMIN_USERNAME=admin-from-file",
		"KEYCHAIN_ADMIN_PASSWORD=secret",
		"KEYCHAIN_SESSION_SECRET=session-secret",
		"KEYCHAIN_RUNTIME_API_TOKEN=runtime-token",
		"KEYCHAIN_DB_PATH=./keychain.db",
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

func TestLoadSupportsLegacyEnvironmentNames(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("ADMIN_PASSWORD", "secret")
	t.Setenv("SESSION_SECRET", "session-secret")
	t.Setenv("RUNTIME_API_TOKEN", "runtime-token")
	t.Setenv("DATABASE_PATH", "./legacy.db")
	t.Setenv("HTTP_ADDR", "127.0.0.1:9091")

	cfg, err := Load(filepath.Join(t.TempDir(), ".env"))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DatabasePath != "./legacy.db" {
		t.Fatalf("DatabasePath = %q, want ./legacy.db", cfg.DatabasePath)
	}
	if cfg.HTTPAddr != "127.0.0.1:9091" {
		t.Fatalf("HTTPAddr = %q, want 127.0.0.1:9091", cfg.HTTPAddr)
	}
}

func TestLoadUsesSafeLocalDefaults(t *testing.T) {
	clearConfigEnv(t)
	envPath := filepath.Join(t.TempDir(), ".env")
	content := strings.Join([]string{
		"KEYCHAIN_ADMIN_PASSWORD=secret",
		"KEYCHAIN_SESSION_SECRET=session-secret",
		"KEYCHAIN_RUNTIME_API_TOKEN=runtime-token",
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
	if cfg.HTTPAddr != "127.0.0.1:8080" {
		t.Fatalf("HTTPAddr = %q, want 127.0.0.1:8080", cfg.HTTPAddr)
	}
	if cfg.DatabasePath != "app.db" {
		t.Fatalf("DatabasePath = %q, want app.db", cfg.DatabasePath)
	}
	if cfg.DataDir != "data" {
		t.Fatalf("DataDir = %q, want data", cfg.DataDir)
	}
}

func TestLoadRejectsMissingRequiredSecrets(t *testing.T) {
	clearConfigEnv(t)

	_, err := Load(filepath.Join(t.TempDir(), ".env"))
	if err == nil {
		t.Fatal("Load() error = nil, want missing config error")
	}
	if !strings.Contains(err.Error(), "ADMIN_PASSWORD") {
		t.Fatalf("Load() error = %q, want missing ADMIN_PASSWORD", err)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"KEYCHAIN_ADMIN_USERNAME",
		"KEYCHAIN_ADMIN_PASSWORD",
		"KEYCHAIN_SESSION_SECRET",
		"KEYCHAIN_RUNTIME_API_TOKEN",
		"KEYCHAIN_DB_PATH",
		"KEYCHAIN_DATA_DIR",
		"KEYCHAIN_ADDR",
		"ADMIN_USERNAME",
		"ADMIN_PASSWORD",
		"SESSION_SECRET",
		"RUNTIME_API_TOKEN",
		"DATABASE_PATH",
		"HTTP_ADDR",
	} {
		t.Setenv(key, "")
	}
}
