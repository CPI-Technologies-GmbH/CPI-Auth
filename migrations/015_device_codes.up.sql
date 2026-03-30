CREATE TABLE device_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    device_code TEXT NOT NULL UNIQUE,
    user_code VARCHAR(10) NOT NULL UNIQUE,
    client_id TEXT NOT NULL,
    scopes TEXT[] DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'authorized', 'denied', 'expired')),
    user_id UUID REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    poll_interval INT NOT NULL DEFAULT 5,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_device_codes_user_code ON device_codes(user_code);
CREATE INDEX idx_device_codes_device_code ON device_codes(device_code);
CREATE INDEX idx_device_codes_expires ON device_codes(expires_at);
