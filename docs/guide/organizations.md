# Organizations

Organizations in CPI Auth represent B2B entities -- companies, teams, or groups -- within a tenant. They provide a way to group users and assign organization-scoped roles.

## Organization Model

```json
{
  "id": "d1e2f3a4-b5c6-7890-defg-h12345678901",
  "tenant_id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Acme Engineering",
  "slug": "acme-engineering",
  "domains": ["acme.com", "acme.io"],
  "settings": {},
  "branding": {
    "primary_color": "#2563EB",
    "logo_url": "https://acme.com/logo.png"
  },
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique organization identifier |
| `tenant_id` | UUID | Parent tenant |
| `name` | string | Display name |
| `slug` | string | URL-safe identifier |
| `domains` | []string | Email domains for auto-join |
| `settings` | JSON | Organization-specific settings |
| `branding` | JSON | Organization-specific branding |

## Organization Members

Members link users to organizations with a role:

```json
{
  "organization_id": "d1e2f3a4-b5c6-7890-defg-h12345678901",
  "user_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "role": "admin",
  "created_at": "2025-01-15T10:30:00Z"
}
```

Common organization roles:

| Role | Description |
|------|-------------|
| `admin` | Full organization management |
| `member` | Standard member access |
| `billing` | Access to billing and subscription |
| `readonly` | View-only access |

## Domain-Based Auto-Join

When an organization has `domains` configured (e.g., `["acme.com"]`), users who register with a matching email domain can be automatically added to the organization.

This is useful for B2B SaaS where you want all employees of a company to automatically join their organization.

## Managing Organizations

### Create an Organization

```bash
curl -X POST http://localhost:5050/api/v1/organizations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Engineering",
    "slug": "acme-engineering",
    "domains": ["acme.com", "acme.io"],
    "branding": {
      "primary_color": "#2563EB",
      "logo_url": "https://acme.com/logo.png"
    }
  }'
```

### List Organizations

```bash
curl "http://localhost:5050/api/v1/organizations?page=1&per_page=20" \
  -H "Authorization: Bearer $TOKEN"
```

### Get an Organization

```bash
curl http://localhost:5050/api/v1/organizations/{org_id} \
  -H "Authorization: Bearer $TOKEN"
```

### Update an Organization

```bash
curl -X PATCH http://localhost:5050/api/v1/organizations/{org_id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Engineering Team",
    "domains": ["acme.com", "acme.io", "acme.dev"]
  }'
```

### Delete an Organization

```bash
curl -X DELETE http://localhost:5050/api/v1/organizations/{org_id} \
  -H "Authorization: Bearer $TOKEN"
```

## Managing Members

### Add a Member

```bash
curl -X POST http://localhost:5050/api/v1/organizations/{org_id}/members \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "role": "member"
  }'
```

### List Members

```bash
curl "http://localhost:5050/api/v1/organizations/{org_id}/members?page=1&per_page=20" \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "data": [
    {
      "organization_id": "d1e2f3a4-b5c6-7890-defg-h12345678901",
      "user_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "role": "admin",
      "created_at": "2025-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "per_page": 20,
  "total_pages": 1
}
```

### Remove a Member

```bash
curl -X DELETE http://localhost:5050/api/v1/organizations/{org_id}/members/{user_id} \
  -H "Authorization: Bearer $TOKEN"
```

## Organization-Scoped Roles

Roles can be assigned to users within the context of a specific organization. This allows a user to have different permissions in different organizations.

```bash
# Assign a role scoped to an organization
curl -X POST http://localhost:5050/api/v1/users/{user_id}/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role_id": "b1c2d3e4-f5a6-7890-bcde-f01234567890",
    "organization_id": "d1e2f3a4-b5c6-7890-defg-h12345678901"
  }'
```

When the user authenticates in the context of that organization, only the organization-scoped permissions are included in their token.

## Organization Branding

Each organization can override the tenant's default branding:

```json
{
  "primary_color": "#2563EB",
  "logo_url": "https://acme.com/logo.png",
  "background_color": "#F0F4F8"
}
```

This allows B2B customers to see their own branding on the login and account pages.

## Use Cases

### SaaS Platform with Team Workspaces

Each paying customer creates an organization. Team members are added as members. Organization admins manage their own users while the platform owner manages everything at the tenant level.

### Enterprise SSO per Company

Different organizations can have different identity providers. Employees of Acme authenticate via Acme's SAML IdP, while employees of Globex use Google Workspace.

### Multi-Tier Support

Combine organizations with the role hierarchy for multi-level access control:
- Platform admins (tenant-level roles)
- Organization admins (org-scoped admin role)
- Organization members (org-scoped member role)

## Next Steps

- [RBAC](./rbac) -- Role and permission model
- [Users](./users) -- User management
- [Tenants](./tenants) -- Tenant-level configuration
