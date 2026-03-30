# Role-Based Access Control

CPI Auth implements a hierarchical RBAC model with dynamic permissions, role inheritance, and application-scoped permission whitelists.

## Concepts

### Roles

A **role** is a named collection of permissions. Roles are scoped to a tenant and can be either system-defined or custom.

```json
{
  "id": "b1c2d3e4-f5a6-7890-bcde-f01234567890",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "editor",
  "description": "Can create and edit content",
  "is_system": false,
  "permissions": ["content:read", "content:write", "media:upload"],
  "parent_role_id": "a0b1c2d3-e4f5-6789-abcd-ef0123456789",
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

### Permissions

A **permission** is a fine-grained access right, identified by a name like `users:read` or `reports:export`. Permissions are grouped for easier management.

```json
{
  "id": "c1d2e3f4-a5b6-7890-cdef-012345678901",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "users:write",
  "display_name": "Manage Users",
  "description": "Create, update, and delete users",
  "group_name": "User Management",
  "is_system": true,
  "created_at": "2025-01-15T10:30:00Z"
}
```

### Role Hierarchy

Roles support single inheritance via `parent_role_id`. A child role inherits all permissions from its parent chain, plus its own permissions.

```
admin
  |-- manager (inherits admin permissions)
        |-- editor (inherits manager + admin permissions)
              |-- viewer (inherits editor + manager + admin permissions)
```

The effective permissions for a user are the **union** of all permissions from all roles assigned to that user, including inherited permissions.

## System Roles

CPI Auth creates the following system roles for each tenant:

| Role | Description |
|------|-------------|
| `admin` | Full access to all tenant resources |
| `manager` | Manage users, applications, and organizations |
| `editor` | Create and modify content and configurations |
| `viewer` | Read-only access to tenant data |

System roles cannot be deleted but their permissions can be customized.

## Token Permission Model

The permissions that appear in a user's access token follow this formula:

```
token_permissions = user_effective_permissions ∩ application_whitelist
```

Where:

- **user_effective_permissions** = union of all permissions from all assigned roles (including inherited permissions)
- **application_whitelist** = the set of permissions configured on the application

If the application whitelist is **empty**, all of the user's effective permissions are included. This means new applications get full permission pass-through by default.

### Example

Consider a user with the `editor` role, which has permissions `[content:read, content:write, media:upload]`.

If the application's permission whitelist is `[content:read, reports:read]`:

```
user_effective = {content:read, content:write, media:upload}
app_whitelist  = {content:read, reports:read}
token_perms    = {content:read}  (intersection)
```

Only `content:read` appears in the token because it is the only permission present in both sets.

## Managing Roles

### Create a Role

```bash
curl -X POST http://localhost:5050/api/v1/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "content-editor",
    "description": "Can manage blog content",
    "permissions": ["content:read", "content:write", "media:upload"],
    "parent_role_id": null
  }'
```

### List Roles

```bash
curl http://localhost:5050/api/v1/roles?page=1&per_page=20 \
  -H "Authorization: Bearer $TOKEN"
```

### Update a Role

```bash
curl -X PATCH http://localhost:5050/api/v1/roles/{role_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated description",
    "permissions": ["content:read", "content:write", "content:publish", "media:upload"]
  }'
```

### Delete a Role

```bash
curl -X DELETE http://localhost:5050/api/v1/roles/{role_id} \
  -H "Authorization: Bearer $TOKEN"
```

::: warning
System roles (`is_system: true`) cannot be deleted.
:::

## Managing Permissions

### Create a Custom Permission

```bash
curl -X POST http://localhost:5050/api/v1/permissions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "invoices:export",
    "display_name": "Export Invoices",
    "description": "Download invoice data as CSV or PDF",
    "group_name": "Billing"
  }'
```

### List Permissions

```bash
curl http://localhost:5050/api/v1/permissions?page=1&per_page=50 \
  -H "Authorization: Bearer $TOKEN"
```

Permissions are returned with their group names, making it easy to render grouped permission pickers in your UI.

### Update a Permission

```bash
curl -X PATCH http://localhost:5050/api/v1/permissions/{perm_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Export Invoices (CSV/PDF)",
    "description": "Download invoice data in CSV or PDF format"
  }'
```

### Delete a Permission

```bash
curl -X DELETE http://localhost:5050/api/v1/permissions/{perm_id} \
  -H "Authorization: Bearer $TOKEN"
```

::: warning
System permissions (`is_system: true`) cannot be deleted. Deleting a custom permission removes it from all roles and application whitelists.
:::

## Assigning Roles to Users

### Assign a Role

```bash
curl -X POST http://localhost:5050/api/v1/users/{user_id}/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role_id": "b1c2d3e4-f5a6-7890-bcde-f01234567890",
    "organization_id": "00000000-0000-0000-0000-000000000000"
  }'
```

The `organization_id` field scopes the role assignment to a specific organization. Use a nil UUID for tenant-wide assignments.

### Get User Roles

```bash
curl http://localhost:5050/api/v1/users/{user_id}/roles \
  -H "Authorization: Bearer $TOKEN"
```

### Remove a Role

```bash
curl -X DELETE http://localhost:5050/api/v1/users/{user_id}/roles/{role_id} \
  -H "Authorization: Bearer $TOKEN"
```

## Permission Naming Convention

CPI Auth uses the `resource:action` naming pattern:

| Permission | Description |
|-----------|-------------|
| `users:read` | View user profiles |
| `users:write` | Create and update users |
| `users:delete` | Delete users |
| `applications:manage` | Full application CRUD |
| `roles:manage` | Create and assign roles |
| `audit:read` | View audit logs |
| `webhooks:manage` | Configure webhooks |
| `settings:manage` | Update tenant settings |

Custom permissions should follow the same convention for consistency.

## Checking Permissions in Your Application

After obtaining an access token, decode it and check the `permissions` array:

```javascript
const token = jwt.decode(accessToken);

function hasPermission(permission) {
  return token.permissions?.includes(permission) ?? false;
}

if (hasPermission('invoices:export')) {
  // Show export button
}
```

For server-side validation, verify the token signature using the JWKS endpoint before checking claims.

## Next Steps

- [Applications](./applications) -- Configure application permission whitelists
- [Organizations](./organizations) -- Organization-scoped roles
- [Auth Flows](./auth-flows) -- How permissions flow into tokens
