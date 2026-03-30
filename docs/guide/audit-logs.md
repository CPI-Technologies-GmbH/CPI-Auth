# Audit Logs

CPI Auth maintains an immutable audit trail of all significant actions within each tenant. Audit logs are essential for compliance, security monitoring, and debugging.

## Audit Log Model

```json
{
  "id": "a1l2g3o4-b5c6-7890-defg-h12345678901",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "actor_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "action": "user.login",
  "target_type": "user",
  "target_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "metadata": {
    "ip": "192.168.1.100",
    "user_agent": "Mozilla/5.0...",
    "method": "password"
  },
  "ip": "192.168.1.100",
  "created_at": "2025-01-15T10:30:00Z"
}
```

### Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique log entry identifier |
| `tenant_id` | UUID | Tenant this event belongs to |
| `actor_id` | UUID | User who performed the action (null for system events) |
| `action` | string | Event type identifier |
| `target_type` | string | Type of resource affected (user, application, etc.) |
| `target_id` | string | ID of the affected resource |
| `metadata` | JSON | Additional context (IP, user agent, details) |
| `ip` | string | IP address of the request |
| `created_at` | timestamp | When the event occurred |

## Logged Actions

### Authentication Events

| Action | Description |
|--------|-------------|
| `user.login` | Successful authentication |
| `user.login_failed` | Failed authentication attempt |
| `user.logout` | User logged out |
| `user.token_issued` | Access token issued |
| `user.token_revoked` | Token revoked |
| `user.password_changed` | Password updated |
| `user.password_reset_requested` | Password reset email sent |
| `user.email_verified` | Email verification completed |
| `user.mfa_enrolled` | MFA method enrolled |
| `user.mfa_challenge_completed` | MFA challenge passed |

### User Management Events

| Action | Description |
|--------|-------------|
| `user.created` | New user account created |
| `user.updated` | User profile updated |
| `user.deleted` | User account deleted |
| `user.blocked` | User account blocked |
| `user.unblocked` | User account unblocked |
| `user.impersonated` | Admin impersonated a user |
| `user.imported` | Users imported via bulk import |

### Application Events

| Action | Description |
|--------|-------------|
| `application.created` | New application created |
| `application.updated` | Application settings changed |
| `application.deleted` | Application removed |
| `application.secret_rotated` | Client secret rotated |

### Organization Events

| Action | Description |
|--------|-------------|
| `organization.created` | New organization created |
| `organization.updated` | Organization settings changed |
| `organization.deleted` | Organization removed |
| `organization.member_added` | Member added |
| `organization.member_removed` | Member removed |

### Configuration Events

| Action | Description |
|--------|-------------|
| `tenant.settings_updated` | Tenant settings changed |
| `tenant.branding_updated` | Tenant branding changed |
| `role.created` | New role created |
| `role.updated` | Role permissions changed |
| `role.deleted` | Role removed |
| `role.assigned` | Role assigned to user |
| `role.removed` | Role removed from user |
| `webhook.created` | Webhook configured |
| `webhook.updated` | Webhook settings changed |
| `webhook.deleted` | Webhook removed |

## Log Partitioning

Audit logs are partitioned by month in PostgreSQL. This design provides:

- **Fast queries** -- Queries for recent events only scan the relevant partition
- **Easy retention** -- Old partitions can be dropped without affecting recent data
- **Efficient storage** -- Partitions can be independently compressed or archived

Partitions are named `audit_logs_YYYY_MM` (e.g., `audit_logs_2025_01`).

## Querying Audit Logs

### List Audit Logs

```bash
curl "http://localhost:5050/api/v1/audit-logs?page=1&per_page=50" \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "data": [
    {
      "id": "a1l2g3o4-b5c6-7890-defg-h12345678901",
      "actor_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "action": "user.login",
      "target_type": "user",
      "target_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "ip": "192.168.1.100",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "total": 1523,
  "page": 1,
  "per_page": 50,
  "total_pages": 31
}
```

### Filter by Action

```bash
curl "http://localhost:5050/api/v1/audit-logs?action=user.login&page=1&per_page=20" \
  -H "Authorization: Bearer $TOKEN"
```

### User-Specific Audit Log

View all audit events for a specific user:

```bash
curl http://localhost:5050/api/v1/users/{user_id}/audit-log \
  -H "Authorization: Bearer $TOKEN"
```

### Legacy Endpoint

The older endpoint is also available:

```bash
curl "http://localhost:5050/api/v1/logs?page=1&per_page=50" \
  -H "Authorization: Bearer $TOKEN"
```

## Exporting Audit Logs

Export audit logs for compliance reporting or external analysis:

### JSON Export

```bash
curl "http://localhost:5050/api/v1/audit-logs/export?format=json" \
  -H "Authorization: Bearer $TOKEN" \
  -o audit-logs.json
```

### CSV Export

```bash
curl "http://localhost:5050/api/v1/audit-logs/export?format=csv" \
  -H "Authorization: Bearer $TOKEN" \
  -o audit-logs.csv
```

The export endpoints support the same filtering parameters as the list endpoint:

```bash
curl "http://localhost:5050/api/v1/audit-logs/export?format=csv&action=user.login&from=2025-01-01&to=2025-01-31" \
  -H "Authorization: Bearer $TOKEN" \
  -o january-logins.csv
```

## PII Masking

Audit log exports support PII masking to protect sensitive user data when sharing logs with third parties or storing them in less secure systems.

When PII masking is enabled:

| Field | Original | Masked |
|-------|----------|--------|
| Email | jane@example.com | j***@e***.com |
| Name | Jane Smith | J*** S**** |
| IP Address | 192.168.1.100 | 192.168.x.x |
| Phone | +1234567890 | +1***890 |

Enable masking in the export request:

```bash
curl "http://localhost:5050/api/v1/audit-logs/export?format=json&mask_pii=true" \
  -H "Authorization: Bearer $TOKEN" \
  -o masked-audit-logs.json
```

## Audit Log Immutability

Audit logs are append-only. There is no API to update or delete individual audit log entries. This ensures the integrity of the audit trail for compliance purposes.

Administrators can manage retention through database-level partition management, but the API intentionally provides no delete capability.

## Dashboard Integration

The Admin UI dashboard includes audit log visualizations:

- **Recent Events** -- Live feed of the latest audit events
- **Login Chart** -- Login success/failure trends over time
- **Auth Methods** -- Breakdown of authentication methods used

Access these through the Admin UI or the dashboard API:

```bash
# Recent events
curl http://localhost:5050/api/v1/dashboard/events \
  -H "Authorization: Bearer $TOKEN"

# Login chart data
curl http://localhost:5050/api/v1/dashboard/logins \
  -H "Authorization: Bearer $TOKEN"

# Auth methods breakdown
curl http://localhost:5050/api/v1/dashboard/auth-methods \
  -H "Authorization: Bearer $TOKEN"
```

## Next Steps

- [Webhooks & Actions](./webhooks-actions) -- React to audit events
- [Users](./users) -- User-specific audit trails
- [Configuration](./configuration) -- Logging configuration
