-- Per-tenant issuer URL.
--
-- Up to v0.1.19 every tenant used the same global issuer (cfg.Security.Issuer)
-- in the iss claim of every issued token. With path-based tenant routing each
-- tenant gets its own canonical issuer (e.g. https://auth.cpi.dev/t/lastsoftware),
-- so RPs can validate tokens against the discovery doc that matches the URL
-- they were redirected to.
--
-- The column is nullable so existing tenants keep using the global issuer
-- until a backfill or admin update assigns one. The runtime falls back to the
-- global issuer when this is NULL or empty.
ALTER TABLE tenants
    ADD COLUMN IF NOT EXISTS issuer_url VARCHAR(512);

COMMENT ON COLUMN tenants.issuer_url IS
    'Per-tenant OIDC issuer URL. NULL means use the global issuer from config.';
