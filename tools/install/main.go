package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ─── Config ──────────────────────────────────────────────────────────────────

type InstallConfig struct {
	// Database
	DBHost     string `json:"db_host"`
	DBPort     string `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`

	// Redis
	RedisHost string `json:"redis_host"`
	RedisPort string `json:"redis_port"`

	// NATS
	NATSUrl string `json:"nats_url"`

	// Core API
	CoreHost string `json:"core_host"`
	CorePort string `json:"core_port"`
	BaseURL  string `json:"base_url"`

	// Admin
	AdminEmail    string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`
	AdminName     string `json:"admin_name"`

	// Tenant
	TenantName string `json:"tenant_name"`
	TenantSlug string `json:"tenant_slug"`

	// Application
	AppName         string `json:"app_name"`
	AppType         string `json:"app_type"`
	AppRedirectURIs string `json:"app_redirect_uris"`

	// Security
	EncryptionKey string `json:"encryption_key"`
	JWTAlgorithm  string `json:"jwt_algorithm"`

	// Ports
	LoginUIPort   string `json:"login_ui_port"`
	AdminUIPort   string `json:"admin_ui_port"`
	AccountUIPort string `json:"account_ui_port"`

	// SMTP
	SMTPHost string `json:"smtp_host"`
	SMTPPort string `json:"smtp_port"`

	// Use Docker
	UseDocker bool `json:"use_docker"`
}

func DefaultConfig() InstallConfig {
	return InstallConfig{
		DBHost:        "localhost",
		DBPort:        "5052",
		DBUser:        "cpi-auth",
		DBPassword:    generatePassword(24),
		DBName:        "cpi-auth",
		RedisHost:     "localhost",
		RedisPort:     "5056",
		NATSUrl:       "nats://localhost:5057",
		CoreHost:      "0.0.0.0",
		CorePort:      "5050",
		BaseURL:       "http://localhost:5050",
		AdminEmail:    "admin@cpi-auth.local",
		AdminPassword: "Admin123!",
		AdminName:     "System Administrator",
		TenantName:    "Default Tenant",
		TenantSlug:    "default",
		AppName:       "Default App",
		AppType:       "web",
		AppRedirectURIs: "http://localhost:3000/callback",
		EncryptionKey: generateHexKey(32),
		JWTAlgorithm:  "RS256",
		LoginUIPort:   "5053",
		AdminUIPort:   "5054",
		AccountUIPort: "5055",
		SMTPHost:      "localhost",
		SMTPPort:      "5060",
		UseDocker:     true,
	}
}

// ─── Main ────────────────────────────────────────────────────────────────────

func main() {
	fmt.Println()
	printBanner()
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	cfg := DefaultConfig()

	// Step 1: Welcome & Prerequisites
	step1Prerequisites()

	// Step 2: Deployment Mode
	step2DeploymentMode(scanner, &cfg)

	// Step 3: Database Configuration
	step3Database(scanner, &cfg)

	// Step 4: Security Configuration
	step4Security(scanner, &cfg)

	// Step 5: Admin Account
	step5AdminAccount(scanner, &cfg)

	// Step 6: Tenant Configuration
	step6Tenant(scanner, &cfg)

	// Step 7: Application Configuration
	step7Application(scanner, &cfg)

	// Step 8: UI Ports
	step8UIPorts(scanner, &cfg)

	// Step 9: Review & Confirm
	if !step9Review(scanner, &cfg) {
		fmt.Println("\nInstallation cancelled.")
		os.Exit(0)
	}

	// Step 10: Execute Installation
	step10Execute(&cfg)

	// Step 11: Verify
	step11Verify(&cfg)

	// Step 12: Done
	step12Complete(&cfg)
}

// ─── Steps ───────────────────────────────────────────────────────────────────

func step1Prerequisites() {
	printStep(1, 12, "Checking Prerequisites")

	checks := []struct {
		name    string
		command string
		args    []string
	}{
		{"Docker", "docker", []string{"--version"}},
		{"Docker Compose", "docker", []string{"compose", "version"}},
		{"Go", "go", []string{"version"}},
	}

	allOk := true
	for _, check := range checks {
		cmd := exec.Command(check.command, check.args...)
		out, err := cmd.Output()
		if err != nil {
			printStatus(check.name, "NOT FOUND", "red")
			allOk = false
		} else {
			version := strings.TrimSpace(string(out))
			if len(version) > 60 {
				version = version[:60] + "..."
			}
			printStatus(check.name, version, "green")
		}
	}

	if !allOk {
		fmt.Println("\n  WARNING: Some prerequisites are missing.")
		fmt.Println("  Docker is required for the default installation.")
	}
	fmt.Println()
}

func step2DeploymentMode(scanner *bufio.Scanner, cfg *InstallConfig) {
	printStep(2, 12, "Deployment Mode")
	fmt.Println("  How would you like to deploy CPI Auth?")
	fmt.Println()
	fmt.Println("  [1] Docker Compose (recommended) - All services in containers")
	fmt.Println("  [2] Manual - Configure external PostgreSQL, Redis, NATS")
	fmt.Println()
	choice := prompt(scanner, "  Choice", "1")

	if choice == "2" {
		cfg.UseDocker = false
	} else {
		cfg.UseDocker = true
	}
	fmt.Println()
}

func step3Database(scanner *bufio.Scanner, cfg *InstallConfig) {
	printStep(3, 12, "Database Configuration")

	if cfg.UseDocker {
		fmt.Println("  PostgreSQL will run in Docker (port 5052).")
		fmt.Println("  Redis will run in Docker (port 5056).")
		fmt.Println("  NATS will run in Docker (port 5057).")
		fmt.Println()
		cfg.DBPassword = prompt(scanner, "  Database password", cfg.DBPassword)
	} else {
		cfg.DBHost = prompt(scanner, "  PostgreSQL host", cfg.DBHost)
		cfg.DBPort = prompt(scanner, "  PostgreSQL port", cfg.DBPort)
		cfg.DBUser = prompt(scanner, "  Database user", cfg.DBUser)
		cfg.DBPassword = prompt(scanner, "  Database password", cfg.DBPassword)
		cfg.DBName = prompt(scanner, "  Database name", cfg.DBName)
		cfg.RedisHost = prompt(scanner, "  Redis host", cfg.RedisHost)
		cfg.RedisPort = prompt(scanner, "  Redis port", cfg.RedisPort)
		cfg.NATSUrl = prompt(scanner, "  NATS URL", cfg.NATSUrl)
	}
	fmt.Println()
}

func step4Security(scanner *bufio.Scanner, cfg *InstallConfig) {
	printStep(4, 12, "Security Configuration")

	fmt.Printf("  Encryption Key (auto-generated): %s...%s\n", cfg.EncryptionKey[:8], cfg.EncryptionKey[len(cfg.EncryptionKey)-4:])
	changeKey := prompt(scanner, "  Change encryption key? (y/N)", "n")
	if strings.ToLower(changeKey) == "y" {
		cfg.EncryptionKey = prompt(scanner, "  Encryption key (64 hex chars)", cfg.EncryptionKey)
	}

	cfg.JWTAlgorithm = prompt(scanner, "  JWT signing algorithm (RS256/ES256)", cfg.JWTAlgorithm)
	fmt.Println()
}

func step5AdminAccount(scanner *bufio.Scanner, cfg *InstallConfig) {
	printStep(5, 12, "Admin Account")

	cfg.AdminEmail = prompt(scanner, "  Admin email", cfg.AdminEmail)
	cfg.AdminPassword = prompt(scanner, "  Admin password", cfg.AdminPassword)
	cfg.AdminName = prompt(scanner, "  Admin display name", cfg.AdminName)

	// Validate password
	if len(cfg.AdminPassword) < 8 {
		fmt.Println("  WARNING: Password must be at least 8 characters!")
	}
	fmt.Println()
}

func step6Tenant(scanner *bufio.Scanner, cfg *InstallConfig) {
	printStep(6, 12, "Tenant Configuration")

	cfg.TenantName = prompt(scanner, "  Tenant name", cfg.TenantName)
	cfg.TenantSlug = prompt(scanner, "  Tenant slug (URL-safe)", cfg.TenantSlug)
	fmt.Println()
}

func step7Application(scanner *bufio.Scanner, cfg *InstallConfig) {
	printStep(7, 12, "OAuth Application")

	cfg.AppName = prompt(scanner, "  Application name", cfg.AppName)
	fmt.Println("  Application types: spa, web, native, m2m")
	cfg.AppType = prompt(scanner, "  Application type", cfg.AppType)
	cfg.AppRedirectURIs = prompt(scanner, "  Redirect URIs (comma-separated)", cfg.AppRedirectURIs)
	fmt.Println()
}

func step8UIPorts(scanner *bufio.Scanner, cfg *InstallConfig) {
	printStep(8, 12, "Service Ports")

	fmt.Println("  Configure ports for all CPI Auth services:")
	fmt.Println()
	cfg.CorePort = prompt(scanner, "  Core API port", cfg.CorePort)
	cfg.LoginUIPort = prompt(scanner, "  Login UI port", cfg.LoginUIPort)
	cfg.AdminUIPort = prompt(scanner, "  Admin Console port", cfg.AdminUIPort)
	cfg.AccountUIPort = prompt(scanner, "  Account Portal port", cfg.AccountUIPort)

	// Check port availability
	fmt.Println()
	ports := map[string]string{
		"Core API":       cfg.CorePort,
		"Login UI":       cfg.LoginUIPort,
		"Admin Console":  cfg.AdminUIPort,
		"Account Portal": cfg.AccountUIPort,
	}
	for name, port := range ports {
		if isPortAvailable(port) {
			printStatus(name+" (port "+port+")", "available", "green")
		} else {
			printStatus(name+" (port "+port+")", "IN USE", "red")
		}
	}
	fmt.Println()
}

func step9Review(scanner *bufio.Scanner, cfg *InstallConfig) bool {
	printStep(9, 12, "Review Configuration")

	fmt.Println("  ┌─────────────────────────────────────────────────────────┐")
	fmt.Println("  │  CPI Auth Installation Summary                        │")
	fmt.Println("  ├─────────────────────────────────────────────────────────┤")
	fmt.Printf("  │  Deployment:   %-40s │\n", modeStr(cfg.UseDocker))
	fmt.Printf("  │  Database:     %-40s │\n", cfg.DBUser+"@"+cfg.DBHost+":"+cfg.DBPort+"/"+cfg.DBName)
	fmt.Printf("  │  JWT:          %-40s │\n", cfg.JWTAlgorithm)
	fmt.Printf("  │  Admin:        %-40s │\n", cfg.AdminEmail)
	fmt.Printf("  │  Tenant:       %-40s │\n", cfg.TenantName+" ("+cfg.TenantSlug+")")
	fmt.Printf("  │  App:          %-40s │\n", cfg.AppName+" ("+cfg.AppType+")")
	fmt.Println("  ├─────────────────────────────────────────────────────────┤")
	fmt.Printf("  │  Core API:     http://localhost:%-24s │\n", cfg.CorePort)
	fmt.Printf("  │  Login UI:     http://localhost:%-24s │\n", cfg.LoginUIPort)
	fmt.Printf("  │  Admin UI:     http://localhost:%-24s │\n", cfg.AdminUIPort)
	fmt.Printf("  │  Account UI:   http://localhost:%-24s │\n", cfg.AccountUIPort)
	fmt.Println("  └─────────────────────────────────────────────────────────┘")
	fmt.Println()

	confirm := prompt(scanner, "  Proceed with installation? (Y/n)", "y")
	return strings.ToLower(confirm) != "n"
}

func step10Execute(cfg *InstallConfig) {
	printStep(10, 12, "Installing CPI Auth")

	// 10a: Generate .env file
	fmt.Println("  [1/5] Generating environment configuration...")
	if err := generateEnvFile(cfg); err != nil {
		printError("Failed to generate .env file: " + err.Error())
		os.Exit(1)
	}
	printStatus("  .env file", "created", "green")

	// 10b: Generate config.yaml
	fmt.Println("  [2/5] Generating server configuration...")
	if err := generateConfigYAML(cfg); err != nil {
		printError("Failed to generate config.yaml: " + err.Error())
		os.Exit(1)
	}
	printStatus("  config.yaml", "created", "green")

	// 10c: Update docker-compose with custom ports
	fmt.Println("  [3/5] Updating Docker Compose configuration...")
	if err := generateDockerComposeOverride(cfg); err != nil {
		printError("Failed to generate docker-compose.override.yml: " + err.Error())
		os.Exit(1)
	}
	printStatus("  docker-compose.override.yml", "created", "green")

	if cfg.UseDocker {
		// 10d: Start infrastructure services
		fmt.Println("  [4/5] Starting infrastructure services...")
		if err := runCommand("docker", "compose", "up", "-d", "postgres", "redis", "nats", "mailhog"); err != nil {
			printError("Failed to start infrastructure: " + err.Error())
			os.Exit(1)
		}
		printStatus("  Infrastructure", "started", "green")

		// Wait for services
		fmt.Println("  Waiting for services to become healthy...")
		if err := waitForService("localhost", cfg.DBPort, 30*time.Second); err != nil {
			printError("PostgreSQL not ready: " + err.Error())
			os.Exit(1)
		}
		printStatus("  PostgreSQL", "healthy", "green")

		if err := waitForService("localhost", cfg.RedisPort, 15*time.Second); err != nil {
			printError("Redis not ready: " + err.Error())
			os.Exit(1)
		}
		printStatus("  Redis", "healthy", "green")
		// Small delay for NATS
		time.Sleep(2 * time.Second)
		printStatus("  NATS", "healthy", "green")

		// 10e: Start application services
		fmt.Println("  [5/5] Starting CPI Auth services...")
		if err := runCommand("docker", "compose", "up", "-d", "--build"); err != nil {
			printError("Failed to start CPI Auth: " + err.Error())
			os.Exit(1)
		}
		printStatus("  CPI Auth Core", "starting", "green")
		printStatus("  Login UI", "starting", "green")
		printStatus("  Admin Console", "starting", "green")
		printStatus("  Account Portal", "starting", "green")
	} else {
		fmt.Println("  [4/5] Skipping Docker setup (manual mode)...")
		fmt.Println("  [5/5] Please start CPI Auth manually with: go run main.go")
	}
	fmt.Println()
}

func step11Verify(cfg *InstallConfig) {
	printStep(11, 12, "Verifying Installation")

	if !cfg.UseDocker {
		fmt.Println("  Skipping verification in manual mode.")
		fmt.Println("  Start the server and run: go test ./tests/e2e/ -v")
		return
	}

	// Wait for core API to be ready
	fmt.Println("  Waiting for Core API to become ready...")
	coreURL := "http://localhost:" + cfg.CorePort

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	ready := false
	for !ready {
		select {
		case <-ctx.Done():
			printError("Core API did not become ready within 2 minutes")
			fmt.Println("  Check logs: docker compose logs core")
			return
		default:
			resp, err := http.Get(coreURL + "/health")
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				ready = true
			} else {
				if resp != nil {
					resp.Body.Close()
				}
				time.Sleep(3 * time.Second)
				fmt.Print(".")
			}
		}
	}
	fmt.Println()
	printStatus("  Core API", "healthy", "green")

	// Verify OIDC discovery
	resp, err := http.Get(coreURL + "/.well-known/openid-configuration")
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		printStatus("  OIDC Discovery", "available", "green")
	} else {
		printStatus("  OIDC Discovery", "ERROR", "red")
	}

	// Verify JWKS
	resp, err = http.Get(coreURL + "/.well-known/jwks.json")
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		printStatus("  JWKS Endpoint", "available", "green")
	} else {
		printStatus("  JWKS Endpoint", "ERROR", "red")
	}

	// Verify Prometheus metrics
	resp, err = http.Get(coreURL + "/metrics")
	if err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		printStatus("  Prometheus Metrics", "available", "green")
	} else {
		printStatus("  Prometheus Metrics", "unavailable", "yellow")
	}

	// Check UI services
	checkUIService("Login UI", "http://localhost:"+cfg.LoginUIPort)
	checkUIService("Admin Console", "http://localhost:"+cfg.AdminUIPort)
	checkUIService("Account Portal", "http://localhost:"+cfg.AccountUIPort)

	fmt.Println()
}

func step12Complete(cfg *InstallConfig) {
	printStep(12, 12, "Installation Complete!")

	fmt.Println("  ╔═════════════════════════════════════════════════════════════╗")
	fmt.Println("  ║  CPI Auth has been successfully installed!                 ║")
	fmt.Println("  ╠═════════════════════════════════════════════════════════════╣")
	fmt.Println("  ║                                                             ║")
	fmt.Printf("  ║  Admin Login:     %-40s ║\n", cfg.AdminEmail)
	fmt.Printf("  ║  Admin Password:  %-40s ║\n", cfg.AdminPassword)
	fmt.Println("  ║                                                             ║")
	fmt.Println("  ╠═════════════════════════════════════════════════════════════╣")
	fmt.Println("  ║  Service URLs:                                              ║")
	fmt.Printf("  ║  Core API:        http://localhost:%-24s ║\n", cfg.CorePort)
	fmt.Printf("  ║  Login UI:        http://localhost:%-24s ║\n", cfg.LoginUIPort)
	fmt.Printf("  ║  Admin Console:   http://localhost:%-24s ║\n", cfg.AdminUIPort)
	fmt.Printf("  ║  Account Portal:  http://localhost:%-24s ║\n", cfg.AccountUIPort)
	fmt.Printf("  ║  MailHog:         http://localhost:%-24s ║\n", "5059")
	fmt.Println("  ║                                                             ║")
	fmt.Println("  ╠═════════════════════════════════════════════════════════════╣")
	fmt.Println("  ║  Quick Commands:                                            ║")
	fmt.Println("  ║  docker compose logs -f core    # Watch API logs            ║")
	fmt.Println("  ║  docker compose down            # Stop all services         ║")
	fmt.Println("  ║  docker compose up -d           # Start all services        ║")
	fmt.Println("  ║  go test ./tests/e2e/ -v        # Run E2E tests             ║")
	fmt.Println("  ║                                                             ║")
	fmt.Println("  ╚═════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// ─── File Generation ─────────────────────────────────────────────────────────

func generateEnvFile(cfg *InstallConfig) error {
	content := fmt.Sprintf(`# CPI Auth Environment Configuration
# Generated by cpi-auth-install at %s

# Database
POSTGRES_DB=%s
POSTGRES_USER=%s
POSTGRES_PASSWORD=%s

# Core API
AF_SERVER_HOST=%s
AF_SERVER_PORT=%s
AF_SERVER_BASE_URL=%s
AF_DATABASE_HOST=postgres
AF_DATABASE_PORT=5432
AF_DATABASE_USER=%s
AF_DATABASE_PASSWORD=%s
AF_DATABASE_NAME=%s
AF_DATABASE_SSLMODE=disable
AF_REDIS_HOST=redis
AF_REDIS_PORT=6379
AF_NATS_URL=nats://nats:4222
AF_SMTP_HOST=mailhog
AF_SMTP_PORT=1025
AF_SMTP_FROM=noreply@cpi-auth.local
AF_SECURITY_ENCRYPTION_KEY=%s
AF_SECURITY_JWT_SIGNING_ALGORITHM=%s
AF_SECURITY_RATE_LIMIT_ENABLED=true
AF_SECURITY_RATE_LIMIT_REQUESTS_PER_SEC=100
AF_LOGGING_LEVEL=info
AF_METRICS_ENABLED=true
AF_WEBAUTHN_RP_DISPLAY_NAME=CPI Auth
AF_WEBAUTHN_RP_ID=localhost

# Login UI
PUBLIC_API_URL=http://core:%s
LOGIN_UI_PORT=%s

# Admin UI
VITE_API_URL=http://localhost:%s
ADMIN_UI_PORT=%s

# Account Portal
ACCOUNT_UI_PORT=%s
PUBLIC_LOGIN_URL=http://localhost:%s/login

# Port Summary:
# %s - Core API
# 5052 - PostgreSQL
# %s - Login UI
# %s - Admin Console
# %s - Account Portal
# 5056 - Redis
# 5057 - NATS
# 5059 - MailHog Web UI
`,
		time.Now().Format(time.RFC3339),
		cfg.DBName, cfg.DBUser, cfg.DBPassword,
		cfg.CoreHost, cfg.CorePort, cfg.BaseURL,
		cfg.DBUser, cfg.DBPassword, cfg.DBName,
		cfg.EncryptionKey, cfg.JWTAlgorithm,
		cfg.CorePort, cfg.LoginUIPort,
		cfg.CorePort, cfg.AdminUIPort,
		cfg.AccountUIPort, cfg.LoginUIPort,
		cfg.CorePort, cfg.LoginUIPort, cfg.AdminUIPort, cfg.AccountUIPort,
	)
	return os.WriteFile(".env.local", []byte(content), 0600)
}

func generateConfigYAML(cfg *InstallConfig) error {
	content := fmt.Sprintf(`# CPI Auth Configuration
# Generated by cpi-auth-install

server:
  host: "%s"
  port: %s
  base_url: "%s"
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 15s

database:
  host: postgres
  port: 5432
  user: %s
  password: "%s"
  name: %s
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 10

redis:
  host: redis
  port: 6379
  password: ""
  db: 0

nats:
  url: "nats://nats:4222"

smtp:
  host: mailhog
  port: 1025
  from: "noreply@cpi-auth.local"
  username: ""
  password: ""

security:
  jwt_signing_algorithm: %s
  jwt_private_key_path: ""
  access_token_ttl: 3600
  refresh_token_ttl: 2592000
  encryption_key: "%s"
  rate_limit_enabled: true
  rate_limit_requests_per_sec: 100
  cors_allowed_origins:
    - "http://localhost:%s"
    - "http://localhost:%s"
    - "http://localhost:%s"
    - "http://localhost:%s"
  csp_header: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' http://localhost:*"
  issuer: "%s"

logging:
  level: info
  format: json

metrics:
  enabled: true
  path: /metrics

webauthn:
  rp_display_name: "CPI Auth"
  rp_id: "localhost"
  rp_origins:
    - "http://localhost:%s"
    - "http://localhost:%s"
`,
		cfg.CoreHost, cfg.CorePort, cfg.BaseURL,
		cfg.DBUser, cfg.DBPassword, cfg.DBName,
		cfg.JWTAlgorithm, cfg.EncryptionKey,
		cfg.CorePort, cfg.LoginUIPort, cfg.AdminUIPort, cfg.AccountUIPort,
		cfg.BaseURL,
		cfg.LoginUIPort, cfg.AccountUIPort,
	)
	return os.WriteFile("config.yaml", []byte(content), 0644)
}

func generateDockerComposeOverride(cfg *InstallConfig) error {
	content := fmt.Sprintf(`# CPI Auth Docker Compose Override
# Generated by cpi-auth-install
version: "3.9"

services:
  postgres:
    environment:
      POSTGRES_PASSWORD: "%s"

  core:
    ports:
      - "%s:5050"
    environment:
      AF_SERVER_PORT: "5050"
      AF_DATABASE_PASSWORD: "%s"
      AF_SECURITY_ENCRYPTION_KEY: "%s"
      AF_SECURITY_JWT_SIGNING_ALGORITHM: "%s"

  login-ui:
    ports:
      - "%s:5053"
    environment:
      PUBLIC_API_URL: "http://core:5050"
      ORIGIN: "http://localhost:%s"

  admin-ui:
    ports:
      - "%s:5054"
    build:
      args:
        VITE_API_URL: "http://localhost:%s"

  account-ui:
    ports:
      - "%s:5055"
    environment:
      PUBLIC_API_URL: "http://core:5050"
      PUBLIC_LOGIN_URL: "http://localhost:%s/login"
`,
		cfg.DBPassword,
		cfg.CorePort, cfg.DBPassword, cfg.EncryptionKey, cfg.JWTAlgorithm,
		cfg.LoginUIPort, cfg.LoginUIPort,
		cfg.AdminUIPort, cfg.CorePort,
		cfg.AccountUIPort, cfg.LoginUIPort,
	)
	return os.WriteFile("docker-compose.override.yml", []byte(content), 0644)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func prompt(scanner *bufio.Scanner, label, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("  %s [%s]: ", label, defaultValue)
	} else {
		fmt.Printf("  %s: ", label)
	}
	scanner.Scan()
	value := strings.TrimSpace(scanner.Text())
	if value == "" {
		return defaultValue
	}
	return value
}

func printBanner() {
	fmt.Println("  ╔═══════════════════════════════════════════════════════╗")
	fmt.Println("  ║                                                       ║")
	fmt.Println("  ║     █████╗ ██╗   ██╗████████╗██╗  ██╗                 ║")
	fmt.Println("  ║    ██╔══██╗██║   ██║╚══██╔══╝██║  ██║                 ║")
	fmt.Println("  ║    ███████║██║   ██║   ██║   ███████║                 ║")
	fmt.Println("  ║    ██╔══██║██║   ██║   ██║   ██╔══██║                 ║")
	fmt.Println("  ║    ██║  ██║╚██████╔╝   ██║   ██║  ██║                 ║")
	fmt.Println("  ║    ╚═╝  ╚═╝ ╚═════╝    ╚═╝   ╚═╝  ╚═╝                 ║")
	fmt.Println("  ║    ███████╗ ██████╗ ██████╗  ██████╗ ███████╗         ║")
	fmt.Println("  ║    ██╔════╝██╔═══██╗██╔══██╗██╔════╝ ██╔════╝         ║")
	fmt.Println("  ║    █████╗  ██║   ██║██████╔╝██║  ███╗█████╗           ║")
	fmt.Println("  ║    ██╔══╝  ██║   ██║██╔══██╗██║   ██║██╔══╝           ║")
	fmt.Println("  ║    ██║     ╚██████╔╝██║  ██║╚██████╔╝███████╗         ║")
	fmt.Println("  ║    ╚═╝      ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚══════╝         ║")
	fmt.Println("  ║                                                       ║")
	fmt.Println("  ║    Identity & Access Management Platform              ║")
	fmt.Println("  ║    Install Wizard v1.0                                ║")
	fmt.Println("  ║                                                       ║")
	fmt.Println("  ╚═══════════════════════════════════════════════════════╝")
}

func printStep(step, total int, title string) {
	fmt.Printf("  ━━━ Step %d/%d: %s ━━━\n\n", step, total, title)
}

func printStatus(name, status, color string) {
	icon := "✓"
	if color == "red" {
		icon = "✗"
	} else if color == "yellow" {
		icon = "!"
	}
	fmt.Printf("  %s %s: %s\n", icon, name, status)
}

func printError(msg string) {
	fmt.Printf("\n  ERROR: %s\n", msg)
}

func modeStr(useDocker bool) string {
	if useDocker {
		return "Docker Compose"
	}
	return "Manual"
}

func isPortAvailable(port string) bool {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

func waitForService(host, port string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", host+":"+port, 2*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("service at %s:%s not ready after %v", host, port, timeout)
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func checkUIService(name, url string) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err == nil && resp.StatusCode < 500 {
		resp.Body.Close()
		printStatus("  "+name, "accessible", "green")
	} else {
		printStatus("  "+name, "starting...", "yellow")
	}
}

func generatePassword(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = chars[r.Intn(len(chars))]
	}
	return string(b)
}

func generateHexKey(byteLen int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, byteLen)
	r.Read(b)
	return hex.EncodeToString(b)
}

// Used by E2E test to generate PKCE challenge
func GeneratePKCEChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// Used by E2E test to make API calls
func APICall(method, url string, body interface{}, token string) (*http.Response, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	return resp, respBody, err
}
