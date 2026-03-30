# Configuration

CPI Auth is configured through a `config.yaml` file and environment variable overrides. This page documents every available option.

## Configuration File

The default configuration file is `config.yaml` in the project root. CPI Auth reads this file on startup.

### Complete Reference

```yaml
# ─── Server ─────────────────────────────────────────────────
server:
  host: "0.0.0.0"              # Bind address
  port: 5050                    # HTTP port
  read_timeout: 30s             # Maximum duration for reading request
  write_timeout: 30s            # Maximum duration for writing response
  shutdown_timeout: 15s         # Graceful shutdown timeout

# ─── Database (PostgreSQL) ──────────────────────────────────
database:
  host: postgres                # Database hostname
  port: 5432                    # Database port
  user: cpi-auth               # Database user
  password: cpi-auth_secret    # Database password
  name: cpi-auth               # Database name
  sslmode: disable              # SSL mode: disable, require, verify-ca, verify-full
  max_open_conns: 25            # Maximum open connections
  max_idle_conns: 10            # Maximum idle connections

# ─── Redis ──────────────────────────────────────────────────
redis:
  host: redis                   # Redis hostname
  port: 6379                    # Redis port
  password: ""                  # Redis password (empty for no auth)
  db: 0                         # Redis database number

# ─── NATS ───────────────────────────────────────────────────
nats:
  url: "nats://nats:4222"       # NATS server URL

# ─── SMTP ───────────────────────────────────────────────────
smtp:
  host: mailhog                 # SMTP server hostname
  port: 1025                    # SMTP server port
  from: "noreply@cpi-auth.local"  # Default sender address
  username: ""                  # SMTP username (empty for no auth)
  password: ""                  # SMTP password

# ─── Security ──────────────────────────────────────────────
security:
  jwt_signing_algorithm: RS256  # JWT algorithm: RS256, ES256
  jwt_private_key_path: ""      # Path to PEM private key (auto-generated if empty)
  access_token_ttl: 3600        # Access token lifetime in seconds (1 hour)
  refresh_token_ttl: 2592000    # Refresh token lifetime in seconds (30 days)
  encryption_key: "change-me-in-production-32bytes!"  # 32-byte encryption key for secrets
  rate_limit_enabled: true      # Enable Redis-based rate limiting
  rate_limit_requests_per_sec: 100  # Max requests per second per IP
  cors_allowed_origins:         # Allowed CORS origins
    - "http://localhost:5050"
    - "http://localhost:5053"
    - "http://localhost:5054"
    - "http://localhost:5055"
  csp_header: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' http://localhost:*"

# ─── Logging ────────────────────────────────────────────────
logging:
  level: info                   # Log level: debug, info, warn, error
  format: json                  # Log format: json, console

# ─── Metrics ────────────────────────────────────────────────
metrics:
  enabled: true                 # Enable Prometheus metrics endpoint
  path: /metrics                # Metrics endpoint path

# ─── WebAuthn ───────────────────────────────────────────────
webauthn:
  rp_display_name: "CPI Auth"  # Relying Party display name
  rp_id: "localhost"            # Relying Party ID (typically the domain)
  rp_origins:                   # Allowed origins for WebAuthn ceremonies
    - "http://localhost:5053"
    - "http://localhost:5055"
```

## Environment Variable Overrides

Every configuration value can be overridden using environment variables with the `AF_` prefix. Nested keys use underscores. The mapping is:

