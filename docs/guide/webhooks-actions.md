# Webhooks and Actions

CPI Auth supports two mechanisms for extending the authentication pipeline: **webhooks** for external notifications and **actions** for inline custom logic.

## Webhooks

Webhooks send HTTP POST requests to external URLs when events occur in CPI Auth. Use them to notify external systems, trigger workflows, or synchronize data.

### Webhook Model

```json
{
  "id": "w1h2k3a4-b5c6-7890-defg-h12345678901",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "url": "https://api.example.com/webhooks/cpi-auth",
  "events": ["user.created", "user.login", "user.deleted"],
  "secret": "whsec_...",
  "active": true,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

### Supported Events

| Event | Trigger |
|-------|---------|
| `user.created` | New user registered |
| `user.updated` | User profile updated |
| `user.deleted` | User deleted |
| `user.blocked` | User account blocked |
| `user.unblocked` | User account unblocked |
| `user.login` | Successful login |
| `user.login_failed` | Failed login attempt |
| `user.password_changed` | Password changed |
| `user.email_verified` | Email verification completed |
| `user.mfa_enrolled` | MFA method enrolled |
| `token.issued` | Access token issued |
| `token.revoked` | Token revoked |
| `application.created` | New application created |
| `application.updated` | Application updated |
| `application.deleted` | Application deleted |
| `organization.created` | New organization created |
| `organization.member_added` | Member added to organization |
| `organization.member_removed` | Member removed from organization |

### Webhook Payload

```json
{
  "id": "evt_abc123",
  "type": "user.created",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-15T10:30:00Z",
  "data": {
    "user": {
      "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "email": "jane@example.com",
      "name": "Jane Smith",
      "status": "active"
    }
  }
}
```

### Webhook Signature

Each webhook request includes an `X-CPI Auth-Signature` header containing an HMAC-SHA256 signature of the payload using the webhook secret. Verify this signature to ensure the request is authentic.

```python
import hmac
import hashlib

def verify_webhook(payload, signature, secret):
    expected = hmac.new(
        secret.encode(),
        payload.encode(),
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(f"sha256={expected}", signature)
```

### Retry Policy

Failed webhook deliveries (non-2xx responses or timeouts) are retried with exponential backoff:

| Attempt | Delay |
|---------|-------|
| 1st retry | 1 minute |
| 2nd retry | 5 minutes |
| 3rd retry | 30 minutes |
| 4th retry | 2 hours |
| 5th retry | 24 hours |

After 5 failed attempts, the delivery is marked as failed. Webhook delivery is powered by NATS JetStream for durability.

### Managing Webhooks

```bash
# Create a webhook
curl -X POST http://localhost:5050/api/v1/webhooks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://api.example.com/webhooks/cpi-auth",
    "events": ["user.created", "user.login", "user.deleted"],
    "active": true
  }'

# List webhooks
curl http://localhost:5050/api/v1/webhooks \
  -H "Authorization: Bearer $TOKEN"

# Update a webhook
curl -X PATCH http://localhost:5050/api/v1/webhooks/{webhook_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "events": ["user.created", "user.login", "user.deleted", "user.blocked"],
    "active": true
  }'

# Delete a webhook
curl -X DELETE http://localhost:5050/api/v1/webhooks/{webhook_id} \
  -H "Authorization: Bearer $TOKEN"

# Test a webhook (sends a test event)
curl -X POST http://localhost:5050/api/v1/webhooks/{webhook_id}/test \
  -H "Authorization: Bearer $TOKEN"

# View delivery history
curl http://localhost:5050/api/v1/webhooks/{webhook_id}/deliveries \
  -H "Authorization: Bearer $TOKEN"
