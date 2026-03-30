package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the root configuration structure.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	NATS     NATSConfig     `yaml:"nats"`
	SMTP     SMTPConfig     `yaml:"smtp"`
	Security SecurityConfig `yaml:"security"`
	Logging  LoggingConfig  `yaml:"logging"`
	Metrics  MetricsConfig  `yaml:"metrics"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	BaseURL         string        `yaml:"base_url"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	TLSCert         string        `yaml:"tls_cert"`
	TLSKey          string        `yaml:"tls_key"`
}

// DatabaseConfig holds PostgreSQL settings.
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	Database        string        `yaml:"database"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxConns        int32         `yaml:"max_conns"`
	MinConns        int32         `yaml:"min_conns"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time"`
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Database, d.SSLMode,
	)
}

// RedisConfig holds Redis settings.
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	TLS      bool   `yaml:"tls"`
}

// Addr returns the Redis address.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// NATSConfig holds NATS settings.
type NATSConfig struct {
	URL      string `yaml:"url"`
	Token    string `yaml:"token"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// SMTPConfig holds email delivery settings.
type SMTPConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	FromName string `yaml:"from_name"`
	TLS      bool   `yaml:"tls"`
}

// SecurityConfig holds security-related settings.
type SecurityConfig struct {
	EncryptionKey          string        `yaml:"encryption_key"`
	JWTSigningAlgorithm    string        `yaml:"jwt_signing_algorithm"`
	JWTPrivateKeyPath      string        `yaml:"jwt_private_key_path"`
	JWTPublicKeyPath       string        `yaml:"jwt_public_key_path"`
	AccessTokenLifetime    time.Duration `yaml:"access_token_lifetime"`
	RefreshTokenLifetime   time.Duration `yaml:"refresh_token_lifetime"`
	IDTokenLifetime        time.Duration `yaml:"id_token_lifetime"`
	AuthCodeLifetime       time.Duration `yaml:"auth_code_lifetime"`
	SessionLifetime        time.Duration `yaml:"session_lifetime"`
	InactivityTimeout      time.Duration `yaml:"inactivity_timeout"`
	Argon2Time             uint32        `yaml:"argon2_time"`
	Argon2Memory           uint32        `yaml:"argon2_memory"`
	Argon2Threads          uint8         `yaml:"argon2_threads"`
	Argon2KeyLen           uint32        `yaml:"argon2_key_len"`
	Argon2SaltLen          int           `yaml:"argon2_salt_len"`
	BcryptCost             int           `yaml:"bcrypt_cost"`
	HIBPEnabled            bool          `yaml:"hibp_enabled"`
	RateLimitEnabled       bool          `yaml:"rate_limit_enabled"`
	RateLimitRequestsPerSec int          `yaml:"rate_limit_requests_per_sec"`
	CORSAllowedOrigins     []string      `yaml:"cors_allowed_origins"`
	CSPHeader              string        `yaml:"csp_header"`
	Issuer                 string        `yaml:"issuer"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// MetricsConfig holds observability settings.
type MetricsConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Path       string `yaml:"path"`
	OTLPEndpoint string `yaml:"otlp_endpoint"`
}

