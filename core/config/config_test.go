package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
	}
	if cfg.Server.BaseURL != "http://localhost:8080" {
		t.Errorf("Server.BaseURL = %q, want %q", cfg.Server.BaseURL, "http://localhost:8080")
	}
	if cfg.Server.ReadTimeout != 30*time.Second {
		t.Errorf("Server.ReadTimeout = %v, want %v", cfg.Server.ReadTimeout, 30*time.Second)
	}
	if cfg.Server.WriteTimeout != 30*time.Second {
		t.Errorf("Server.WriteTimeout = %v, want %v", cfg.Server.WriteTimeout, 30*time.Second)
	}
	if cfg.Server.ShutdownTimeout != 15*time.Second {
		t.Errorf("Server.ShutdownTimeout = %v, want %v", cfg.Server.ShutdownTimeout, 15*time.Second)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %q, want %q", cfg.Database.Host, "localhost")
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %d, want %d", cfg.Database.Port, 5432)
	}
	if cfg.Database.User != "cpi-auth" {
		t.Errorf("Database.User = %q, want %q", cfg.Database.User, "cpi-auth")
	}
	if cfg.Database.Database != "cpi-auth" {
		t.Errorf("Database.Database = %q, want %q", cfg.Database.Database, "cpi-auth")
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("Database.SSLMode = %q, want %q", cfg.Database.SSLMode, "disable")
	}
	if cfg.Database.MaxConns != 25 {
		t.Errorf("Database.MaxConns = %d, want %d", cfg.Database.MaxConns, 25)
	}

	if cfg.Redis.Host != "localhost" {
		t.Errorf("Redis.Host = %q, want %q", cfg.Redis.Host, "localhost")
	}
	if cfg.Redis.Port != 6379 {
		t.Errorf("Redis.Port = %d, want %d", cfg.Redis.Port, 6379)
	}

	if cfg.Security.JWTSigningAlgorithm != "RS256" {
		t.Errorf("Security.JWTSigningAlgorithm = %q, want %q", cfg.Security.JWTSigningAlgorithm, "RS256")
	}
	if cfg.Security.AccessTokenLifetime != 15*time.Minute {
		t.Errorf("Security.AccessTokenLifetime = %v, want %v", cfg.Security.AccessTokenLifetime, 15*time.Minute)
	}
	if cfg.Security.RefreshTokenLifetime != 7*24*time.Hour {
		t.Errorf("Security.RefreshTokenLifetime = %v, want %v", cfg.Security.RefreshTokenLifetime, 7*24*time.Hour)
	}
	if cfg.Security.IDTokenLifetime != 1*time.Hour {
		t.Errorf("Security.IDTokenLifetime = %v, want %v", cfg.Security.IDTokenLifetime, 1*time.Hour)
	}
	if cfg.Security.AuthCodeLifetime != 10*time.Minute {
		t.Errorf("Security.AuthCodeLifetime = %v, want %v", cfg.Security.AuthCodeLifetime, 10*time.Minute)
	}
	if !cfg.Security.HIBPEnabled {
		t.Error("Security.HIBPEnabled should be true by default")
	}
	if !cfg.Security.RateLimitEnabled {
		t.Error("Security.RateLimitEnabled should be true by default")
	}
	if cfg.Security.Argon2Time != 1 {
		t.Errorf("Security.Argon2Time = %d, want %d", cfg.Security.Argon2Time, 1)
	}
	if cfg.Security.Argon2Memory != 64*1024 {
		t.Errorf("Security.Argon2Memory = %d, want %d", cfg.Security.Argon2Memory, 64*1024)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %q, want %q", cfg.Logging.Level, "info")
	}
	if cfg.Logging.Format != "json" {
		t.Errorf("Logging.Format = %q, want %q", cfg.Logging.Format, "json")
	}

	if !cfg.Metrics.Enabled {
		t.Error("Metrics.Enabled should be true by default")
	}
	if cfg.Metrics.Path != "/metrics" {
		t.Errorf("Metrics.Path = %q, want %q", cfg.Metrics.Path, "/metrics")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	// Should return defaults when no file is provided
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %d, want %d", cfg.Database.Port, 5432)
	}
}

