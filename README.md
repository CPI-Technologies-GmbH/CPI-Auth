# CPI Auth

Open-source, self-hosted Identity & Access Management platform.

[![CI](https://github.com/CPI-Technologies-GmbH/CPI-Auth/actions/workflows/ci.yaml/badge.svg)](https://github.com/CPI-Technologies-GmbH/CPI-Auth/actions/workflows/ci.yaml)

## Features

- **OAuth 2.0 + OpenID Connect** — Full PKCE authorization code flow, JWKS rotation, discovery
- **Multi-Tenant Architecture** — Complete data isolation, custom domains, per-tenant branding
- **Multi-Factor Authentication** — TOTP, SMS, Email, WebAuthn/FIDO2, recovery codes
- **Social Login** — Google, GitHub, Microsoft, Apple, Facebook, Twitter
- **SAML 2.0 Federation** — SP metadata, SSO, assertion consumer service
- **Role-Based Access Control** — Hierarchical roles, dynamic permissions, app-scoped whitelists
- **Organizations (B2B)** — Member management, domain-based auto-join
- **Page Template Engine** — Customizable auth pages with i18n, design tokens, and CLI tooling
- **Webhooks & Actions** — 17 event types, 8 auth pipeline hooks with custom JavaScript
- **Admin Console** — Full-featured React UI (18 pages) for managing everything
- **CLI & SDK** — Build, preview, and deploy design systems locally

## Installation

### Option 1: Docker Compose (Quickstart)

```bash
git clone https://github.com/CPI-Technologies-GmbH/CPI-Auth.git
cd CPI-Auth
docker compose up -d
```

| Service | URL | Purpose |
|---------|-----|---------|
| Backend API | http://localhost:5050 | Go backend (131 endpoints) |
| Login UI | http://localhost:5053 | Authentication pages |
| Admin Console | http://localhost:5054 | Admin management |
| Account Portal | http://localhost:5055 | User self-service |
| MailHog | http://localhost:5059 | Dev email viewer |

**Default admin:** `admin@cpi-auth.local` / `admin123!`

### Option 2: Docker Pull

```bash
# Pull individual images
docker pull ghcr.io/cpi-technologies-gmbh/cpi-auth:latest
docker pull ghcr.io/cpi-technologies-gmbh/cpi-auth-login-ui:latest
docker pull ghcr.io/cpi-technologies-gmbh/cpi-auth-admin-ui:latest
docker pull ghcr.io/cpi-technologies-gmbh/cpi-auth-account-ui:latest

# Run the backend
docker run -p 5050:5050 \
  -e AF_DB_HOST=your-postgres-host \
  -e AF_REDIS_HOST=your-redis-host \
  ghcr.io/cpi-technologies-gmbh/cpi-auth:latest
```

### Option 3: Helm Chart (Kubernetes)

```bash
# Add the Helm repository
helm repo add cpi-auth https://cpi-technologies-gmbh.github.io/CPI-Auth/charts
helm repo update

# Install with default settings (includes PostgreSQL, Redis, NATS)
helm install cpi-auth cpi-auth/cpi-auth \
  --set global.domain=auth.example.com

# Install with external database
helm install cpi-auth cpi-auth/cpi-auth \
  --set global.domain=auth.example.com \
  --set postgresql.enabled=false \
  --set postgresql.external.host=your-rds-host.amazonaws.com \
  --set postgresql.external.existingSecret=db-credentials \
  --set redis.enabled=false \
  --set redis.external.host=your-elasticache.amazonaws.com

# Production with autoscaling and TLS
helm install cpi-auth cpi-auth/cpi-auth \
  --set global.domain=auth.mycompany.com \
  --set global.tls=true \
  --set core.replicas=3 \
  --set core.autoscaling.enabled=true \
  --set core.autoscaling.maxReplicas=10 \
  --set ingress.annotations."cert-manager\.io/cluster-issuer"=letsencrypt-prod
```

#### Helm Values Reference

| Parameter | Default | Description |
|-----------|---------|-------------|
| `global.domain` | `auth.example.com` | Ingress hostname |
| `global.tls` | `true` | Enable TLS via cert-manager |
| `core.replicas` | `2` | Backend pod replicas |
| `core.image.tag` | `latest` | Backend image tag |
| `core.autoscaling.enabled` | `false` | Enable HPA |
| `core.config.security.encryptionKey` | auto-generated | 32-byte encryption key |
| `loginUI.enabled` | `true` | Deploy login UI |
| `adminUI.enabled` | `true` | Deploy admin console |
| `accountUI.enabled` | `true` | Deploy account portal |
| `postgresql.enabled` | `true` | Deploy built-in PostgreSQL |
| `postgresql.external.host` | `""` | External PostgreSQL host |
| `redis.enabled` | `true` | Deploy built-in Redis |
| `nats.enabled` | `true` | Deploy built-in NATS |
| `ingress.enabled` | `true` | Create Ingress resource |
| `ingress.className` | `nginx` | Ingress controller class |

### Option 4: CLI Tool

```bash
# Install globally
npm install -g @cpi-auth/cli

# Or use directly with npx
npx @cpi-auth/cli init
npx @cpi-auth/cli dev
npx @cpi-auth/cli push
```

### Option 5: SDK (TypeScript)

```bash
npm install @cpi-auth/sdk
```

```typescript
import { AuthForge } from '@cpi-auth/sdk'

const auth = new AuthForge({
  server: 'https://auth.example.com',
  credentials: { email: 'admin@example.com', password: '...' }
})

const users = await auth.client.request('GET', '/admin/users')
```

## Architecture

```
┌─────────────┐  ┌─────────────┐  ┌──────────────┐
│  Login UI   │  │  Admin UI   │  │  Account UI  │
│  SvelteKit  │  │  React/Vite │  │  SvelteKit   │
│  :5053      │  │  :5054      │  │  :5055       │
└──────┬──────┘  └──────┬──────┘  └──────┬───────┘
       │                │                │
       └────────────────┼────────────────┘
                        │
                 ┌──────▼──────┐
                 │  CPI Auth   │
                 │  Go Backend │
                 │  :5050      │
                 └──┬───┬───┬──┘
                    │   │   │
              ┌─────┘   │   └─────┐
              ▼         ▼         ▼
        ┌──────────┐ ┌─────┐ ┌──────┐
        │PostgreSQL│ │Redis│ │ NATS │
        │  :5052   │ │:5056│ │:5057 │
        └──────────┘ └─────┘ └──────┘
```

## CLI Commands

```bash
authforge init                    # Scaffold design project
authforge login                   # Authenticate with server
authforge dev                     # Hot-reload preview server
authforge pull                    # Pull templates & strings from server
authforge push [--dry-run]        # Deploy to server
authforge diff                    # Show local vs. server differences
authforge strings list            # List language strings
authforge strings sync            # Find missing translations
authforge tokens build            # Generate CSS custom properties
authforge tokens validate         # WCAG contrast check
authforge validate                # Full validation (templates + strings + tokens)
```

## Development

```bash
# Backend
go run .

# Admin UI
cd admin-ui && npm run dev

# Login UI
cd login-ui && npm run dev

# Run all tests
go test ./...                                    # Go unit + integration
cd admin-ui && npx vitest run                    # Admin UI (155 tests)
cd login-ui && npx vitest run                    # Login UI (130 tests)
cd account-ui && npx vitest run                  # Account UI (122 tests)
cd sdks/typescript && npx vitest run             # SDK (39 tests)
cd tests/browser && npx playwright test          # E2E (212 tests)
```

## Documentation

```bash
cd docs && npm install && npm run dev
```

Opens VitePress documentation at http://localhost:5173 with:
- Getting Started guide
- API Reference (131 endpoints)
- Admin Console guide
- CLI & SDK reference

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go 1.23, chi router, pgx/v5, JWT (RS256) |
| Admin UI | React 19, Vite, TanStack Query |
| Login UI | SvelteKit, i18n (en/de/fr/es) |
| Account UI | SvelteKit |
| Database | PostgreSQL 16 (partitioned audit logs) |
| Cache | Redis 7 (sessions, rate limiting) |
| Events | NATS JetStream (webhooks, async) |
| Tests | 750+ tests (Go, Vitest, Playwright) |

## Registries

| Artifact | Registry | Install |
|----------|----------|---------|
| Docker Images | [ghcr.io](https://ghcr.io/cpi-technologies-gmbh/cpi-auth) | `docker pull ghcr.io/cpi-technologies-gmbh/cpi-auth` |
| Helm Chart | [GitHub Pages](https://cpi-technologies-gmbh.github.io/CPI-Auth/charts) | `helm repo add cpi-auth ...` |
| CLI | [npm](https://www.npmjs.com/package/@cpi-auth/cli) | `npx @cpi-auth/cli` |
| SDK | [npm](https://www.npmjs.com/package/@cpi-auth/sdk) | `npm i @cpi-auth/sdk` |
| Go Module | [pkg.go.dev](https://pkg.go.dev/github.com/CPI-Technologies-GmbH/CPI-Auth) | `go get github.com/CPI-Technologies-GmbH/CPI-Auth` |

## License

MIT
