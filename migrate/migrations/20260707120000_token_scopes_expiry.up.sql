-- API token scoping and expiry. Both columns are nullable so existing tokens
-- keep working: a NULL/empty scopes list means full access (legacy behaviour),
-- and a NULL expires_at means the token never expires.
ALTER TABLE token ADD COLUMN IF NOT EXISTS scopes JSONB;
ALTER TABLE token ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ;
