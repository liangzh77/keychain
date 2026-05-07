package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	AdminUsername   string
	AdminPassword   string
	SessionSecret   string
	RuntimeAPIToken string
	DatabasePath    string
	DataDir         string
	HTTPAddr        string
}

func Load(envPath string) (Config, error) {
	fileValues, err := readEnvFile(envPath)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AdminUsername:   getConfigValue(fileValues, "KEYCHAIN_ADMIN_USERNAME", "ADMIN_USERNAME"),
		AdminPassword:   getConfigValue(fileValues, "KEYCHAIN_ADMIN_PASSWORD", "ADMIN_PASSWORD"),
		SessionSecret:   getConfigValue(fileValues, "KEYCHAIN_SESSION_SECRET", "SESSION_SECRET"),
		RuntimeAPIToken: getConfigValue(fileValues, "KEYCHAIN_RUNTIME_API_TOKEN", "RUNTIME_API_TOKEN"),
		DatabasePath:    getConfigValue(fileValues, "KEYCHAIN_DB_PATH", "DATABASE_PATH"),
		DataDir:         getConfigValue(fileValues, "KEYCHAIN_DATA_DIR"),
		HTTPAddr:        getConfigValue(fileValues, "KEYCHAIN_ADDR", "HTTP_ADDR"),
	}
	if cfg.AdminUsername == "" {
		cfg.AdminUsername = "admin"
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = "127.0.0.1:8080"
	}
	if cfg.DatabasePath == "" {
		cfg.DatabasePath = "app.db"
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "data"
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (cfg Config) Validate() error {
	var missing []string
	if cfg.AdminUsername == "" {
		missing = append(missing, "ADMIN_USERNAME")
	}
	if cfg.AdminPassword == "" {
		missing = append(missing, "ADMIN_PASSWORD")
	}
	if cfg.SessionSecret == "" {
		missing = append(missing, "SESSION_SECRET")
	}
	if cfg.RuntimeAPIToken == "" {
		missing = append(missing, "RUNTIME_API_TOKEN")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required config: %s", strings.Join(missing, ", "))
	}
	return nil
}

func readEnvFile(path string) (map[string]string, error) {
	values := make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return values, nil
		}
		return nil, fmt.Errorf("read env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("invalid env line %d: missing =", lineNumber)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, fmt.Errorf("invalid env line %d: empty key", lineNumber)
		}
		values[key] = trimEnvValue(value)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan env file: %w", err)
	}
	return values, nil
}

func trimEnvValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		first := value[0]
		last := value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func getConfigValue(fileValues map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	for _, key := range keys {
		if value := strings.TrimSpace(fileValues[key]); value != "" {
			return value
		}
	}
	return ""
}
