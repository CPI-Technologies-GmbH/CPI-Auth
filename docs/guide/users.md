# Users

Users are the core entity in CPI Auth. Every user belongs to exactly one tenant and can authenticate through passwords, social providers, passwordless links, or WebAuthn credentials.

## User Model

```json
{
  "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "jane@example.com",
  "phone": "+1234567890",
  "name": "Jane Smith",
  "avatar_url": "https://example.com/avatar.jpg",
  "locale": "en",
  "metadata": { "company": "Acme", "department": "Engineering" },
  "app_metadata": { "plan": "enterprise", "stripe_id": "cus_123" },
  "status": "active",
  "email_verified": true,
  "phone_verified": false,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-02-01T08:00:00Z"
}
```

### Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique user identifier |
| `tenant_id` | UUID | Tenant this user belongs to |
| `email` | string | Primary email (unique per tenant) |
| `phone` | string | Phone number (optional) |
| `name` | string | Display name |
| `avatar_url` | string | Profile image URL |
| `locale` | string | Preferred locale (e.g., `en`, `de`, `fr`, `es`) |
| `metadata` | JSON | User-editable metadata (visible to the user) |
| `app_metadata` | JSON | Application-controlled metadata (not user-editable) |
| `status` | string | Account status |
| `email_verified` | bool | Whether the email has been verified |
| `phone_verified` | bool | Whether the phone has been verified |

### Metadata vs App Metadata

- **`metadata`** -- Data the user can read and update through the Account UI or profile API. Use this for preferences, display settings, and non-sensitive profile fields.
- **`app_metadata`** -- Data only writable by administrators and the backend API. Use this for subscription plans, internal IDs, feature flags, and authorization-related data.

## User Statuses

| Status | Description |
|--------|-------------|
| `active` | Normal operating state; user can authenticate |
| `inactive` | Account created but not yet activated (pending verification) |
| `blocked` | Administratively blocked; user cannot authenticate |
| `suspended` | Temporarily suspended (e.g., due to policy violation) |
| `deleted` | Soft-deleted; data retained for audit trail |

## User Lifecycle

```
Create Account
    |
    v
[inactive] ---(verify email)---> [active]
    |                                |
    |                                |---(admin blocks)---> [blocked]
    |                                |                          |
    |                                |                    (admin unblocks)
    |                                |                          |
    |                                |<-------------------------+
    |                                |
    |                                |---(admin suspends)---> [suspended]
    |                                |
    |                                |---(delete)---> [deleted]
```

## Managing Users via API

### Create a User

```bash
curl -X POST http://localhost:5050/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "password": "SecureP@ss123",
    "name": "Jane Smith",
    "phone": "+1234567890",
    "locale": "en",
    "metadata": {
      "company": "Acme"
    }
  }'
```

### List Users

```bash
curl "http://localhost:5050/api/v1/users?page=1&per_page=20&search=jane" \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "data": [ ... ],
  "total": 42,
  "page": 1,
  "per_page": 20,
  "total_pages": 3
}
```

### Get a User

```bash
curl http://localhost:5050/api/v1/users/{user_id} \
  -H "Authorization: Bearer $TOKEN"
```

### Update a User

```bash
curl -X PATCH http://localhost:5050/api/v1/users/{user_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane M. Smith",
    "metadata": { "company": "Acme Corp" },
    "app_metadata": { "plan": "enterprise" }
  }'
```

### Block a User

```bash
curl -X POST http://localhost:5050/api/v1/users/{user_id}/block \
  -H "Authorization: Bearer $TOKEN"
```

### Unblock a User

```bash
curl -X POST http://localhost:5050/api/v1/users/{user_id}/unblock \
  -H "Authorization: Bearer $TOKEN"
```

### Delete a User

```bash
curl -X DELETE http://localhost:5050/api/v1/users/{user_id} \
  -H "Authorization: Bearer $TOKEN"
```

### Reset a User's Password

```bash
curl -X POST http://localhost:5050/api/v1/users/{user_id}/reset-password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "password": "NewSecureP@ss456"
  }'
```

## Session Management

### List Active Sessions

```bash
curl http://localhost:5050/api/v1/users/{user_id}/sessions \
  -H "Authorization: Bearer $TOKEN"
```

### Revoke a Specific Session

```bash
curl -X DELETE http://localhost:5050/api/v1/users/{user_id}/sessions/{session_id} \
  -H "Authorization: Bearer $TOKEN"
```

### Force Logout (All Sessions)

```bash
curl -X POST http://localhost:5050/api/v1/users/{user_id}/force-logout \
  -H "Authorization: Bearer $TOKEN"
```

## Bulk Operations

### Import Users

```bash
curl -X POST http://localhost:5050/api/v1/users/import \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "users": [
      { "email": "user1@example.com", "name": "User One", "password": "Pass123!" },
      { "email": "user2@example.com", "name": "User Two", "password": "Pass456!" }
    ]
  }'
```

### Export Users

```bash
curl http://localhost:5050/api/v1/users/export \
  -H "Authorization: Bearer $TOKEN" \
  -o users.json
```

### Bulk Block

```bash
curl -X POST http://localhost:5050/api/v1/users/bulk/block \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_ids": [
      "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "8a1b2c3d-4e5f-6789-abcd-ef0123456789"
    ]
  }'
```

### Bulk Delete

```bash
curl -X POST http://localhost:5050/api/v1/users/bulk/delete \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_ids": [
      "7c9e6679-7425-40de-944b-e07fc1f90ae7"
    ]
  }'
```

## Impersonation

Administrators can generate a token that acts as a specific user. This is useful for debugging and support.

```bash
curl -X POST http://localhost:5050/api/v1/users/{user_id}/impersonate \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

The impersonation token includes a claim indicating it was issued via impersonation for audit trail purposes.

## User Roles and Identities

### Get User Roles

```bash
curl http://localhost:5050/api/v1/users/{user_id}/roles \
  -H "Authorization: Bearer $TOKEN"
```

### Get User Identities (Social Logins)

```bash
curl http://localhost:5050/api/v1/users/{user_id}/identities \
  -H "Authorization: Bearer $TOKEN"
```

### Get User MFA Enrollments

```bash
curl http://localhost:5050/api/v1/users/{user_id}/mfa \
  -H "Authorization: Bearer $TOKEN"
```

### Get User Audit Log

```bash
curl http://localhost:5050/api/v1/users/{user_id}/audit-log \
  -H "Authorization: Bearer $TOKEN"
```

## Next Steps

- [Applications](./applications) -- Configure OAuth clients
- [RBAC](./rbac) -- Assign roles and permissions to users
- [Custom Fields](./custom-fields) -- Add custom profile fields
- [MFA](./mfa) -- Set up multi-factor authentication
