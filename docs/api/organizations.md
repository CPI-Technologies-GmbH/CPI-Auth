# Organizations API

Organizations allow you to group users within a tenant. They are useful for B2B scenarios where your customers have their own teams and need scoped access control.

## Base URL

```
http://localhost:5054/admin/organizations
```

---

## List Organizations

### GET /admin/organizations

Retrieve all organizations in the current tenant.

**Query Parameters:**

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `page` | integer | 1 | Page number |
| `per_page` | integer | 20 | Items per page (max 100) |
| `search` | string | | Search by name |

```bash
curl "http://localhost:5054/admin/organizations?page=1&per_page=20" \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "id": "org-uuid-1",
      "name": "Acme Corporation",
      "display_name": "Acme Corp",
      "metadata": {
        "plan": "enterprise",
        "industry": "technology"
      },
      "member_count": 25,
      "created_at": "2025-06-15T10:00:00Z",
      "updated_at": "2026-03-28T08:00:00Z"
    },
    {
      "id": "org-uuid-2",
      "name": "Globex Inc",
      "display_name": "Globex",
      "metadata": {
        "plan": "pro"
      },
      "member_count": 8,
      "created_at": "2025-09-01T10:00:00Z",
      "updated_at": "2026-02-15T14:00:00Z"
    }
  ],
  "total": 2,
  "page": 1,
  "per_page": 20,
  "total_pages": 1
}
```

---

## Create Organization

### POST /admin/organizations

```bash
curl -X POST http://localhost:5054/admin/organizations \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "New Startup",
    "display_name": "New Startup Inc",
    "metadata": {
      "plan": "free",
      "industry": "fintech"
    }
  }'
```

**Response 201 Created:**

```json
{
  "id": "org-uuid-3",
  "name": "New Startup",
  "display_name": "New Startup Inc",
  "metadata": {
    "plan": "free",
    "industry": "fintech"
  },
  "member_count": 0,
  "created_at": "2026-03-28T12:00:00Z",
  "updated_at": "2026-03-28T12:00:00Z"
}
```

---

## Get Organization

### GET /admin/organizations/:id

```bash
curl http://localhost:5054/admin/organizations/org-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

Returns the full organization object.

---

## Update Organization

### PATCH /admin/organizations/:id

```bash
curl -X PATCH http://localhost:5054/admin/organizations/org-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "display_name": "Acme Corporation LLC",
    "metadata": {
      "plan": "enterprise",
      "industry": "technology",
      "region": "us-west"
    }
  }'
```

**Response 200 OK:**

Returns the full updated organization object.

---

## Delete Organization

### DELETE /admin/organizations/:id

Remove an organization. Members are not deleted but lose their organization membership.

```bash
curl -X DELETE http://localhost:5054/admin/organizations/org-uuid-1 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 204 No Content:**

---

## Organization Members

### GET /admin/organizations/:id/members

List all members of an organization.

```bash
curl http://localhost:5054/admin/organizations/org-uuid-1/members \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

**Response 200 OK:**

```json
{
  "data": [
    {
      "user_id": "user-uuid-1",
      "email": "alice@acme.com",
      "name": "Alice Johnson",
      "role": "admin",
      "joined_at": "2025-06-15T10:00:00Z"
    },
    {
      "user_id": "user-uuid-2",
      "email": "bob@acme.com",
      "name": "Bob Smith",
      "role": "member",
      "joined_at": "2025-07-01T08:00:00Z"
    }
  ],
  "total": 2
}
```

### POST /admin/organizations/:id/members

Add users to an organization.

```bash
curl -X POST http://localhost:5054/admin/organizations/org-uuid-1/members \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "user_ids": ["user-uuid-3", "user-uuid-4"],
    "role": "member"
  }'
```

**Response 200 OK:**

```json
{
  "added": 2
}
```

### DELETE /admin/organizations/:id/members

Remove users from an organization.

```bash
curl -X DELETE http://localhost:5054/admin/organizations/org-uuid-1/members \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "user_ids": ["user-uuid-3"]
  }'
```

**Response 200 OK:**

```json
{
  "removed": 1
}
```

---

## Fields Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique organization identifier |
| `display_name` | string | No | Human-readable display name |
| `metadata` | object | No | Arbitrary key-value metadata |

## Member Roles

| Role | Description |
|------|-------------|
| `admin` | Can manage organization members and settings |
| `member` | Standard organization member |
