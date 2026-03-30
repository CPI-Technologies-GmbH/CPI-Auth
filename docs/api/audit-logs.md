# Audit Logs API

CPI Auth automatically records an audit trail of all significant actions. Audit logs are immutable and cannot be modified or deleted through the API.

## Base URL

```
http://localhost:5054/admin/audit-logs
```

---

## List Audit Logs

### GET /admin/audit-logs

Retrieve a paginated list of audit log entries with optional filtering.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `per_page` | integer | 20 | Items per page (max 100) |
| `action` | string | | Filter by action type |
| `actor_id` | string | | Filter by the user who performed the action |
| `target_id` | string | | Filter by the target entity ID |
| `from` | string | | Start date (ISO 8601) |
| `to` | string | | End date (ISO 8601) |
| `sort` | string | `created_at` | Sort field |
| `order` | string | `desc` | Sort order: `asc` or `desc` |

### curl Example

```bash
curl "http://localhost:5054/admin/audit-logs?page=1&per_page=20&action=login.success" \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "log-uuid-1",
      "action": "login.success",
      "actor_id": "user-uuid-1",
      "actor_email": "jane@example.com",
      "actor_name": "Jane Doe",
      "target_type": "user",
      "target_id": "user-uuid-1",
      "ip_address": "192.168.1.100",
      "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)...",
      "metadata": {
        "method": "password",
        "application": "My SPA"
      },
      "created_at": "2026-03-28T12:00:00Z"
    },
    {
      "id": "log-uuid-2",
      "action": "user.updated",
      "actor_id": "admin-uuid-1",
      "actor_email": "admin@example.com",
      "actor_name": "Admin User",
      "target_type": "user",
      "target_id": "user-uuid-2",
      "ip_address": "10.0.0.1",
      "user_agent": "Mozilla/5.0...",
      "metadata": {
        "changes": {
          "name": {
            "old": "Bob Smith",
            "new": "Robert Smith"
          }
        }
      },
      "created_at": "2026-03-28T11:45:00Z"
    }
  ],
  "total": 1547,
  "page": 1,
  "per_page": 20,
  "total_pages": 78
}
```

### Filtering by Date Range

```bash
curl "http://localhost:5054/admin/audit-logs?from=2026-03-01T00:00:00Z&to=2026-03-28T23:59:59Z" \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

### Filtering by Actor

```bash
curl "http://localhost:5054/admin/audit-logs?actor_id=admin-uuid-1" \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

### Filtering by Multiple Actions

```bash
curl "http://localhost:5054/admin/audit-logs?action=login.failed" \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

---

## Export Audit Logs

### GET /admin/audit-logs/export

Export audit logs as a CSV file. Supports the same query parameters as the list endpoint for filtering.

```bash
curl "http://localhost:5054/admin/audit-logs/export?from=2026-03-01T00:00:00Z&to=2026-03-28T23:59:59Z" \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}" \
  -o audit-logs.csv
```

### CSV Format

```csv
id,action,actor_id,actor_email,target_type,target_id,ip_address,created_at
log-uuid-1,login.success,user-uuid-1,jane@example.com,user,user-uuid-1,192.168.1.100,2026-03-28T12:00:00Z
log-uuid-2,user.updated,admin-uuid-1,admin@example.com,user,user-uuid-2,10.0.0.1,2026-03-28T11:45:00Z
```

### Filtered Export

```bash
# Export only failed logins for the past week
curl "http://localhost:5054/admin/audit-logs/export?action=login.failed&from=2026-03-21T00:00:00Z" \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}" \
  -o failed-logins.csv
```

---

## Action Types

### Authentication Actions

| Action | Description |
|--------|-------------|
| `login.success` | Successful user login |
| `login.failed` | Failed login attempt |
| `logout` | User logged out |
| `token.refreshed` | Access token refreshed |
| `token.revoked` | Token was revoked |
| `password.changed` | Password was changed |
| `password.reset_requested` | Password reset email sent |
| `password.reset_completed` | Password was reset via link |

### User Management Actions

| Action | Description |
|--------|-------------|
| `user.created` | New user account created |
| `user.updated` | User profile updated |
| `user.deleted` | User account deleted |
| `user.blocked` | User was blocked |
| `user.unblocked` | User was unblocked |
| `user.impersonated` | Admin impersonated user |
| `user.imported` | Users imported via bulk import |
| `user.exported` | Users exported to CSV |

### MFA Actions

| Action | Description |
|--------|-------------|
| `mfa.enabled` | MFA was enabled |
| `mfa.disabled` | MFA was disabled |
| `mfa.challenge.success` | MFA verification succeeded |
| `mfa.challenge.failed` | MFA verification failed |

### Application Actions

| Action | Description |
|--------|-------------|
| `application.created` | Application registered |
| `application.updated` | Application settings modified |
| `application.deleted` | Application removed |
| `application.secret_rotated` | Client secret was rotated |

### Role & Permission Actions

| Action | Description |
|--------|-------------|
| `role.created` | New role created |
| `role.updated` | Role modified |
| `role.deleted` | Role deleted |
| `role.assigned` | Role assigned to user |
| `role.unassigned` | Role removed from user |

### Settings Actions

| Action | Description |
|--------|-------------|
| `settings.updated` | Tenant settings modified |
| `branding.updated` | Branding settings modified |
| `tenant.force_logout` | Force logout for all tenant users |

---

## Log Entry Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique log entry ID |
| `action` | string | The action that was performed |
| `actor_id` | string | UUID of the user who performed the action |
| `actor_email` | string | Email of the acting user |
| `actor_name` | string | Name of the acting user |
| `target_type` | string | Entity type: `user`, `application`, `role`, etc. |
| `target_id` | string | UUID of the affected entity |
| `ip_address` | string | IP address of the request |
| `user_agent` | string | Browser/client user agent string |
| `metadata` | object | Additional context (changes, method, etc.) |
| `created_at` | string | ISO 8601 timestamp |
