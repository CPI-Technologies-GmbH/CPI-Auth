# Tenants

CPI Auth is built as a multi-tenant platform from the ground up. Every piece of data -- users, applications, roles, templates -- is scoped to a tenant. This page explains the tenant model and how to manage tenants.

## Multi-Tenant Model

A **tenant** represents an isolated environment within CPI Auth. Each tenant has its own:

- Users and their credentials
- Applications (OAuth clients)
- Roles and permissions
- Organizations
- Page and email templates
- Language strings
- Webhooks and actions
- Custom field definitions
- Audit logs
- Branding and settings

Tenants share the same database but are logically isolated. Every database query includes a `tenant_id` filter to ensure strict data separation.

## Tenant Structure

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Acme Corp",
  "slug": "acme",
  "domain": "auth.acme.com",
  "parent_id": null,
  "settings": { ... },
  "branding": { ... },
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

| Field | Description |
|-------|-------------|
| `id` | UUID primary key |
| `name` | Human-readable tenant name |
| `slug` | URL-safe identifier used for tenant resolution |
| `domain` | Custom domain for this tenant (optional) |
| `parent_id` | Parent tenant UUID for hierarchical tenants (optional) |
| `settings` | JSON object with tenant-level configuration |
| `branding` | JSON object with visual customization |

## Tenant Settings

Each tenant has configurable settings that control security policies and user behavior:

```json
{
  "password_min_length": 8,
  "password_require_upper": true,
  "password_require_lower": true,
  "password_require_digit": true,
  "password_require_symbol": false,
  "password_history_count": 5,
  "mfa_required": false,
  "allowed_mfa_methods": ["totp", "email", "webauthn"],
  "session_duration_minutes": 1440,
  "inactivity_timeout_minutes": 60,
  "allowed_signup_domains": [],
  "enable_self_signup": true
}
```

### Settings Reference

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `password_min_length` | int | 8 | Minimum password length |
| `password_require_upper` | bool | true | Require uppercase letter |
| `password_require_lower` | bool | true | Require lowercase letter |
| `password_require_digit` | bool | true | Require numeric digit |
| `password_require_symbol` | bool | false | Require special character |
| `password_history_count` | int | 5 | Number of previous passwords to check against |
| `mfa_required` | bool | false | Force MFA enrollment for all users |
| `allowed_mfa_methods` | []string | ["totp","email","webauthn"] | Which MFA methods users can enroll |
| `session_duration_minutes` | int | 1440 | Maximum session lifetime (24h default) |
| `inactivity_timeout_minutes` | int | 60 | Session timeout after inactivity |
| `allowed_signup_domains` | []string | [] | Email domains allowed for self-signup (empty = all) |
| `enable_self_signup` | bool | true | Allow users to register themselves |

## Tenant Branding

Customize the look and feel of login and account pages per tenant:

```json
{
  "primary_color": "#4F46E5",
  "logo_url": "https://acme.com/logo.png",
  "favicon_url": "https://acme.com/favicon.ico",
  "background_color": "#F9FAFB",
  "layout_mode": "centered"
}
```

| Property | Description |
|----------|-------------|
| `primary_color` | Accent color for buttons and links |
| `logo_url` | URL of the tenant logo displayed on auth pages |
| `favicon_url` | Browser tab icon |
| `background_color` | Page background color |
| `layout_mode` | Layout style: `centered`, `split`, or `fullwidth` |

## Custom Domains

Each tenant can have a custom domain for white-label authentication. See the [Custom Domains](./custom-domains) guide for the full verification flow.

When a custom domain is configured and verified, users access the login UI at that domain instead of the default CPI Auth URL.

## Managing Tenants

### Create a Tenant

```bash
curl -X POST http://localhost:5050/api/v1/tenants \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "slug": "acme",
    "settings": {
      "password_min_length": 10,
      "mfa_required": true,
      "allowed_mfa_methods": ["totp", "webauthn"],
      "session_duration_minutes": 720
    },
    "branding": {
      "primary_color": "#4F46E5",
      "logo_url": "https://acme.com/logo.png",
      "layout_mode": "split"
    }
  }'
```

### List Tenants

```bash
curl http://localhost:5050/api/v1/tenants \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Acme Corp",
      "slug": "acme",
      "domain": "",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 20,
  "total_pages": 1
}
```

### Update a Tenant

```bash
curl -X PATCH http://localhost:5050/api/v1/tenants/{tenant_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation",
    "settings": {
      "password_min_length": 12,
      "password_require_symbol": true
    }
  }'
```

### Delete a Tenant

```bash
curl -X DELETE http://localhost:5050/api/v1/tenants/{tenant_id} \
  -H "Authorization: Bearer $TOKEN"
```

::: danger
Deleting a tenant removes all associated users, applications, and data. This action cannot be undone.
:::

### Force Logout All Users in a Tenant

```bash
curl -X POST http://localhost:5050/api/v1/tenants/{tenant_id}/force-logout \
  -H "Authorization: Bearer $TOKEN"
```

This revokes all active sessions for every user in the tenant.

## Tenant Resolution

CPI Auth resolves the current tenant from incoming requests in this order:

1. **`X-Tenant-ID` header** -- Explicit tenant ID in the request
2. **Custom domain** -- The `Host` header is matched against verified tenant domains
3. **JWT claim** -- The `tenant_id` claim in the authenticated access token

## Managing Tenants in the Admin UI

1. Navigate to **Settings** in the Admin UI sidebar
2. The **General** tab shows the current tenant name and slug
3. The **Security** tab exposes all password, MFA, and session settings
4. The **Branding** tab lets you configure colors, logo, and layout
5. The **Domain** tab handles custom domain verification

## Next Steps

- [Users](./users) -- Manage users within a tenant
- [Custom Domains](./custom-domains) -- Set up white-label domains
- [Page Templates](./page-templates) -- Customize auth pages per tenant