| Config Path | Environment Variable |
|------------|---------------------|
| `server.host` | `AF_SERVER_HOST` |
| `server.port` | `AF_SERVER_PORT` |
| `database.host` | `AF_DATABASE_HOST` |
| `database.port` | `AF_DATABASE_PORT` |
| `database.user` | `AF_DATABASE_USER` |
| `database.password` | `AF_DATABASE_PASSWORD` |
| `database.name` | `AF_DATABASE_NAME` |
| `database.sslmode` | `AF_DATABASE_SSLMODE` |
| `redis.host` | `AF_REDIS_HOST` |
| `redis.port` | `AF_REDIS_PORT` |
| `redis.password` | `AF_REDIS_PASSWORD` |
| `nats.url` | `AF_NATS_URL` |
| `smtp.host` | `AF_SMTP_HOST` |
| `smtp.port` | `AF_SMTP_PORT` |
| `smtp.from` | `AF_SMTP_FROM` |
| `smtp.username` | `AF_SMTP_USERNAME` |
| `smtp.password` | `AF_SMTP_PASSWORD` |
| `security.jwt_signing_algorithm` | `AF_SECURITY_JWT_SIGNING_ALGORITHM` |
| `security.access_token_ttl` | `AF_SECURITY_ACCESS_TOKEN_TTL` |
| `security.refresh_token_ttl` | `AF_SECURITY_REFRESH_TOKEN_TTL` |
| `security.encryption_key` | `AF_SECURITY_ENCRYPTION_KEY` |
| `security.rate_limit_enabled` | `AF_SECURITY_RATE_LIMIT_ENABLED` |
| `security.rate_limit_requests_per_sec` | `AF_SECURITY_RATE_LIMIT_REQUESTS_PER_SEC` |
| `logging.level` | `AF_LOGGING_LEVEL` |
| `metrics.enabled` | `AF_METRICS_ENABLED` |
| `webauthn.rp_display_name` | `AF_WEBAUTHN_RP_DISPLAY_NAME` |
| `webauthn.rp_id` | `AF_WEBAUTHN_RP_ID` |

Environment variables take precedence over `config.yaml` values. This is the recommended approach for production deployments and container orchestration.

### Example: Docker Compose Override

```yaml
services:
  core:
    environment:
      AF_DATABASE_HOST: my-postgres.internal
      AF_DATABASE_PASSWORD: ${DB_PASSWORD}
      AF_SECURITY_ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      AF_SECURITY_CORS_ALLOWED_ORIGINS: "https://app.example.com,https://admin.example.com"
```

## Production Checklist

Before deploying CPI Auth to production, address every item on this list:

### Security

- [ ] **Change the encryption key** -- Replace `change-me-in-production-32bytes!` with a cryptographically random 32-byte string
  ```bash
  openssl rand -base64 32 | head -c 32
  ```
- [ ] **Set a strong admin password** -- Change the default `admin123!` immediately after first login
- [ ] **Configure CORS origins** -- List only the exact origins that need access (no wildcards)
- [ ] **Enable TLS** -- Terminate TLS at your load balancer or reverse proxy
- [ ] **Set CSP headers** -- Restrict content sources to your known domains
- [ ] **Use a dedicated JWT key** -- Provide a PEM private key via `jwt_private_key_path` instead of relying on auto-generation

### Database

- [ ] **Use SSL mode `verify-full`** -- Never use `disable` in production
  ```yaml
  database:
    sslmode: verify-full
  ```
- [ ] **Set strong database credentials** -- Use a unique, long password
- [ ] **Configure connection pooling** -- Tune `max_open_conns` based on your expected load

### SMTP

- [ ] **Configure a real SMTP server** -- Replace MailHog with your transactional email service (SES, SendGrid, Postmark)
  ```yaml
  smtp:
    host: smtp.sendgrid.net
    port: 587
    from: "noreply@yourdomain.com"
    username: "apikey"
    password: "${SENDGRID_API_KEY}"
  ```

### Monitoring

- [ ] **Enable metrics** -- Scrape the `/metrics` endpoint with Prometheus
- [ ] **Set log level to `info` or `warn`** -- Avoid `debug` in production
- [ ] **Forward logs** -- Ship structured JSON logs to your observability platform

### WebAuthn

- [ ] **Set the correct RP ID** -- Must match your production domain
  ```yaml
  webauthn:
    rp_id: "yourdomain.com"
    rp_origins:
      - "https://login.yourdomain.com"
      - "https://account.yourdomain.com"
  ```

### Infrastructure

- [ ] **Run Redis with persistence** -- Use AOF or RDB snapshots
- [ ] **Secure NATS** -- Enable authentication and TLS for the NATS connection
- [ ] **Set up backups** -- Schedule regular PostgreSQL backups
- [ ] **Configure rate limiting** -- Tune `rate_limit_requests_per_sec` for your traffic patterns

## Next Steps

- [Tenants](./tenants) -- Configure multi-tenant settings
- [Applications](./applications) -- Set up OAuth applications
- [Auth Flows](./auth-flows) -- Understand token lifetimes and flow configuration
