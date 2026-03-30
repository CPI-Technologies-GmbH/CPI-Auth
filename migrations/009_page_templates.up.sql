-- Page templates for customizable auth-flow pages (login, signup, etc.)
CREATE TABLE page_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    page_type VARCHAR(50) NOT NULL,
    locale VARCHAR(10) NOT NULL DEFAULT 'en',
    name VARCHAR(255) NOT NULL,
    html_content TEXT NOT NULL DEFAULT '',
    css_content TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, page_type, locale)
);

CREATE INDEX idx_page_templates_tenant ON page_templates(tenant_id);
CREATE INDEX idx_page_templates_type ON page_templates(tenant_id, page_type);
