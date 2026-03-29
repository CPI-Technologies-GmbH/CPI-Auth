ALTER TABLE applications DROP COLUMN IF EXISTS access_token_ttl;
ALTER TABLE applications DROP COLUMN IF EXISTS refresh_token_ttl;
ALTER TABLE applications DROP COLUMN IF EXISTS id_token_ttl;
