# CPI Auth

Open-source, self-hosted Identity & Access Management platform.

## Features

- **OAuth 2.0 + OpenID Connect** — Full PKCE authorization code flow, JWKS rotation, discovery
- **Multi-Tenant Architecture** — Complete data isolation, custom domains, per-tenant branding
- **Multi-Factor Authentication** — TOTP, SMS, Email, WebAuthn/FIDO2, recovery codes
- **Social Login** — Google, GitHub, Microsoft, Apple, Facebook, Twitter
- **SAML 2.0 Federation** — SP metadata, SSO, assertion consumer service
- **Role-Based Access Control** — Hierarchical roles, dynamic permissions, application-scoped whitelists
- **Organizations (B2B)** — Member management, domain-based auto-join
- **Page Template Engine** — Customizable login/signup pages with i18n, design tokens, and CLI tooling
- **Webhooks & Actions** — 17 event types, 8 auth pipeline hooks with custom JavaScript
- **Admin Console** — Full-featured React UI for managing users, apps, roles, templates, and settings
- **CLI & SDK** — Build, preview, and deploy design systems locally with hot-reload

## Quick Start

```bash
git clone https://github.com/CPI-Technologies-GmbH/CPI-Auth.git
cd CPI-Auth
docker compose up -d
```

| Service | URL | Purpose |
|---------|-----|---------|
| Backend API | http://localhost:5050 | Go backend |
| Login UI | http://localhost:5053 | Authentication pages |
| Admin Console | http://localhost:5054 | Admin management |
| Account Portal | http://localhost:5055 | User self-service |
| MailHog | http://localhost:5059 | Dev email viewer |

**Default admin login:** `admin@cpi-auth.local` / `admin123!`

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

## Development

```bash
# Backend
go run .

# Admin UI
cd admin-ui && npm run dev

# Login UI
cd login-ui && npm run dev

# Account UI
cd account-ui && npm run dev

# Run tests
go test ./...
cd admin-ui && npx vitest run
cd tests/browser && npx playwright test
```

## CLI & SDK

```bash
# Initialize a design project
npx @cpi-auth/cli init

# Start local preview server
npx @cpi-auth/cli dev

# Deploy to your tenant
npx @cpi-auth/cli push
```

## Documentation

See the [docs/](./docs/) directory or run:

```bash
cd docs && npm run dev
```

## Tech Stack

- **Backend:** Go (chi router, pgx/v5, JWT)
- **Admin UI:** React, Vite, TanStack Query
- **Login UI:** SvelteKit, i18n (en/de/fr/es)
- **Account UI:** SvelteKit
- **Database:** PostgreSQL 16
- **Cache:** Redis 7
- **Events:** NATS JetStream
- **Tests:** Go tests, Vitest, Playwright E2E (750+ tests)

## License

MIT
