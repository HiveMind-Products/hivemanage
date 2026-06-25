ALTER TABLE asset DROP CONSTRAINT IF EXISTS asset_organization_id_fk;
ALTER TABLE organization_member DROP CONSTRAINT IF EXISTS organization_member_user_id_fk;
ALTER TABLE organization_member DROP CONSTRAINT IF EXISTS organization_member_organization_id_fk;
ALTER TABLE token DROP CONSTRAINT IF EXISTS token_user_id_fk;
ALTER TABLE token DROP CONSTRAINT IF EXISTS token_organization_id_fk;
ALTER TABLE session DROP CONSTRAINT IF EXISTS session_user_id_fk;

DROP INDEX IF EXISTS asset_organization_created_idx;
DROP INDEX IF EXISTS organization_member_user_id_idx;
DROP INDEX IF EXISTS token_organization_id_idx;
DROP INDEX IF EXISTS session_expires_at_idx;
DROP INDEX IF EXISTS dataset_org_name_unique_idx;
DROP INDEX IF EXISTS organization_member_unique_idx;
DROP INDEX IF EXISTS token_hash_unique_idx;
DROP INDEX IF EXISTS user_email_unique_idx;
DROP INDEX IF EXISTS user_username_unique_idx;

ALTER TABLE asset DROP COLUMN IF EXISTS original_name;
