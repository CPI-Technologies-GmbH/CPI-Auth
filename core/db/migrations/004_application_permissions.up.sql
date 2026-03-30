-- Application-scoped permission whitelist
-- Acts as a whitelist: token = intersection(user permissions, app permissions)
-- If app has 0 entries, all user permissions are included (backward-compatible default)
CREATE TABLE application_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    permission_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(application_id, permission_name)
);

CREATE INDEX idx_app_permissions_app ON application_permissions(application_id);
