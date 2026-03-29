CREATE TABLE domain_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    domain VARCHAR(255) NOT NULL,
    verification_token VARCHAR(255) NOT NULL,
    verification_method VARCHAR(20) NOT NULL DEFAULT 'TXT',
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(domain)
);

CREATE INDEX idx_domain_verifications_tenant ON domain_verifications(tenant_id);
CREATE INDEX idx_domain_verifications_domain ON domain_verifications(domain);