```

## Actions (Hooks)

Actions let you run custom JavaScript code at specific points in the authentication pipeline. Unlike webhooks, actions run **synchronously** and can modify the flow -- blocking logins, enriching tokens, or validating data.

### Supported Triggers

| Trigger | When It Runs | Can Block |
|---------|-------------|-----------|
| `pre-registration` | Before user is created | Yes |
| `post-registration` | After user is created | No |
| `pre-login` | Before credentials are verified | Yes |
| `post-login` | After successful login | No |
| `pre-token` | Before token is issued | Yes (can modify claims) |
| `post-change-password` | After password is changed | No |
| `pre-user-update` | Before user profile is updated | Yes |
| `post-user-delete` | After user is deleted | No |

### Action Model

```json
{
  "id": "a1c2t3n4-b5c6-7890-defg-h12345678901",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "trigger": "pre-login",
  "name": "Block disposable emails",
  "code": "module.exports = async function(event) { ... }",
  "enabled": true,
  "order": 1,
  "timeout_ms": 5000,
  "created_at": "2025-01-15T10:30:00Z"
}
```

### Writing Action Code

Actions are JavaScript functions that receive an `event` object and return a result:

```javascript
// pre-registration: Block disposable email domains
module.exports = async function(event) {
  const blockedDomains = ['tempmail.com', 'throwaway.email', 'guerrillamail.com'];
  const domain = event.user.email.split('@')[1];

  if (blockedDomains.includes(domain)) {
    return {
      error: 'Registration with disposable email addresses is not allowed.'
    };
  }

  return { user: event.user };
};
```

```javascript
// pre-token: Add custom claims to the access token
module.exports = async function(event) {
  return {
    claims: {
      ...event.claims,
      'https://myapp.com/plan': event.user.app_metadata?.plan || 'free',
      'https://myapp.com/org_id': event.user.app_metadata?.org_id
    }
  };
};
```

```javascript
// post-login: Log login event to external analytics
module.exports = async function(event) {
  await fetch('https://analytics.example.com/events', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      event: 'user.login',
      user_id: event.user.id,
      ip: event.request.ip,
      timestamp: new Date().toISOString()
    })
  });
};
```

### Event Object

The event object passed to actions contains:

| Property | Description |
|----------|-------------|
| `event.user` | The user object (if applicable) |
| `event.user.email` | User's email |
| `event.user.name` | User's name |
| `event.user.metadata` | User's metadata |
| `event.user.app_metadata` | User's app metadata |
| `event.request.ip` | Client IP address |
| `event.request.user_agent` | Client user agent |
| `event.claims` | Token claims (for pre-token trigger) |
| `event.application` | The OAuth application |

### Blocking an Action

For `pre-*` triggers, return an `error` property to block the operation:

```javascript
module.exports = async function(event) {
  if (someCondition) {
    return { error: 'Operation not allowed' };
  }
  // Allow the operation to proceed
  return {};
};
```

### Execution Order

When multiple actions are configured for the same trigger, they execute in the order defined by the `order` field. Use the reorder endpoint to change execution order:

```bash
curl -X POST http://localhost:5050/api/v1/actions/reorder \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "action_ids": [
      "first-action-id",
      "second-action-id",
      "third-action-id"
    ]
  }'
```

### Timeouts

Each action has a configurable timeout (`timeout_ms`, default 5000ms). If an action exceeds its timeout, it is terminated and:

- For `pre-*` triggers: the operation proceeds as if the action was not present
- For `post-*` triggers: the timeout is logged but does not affect the operation

### Managing Actions

```bash
# Create an action
curl -X POST http://localhost:5050/api/v1/actions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "trigger": "pre-registration",
    "name": "Block disposable emails",
    "code": "module.exports = async function(event) { const blocked = [\"tempmail.com\"]; const domain = event.user.email.split(\"@\")[1]; if (blocked.includes(domain)) { return { error: \"Disposable emails not allowed\" }; } return {}; }",
    "enabled": true,
    "timeout_ms": 3000
  }'

# List actions
curl http://localhost:5050/api/v1/actions \
  -H "Authorization: Bearer $TOKEN"

# Update an action
curl -X PATCH http://localhost:5050/api/v1/actions/{action_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": false
  }'

# Delete an action
curl -X DELETE http://localhost:5050/api/v1/actions/{action_id} \
  -H "Authorization: Bearer $TOKEN"
```

## Next Steps

- [Auth Flows](./auth-flows) -- Understand where actions fit in the flow
- [Audit Logs](./audit-logs) -- Track webhook deliveries and action executions
- [Configuration](./configuration) -- NATS configuration for webhook delivery
