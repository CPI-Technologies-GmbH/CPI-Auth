# Tenants API

CPI Auth supports multi-tenancy. Each tenant is an isolated environment with its own users, applications, settings, and branding. Tenants are managed through the admin API.

## Base URL

```
http://localhost:5054/admin/tenants
```

---

## List Tenants

### GET /admin/tenants

Retrieve all tenants accessible to the current admin user.

```bash
curl http://localhost:5054/admin/tenants \
  -H "Authorization: Bearer {token}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "tenant-uuid-1",
      "name": "Production",
      "slug": "production",
      "domain": "auth.myapp.com",
      "is_active": true,
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2026-03-28T08:00:00Z"
    },
    {
      "id": "tenant-uuid-2",
      "name": "Staging",
      "slug": "staging",
      "domain": "auth-staging.myapp.com",
      "is_active": true,
      "created_at": "2025-03-01T00:00:00Z",
      "updated_at": "2026-03-20T12:00:00Z"
    }
  ],
  "total": 2
}
```

---

## Create Tenant

### POST /admin/tenants

```bash
curl -X POST http://localhost:5054/admin/tenants \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Development",
    "slug": "development",
    "domain": "auth-dev.myapp.com"
  }'
```

**Response 201 Created:**

```json
{
  "id": "tenant-uuid-3",
  "name": "Development",
  "slug": "development",
  "domain": "auth-dev.myapp.com",
  "is_active": true,
  "created_at": "2026-03-28T12:00:00Z",
  "updated_at": "2026-03-28T12:00:00Z"
}
```

---

## Get Tenant

### GET /admin/tenants/:id

```bash
curl http://localhost:5054/admin/tenants/tenant-uuid-1 \
  -H "Authorization: Bearer {token}"
```

**Response 200 OK:**

```json
{
  "id": "tenant-uuid-1",
  "name": "Production",
  "slug": "production",
  "domain": "auth.myapp.com",
  "is_active": true,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2026-03-28T08:00:00Z"
}
```

---

## Update Tenant

### PATCH /admin/tenants/:id

```bash
curl -X PATCH http://localhost:5054/admin/tenants/tenant-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Production US",
    "domain": "auth-us.myapp.com"
  }'
```

**Response 200 OK:**

Returns the full updated tenant object.

---

## Delete Tenant

### DELETE /admin/tenants/:id

Permanently delete a tenant and all associated data. This action is irreversible.

```bash
curl -X DELETE http://localhost:5054/admin/tenants/tenant-uuid-2 \
  -H "Authorization: Bearer {token}"
```

**Response 204 No Content:**

::: danger
Deleting a tenant removes all users, applications, settings, roles, permissions, and audit logs within that tenant. This cannot be undone.
:::

---

## Force Logout

### POST /admin/tenants/:id/force-logout

Revoke all active sessions across the entire tenant. Every user will be required to re-authenticate.

```bash
curl -X POST http://localhost:5054/admin/tenants/tenant-uuid-1/force-logout \
  -H "Authorization: Bearer {token}"
```

**Response 200 OK:**

```json
{
  "sessions_revoked": 1247
}
```

::: warning
This is a disruptive action that will log out every user in the tenant. Use with caution.
:::

---

## Tenant Settings

### PATCH /admin/settings

Update general settings for the current tenant (resolved from the `X-Tenant-ID` header or JWT).

```bash
curl -X PATCH http://localhost:5054/admin/settings \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "session_lifetime": 86400,
    "idle_session_lifetime": 1800,
    "enable_signup": true,
    "require_email_verification": true,
    "password_policy": {
      "min_length": 8,
      "require_uppercase": true,
      "require_lowercase": true,
      "require_numbers": true,
      "require_special": false
    },
    "mfa_policy": {
      "enabled": true,
      "required": false,
      "methods": ["totp", "email"]
    },
    "default_locale": "en",
    "supported_locales": ["en", "de", "fr", "es"]
  }'
```

**Response 200 OK:**

```json
{
  "session_lifetime": 86400,
  "idle_session_lifetime": 1800,
  "enable_signup": true,
  "require_email_verification": true,
  "password_policy": {
    "min_length": 8,
    "require_uppercase": true,
    "require_lowercase": true,
    "require_numbers": true,
    "require_special": false
  },
  "mfa_policy": {
    "enabled": true,
    "required": false,
    "methods": ["totp", "email"]
  },
  "default_locale": "en",
  "supported_locales": ["en", "de", "fr", "es"],
  "updated_at": "2026-03-28T12:00:00Z"
}
```

---

## Branding Settings

### PATCH /admin/settings/branding

Update the visual branding for the tenant's login and signup pages.

```bash
curl -X PATCH http://localhost:5054/admin/settings/branding \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "colors": {
      "primary": "#4F46E5",
      "secondary": "#7C3AED",
      "background": "#FFFFFF",
      "text": "#1F2937"
    },
    "logo_url": "https://cdn.myapp.com/logo.png",
    "logo_dark_url": "https://cdn.myapp.com/logo-dark.png",
    "font_family": "Inter",
    "border_radius": 8,
    "layout_mode": "centered"
  }'
```

**Response 200 OK:**

```json
{
  "colors": {
    "primary": "#4F46E5",
    "secondary": "#7C3AED",
    "background": "#FFFFFF",
    "text": "#1F2937"
  },
  "logo_url": "https://cdn.myapp.com/logo.png",
  "logo_dark_url": "https://cdn.myapp.com/logo-dark.png",
  "font_family": "Inter",
  "border_radius": 8,
  "layout_mode": "centered",
  "updated_at": "2026-03-28T12:00:00Z"
}
```

### Branding Fields

| Field | Type | Description |
|-------|------|-------------|
| `colors.primary` | string | Primary accent color (hex) |
| `colors.secondary` | string | Secondary accent color (hex) |
| `colors.background` | string | Page background color (hex) |
| `colors.text` | string | Primary text color (hex) |
| `logo_url` | string | Logo URL for light backgrounds |
| `logo_dark_url` | string | Logo URL for dark backgrounds |
| `font_family` | string | CSS font family name |
| `border_radius` | integer | Border radius in pixels (0-24) |
| `layout_mode` | string | `centered`, `split-screen`, or `sidebar` |
