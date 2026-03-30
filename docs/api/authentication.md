# Authentication

The CPI Auth Admin API uses JWT bearer tokens for authentication. All admin endpoints require a valid access token obtained through the login flow.

## Base URL

```
http://localhost:5054/admin/auth
```

## Headers

All authenticated requests must include:

| Header | Description |
|--------|-------------|
| `Authorization` | `Bearer {access_token}` |
| `X-Tenant-ID` | Tenant UUID (optional if embedded in JWT) |
| `Content-Type` | `application/json` |

## Rate Limiting

All authentication endpoints are rate-limited to **100 requests per second** per IP address. Exceeding this limit returns a `429 Too Many Requests` response.

## Error Format

All error responses follow a consistent structure:

```json
{
  "error": "unauthorized",
  "error_description": "Invalid email or password"
}
```

Common error codes:

| HTTP Status | Error | Description |
|-------------|-------|-------------|
| 400 | `bad_request` | Malformed request body |
| 401 | `unauthorized` | Invalid or expired token |
| 403 | `forbidden` | Insufficient permissions |
| 429 | `too_many_requests` | Rate limit exceeded |

---

## POST /admin/auth/login

Authenticate an admin user with email and password. Returns an access token, refresh token, and expiration time.

**Request:**

```json
{
  "email": "admin@example.com",
  "password": "your-secure-password"
}
```

**Response 200 OK:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

### curl Example

```bash
curl -X POST http://localhost:5054/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "your-secure-password"
  }'
```

### Error Responses

| Status | Condition |
|--------|-----------|
| 401 | Invalid credentials |
| 422 | Missing email or password |
| 429 | Too many login attempts |

---

## POST /admin/auth/refresh

Exchange a refresh token for a new access token. The old refresh token is invalidated (rotation).

**Request:**

```json
{
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4..."
}
```

**Response 200 OK:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "bmV3IHJlZnJlc2ggdG9rZW4...",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

### curl Example

```bash
curl -X POST http://localhost:5054/admin/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4..."
  }'
```

### Error Responses

| Status | Condition |
|--------|-----------|
| 401 | Invalid or expired refresh token |
| 401 | Refresh token already used (replay detection) |

---

## GET /admin/auth/me

Returns the profile and permissions of the currently authenticated admin user.

**Response 200 OK:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "admin@example.com",
  "name": "Admin User",
  "role": "admin",
  "permissions": [
    "users:read",
    "users:write",
    "applications:read",
    "applications:write",
    "settings:read",
    "settings:write"
  ],
  "tenant_id": "tenant-uuid",
  "created_at": "2025-01-15T10:30:00Z",
  "last_login": "2026-03-28T08:00:00Z"
}
```

### curl Example

```bash
curl -X GET http://localhost:5054/admin/auth/me \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIs..." \
  -H "X-Tenant-ID: tenant-uuid"
```

---

## POST /admin/auth/logout

Invalidates the current access token and its associated refresh token. The token is added to a deny list.

**Request:**

No request body required. The token to invalidate is read from the `Authorization` header.

**Response 204 No Content:**

No response body.

### curl Example

```bash
curl -X POST http://localhost:5054/admin/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIs..."
```

---

## Tenant Context

CPI Auth is multi-tenant. The tenant context is resolved in the following order of priority:

1. **X-Tenant-ID header** -- explicitly set per request
2. **JWT `tenant_id` claim** -- embedded during login
3. **Default tenant** -- if only one tenant exists

```bash
# Explicit tenant header
curl -X GET http://localhost:5054/admin/users \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000"
```

## Token Lifecycle

| Token | Default TTL | Storage |
|-------|-------------|---------|
| Access token | 1 hour | JWT (stateless) |
| Refresh token | 7 days | Redis (server-side) |

Access tokens are RS256-signed JWTs. Refresh tokens are opaque strings stored server-side in Redis with automatic expiry.

## Typical Authentication Flow

```
1. POST /admin/auth/login    → get access_token + refresh_token
2. Use access_token in Authorization header for all requests
3. When access_token expires, POST /admin/auth/refresh
4. POST /admin/auth/logout to end the session
```
