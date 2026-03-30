# Architecture

This page describes the internal architecture of CPI Auth, its tech stack, and the design patterns used throughout the codebase.

## System Overview

```
                         +------------------+
                         |   Load Balancer  |
                         +--------+---------+
                                  |
            +---------------------+---------------------+
            |                     |                     |
   +--------v-------+   +--------v-------+   +---------v------+
   |   Login UI     |   |   Admin UI     |   |  Account UI    |
   |  (SvelteKit)   |   |   (React)      |   |  (SvelteKit)   |
   |   :5053        |   |    :5054       |   |    :5055       |
   +--------+-------+   +--------+-------+   +---------+------+
            |                     |                     |
            +---------------------+---------------------+
                                  |
                         +--------v---------+
                         |  Go Backend API  |
                         |     :5050        |
                         +---+----+----+----+
                             |    |    |
                +------------+    |    +------------+
                |                 |                 |
       +--------v---+    +-------v----+    +-------v----+
       | PostgreSQL  |    |   Redis    |    |   NATS     |
       |   :5052     |    |   :5056    |    |   :5057    |
       +-------------+    +------------+    +------------+
```

## Tech Stack

### Backend (Go)

| Component | Library | Purpose |
|-----------|---------|---------|
| HTTP Router | chi v5 | Request routing and middleware chain |
| Database | pgx v5 | PostgreSQL driver with connection pooling |
| JWT | golang-jwt v5 | Token signing and verification (RS256/ES256) |
| Logging | zap | Structured JSON logging |
| Validation | go-playground/validator | Request input validation |
| WebAuthn | go-webauthn | FIDO2/passkey credential handling |
| SAML | crewjam/saml | SAML 2.0 SP implementation |
| Crypto | stdlib + nacl | Password hashing (bcrypt), encryption (NaCl secretbox) |
| Events | nats.go | JetStream publish/subscribe |
| Cache | go-redis | Session storage and rate limiting |

### Admin UI (React)

| Component | Library | Purpose |
|-----------|---------|---------|
| Build Tool | Vite | Fast development server and bundling |
| Data Fetching | TanStack Query | Server state management with caching |
| Routing | React Router | SPA navigation |
| UI | Tailwind CSS | Utility-first styling |

### Login UI and Account UI (SvelteKit)

| Component | Library | Purpose |
|-----------|---------|---------|
| Framework | SvelteKit | SSR + client-side hydration |
| i18n | Custom | Locale-based string resolution (en/de/fr/es) |
| Styling | Tailwind CSS | Consistent design system |

## Database: PostgreSQL

CPI Auth uses PostgreSQL 16 as its primary data store. Key design choices:

- **UUID primary keys** on all tables for global uniqueness and safe cross-tenant references
- **Partitioned audit logs** -- the `audit_logs` table is partitioned by month for efficient querying and retention management
- **JSONB columns** for flexible data like `metadata`, `app_metadata`, `settings`, and `branding`
- **Array columns** for multi-value fields like `redirect_uris`, `grant_types`, `permissions`
- **Connection pooling** via pgx with configurable `max_open_conns` (default 25) and `max_idle_conns` (default 10)

### Schema Highlights

```
tenants
  |-- users (tenant_id FK)
  |     |-- identities
  |     |-- mfa_enrollments
  |     |-- recovery_codes
  |     |-- webauthn_credentials
  |     |-- sessions
  |     |-- password_history
  |-- applications (tenant_id FK)
  |     |-- application_permissions
  |-- organizations (tenant_id FK)
  |     |-- organization_members
  |-- roles (tenant_id FK)
  |-- permissions (tenant_id FK)
  |-- webhooks (tenant_id FK)
  |-- actions (tenant_id FK)
  |-- email_templates (tenant_id FK)
  |-- page_templates (tenant_id FK)
  |-- language_strings (tenant_id FK)
  |-- custom_field_definitions (tenant_id FK)
  |-- domain_verifications (tenant_id FK)
  |-- audit_logs (tenant_id FK, partitioned by month)
  |-- api_keys (tenant_id FK)
```

## Cache: Redis

Redis serves two primary purposes:

