# Webhooks API

Webhooks allow you to receive HTTP callbacks when events occur in CPI Auth. Configure webhook endpoints to integrate with external services, trigger workflows, or maintain audit trails.

## Base URL

```
http://localhost:5054/admin/webhooks
```

---

## List Webhooks

### GET /admin/webhooks

```bash
curl http://localhost:5054/admin/webhooks \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "wh-uuid-1",
      "name": "User Sync",
      "url": "https://api.myapp.com/webhooks/cpi-auth",
      "events": ["user.created", "user.updated", "user.deleted"],
      "is_active": true,
      "secret": "whsec_abc123...",
      "created_at": "2025-06-15T10:00:00Z",
      "updated_at": "2026-03-28T08:00:00Z"
    }
  ],
  "total": 1
}
```

---

## Create Webhook

### POST /admin/webhooks

```bash
curl -X POST http://localhost:5054/admin/webhooks \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "Security Alerts",
    "url": "https://api.myapp.com/webhooks/security",
    "events": [
      "login.failed",
      "user.blocked",
      "mfa.disabled"
    ],
    "is_active": true
  }'
```

**Response 201 Created:**

```json
{
  "id": "wh-uuid-2",
  "name": "Security Alerts",
  "url": "https://api.myapp.com/webhooks/security",
  "events": ["login.failed", "user.blocked", "mfa.disabled"],
  "is_active": true,
  "secret": "whsec_new_secret_xyz789",
  "created_at": "2026-03-28T12:00:00Z",
  "updated_at": "2026-03-28T12:00:00Z"
}
```

::: tip
Store the `secret` value securely. Use it to verify webhook signatures on your receiving endpoint.
:::

---

## Get Webhook

### GET /admin/webhooks/:id

```bash
curl http://localhost:5054/admin/webhooks/wh-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

---

## Update Webhook

### PATCH /admin/webhooks/:id

```bash
curl -X PATCH http://localhost:5054/admin/webhooks/wh-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "events": [
      "user.created",
      "user.updated",
      "user.deleted",
      "login.success",
      "login.failed"
    ],
    "is_active": true
  }'
```

**Response 200 OK:**

Returns the full updated webhook object.

---

## Delete Webhook

### DELETE /admin/webhooks/:id

```bash
curl -X DELETE http://localhost:5054/admin/webhooks/wh-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 204 No Content:**

---

## Test Webhook

### POST /admin/webhooks/:id/test

Send a test event to the webhook endpoint to verify connectivity.

```bash
curl -X POST http://localhost:5054/admin/webhooks/wh-uuid-1/test \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "success": true,
  "status_code": 200,
  "response_time_ms": 145,
  "response_body": "OK"
}
```

**Response:** on Failure

```json
{
  "success": false,
  "status_code": 500,
  "response_time_ms": 3012,
  "error": "Connection timeout after 3000ms"
}
```

---

## Webhook Deliveries

### GET /admin/webhooks/:id/deliveries

View the delivery history for a webhook, including successes and failures.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `per_page` | integer | 20 | Items per page |
| `status` | string | | Filter: `success`, `failed` |

```bash
curl "http://localhost:5054/admin/webhooks/wh-uuid-1/deliveries?page=1&per_page=10" \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "delivery-uuid-1",
      "event": "user.created",
      "status": "success",
      "status_code": 200,
      "request_body": "{\"event\":\"user.created\",\"data\":{...}}",
      "response_body": "OK",
      "response_time_ms": 89,
      "attempts": 1,
      "delivered_at": "2026-03-28T12:00:00Z"
    },
    {
      "id": "delivery-uuid-2",
      "event": "login.failed",
      "status": "failed",
      "status_code": 503,
      "request_body": "{\"event\":\"login.failed\",\"data\":{...}}",
      "response_body": "Service Unavailable",
      "response_time_ms": 5000,
      "attempts": 3,
      "delivered_at": "2026-03-28T11:30:00Z"
    }
  ],
  "total": 2
}
```

---

## Events Reference

### User Events

| Event | Description |
|-------|-------------|
| `user.created` | A new user was registered |
| `user.updated` | User profile was modified |
| `user.deleted` | User was permanently deleted |
| `user.blocked` | User was blocked |
| `user.unblocked` | User was unblocked |

### Authentication Events

| Event | Description |
|-------|-------------|
| `login.success` | Successful login |
| `login.failed` | Failed login attempt |
| `logout` | User logged out |
| `token.refreshed` | Token was refreshed |
| `token.revoked` | Token was revoked |

### MFA Events

| Event | Description |
|-------|-------------|
| `mfa.enabled` | MFA was enabled for a user |
| `mfa.disabled` | MFA was disabled for a user |
| `mfa.challenge.success` | MFA challenge passed |
| `mfa.challenge.failed` | MFA challenge failed |

### Application Events

| Event | Description |
|-------|-------------|
| `application.created` | New application registered |
| `application.updated` | Application settings changed |
| `application.deleted` | Application was removed |

### Organization Events

| Event | Description |
|-------|-------------|
| `organization.created` | New organization created |
| `organization.member.added` | Member added to organization |
| `organization.member.removed` | Member removed from organization |

---

## Webhook Payload Format

All webhook payloads follow this structure:

```json
{
  "id": "evt-uuid-unique",
  "event": "user.created",
  "timestamp": "2026-03-28T12:00:00Z",
  "tenant_id": "tenant-uuid",
  "data": {
    "id": "user-uuid",
    "email": "newuser@example.com",
    "name": "New User",
    "created_at": "2026-03-28T12:00:00Z"
  }
}
```

## Signature Verification

Webhooks include an `X-CPI Auth-Signature` header containing an HMAC-SHA256 signature of the request body using the webhook secret.

```python
import hmac
import hashlib

def verify_webhook(payload: bytes, signature: str, secret: str) -> bool:
    expected = hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(f"sha256={expected}", signature)
```

```javascript
const crypto = require('crypto');

function verifyWebhook(payload, signature, secret) {
  const expected = crypto
    .createHmac('sha256', secret)
    .update(payload)
    .digest('hex');
  return signature === `sha256=${expected}`;
}
```

## Retry Policy

Failed webhook deliveries are retried with exponential backoff:

| Attempt | Delay |
|---------|-------|
| 1 | Immediate |
| 2 | 1 minute |
| 3 | 5 minutes |

After 3 failed attempts, the delivery is marked as failed. You can monitor delivery status through the deliveries endpoint.
