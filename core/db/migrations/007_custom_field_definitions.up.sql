-- Custom field definitions for tenant-configurable user fields
CREATE TABLE custom_field_definitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    label VARCHAR(255) NOT NULL,
    field_type VARCHAR(20) NOT NULL CHECK (field_type IN ('text', 'number', 'select', 'checkbox', 'date', 'tel', 'url', 'email', 'textarea')),
    placeholder VARCHAR(255),
    description TEXT,
    options JSONB,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    visible_on VARCHAR(20) NOT NULL DEFAULT 'both' CHECK (visible_on IN ('registration', 'profile', 'both')),
    position INTEGER NOT NULL DEFAULT 0,
    validation_rules JSONB,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);
