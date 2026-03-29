# Roles & Permissions API

CPI Auth uses a role-based access control (RBAC) system with support for hierarchical roles. Permissions define granular access rights, and roles bundle permissions together for easy assignment to users.

## Key Concepts

- **Permissions** are granular access rights (e.g., `users:read`, `posts:write`)
- **Roles** group permissions together (e.g., `editor` has `posts:read` + `posts:write`)
- **Role hierarchy** allows a parent role to inherit all permissions from child roles
- **System permissions** are built-in and cannot be deleted
- Token permissions = user effective permissions intersected with application whitelist

---

## Roles

### Base URL

```
http://localhost:5054/admin/roles
```

### List Roles

#### GET /admin/roles

```bash
curl http://localhost:5054/admin/roles \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

#### Response `200 OK`

```json
{
  "data": [
    {
      "id": "role-uuid-1",
      "name": "admin",
      "description": "Full administrative access",
      "parent_role_id": null,
      "is_system": true,
      "permissions": [
        {
          "id": "perm-uuid-1",
          "name": "users:read"
        },
        {
          "id": "perm-uuid-2",
          "name": "users:write"
        },
        {
          "id": "perm-uuid-3",
          "name": "settings:read"
        },
        {
          "id": "perm-uuid-4",
          "name": "settings:write"
        }
      ],
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": "role-uuid-2",
      "name": "editor",
      "description": "Can manage content",
      "parent_role_id": "role-uuid-1",
      "is_system": false,
      "permissions": [
        {
          "id": "perm-uuid-5",
          "name": "posts:read"
        },
        {
          "id": "perm-uuid-6",
          "name": "posts:write"
        }
      ],
      "created_at": "2025-06-15T10:00:00Z",
      "updated_at": "2026-03-20T08:00:00Z"
    }
  ],
  "total": 2
}
```

### Create Role

#### POST /admin/roles

```bash
curl -X POST http://localhost:5054/admin/roles \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "moderator",
    "description": "Can moderate user content",
    "parent_role_id": "role-uuid-2",
    "permission_ids": ["perm-uuid-5", "perm-uuid-7"]
  }'
```

#### Response `201 Created`

```json
{
  "id": "role-uuid-3",
  "name": "moderator",
  "description": "Can moderate user content",
  "parent_role_id": "role-uuid-2",
  "is_system": false,
  "permissions": [
    {
      "id": "perm-uuid-5",
      "name": "posts:read"
    },
    {
      "id": "perm-uuid-7",
      "name": "comments:moderate"
    }
  ],
  "created_at": "2026-03-28T12:00:00Z",
  "updated_at": "2026-03-28T12:00:00Z"
}
```

### Update Role

#### PATCH /admin/roles/:id

```bash
curl -X PATCH http://localhost:5054/admin/roles/role-uuid-3 \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "description": "Can moderate all user-generated content",
    "permission_ids": ["perm-uuid-5", "perm-uuid-7", "perm-uuid-8"]
  }'
```

#### Response `200 OK`

Returns the full updated role object.

### Delete Role

#### DELETE /admin/roles/:id

Remove a custom role. Users with this role lose its permissions. System roles cannot be deleted.

```bash
curl -X DELETE http://localhost:5054/admin/roles/role-uuid-3 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

#### Response `204 No Content`

#### Error: Deleting System Role

```json
{
  "error": "forbidden",
  "error_description": "System roles cannot be deleted"
}
```

---

## Permissions

### Base URL

```
http://localhost:5054/admin/permissions
```

### List Permissions

#### GET /admin/permissions

```bash
curl http://localhost:5054/admin/permissions \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

#### Response `200 OK`

```json
{
  "data": [
    {
      "id": "perm-uuid-1",
      "name": "users:read",
      "description": "View user profiles and lists",
      "is_system": true,
      "created_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": "perm-uuid-2",
      "name": "users:write",
      "description": "Create and update users",
      "is_system": true,
      "created_at": "2025-01-01T00:00:00Z"
    },
    {
      "id": "perm-uuid-10",
      "name": "reports:generate",
      "description": "Generate analytics reports",
      "is_system": false,
      "created_at": "2026-01-15T10:00:00Z"
    }
  ],
  "total": 10
}
```

### Create Permission

#### POST /admin/permissions

```bash
curl -X POST http://localhost:5054/admin/permissions \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "invoices:write",
    "description": "Create and update invoices"
  }'
```

#### Response `201 Created`

```json
{
  "id": "perm-uuid-11",
  "name": "invoices:write",
  "description": "Create and update invoices",
  "is_system": false,
  "created_at": "2026-03-28T12:00:00Z"
}
```

### Update Permission

#### PATCH /admin/permissions/:id

```bash
curl -X PATCH http://localhost:5054/admin/permissions/perm-uuid-11 \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "description": "Create, update, and send invoices"
  }'
```

### Delete Permission

#### DELETE /admin/permissions/:id

System permissions cannot be deleted and return `403 Forbidden`.

```bash
curl -X DELETE http://localhost:5054/admin/permissions/perm-uuid-11 \
  -H "Authorization: Bearer {token}" \
  -H "X-Tenant-ID: {tenant_id}"
```

#### Response `204 No Content`

#### Error: Deleting System Permission

```json
{
  "error": "forbidden",
  "error_description": "System permissions cannot be deleted"
}
```

---

## Role Hierarchy

Roles can be organized in a parent-child hierarchy using `parent_role_id`. A user assigned to a parent role effectively inherits all permissions of its children.

```
admin (parent)
├── editor
│   ├── posts:read
│   └── posts:write
└── moderator
    ├── posts:read
    └── comments:moderate
```

When a user is assigned the `admin` role, they receive all permissions from `editor` and `moderator` in addition to any permissions directly assigned to `admin`.

### Setting Up a Hierarchy

```bash
# Create base role
curl -X POST http://localhost:5054/admin/roles \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "viewer",
    "description": "Read-only access",
    "permission_ids": ["perm-uuid-5"]
  }'

# Create child role with parent
curl -X POST http://localhost:5054/admin/roles \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: {tenant_id}" \
  -d '{
    "name": "contributor",
    "description": "Can view and create content",
    "parent_role_id": "viewer-role-uuid",
    "permission_ids": ["perm-uuid-6"]
  }'
```

---

## Naming Conventions

Permissions follow a `resource:action` pattern:

| Pattern | Examples |
|---------|----------|
| `resource:read` | `users:read`, `posts:read` |
| `resource:write` | `users:write`, `settings:write` |
| `resource:delete` | `users:delete`, `posts:delete` |
| `resource:action` | `users:impersonate`, `reports:generate` |
