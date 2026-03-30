# CPI Auth — Features, SWOT Analysis & USPs

> Open-source, self-hosted Identity & Access Management platform built with Go, React, and SvelteKit.

---

## Table of Contents

1. [Complete Feature Matrix](#complete-feature-matrix)
2. [Platform Statistics](#platform-statistics)
3. [SWOT Analysis](#swot-analysis)
4. [Unique Selling Propositions (USPs)](#unique-selling-propositions)
5. [Competitive Comparison](#competitive-comparison)
6. [Target Audience](#target-audience)

---

## Complete Feature Matrix

### Authentication & Federation

| Feature | Status | Details |
|---------|--------|---------|
| OAuth 2.0 Authorization Code + PKCE | Implemented | Full RFC 7636 compliance, enforced PKCE |
| OpenID Connect (OIDC) | Implemented | Discovery, JWKS, Userinfo, ID Tokens |
| SAML 2.0 Federation | Implemented | SP metadata, ACS endpoint, SSO initiation |
| Social Login | Implemented | Google, GitHub, Microsoft, Apple, Facebook, Twitter |
| Passwordless Authentication | Implemented | Email magic links, SMS OTP |
| WebAuthn / FIDO2 / Passkeys | Implemented | Biometric + hardware security key support |
| Client Credentials Flow | Implemented | Machine-to-machine (M2M) authentication |
| Refresh Token Rotation | Implemented | Automatic rotation with reuse detection |
| Token Revocation | Implemented | Individual and bulk token revocation |
| JWT Signing (RS256) | Implemented | Automatic key rotation via JWKS |

### Multi-Factor Authentication (MFA)

| Feature | Status | Details |
|---------|--------|---------|
| TOTP (Authenticator Apps) | Implemented | Google Authenticator, Authy, 1Password compatible |
| SMS-based MFA | Implemented | Via configurable SMS gateway |
| Email-based MFA | Implemented | One-time codes via email |
| WebAuthn as second factor | Implemented | Hardware keys (YubiKey, Titan) |
| Recovery Codes | Implemented | One-time backup codes for account recovery |
| Tenant-level MFA enforcement | Implemented | Require MFA for all users in a tenant |
| Per-user MFA enrollment | Implemented | Self-service enrollment via account portal |

### Multi-Tenant Architecture

| Feature | Status | Details |
|---------|--------|---------|
| Complete data isolation | Implemented | Every entity scoped to tenant_id |
| Custom domains per tenant | Implemented | DNS TXT verification workflow |
| Per-tenant branding | Implemented | Logo, colors, fonts, layout mode |
| Per-tenant password policies | Implemented | Min length, complexity, history |
| Per-tenant session settings | Implemented | TTL, concurrent session limits |
| Per-tenant MFA requirements | Implemented | Force MFA enrollment for all users |
| Hierarchical tenants | Implemented | Parent-child tenant relationships |
| Tenant force-logout | Implemented | Revoke all sessions tenant-wide |

### User Management

| Feature | Status | Details |
|---------|--------|---------|
| User CRUD | Implemented | Create, read, update, delete via API + UI |
| User statuses | Implemented | Active, inactive, blocked, deleted |
| Email verification | Implemented | Verification codes and magic links |
| Phone verification | Implemented | SMS verification codes |
| User metadata | Implemented | Arbitrary JSON metadata + app_metadata |
| User locale | Implemented | Per-user language preference |
| Custom fields | Implemented | 9 field types, configurable per registration/profile |
| User search & filtering | Implemented | Full-text search with pagination |
| Bulk operations | Implemented | Import, export, bulk block, bulk delete |
| User impersonation | Implemented | Admin impersonates user for debugging |
| Password reset | Implemented | Admin-initiated and self-service flows |
| Session management | Implemented | View, revoke individual or all sessions |
| Login history | Implemented | Per-user audit trail |

### Role-Based Access Control (RBAC)

| Feature | Status | Details |
|---------|--------|---------|
| Hierarchical roles | Implemented | Roles inherit from parent roles |
| Dynamic permissions | Implemented | DB-stored, tenant-scoped permissions |
| System roles | Implemented | admin, manager, editor, viewer (non-deletable) |
| Custom roles | Implemented | Create unlimited custom roles |
| Permission groups | Implemented | Grouped by functional area |
| Application permission whitelists | Implemented | Token perms = user perms ∩ app whitelist |
| Role assignment API | Implemented | Assign/remove roles via API + UI |
| JWT permission claims | Implemented | Permissions embedded in access tokens |

### Organizations (B2B)

| Feature | Status | Details |
|---------|--------|---------|
| Organization CRUD | Implemented | Create, manage, delete organizations |
| Member management | Implemented | Add/remove members with roles |
| Domain-based auto-join | Implemented | Automatic org membership by email domain |
| Organization-scoped roles | Implemented | Different roles per organization |
| Organization branding | Implemented | Per-org branding overrides |
| Organization settings | Implemented | Custom settings per organization |

### Applications (OAuth Clients)

| Feature | Status | Details |
|---------|--------|---------|
| Application types | Implemented | SPA, Native, Web, M2M |
| Client credentials | Implemented | Client ID + Secret with rotation |
| Redirect URI management | Implemented | Multiple URIs with validation |
| Allowed origins (CORS) | Implemented | Per-app CORS configuration |
| Post-logout redirect URIs | Implemented | Configured logout URLs |
| Grant type configuration | Implemented | Per-app: auth_code, client_credentials, refresh, implicit, password |
| Token lifetime configuration | Implemented | Per-app access, refresh, and ID token TTLs |
| Application-scoped permissions | Implemented | Whitelist model for token permissions |
| Client secret rotation | Implemented | Generate new secret, invalidate old |
| Application status toggle | Implemented | Enable/disable applications |

### Page Template Engine

| Feature | Status | Details |
|---------|--------|---------|
| Default templates | Implemented | 8 professional templates (login, signup, verification, password reset, MFA, error, consent, profile) |
| Custom templates | Implemented | Full HTML/CSS editor per page type |
| Template variables | Implemented | user.name, user.email, application.name, tenant.name, code, link, error, custom_fields, profile_fields |
| Language strings (i18n) | Implemented | Key-value strings with locale support, referenced via {{t.key}} |
| Template duplication | Implemented | Clone default templates to customize |
| Default template protection | Implemented | Readonly defaults, 403 on edit/delete |
| Live preview | Implemented | Rendered preview with sample data |
| Custom fields integration | Implemented | {{custom_fields}} and {{profile_fields}} auto-render |
| Custom page type | Implemented | Arbitrary pages for linking |
| Template search | Implemented | Filter by name and page type |

### Email Templates

| Feature | Status | Details |
|---------|--------|---------|
| Template types | Implemented | Verification, password reset, MFA, welcome, invitation, magic link |
| MJML support | Implemented | Responsive email markup language |
| Per-locale templates | Implemented | Different content per language |
| Template variables | Implemented | User, tenant, application data injection |
| Test email sending | Implemented | Send test email from admin UI |

### Webhooks & Actions

| Feature | Status | Details |
|---------|--------|---------|
| Webhook CRUD | Implemented | URL, events, secret, active toggle |
| Event subscription | Implemented | 17 event types (user.*, login.*, mfa.*, session.*, etc.) |
| Webhook signatures | Implemented | HMAC-SHA256 signature verification |
| Delivery history | Implemented | Success/failure tracking per webhook |
| Test delivery | Implemented | Send test event to webhook URL |
| Retry policy | Implemented | Automatic retries with exponential backoff |
| Custom actions | Implemented | JavaScript code at 8 auth pipeline stages |
| Action ordering | Implemented | Configurable execution order |
| Action timeouts | Implemented | Per-action timeout (default 5s) |
| Pre/post hooks | Implemented | pre-registration, post-registration, pre-login, post-login, pre-token, post-change-password, pre-user-update, post-user-delete |

### Security & Compliance

| Feature | Status | Details |
|---------|--------|---------|
| Immutable audit logs | Implemented | Partitioned by month, exportable |
| PII masking | Implemented | Automatic masking in audit trails |
| Rate limiting | Implemented | Configurable per-endpoint |
| CORS configuration | Implemented | Per-origin allowlisting |
| CSP headers | Implemented | Content Security Policy enforcement |
| Password policy engine | Implemented | Length, complexity, history requirements |
| Brute force protection | Implemented | Login attempt tracking and blocking |
| Encryption at rest | Implemented | Configurable encryption key for secrets |
| Session management | Implemented | Forced logout, concurrent limits |

### Developer Tools

| Feature | Status | Details |
|---------|--------|---------|
| TypeScript SDK | Implemented | Full API client with auth, sync, preview |
| Go SDK | Implemented | Native Go client library |
| Python SDK | Implemented | Python package |
| CLI tool | Implemented | 9 commands: init, login, dev, pull, push, diff, strings, tokens, validate |
| Design tokens | Implemented | YAML-based tokens, CSS custom property generation |
| WCAG contrast validation | Implemented | Automated color accessibility checks |
| Local dev server | Implemented | Hot-reload preview with locale/viewport switching |
| Template sync | Implemented | Bidirectional push/pull with diff |
| API key management | Implemented | Scoped API keys with rate limits |

### Admin Console (React)

| Feature | Status | Details |
|---------|--------|---------|
| Dashboard | Implemented | Metrics, charts, recent events, error rates |
| User management | Implemented | Search, CRUD, detail with 7 tabs |
| Application management | Implemented | 5 tabs: Overview, Settings, Permissions, Connections, API |
| Tenant management | Implemented | Multi-tenant administration |
| Organization management | Implemented | B2B org + member management |
| Role & permission management | Implemented | Visual RBAC configuration |
| Branding editor | Implemented | Color pickers, logo, layout mode |
| Webhook management | Implemented | CRUD + test + delivery history |
| Action editor | Implemented | JavaScript code editor with trigger config |
| Email template editor | Implemented | MJML editor with test sending |
| Page template editor | Implemented | HTML/CSS/Preview with variable toolbar |
| API key management | Implemented | Create, revoke, configure |
| Custom field management | Implemented | 9 field types, visibility config |
| Audit log viewer | Implemented | Searchable, filterable, exportable |
| Settings page | Implemented | Security, SMTP, domain, branding config |
| Language switcher | Implemented | EN, DE, FR, ES |
| Internationalization | Implemented | Full i18n across sidebar, pages, metrics |

### User-Facing UIs

| Feature | Status | Details |
|---------|--------|---------|
| Login page | Implemented | Email/password, social buttons, MFA |
| Registration page | Implemented | With custom fields support |
| Password reset flow | Implemented | Request + reset form |
| Email verification | Implemented | Code entry + magic link |
| MFA enrollment | Implemented | TOTP QR code, backup codes |
| MFA challenge | Implemented | Code entry for TOTP/SMS/Email |
| OAuth consent screen | Implemented | Scope display, approve/deny |
| Error page | Implemented | Graceful error display |
| Account profile | Implemented | Edit name, view email |
| Security settings | Implemented | Password change, MFA management |
| Active sessions | Implemented | View and revoke sessions |
| Linked accounts | Implemented | Social provider connections |
| Login history | Implemented | Activity timeline |
| Privacy settings | Implemented | Data management options |

---

## Platform Statistics

| Metric | Count |
|--------|-------|
| API Endpoints | 131 (20 public auth + 111 admin) |
| Admin UI Pages | 18 |
| Login UI Pages | 11 |
| Account UI Pages | 8 |
| Database Tables | 35+ |
| Data Models | 32 |
| Event Types | 17 |
| Action Pipeline Hooks | 8 |
| Custom Field Types | 9 |
| Page Template Types | 9 |
| MFA Methods | 4 |
| Social Providers | 6 |
| Application Types | 4 |
| SDKs | 3 (Go, Python, TypeScript) |
| CLI Commands | 9 |
| Database Migrations | 13 |
| Test Files | 178 |
| Test Cases | ~750+ |
| Documentation Pages | 43 (11,000+ lines) |

---

## SWOT Analysis

### Strengths

| # | Strength | Impact |
|---|----------|--------|
| S1 | **Complete self-hosted solution** — No vendor lock-in, full data sovereignty. All data stays on your infrastructure. | Differentiator for enterprises with compliance requirements (GDPR, HIPAA, SOC2) |
| S2 | **Modern tech stack** — Go backend (low memory ~50MB), React admin UI, SvelteKit auth pages. Fast, efficient, maintainable. | Lower infrastructure costs vs. Java-based alternatives (Keycloak ~500MB+) |
| S3 | **Multi-tenant from day one** — Complete data isolation, per-tenant branding, domains, and policies. Not bolted on. | B2B SaaS platforms can offer white-label auth to their customers |
| S4 | **Full OAuth 2.0 + OIDC compliance** — PKCE enforced, JWKS rotation, discovery endpoint, userinfo. Standards-first design. | Drop-in replacement for Auth0/Okta in existing OAuth integrations |
| S5 | **Page template engine with CLI** — Build, preview, and deploy custom auth UIs locally with hot-reload, design tokens, and i18n. | Unique developer experience not offered by any major competitor |
| S6 | **131 API endpoints** — Comprehensive coverage of every IAM operation. Nothing requires console-only access. | Full automation and CI/CD integration capability |
| S7 | **Extensible action pipeline** — Custom JavaScript hooks at 8 stages of the auth flow. | Custom business logic without forking: fraud detection, enrichment, consent |
| S8 | **Enterprise security features** — Immutable audit logs with PII masking, SAML federation, WebAuthn, MFA enforcement. | Enterprise procurement checklist coverage |
| S9 | **Comprehensive testing** — 750+ tests across unit, integration, and E2E. Every feature has positive and negative test coverage. | Reliability and confidence in production deployments |
| S10 | **3 purpose-built UIs** — Admin console, login/registration flow, user account portal. Not a single monolithic UI. | Better UX and separation of concerns for different user personas |

### Weaknesses

| # | Weakness | Mitigation |
|---|----------|------------|
| W1 | **Young project** — Less battle-tested than Auth0 (10+ years) or Keycloak (8+ years). Fewer edge cases encountered. | Comprehensive test suite; incremental production rollouts recommended |
| W2 | **Single-node architecture** — No built-in horizontal scaling or clustering. Redis and NATS provide some distribution. | Stateless Go binary allows load balancer placement; DB is the bottleneck |
| W3 | **No managed hosting option** — Self-hosted only. Requires DevOps expertise to deploy and maintain. | Docker Compose for dev; Kubernetes Helm chart roadmap |
| W4 | **Limited social provider catalog** — 6 providers vs. Auth0's 30+. Missing: LinkedIn, Slack, Discord, OIDC generic. | Extensible provider architecture; priority additions based on demand |
| W5 | **No native mobile SDKs** — TypeScript, Go, Python only. No Swift/Kotlin SDKs. | Standard OAuth2 PKCE flow works with any HTTP client; community SDKs welcome |
| W6 | **Documentation coverage gaps** — Some advanced features (SAML config, SCIM) lack step-by-step guides. | Active documentation effort; 43 pages and growing |
| W7 | **No built-in user migration tool** — No automated import from Auth0, Firebase, Cognito, or Keycloak. | CSV/JSON import API available; migration scripts can be built with SDK |

### Opportunities

| # | Opportunity | Strategy |
|---|-------------|----------|
| O1 | **Data sovereignty regulations** — GDPR, Digital Sovereignty Acts driving demand for self-hosted IAM. | Position as the modern, lightweight alternative to Keycloak for EU/regulated markets |
| O2 | **Developer experience trend** — Developers choose tools with great DX (Vercel, Supabase model). | CLI + SDK + hot-reload dev server is a differentiator. Invest further in DX |
| O3 | **B2B SaaS authentication market** — Growing need for multi-tenant, white-label auth in SaaS platforms. | First-class multi-tenant + organization support is already built |
| O4 | **Passkeys / WebAuthn adoption** — Passwordless authentication becoming mainstream (Apple, Google push). | Already implemented. Market as passwordless-ready platform |
| O5 | **Open-source community** — Growing distrust of proprietary auth vendors after breaches (Okta 2023). | Open-source transparency builds trust. Community contributions accelerate development |
| O6 | **AI agent authentication** — M2M auth for AI agents, tool-use, and autonomous systems. | Client credentials flow + API keys already support this. Expand with agent-specific features |
| O7 | **Edge deployment** — Go binary compiles to single file, runs anywhere. Edge/IoT potential. | Single-binary deployment story for edge computing scenarios |

### Threats

| # | Threat | Mitigation |
|---|--------|------------|
| T1 | **Auth0/Okta market dominance** — Established brands with massive sales teams and integrations. | Focus on self-hosted niche, developer community, and open-source trust |
| T2 | **Keycloak momentum** — Large existing community, Red Hat backing, Kubernetes-native. | Better DX, lower resource usage, modern tech stack as differentiators |
| T3 | **Cloud-native auth solutions** — AWS Cognito, Firebase Auth, Supabase Auth — free tier + managed. | Self-hosted value prop: no vendor lock-in, data sovereignty, unlimited users |
| T4 | **Security vulnerability risk** — Auth systems are high-value targets. Single vulnerability can be catastrophic. | Security-first development, dependency auditing, responsible disclosure program |
| T5 | **Maintenance burden** — Self-hosted requires ongoing updates, monitoring, and security patches. | Automated update notifications, Docker image publishing, migration tooling |

---

## Unique Selling Propositions

### USP 1: Template Engine + CLI Developer Workflow

**No other IAM platform offers a local development workflow for auth page customization.**

```
cpi-auth init        # Scaffold project with design tokens
cpi-auth dev         # Hot-reload preview server
cpi-auth validate    # WCAG contrast check + string validation
cpi-auth push        # Deploy to production
```

Developers design auth pages like they design any frontend — locally, with version control, design tokens, and automated validation. Auth0 requires editing in a web browser. Keycloak requires theme JAR files.

### USP 2: True Multi-Tenancy with White-Label Auth

**Built for B2B SaaS platforms that need to offer authentication to their customers.**

- Complete data isolation per tenant
- Custom domains with DNS verification
- Per-tenant branding (logo, colors, fonts, layout)
- Per-tenant policies (password rules, MFA, sessions)
- Tenant-scoped everything: users, apps, roles, templates, strings

Auth0 charges per-tenant. Keycloak's realm model has performance limitations at scale. CPI Auth's tenant model is native to every database query.

### USP 3: Lightweight Yet Complete

**50MB memory footprint with 131 API endpoints, 4 MFA methods, SAML, and WebAuthn.**

| Platform | Language | Memory | API Surface |
|----------|----------|--------|------------|
| CPI Auth | Go | ~50MB | 131 endpoints |
| Keycloak | Java | ~500MB+ | ~80 endpoints |
| Auth0 | Proprietary | Managed | ~100 endpoints |
| Supabase Auth | Go | ~30MB | ~20 endpoints |

CPI Auth delivers enterprise-grade features without enterprise-grade infrastructure costs. A single $5/month VPS can run CPI Auth with thousands of users.

### USP 4: Extensible Action Pipeline

**Custom JavaScript at every stage of the authentication lifecycle.**

8 pipeline hooks allow custom business logic without forking:
- **Pre-registration**: Validate email domains, check blocklists
- **Post-registration**: Sync to CRM, send Slack notification
- **Pre-login**: Geo-fence, device fingerprint check
- **Post-login**: Update analytics, trigger welcome flow
- **Pre-token**: Enrich claims, add custom scopes
- **Post-change-password**: Notify security team, revoke sessions

No other self-hosted solution offers this level of extensibility through code injection.

### USP 5: Three Purpose-Built UIs

**Separation of concerns across admin, user, and account portals.**

| UI | Users | Purpose | Tech |
|----|-------|---------|------|
| Admin Console | IT admins, DevOps | Manage users, apps, policies | React + Vite |
| Login UI | End users | Authenticate, register, MFA | SvelteKit |
| Account Portal | End users | Self-service profile, security | SvelteKit |

Each UI is independently deployable, themeable, and replaceable. Auth0 bundles everything into Universal Login. Keycloak's account console is limited.

### USP 6: Application-Scoped Permission Whitelists

**Fine-grained control over which permissions appear in tokens per application.**

```
Token Permissions = User Effective Permissions ∩ Application Whitelist
```

If an application has no whitelist entries, all user permissions are included (backward-compatible). When a whitelist is configured, only those permissions appear in tokens — even if the user has broader access.

This prevents over-privileged tokens without modifying user roles. Auth0 has this concept but charges extra. Keycloak requires complex client scope mapping.

---

## Competitive Comparison

| Capability | CPI Auth | Auth0 | Keycloak | Firebase Auth | Supabase Auth |
|------------|-----------|-------|----------|---------------|---------------|
| **Self-hosted** | Yes | No | Yes | No | Partial |
| **Open source** | Yes (MIT) | No | Yes (Apache) | No | Yes (Apache) |
| **OAuth 2.0 + OIDC** | Full | Full | Full | Partial | Partial |
| **SAML Federation** | Yes | Yes | Yes | No | No |
| **WebAuthn/Passkeys** | Yes | Yes | Yes | No | No |
| **Multi-Tenant** | Native | Paid add-on | Realms | No | No |
| **Custom Domains** | Yes | Paid | Manual | No | No |
| **MFA (TOTP/SMS/Email/WebAuthn)** | 4 methods | 4 methods | 3 methods | 2 methods | 1 method |
| **Custom Auth Pages** | HTML/CSS + CLI | Limited editor | Theme JARs | Firebase UI | Limited |
| **i18n Page Strings** | Yes + CLI sync | Partial | Properties files | No | No |
| **Design Token System** | Yes (WCAG) | No | No | No | No |
| **Action Pipeline Hooks** | 8 triggers | Yes (paid) | SPI (Java) | Cloud Functions | Edge Functions |
| **Organizations (B2B)** | Yes | Paid add-on | Partial | No | No |
| **Custom User Fields** | 9 types | Yes | Yes | No | No |
| **Audit Logs** | Partitioned | Yes (paid) | Events | No | No |
| **User Impersonation** | Yes | Yes | Yes | No | No |
| **CLI Developer Tool** | Yes | No | No | Firebase CLI | Supabase CLI |
| **Memory Footprint** | ~50MB | Managed | ~500MB+ | Managed | ~30MB |
| **Pricing** | Free | $23/1K MAU | Free | Free to 50K | Free to 50K |

---

## Target Audience

### Primary: B2B SaaS Builders
- Building multi-tenant SaaS products
- Need white-label authentication for their customers
- Want to own their auth infrastructure
- Compliance-conscious (GDPR, SOC2, HIPAA)

### Secondary: Enterprise Development Teams
- Replacing Auth0/Okta to reduce costs at scale
- Data sovereignty requirements (on-prem or private cloud)
- Custom authentication flows (action pipeline)
- Complex RBAC with organization hierarchies

### Tertiary: Developer Tool Builders
- Building platforms with developer-facing auth
- Need M2M (client credentials) alongside user auth
- API key management for their SDKs
- Custom branding per customer (multi-tenant)

---

*CPI Auth — Authentication infrastructure you own, customize, and control.*
