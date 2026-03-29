# Users API

Manage users through the CPI Auth Admin API. All endpoints require admin authentication and appropriate permissions.

## Base URL

```
http://localhost:5054/admin/users
```

## Headers

```
Authorization: Bearer {access_token}
Content-Type: application/json
X-Tenant-ID: {tenant_id}
```

---

## List Users

### GET /admin/users

Retrieve a paginated list of users with optional search filtering.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `per_page` | integer | 20 | Items per page (max 100) |
| `search` | string | | Search by email or name |
| `status` | string | | Filter: `active`, `blocked`, `unverified` |
| `sort` | string | `created_at` | Sort field |
| `order` | string | `desc` | Sort order: `asc` or `desc` |

### curl Example

```bash
curl "http://localhost:5054/admin/users?page=1&per_page=20&search=jane" \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "jane@example.com",
      "name": "Jane Doe",
      "phone": "+1234567890",
      "locale": "en",
      "email_verified": true,
      "blocked": false,
      "last_login": "2026-03-28T08:00:00Z",
      "login_count": 42,
      "created_at": "2025-06-15T10:00:00Z",
      "updated_at": "2026-03-28T08:00:00Z"
    }
  ],
  "total": 156,
  "page": 1,
  "per_page": 20,
  "total_pages": 8
}
```

---

## Create User

### POST /admin/users

Create a new user account.

**Request:**

```json
{
  "email": "newuser@example.com",
  "password": "SecureP@ss123",
  "name": "New User",
  "phone": "+1234567890",
  "locale": "en",
  "metadata": {
    "company": "Acme Corp",
    "department": "Engineering"
  }
}
```

### curl Example

```bash
curl -X POST http://localhost:5054/admin/users \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "email": "newuser@example.com",
    "password": "SecureP@ss123",
    "name": "New User",
    "phone": "+1234567890",
    "locale": "en",
    "metadata": {
      "company": "Acme Corp"
    }
  }'
```

**Response 201 Created:**

```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "email": "newuser@example.com",
  "name": "New User",
  "phone": "+1234567890",
  "locale": "en",
  "email_verified": false,
  "blocked": false,
  "metadata": {
    "company": "Acme Corp"
  },
  "created_at": "2026-03-28T12:00:00Z",
  "updated_at": "2026-03-28T12:00:00Z"
}
```

---

## Get User

### GET /admin/users/:id

Retrieve a single user by ID.

```bash
curl http://localhost:5054/admin/users/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "jane@example.com",
  "name": "Jane Doe",
  "phone": "+1234567890",
  "locale": "en",
  "email_verified": true,
  "blocked": false,
  "last_login": "2026-03-28T08:00:00Z",
  "login_count": 42,
  "metadata": {
    "company": "Acme Corp"
  },
  "created_at": "2025-06-15T10:00:00Z",
  "updated_at": "2026-03-28T08:00:00Z"
}
```

---

## Update User

### PATCH /admin/users/:id

Partially update a user. Only provided fields are modified.

```bash
curl -X PATCH http://localhost:5054/admin/users/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "Jane Smith",
    "phone": "+0987654321",
    "metadata": {
      "company": "New Corp"
    }
  }'
```

**Response 200 OK:**

Returns the full updated user object.

---

## Delete User

### DELETE /admin/users/:id

Permanently delete a user and all associated data (sessions, roles, etc.).

```bash
curl -X DELETE http://localhost:5054/admin/users/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 204 No Content:**

---

## Block User

### POST /admin/users/:id/block

Block a user, preventing them from logging in. Active sessions are revoked.

```bash
curl -X POST http://localhost:5054/admin/users/550e8400-e29b-41d4-a716-446655440000/block \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "blocked": true
}
```

## Unblock User

### POST /admin/users/:id/unblock

```bash
curl -X POST http://localhost:5054/admin/users/550e8400-e29b-41d4-a716-446655440000/unblock \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "blocked": false
}
```

---

## User Sessions

### GET /admin/users/:id/sessions

List all active sessions for a user.

```bash
curl http://localhost:5054/admin/users/550e8400-e29b-41d4-a716-446655440000/sessions \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "session-uuid-1",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)...",
      "last_active": "2026-03-28T12:00:00Z",
      "created_at": "2026-03-28T08:00:00Z"
    }
  ]
}
```

### DELETE /admin/users/:id/sessions

Revoke all active sessions for a user (force logout).

```bash
curl -X DELETE http://localhost:5054/admin/users/550e8400-e29b-41d4-a716-446655440000/sessions \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 204 No Content:**

---

## User Roles

### GET /admin/users/:id/roles

List roles assigned to a user.

```bash
curl http://localhost:5054/admin/users/550e8400-e29b-41d4-a716-446655440000/roles \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "role-uuid-1",
      "name": "editor",
      "description": "Can edit content",
      "permissions": ["posts:read", "posts:write"]
    }
  ]
}
```

### POST /admin/users/:id/roles

Assign roles to a user.

```bash
curl -X POST http://localhost:5054/admin/users/550e8400-e29b-41d4-a716-446655440000/roles \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "role_ids": ["role-uuid-1", "role-uuid-2"]
  }'
```

### DELETE /admin/users/:id/roles

Remove roles from a user.

```bash
curl -X DELETE http://localhost:5054/admin/users/550e8400-e29b-41d4-a716-446655440000/roles \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "role_ids": ["role-uuid-1"]
  }'
```

---

## Impersonate User

### POST /admin/users/:id/impersonate

Generate a token to act as the specified user. Useful for debugging user-reported issues.

```bash
curl -X POST http://localhost:5054/admin/users/550e8400-e29b-41d4-a716-446655440000/impersonate \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "expires_in": 3600,
  "impersonating": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "jane@example.com"
  }
}
```

::: warning
Impersonation is logged in the audit trail and requires the `users:impersonate` permission.
:::

---

## Bulk Operations

### POST /admin/users/bulk/block

Block multiple users at once.

```bash
curl -X POST http://localhost:5054/admin/users/bulk/block \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "user_ids": [
      "550e8400-e29b-41d4-a716-446655440000",
      "660e8400-e29b-41d4-a716-446655440001"
    ]
  }'
```

### POST /admin/users/bulk/delete

Delete multiple users at once.

```bash
curl -X POST http://localhost:5054/admin/users/bulk/delete \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "user_ids": [
      "550e8400-e29b-41d4-a716-446655440000",
      "660e8400-e29b-41d4-a716-446655440001"
    ]
  }'
```

**Response 200 OK:**

```json
{
  "affected": 2
}
```

---

## Export Users

### GET /admin/users/export

Export all users as a CSV file.

```bash
curl http://localhost:5054/admin/users/export \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}" \
  -o users.csv
```

**Response:**

Returns a CSV file with headers: `id, email, name, phone, locale, email_verified, blocked, created_at, last_login`.

---

## Import Users

### POST /admin/users/import

Bulk import users from a CSV or JSON file.

```bash
curl -X POST http://localhost:5054/admin/users/import \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}" \
  -F "file=@users.csv"
```

**Response 200 OK:**

```json
{
  "imported": 45,
  "skipped": 3,
  "errors": [
    {
      "row": 12,
      "email": "duplicate@example.com",
      "error": "Email already exists"
    }
  ]
}
```

### CSV Format

```csv
email,password,name,phone,locale
john@example.com,TempPass123!,John Doe,+1234567890,en
jane@example.com,TempPass456!,Jane Doe,+0987654321,de
```
