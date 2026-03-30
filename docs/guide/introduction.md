# Introduction

## What is CPI Auth?

CPI Auth is an open-source Identity and Access Management (IAM) platform built for developers who need full control over their authentication infrastructure. It provides a complete, self-hosted alternative to managed identity providers like Auth0, Okta, and Firebase Auth, while remaining simpler to operate than enterprise solutions like Keycloak.

CPI Auth handles everything from user sign-up and login to fine-grained authorization, multi-factor authentication, and multi-tenant isolation -- all through a clean REST API backed by a modern tech stack.

## Key Features

- **OAuth 2.0 / OpenID Connect** -- Full OIDC-compliant authorization server with PKCE, discovery, and JWKS endpoints
- **SAML Federation** -- SP-initiated SSO with metadata and ACS endpoints for enterprise integrations
- **Multi-Factor Authentication** -- TOTP, SMS, email, and WebAuthn/FIDO2 with recovery codes
- **Multi-Tenant Architecture** -- Complete data isolation per tenant with individual settings, branding, and domains
- **Role-Based Access Control** -- Hierarchical roles, grouped permissions, and application-scoped permission whitelists
- **WebAuthn / Passkeys** -- Passwordless login with biometrics and hardware security keys
- **Social Login** -- Google, GitHub, Microsoft, Apple, Facebook, and Twitter identity providers
- **Custom Domains** -- DNS-verified custom domains per tenant for white-label authentication
- **Page Templates** -- Fully customizable login, signup, and profile pages with template variables and i18n
- **Email Templates** -- MJML-powered, per-locale email templates for verification, password reset, and more
- **Webhooks and Actions** -- Event-driven hooks with custom JavaScript actions at every stage of the auth pipeline
- **Organizations** -- B2B organization model with domain-based auto-join and org-scoped roles
- **Custom User Fields** -- Configurable field definitions for registration and profile forms
- **Audit Logging** -- Immutable, partitioned audit trail with export and PII masking
- **CLI Tooling** -- Command-line tools for managing templates, configuration, and deployment
- **Passwordless Auth** -- Magic link authentication via email and SMS

## Architecture Overview

CPI Auth is composed of four services:

| Service | Technology | Purpose |
|---------|-----------|---------|
| **Backend API** | Go (chi router, pgx, JWT) | Core REST API, OAuth server, business logic |
| **Admin UI** | React + Vite + TanStack Query | Management console for administrators |
| **Login UI** | SvelteKit with i18n (en/de/fr/es) | End-user authentication pages |
| **Account UI** | SvelteKit | Self-service account management portal |

Supporting infrastructure:

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Database | PostgreSQL 16 | Primary data store with partitioned audit logs |
| Cache | Redis 7 | Session storage, rate limiting, temporary codes |
| Event Bus | NATS JetStream | Webhook delivery, async event processing |
| Dev Email | MailHog | Email capture during development |

## How CPI Auth Compares

| Capability | CPI Auth | Auth0 | Keycloak |
|-----------|-----------|-------|----------|
| Self-hosted | Yes | No (SaaS) | Yes |
| Open source | Yes (MIT) | No | Yes (Apache 2.0) |
| Language | Go | N/A | Java |
| Setup time | Minutes (Docker Compose) | Minutes (SaaS) | Hours |
| Multi-tenant | Native | Per-tenant pricing | Realms |
| Custom login UI | SvelteKit templates | Universal Login | FreeMarker themes |
| Admin UI | Modern React SPA | Dashboard | Legacy JSP + React |
| Memory footprint | ~50 MB | N/A | ~500 MB+ |
| SAML + OIDC | Both | Both | Both |
| WebAuthn | Yes | Yes | Yes |
| Custom actions | JavaScript hooks | Node.js actions | Java SPI |
| Audit logs | Partitioned, exportable | Streams | Events |

CPI Auth is designed for teams that want:

- **Full data sovereignty** -- everything runs on your infrastructure
- **Modern developer experience** -- clean APIs, typed SDKs, Docker-first workflow
- **Simplicity** -- a single Go binary plus frontends, not a Java application server
- **Flexibility** -- customize every page, email, and auth flow without vendor lock-in

## Who Should Use CPI Auth?

- **Startups** building a SaaS product who need authentication on day one without paying per-user fees
- **Enterprises** requiring on-premise identity management with full audit trails and compliance controls
- **Agencies** building multi-tenant platforms where each client needs isolated users, branding, and domains
- **Developers** tired of fighting opaque managed services and want to own their auth stack

## License

CPI Auth is released under the MIT License. You are free to use, modify, and distribute it in commercial and non-commercial projects.

## Next Steps

- [Getting Started](./getting-started) -- Run CPI Auth locally in under 5 minutes
- [Architecture](./architecture) -- Understand the system design
- [Configuration](./configuration) -- Tune every setting
