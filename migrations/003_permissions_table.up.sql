-- Dynamic permission definitions
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    group_name VARCHAR(100) NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_permissions_tenant ON permissions(tenant_id);
CREATE INDEX idx_permissions_group ON permissions(tenant_id, group_name);

-- Add description and is_system columns to roles table
ALTER TABLE roles ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';
ALTER TABLE roles ADD COLUMN IF NOT EXISTS is_system BOOLEAN NOT NULL DEFAULT FALSE;

-- Add organization_id to user_roles if not present (for org-scoped defaults)
ALTER TABLE user_roles ADD COLUMN IF NOT EXISTS organization_id UUID DEFAULT '00000000-0000-0000-0000-000000000000';

-- Seed system permissions for the default tenant
INSERT INTO permissions (id, tenant_id, name, display_name, description, group_name, is_system)
SELECT gen_random_uuid(), t.id, p.name, p.display_name, p.description, p.group_name, TRUE
FROM tenants t
CROSS JOIN (VALUES
    ('users:read', 'Read Users', 'View user profiles and details', 'Users'),
    ('users:write', 'Write Users', 'Create, update, and delete users', 'Users'),
    ('users:block', 'Block Users', 'Block and unblock user accounts', 'Users'),
    ('applications:read', 'Read Applications', 'View application configurations', 'Applications'),
    ('applications:write', 'Write Applications', 'Create, update, and delete applications', 'Applications'),
    ('organizations:read', 'Read Organizations', 'View organizations and members', 'Organizations'),
    ('organizations:write', 'Write Organizations', 'Create, update, and delete organizations', 'Organizations'),
    ('roles:read', 'Read Roles', 'View roles and permissions', 'Roles'),
    ('roles:write', 'Write Roles', 'Create, update, and delete roles', 'Roles'),
    ('tenants:read', 'Read Tenants', 'View tenant configurations', 'Tenants'),
    ('tenants:write', 'Write Tenants', 'Create, update, and delete tenants', 'Tenants'),
    ('webhooks:read', 'Read Webhooks', 'View webhook configurations', 'Webhooks'),
    ('webhooks:write', 'Write Webhooks', 'Create, update, and delete webhooks', 'Webhooks'),
    ('logs:read', 'Read Logs', 'View audit logs', 'Logs'),
    ('settings:read', 'Read Settings', 'View platform settings', 'Settings'),
    ('settings:write', 'Write Settings', 'Modify platform settings', 'Settings'),
    ('admin:access', 'Admin Access', 'Access admin dashboard', 'Admin'),
    ('*', 'Super Admin', 'Full access to all resources', 'Admin')
) AS p(name, display_name, description, group_name)
ON CONFLICT (tenant_id, name) DO NOTHING;

-- Mark existing seed roles as system roles
UPDATE roles SET is_system = TRUE WHERE name IN ('admin', 'manager', 'editor', 'viewer');