func TestLoadConfig_NonExistentFile(t *testing.T) {
	// Non-existent file should not cause an error, just use defaults
	cfg, err := Load("/tmp/nonexistent-cpi-auth-config-12345.yaml")
	if err != nil {
		t.Fatalf("Load returned error for non-existent file: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
	}
}

func TestLoadConfig_YAMLFile(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")

	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9090
database:
  host: "db.example.com"
  port: 5433
  user: "testuser"
  password: "testpass"
  database: "testdb"
  ssl_mode: "require"
redis:
  host: "redis.example.com"
  port: 6380
security:
  issuer: "https://auth.example.com"
logging:
  level: "debug"
  format: "text"
`
	if err := os.WriteFile(configFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "127.0.0.1")
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 9090)
	}
	if cfg.Database.Host != "db.example.com" {
		t.Errorf("Database.Host = %q, want %q", cfg.Database.Host, "db.example.com")
	}
	if cfg.Database.Port != 5433 {
		t.Errorf("Database.Port = %d, want %d", cfg.Database.Port, 5433)
	}
	if cfg.Database.SSLMode != "require" {
		t.Errorf("Database.SSLMode = %q, want %q", cfg.Database.SSLMode, "require")
	}
	if cfg.Redis.Host != "redis.example.com" {
		t.Errorf("Redis.Host = %q, want %q", cfg.Redis.Host, "redis.example.com")
	}
	if cfg.Redis.Port != 6380 {
		t.Errorf("Redis.Port = %d, want %d", cfg.Redis.Port, 6380)
	}
	if cfg.Security.Issuer != "https://auth.example.com" {
		t.Errorf("Security.Issuer = %q, want %q", cfg.Security.Issuer, "https://auth.example.com")
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want %q", cfg.Logging.Level, "debug")
	}
	if cfg.Logging.Format != "text" {
		t.Errorf("Logging.Format = %q, want %q", cfg.Logging.Format, "text")
	}
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"AF_SERVER_HOST":   "0.0.0.0",
		"AF_SERVER_PORT":   "3000",
		"AF_SERVER_BASE_URL": "https://myapp.com",
		"AF_DB_HOST":       "prod-db.example.com",
		"AF_DB_PORT":       "5434",
		"AF_DB_USER":       "prod_user",
		"AF_DB_PASSWORD":   "prod_secret",
		"AF_DB_DATABASE":   "prod_db",
		"AF_DB_SSL_MODE":   "verify-full",
		"AF_REDIS_HOST":    "prod-redis.example.com",
		"AF_REDIS_PORT":    "6381",
		"AF_ENCRYPTION_KEY": "test-key-123",
		"AF_JWT_ALGORITHM": "ES256",
		"AF_ISSUER":        "https://issuer.example.com",
		"AF_LOG_LEVEL":     "warn",
		"AF_LOG_FORMAT":    "text",
		"AF_HIBP_ENABLED":  "false",
		"AF_METRICS_ENABLED": "false",
	}

	for k, v := range envVars {
		t.Setenv(k, v)
	}

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 3000)
	}
	if cfg.Server.BaseURL != "https://myapp.com" {
		t.Errorf("Server.BaseURL = %q, want %q", cfg.Server.BaseURL, "https://myapp.com")
	}
	if cfg.Database.Host != "prod-db.example.com" {
		t.Errorf("Database.Host = %q, want %q", cfg.Database.Host, "prod-db.example.com")
	}
	if cfg.Database.Port != 5434 {
		t.Errorf("Database.Port = %d, want %d", cfg.Database.Port, 5434)
	}
	if cfg.Database.User != "prod_user" {
		t.Errorf("Database.User = %q, want %q", cfg.Database.User, "prod_user")
	}
	if cfg.Database.Password != "prod_secret" {
		t.Errorf("Database.Password = %q, want %q", cfg.Database.Password, "prod_secret")
	}
	if cfg.Database.SSLMode != "verify-full" {
		t.Errorf("Database.SSLMode = %q, want %q", cfg.Database.SSLMode, "verify-full")
	}
	if cfg.Redis.Host != "prod-redis.example.com" {
		t.Errorf("Redis.Host = %q, want %q", cfg.Redis.Host, "prod-redis.example.com")
	}
	if cfg.Redis.Port != 6381 {
		t.Errorf("Redis.Port = %d, want %d", cfg.Redis.Port, 6381)
	}
	if cfg.Security.EncryptionKey != "test-key-123" {
		t.Errorf("Security.EncryptionKey = %q, want %q", cfg.Security.EncryptionKey, "test-key-123")
	}
	if cfg.Security.JWTSigningAlgorithm != "ES256" {
		t.Errorf("Security.JWTSigningAlgorithm = %q, want %q", cfg.Security.JWTSigningAlgorithm, "ES256")
	}
	if cfg.Security.Issuer != "https://issuer.example.com" {
		t.Errorf("Security.Issuer = %q, want %q", cfg.Security.Issuer, "https://issuer.example.com")
	}
	if cfg.Security.HIBPEnabled {
		t.Error("Security.HIBPEnabled should be false after env override")
	}
	if cfg.Logging.Level != "warn" {
		t.Errorf("Logging.Level = %q, want %q", cfg.Logging.Level, "warn")
	}
	if cfg.Metrics.Enabled {
		t.Error("Metrics.Enabled should be false after env override")
	}
}

func TestLoadConfig_EnvOverrides_InvalidInt(t *testing.T) {
	t.Setenv("AF_SERVER_PORT", "not-a-number")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	// Should retain the default when env var is not a valid int
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want %d (default when env is invalid)", cfg.Server.Port, 8080)
	}
}

func TestLoadConfig_EnvBool_Variants(t *testing.T) {
	tests := []struct {
		envValue string
		want     bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"1", true},
		{"false", false},
		{"0", false},
		{"no", false},
		{"anything", false},
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			t.Setenv("AF_HIBP_ENABLED", tt.envValue)
			cfg, err := Load("")
			if err != nil {
				t.Fatalf("Load returned error: %v", err)
			}
			if cfg.Security.HIBPEnabled != tt.want {
				t.Errorf("HIBPEnabled with env %q = %v, want %v", tt.envValue, cfg.Security.HIBPEnabled, tt.want)
			}
		})
	}
}

func TestDatabaseDSN(t *testing.T) {
	db := DatabaseConfig{
		Host:     "db.example.com",
		Port:     5432,
		User:     "myuser",
		Password: "mypass",
		Database: "mydb",
		SSLMode:  "require",
	}

	got := db.DSN()
	want := "postgres://myuser:mypass@db.example.com:5432/mydb?sslmode=require"

	if got != want {
		t.Errorf("DSN() = %q, want %q", got, want)
	}
}

func TestDatabaseDSN_DefaultValues(t *testing.T) {
	cfg := DefaultConfig()
	dsn := cfg.Database.DSN()

	want := "postgres://authforge:authforge_secret@localhost:5432/authforge?sslmode=disable"
	if dsn != want {
		t.Errorf("DSN() = %q, want %q", dsn, want)
	}
}

func TestRedisAddr(t *testing.T) {
	r := RedisConfig{
		Host: "redis.example.com",
		Port: 6380,
	}

	got := r.Addr()
	want := "redis.example.com:6380"

	if got != want {
		t.Errorf("Addr() = %q, want %q", got, want)
	}
}

func TestRedisAddr_DefaultValues(t *testing.T) {
	cfg := DefaultConfig()
	addr := cfg.Redis.Addr()

	want := "localhost:6379"
	if addr != want {
		t.Errorf("Addr() = %q, want %q", addr, want)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "bad.yaml")

	if err := os.WriteFile(configFile, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err := Load(configFile)
	if err == nil {
		t.Error("Load should return error for invalid YAML")
	}
}
