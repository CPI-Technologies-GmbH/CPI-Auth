-- Per-application token lifetime overrides (in seconds). NULL = use global config.
ALTER TABLE applications ADD COLUMN IF NOT EXISTS access_token_ttl INTEGER;
ALTER TABLE applications ADD COLUMN IF NOT EXISTS refresh_token_ttl INTEGER;
ALTER TABLE applications ADD COLUMN IF NOT EXISTS id_token_ttl INTEGER;