1. **Session storage** -- Active user sessions with TTL-based expiry
2. **Rate limiting** -- Sliding window counters for API rate limiting

Redis is configured with `allkeys-lru` eviction to ensure the cache stays within its memory budget (default 256 MB).

## Event Bus: NATS JetStream

NATS JetStream provides durable messaging for:

- **Webhook delivery** -- Events are published to a stream, and a consumer dispatches HTTP POST requests to registered webhook URLs with configurable retries
- **Async event processing** -- Audit log writes, email sending, and other non-critical-path operations are offloaded to the event bus
- **Action triggers** -- Custom JavaScript actions subscribe to pipeline events

## Repository Pattern

CPI Auth uses the repository pattern to decouple business logic from data access:

```
core/models/repository.go    -- Interface definitions
core/db/                      -- PostgreSQL implementations
core/users/                   -- Business logic (Service layer)
api/admin/                    -- HTTP handlers
```

Every data entity has a corresponding repository interface in `core/models/repository.go`. Implementations live in `core/db/` and are injected into service constructors. This makes testing straightforward -- services can be tested with mock repositories.

Example interface:

```go
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, tenantID, id uuid.UUID) (*User, error)
    GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, tenantID, id uuid.UUID) error
    List(ctx context.Context, tenantID uuid.UUID, params PaginationParams, search string) (*PaginatedResult[User], error)
    Block(ctx context.Context, tenantID, id uuid.UUID) error
    Unblock(ctx context.Context, tenantID, id uuid.UUID) error
}
```

## Middleware Chain

The Go backend uses a layered middleware chain applied to every request:

```
Request
  |-> CORS (allowed origins per config)
  |-> Rate Limiting (Redis sliding window)
  |-> Request ID (UUID per request)
  |-> Structured Logging (zap)
  |-> Tenant Resolution (from header, domain, or token)
  |-> Authentication (JWT validation)
  |-> RBAC Authorization (permission check)
  |-> Handler
```

### Tenant Resolution

Multi-tenancy is resolved via:

1. `X-Tenant-ID` header (explicit)
2. Custom domain lookup (DNS-verified domain -> tenant mapping)
3. JWT `tenant_id` claim (from authenticated token)

### RBAC Enforcement

After authentication, the middleware extracts the user's effective permissions (union of all role permissions) and checks them against the required permission for the endpoint. Admin routes require permissions like `users:read`, `users:write`, `applications:manage`, etc.

## Project Structure

```
cpi-auth/
  main.go                   -- Entry point
  config.yaml               -- Default configuration
  docker-compose.yml        -- Development environment
  Dockerfile                -- Production container build
  api/
    auth/handlers.go        -- OAuth, login, MFA, WebAuthn, SAML endpoints
    admin/handlers.go       -- Admin CRUD routes
    admin/handlers_extra.go -- Admin handler implementations
    user/handlers.go        -- Self-service user endpoints
    middleware/              -- HTTP middleware
  core/
    config/                 -- Configuration loading
    crypto/                 -- Encryption and hashing utilities
    db/                     -- Repository implementations (PostgreSQL)
    models/                 -- Domain models and repository interfaces
    oauth/                  -- OAuth 2.0 service logic
    tokens/                 -- JWT creation and validation
    sessions/               -- Session management
    users/                  -- User service (CRUD, password policy)
    policy/                 -- RBAC service
    flows/                  -- MFA flow orchestration
    federation/             -- WebAuthn and social login
    saml/                   -- SAML 2.0 integration
    actions/                -- Custom action pipeline (JS execution)
    events/                 -- NATS event publishing
    email/                  -- Email sending (SMTP + templates)
    domains/                -- Custom domain verification
  migrations/               -- SQL migration files
  admin-ui/                 -- React admin console
  login-ui/                 -- SvelteKit login pages
  account-ui/               -- SvelteKit account portal
  tests/browser/            -- Playwright E2E tests
  sdks/                     -- Client SDKs (TypeScript)
  tools/                    -- CLI utilities
```

## Next Steps

- [Configuration](./configuration) -- Reference for all configuration options
- [Auth Flows](./auth-flows) -- Detailed authentication flow documentation
