# Applications

Applications in CPI Auth represent OAuth 2.0 / OpenID Connect clients. Every frontend, backend service, or third-party integration that authenticates users or requests API access is registered as an application.

## Application Types

| Type | Code | Use Case | Has Client Secret |
|------|------|----------|-------------------|
| Single Page Application | `spa` | Browser-based apps (React, Vue, Angular) | No |
| Native | `native` | Mobile and desktop apps | No |
| Web | `web` | Server-rendered apps (Next.js, Rails, Django) | Yes |
| Machine-to-Machine | `m2m` | Backend services, cron jobs, microservices | Yes |

### Type Recommendations

- **SPA** -- Use Authorization Code with PKCE. No client secret is issued because browser code cannot keep secrets.
- **Native** -- Same as SPA. Use PKCE and deep links or custom URI schemes for redirect URIs.
- **Web** -- Server-side code can safely store a client secret. Use Authorization Code flow (PKCE still recommended).
- **M2M** -- Use Client Credentials flow. No user interaction; the application authenticates with its own client ID and secret.

## Application Model

```json
{
  "id": "a1b2c3d4-e5f6-7890-abcd-ef0123456789",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My React App",
  "description": "Customer-facing SPA",
  "type": "spa",
  "client_id": "af_spa_xK9mQ2rT5wZ8",
  "logo_url": "https://example.com/app-logo.png",
  "redirect_uris": ["https://app.example.com/callback"],
  "allowed_origins": ["https://app.example.com"],
  "allowed_logout_urls": ["https://app.example.com/logout"],
  "grant_types": ["authorization_code", "refresh_token"],
  "access_token_ttl": 3600,
  "refresh_token_ttl": 2592000,
  "id_token_ttl": 3600,
  "is_active": true,
  "settings": {},
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

## Client ID and Client Secret

- **Client ID** -- A public identifier generated automatically when an application is created. Safe to include in frontend code.
- **Client Secret** -- A confidential credential generated only for `web` and `m2m` application types. The raw secret is returned once at creation time and during rotation. CPI Auth stores only the hash.

::: warning
The client secret is shown only once when the application is created or when it is rotated. Store it securely.
:::

## Redirect URIs

Redirect URIs are the allowed callback URLs where CPI Auth sends authorization codes after authentication. They must be registered in advance to prevent open redirect attacks.

Rules:
- Exact match required (no wildcards in production)
- `http://localhost` is allowed for development
- Must use HTTPS in production (except `localhost`)

## Allowed Origins

For SPA applications, allowed origins define which browser origins can make cross-origin requests to CPI Auth endpoints. These are used for CORS validation.

## Logout URLs

Post-logout redirect URLs define where users are sent after signing out. Only registered URLs are accepted.

## Grant Types

| Grant Type | Description | Application Types |
|------------|-------------|-------------------|
| `authorization_code` | Standard OAuth 2.0 authorization code flow | spa, native, web |
| `client_credentials` | Service-to-service authentication | m2m |
| `refresh_token` | Exchange a refresh token for new tokens | spa, native, web |
| `implicit` | Legacy browser flow (not recommended) | spa |
| `password` | Direct username/password exchange (not recommended) | web, native |

::: tip
Always prefer `authorization_code` with PKCE over `implicit` or `password` grants.
:::

## Token Lifetimes

Each application can override the global token TTLs:

| Token | Field | Default | Description |
|-------|-------|---------|-------------|
| Access Token | `access_token_ttl` | 3600 (1 hour) | Seconds until the access token expires |
| Refresh Token | `refresh_token_ttl` | 2592000 (30 days) | Seconds until the refresh token expires |
| ID Token | `id_token_ttl` | 3600 (1 hour) | Seconds until the ID token expires |

If an application does not specify a TTL, the global values from `config.yaml` are used.

## Application-Scoped Permissions (Whitelist)

Each application can define a **permission whitelist** that limits which permissions appear in tokens issued for that application. This follows the formula:

```
token_permissions = user_effective_permissions ∩ app_whitelist
```

- If the whitelist is **empty**, all of the user's effective permissions are included in the token.
- If the whitelist has entries, only permissions that appear in both the user's effective permissions and the whitelist are included.

This lets you create narrowly scoped applications. For example, a reporting dashboard might only whitelist `reports:read` and `dashboards:read`, even if the user has broader permissions.

### Set Application Permissions

```bash
curl -X PUT http://localhost:5050/api/v1/applications/{app_id}/permissions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "permissions": ["reports:read", "dashboards:read"]
  }'
```

### Get Application Permissions

```bash
curl http://localhost:5050/api/v1/applications/{app_id}/permissions \
  -H "Authorization: Bearer $TOKEN"
```

## Managing Applications

### Create an Application

```bash
curl -X POST http://localhost:5050/api/v1/applications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My React App",
    "type": "spa",
    "redirect_uris": ["https://app.example.com/callback"],
    "allowed_origins": ["https://app.example.com"],
    "allowed_logout_urls": ["https://app.example.com/logout"],
    "grant_types": ["authorization_code", "refresh_token"],
    "access_token_ttl": 1800,
    "refresh_token_ttl": 604800
  }'
```

### List Applications

```bash
curl http://localhost:5050/api/v1/applications?page=1&per_page=20 \
  -H "Authorization: Bearer $TOKEN"
```

### Update an Application

```bash
curl -X PATCH http://localhost:5050/api/v1/applications/{app_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My React App (Production)",
    "redirect_uris": [
      "https://app.example.com/callback",
      "https://staging.example.com/callback"
    ]
  }'
```

### Delete an Application

```bash
curl -X DELETE http://localhost:5050/api/v1/applications/{app_id} \
  -H "Authorization: Bearer $TOKEN"
```

## Secret Rotation

For `web` and `m2m` applications, you can rotate the client secret without downtime:

```bash
curl -X POST http://localhost:5050/api/v1/applications/{app_id}/rotate-secret \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "client_secret": "af_secret_newRandomValue123..."
}
```

The old secret is immediately invalidated. Update your backend services with the new secret promptly.

## Next Steps

- [Auth Flows](./auth-flows) -- Step-by-step flow documentation
- [RBAC](./rbac) -- Permission model and application whitelists
- [Webhooks & Actions](./webhooks-actions) -- React to application events
