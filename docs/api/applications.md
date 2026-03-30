# Applications API

Manage OAuth applications (clients) registered with CPI Auth. Applications define how external services authenticate users.

## Base URL

```
http://localhost:5054/admin/applications
```

## Application Types

| Type | Description | Has Secret | PKCE Required |
|------|-------------|------------|---------------|
| `spa` | Single Page Application (browser) | No | Yes |
| `native` | Mobile or desktop app | No | Yes |
| `web` | Server-side web application | Yes | No |
| `m2m` | Machine-to-machine (service) | Yes | N/A |

---

## List Applications

### GET /admin/applications

```bash
curl http://localhost:5054/admin/applications \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "app-uuid-1",
      "name": "My SPA",
      "type": "spa",
      "client_id": "app_abc123",
      "is_active": true,
      "description": "Frontend single page application",
      "created_at": "2025-06-15T10:00:00Z",
      "updated_at": "2026-03-28T08:00:00Z"
    }
  ],
  "total": 3
}
```

---

## Create Application

### POST /admin/applications

### SPA Example

```bash
curl -X POST http://localhost:5054/admin/applications \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "My SPA",
    "type": "spa",
    "redirect_uris": [
      "http://localhost:3000/callback",
      "https://myapp.com/callback"
    ],
    "allowed_origins": [
      "http://localhost:3000",
      "https://myapp.com"
    ],
    "allowed_logout_urls": [
      "http://localhost:3000",
      "https://myapp.com"
    ],
    "grant_types": ["authorization_code", "refresh_token"],
    "access_token_ttl": 3600,
    "refresh_token_ttl": 604800,
    "id_token_ttl": 3600,
    "description": "React frontend application"
  }'
```

### M2M Example

```bash
curl -X POST http://localhost:5054/admin/applications \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "Backend Service",
    "type": "m2m",
    "grant_types": ["client_credentials"],
    "access_token_ttl": 86400,
    "description": "Microservice for order processing"
  }'
```

### Web Application Example

```bash
curl -X POST http://localhost:5054/admin/applications \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "Admin Portal",
    "type": "web",
    "redirect_uris": ["https://admin.myapp.com/callback"],
    "allowed_origins": ["https://admin.myapp.com"],
    "allowed_logout_urls": ["https://admin.myapp.com/logout"],
    "grant_types": ["authorization_code", "refresh_token"],
    "access_token_ttl": 3600,
    "refresh_token_ttl": 2592000,
    "id_token_ttl": 3600,
    "description": "Server-rendered admin portal"
  }'
```

**Response 201 Created:**

```json
{
  "id": "app-uuid-new",
  "name": "My SPA",
  "type": "spa",
  "client_id": "app_k8f2m9x1",
  "client_secret": null,
  "redirect_uris": [
    "http://localhost:3000/callback",
    "https://myapp.com/callback"
  ],
  "allowed_origins": [
    "http://localhost:3000",
    "https://myapp.com"
  ],
  "allowed_logout_urls": [
    "http://localhost:3000",
    "https://myapp.com"
  ],
  "grant_types": ["authorization_code", "refresh_token"],
  "access_token_ttl": 3600,
  "refresh_token_ttl": 604800,
  "id_token_ttl": 3600,
  "is_active": true,
  "description": "React frontend application",
  "created_at": "2026-03-28T12:00:00Z",
  "updated_at": "2026-03-28T12:00:00Z"
}
```

::: tip
For `web` and `m2m` types, the `client_secret` is returned only on creation. Store it securely -- it cannot be retrieved again.
:::

---

## Get Application

### GET /admin/applications/:id

```bash
curl http://localhost:5054/admin/applications/app-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

Returns the full application object (same structure as create response, but `client_secret` is masked).

---

## Update Application

### PATCH /admin/applications/:id

Only provided fields are modified.

```bash
curl -X PATCH http://localhost:5054/admin/applications/app-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "My SPA v2",
    "redirect_uris": [
      "http://localhost:3000/callback",
      "https://myapp.com/callback",
      "https://staging.myapp.com/callback"
    ],
    "access_token_ttl": 1800,
    "is_active": true
  }'
```

**Response 200 OK:**

Returns the full updated application object.

---

## Delete Application

### DELETE /admin/applications/:id

Permanently delete an application. All tokens issued to this application become invalid.

```bash
curl -X DELETE http://localhost:5054/admin/applications/app-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 204 No Content:**

---

## Rotate Client Secret

### POST /admin/applications/:id/rotate-secret

Generate a new client secret. The old secret is immediately invalidated.

```bash
curl -X POST http://localhost:5054/admin/applications/app-uuid-1/rotate-secret \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "client_id": "app_k8f2m9x1",
  "client_secret": "new_secret_value_xyz789"
}
```

::: warning
Store the new secret immediately. It will not be shown again. All services using the old secret must be updated.
:::

---

## Application Permissions

### GET /admin/applications/:id/permissions

Get the permission whitelist for an application. An empty list means all permissions are available.

```bash
curl http://localhost:5054/admin/applications/app-uuid-1/permissions \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "permissions": [
    {
      "id": "perm-uuid-1",
      "name": "users:read",
      "description": "Read user profiles"
    },
    {
      "id": "perm-uuid-2",
      "name": "posts:write",
      "description": "Create and edit posts"
    }
  ]
}
```

### PUT /admin/applications/:id/permissions

Set the permission whitelist for an application. Tokens issued to this app will only include permissions from this list.

```bash
curl -X PUT http://localhost:5054/admin/applications/app-uuid-1/permissions \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "permission_ids": ["perm-uuid-1", "perm-uuid-2"]
  }'
```

**Response 200 OK:**

```json
{
  "permissions": [
    {
      "id": "perm-uuid-1",
      "name": "users:read",
      "description": "Read user profiles"
    },
    {
      "id": "perm-uuid-2",
      "name": "posts:write",
      "description": "Create and edit posts"
    }
  ]
}
```

::: info Permission Model
Token permissions = user's effective permissions **intersected with** the application's permission whitelist. If the whitelist is empty, all of the user's permissions are granted.
:::

---

## Fields Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Display name |
| `type` | string | Yes | `spa`, `native`, `web`, or `m2m` |
| `client_id` | string | Auto | Generated on creation |
| `client_secret` | string | Auto | Generated for `web` and `m2m` types |
| `redirect_uris` | string[] | Varies | Required for `spa`, `native`, `web` |
| `allowed_origins` | string[] | No | CORS origins for browser apps |
| `allowed_logout_urls` | string[] | No | Post-logout redirect URLs |
| `grant_types` | string[] | Yes | `authorization_code`, `refresh_token`, `client_credentials` |
| `access_token_ttl` | integer | No | Seconds (default: 3600) |
| `refresh_token_ttl` | integer | No | Seconds (default: 604800) |
| `id_token_ttl` | integer | No | Seconds (default: 3600) |
| `is_active` | boolean | No | Default: true |
| `description` | string | No | Human-readable description |