// Load reads the configuration from a YAML file and applies environment overrides.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf("reading config file: %w", err)
			}
		} else {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parsing config file: %w", err)
			}
		}
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			BaseURL:         "http://localhost:8080",
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 15 * time.Second,
		},
		Database: DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "authforge",
			Password:        "authforge_secret",
			Database:        "authforge",
			SSLMode:         "disable",
			MaxConns:        25,
			MinConns:        5,
			MaxConnLifetime: 30 * time.Minute,
			MaxConnIdleTime: 5 * time.Minute,
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		NATS: NATSConfig{
			URL: "nats://localhost:4222",
		},
		SMTP: SMTPConfig{
			Host:     "localhost",
			Port:     1025,
			From:     "noreply@cpi-auth.local",
			FromName: "CPI Auth",
		},
		Security: SecurityConfig{
			EncryptionKey:          "",
			JWTSigningAlgorithm:    "RS256",
			AccessTokenLifetime:    15 * time.Minute,
			RefreshTokenLifetime:   7 * 24 * time.Hour,
			IDTokenLifetime:        1 * time.Hour,
			AuthCodeLifetime:       10 * time.Minute,
			SessionLifetime:        24 * time.Hour,
			InactivityTimeout:      1 * time.Hour,
			Argon2Time:             1,
			Argon2Memory:           64 * 1024,
			Argon2Threads:          4,
			Argon2KeyLen:           32,
			Argon2SaltLen:          16,
			BcryptCost:             12,
			HIBPEnabled:            true,
			RateLimitEnabled:       true,
			RateLimitRequestsPerSec: 100,
			CORSAllowedOrigins:     []string{"*"},
			CSPHeader:              "default-src 'self'",
			Issuer:                 "http://localhost:8080",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
	}
}

func applyEnvOverrides(cfg *Config) {
	// Server
	envStr("AF_SERVER_HOST", &cfg.Server.Host)
	envInt("AF_SERVER_PORT", &cfg.Server.Port)
	envStr("AF_SERVER_BASE_URL", &cfg.Server.BaseURL)

	// Database
	envStr("AF_DB_HOST", &cfg.Database.Host)
	envInt("AF_DB_PORT", &cfg.Database.Port)
	envStr("AF_DB_USER", &cfg.Database.User)
	envStr("AF_DB_PASSWORD", &cfg.Database.Password)
	envStr("AF_DB_DATABASE", &cfg.Database.Database)
	envStr("AF_DB_SSL_MODE", &cfg.Database.SSLMode)

	// Redis
	envStr("AF_REDIS_HOST", &cfg.Redis.Host)
	envInt("AF_REDIS_PORT", &cfg.Redis.Port)
	envStr("AF_REDIS_PASSWORD", &cfg.Redis.Password)

	// NATS
	envStr("AF_NATS_URL", &cfg.NATS.URL)

	// SMTP
	envStr("AF_SMTP_HOST", &cfg.SMTP.Host)
	envInt("AF_SMTP_PORT", &cfg.SMTP.Port)
	envStr("AF_SMTP_USER", &cfg.SMTP.User)
	envStr("AF_SMTP_PASSWORD", &cfg.SMTP.Password)
	envStr("AF_SMTP_FROM", &cfg.SMTP.From)

	// Security
	envStr("AF_ENCRYPTION_KEY", &cfg.Security.EncryptionKey)
	envStr("AF_JWT_ALGORITHM", &cfg.Security.JWTSigningAlgorithm)
	envStr("AF_JWT_PRIVATE_KEY_PATH", &cfg.Security.JWTPrivateKeyPath)
	envStr("AF_JWT_PUBLIC_KEY_PATH", &cfg.Security.JWTPublicKeyPath)
	envStr("AF_ISSUER", &cfg.Security.Issuer)
	envBool("AF_HIBP_ENABLED", &cfg.Security.HIBPEnabled)

	// Logging
	envStr("AF_LOG_LEVEL", &cfg.Logging.Level)
	envStr("AF_LOG_FORMAT", &cfg.Logging.Format)

	// Metrics
	envBool("AF_METRICS_ENABLED", &cfg.Metrics.Enabled)
	envStr("AF_OTLP_ENDPOINT", &cfg.Metrics.OTLPEndpoint)
}

func envStr(key string, target *string) {
	if v := os.Getenv(key); v != "" {
		*target = v
	}
}

func envInt(key string, target *int) {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			*target = i
		}
	}
}

func envBool(key string, target *bool) {
	if v := os.Getenv(key); v != "" {
		*target = strings.EqualFold(v, "true") || v == "1"
	}
}
