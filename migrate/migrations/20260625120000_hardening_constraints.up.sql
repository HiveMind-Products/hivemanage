ALTER TABLE asset ADD COLUMN IF NOT EXISTS original_name VARCHAR(255) NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS user_username_unique_idx ON "user" (username) WHERE username <> '';
CREATE UNIQUE INDEX IF NOT EXISTS user_email_unique_idx ON "user" (email) WHERE email <> '';
CREATE UNIQUE INDEX IF NOT EXISTS token_hash_unique_idx ON token (token_hash);
CREATE UNIQUE INDEX IF NOT EXISTS organization_member_unique_idx ON organization_member (organization_id, user_id);
CREATE UNIQUE INDEX IF NOT EXISTS dataset_org_name_unique_idx ON dataset (organization_id, name);

CREATE INDEX IF NOT EXISTS session_expires_at_idx ON session (expires_at);
CREATE INDEX IF NOT EXISTS token_organization_id_idx ON token (organization_id);
CREATE INDEX IF NOT EXISTS organization_member_user_id_idx ON organization_member (user_id);
CREATE INDEX IF NOT EXISTS asset_organization_created_idx ON asset (organization_id, created_at DESC);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'session_user_id_fk') THEN
        ALTER TABLE session ADD CONSTRAINT session_user_id_fk FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE NOT VALID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'token_organization_id_fk') THEN
        ALTER TABLE token ADD CONSTRAINT token_organization_id_fk FOREIGN KEY (organization_id) REFERENCES organization(id) ON DELETE CASCADE NOT VALID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'token_user_id_fk') THEN
        ALTER TABLE token ADD CONSTRAINT token_user_id_fk FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE SET NULL NOT VALID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'organization_member_organization_id_fk') THEN
        ALTER TABLE organization_member ADD CONSTRAINT organization_member_organization_id_fk FOREIGN KEY (organization_id) REFERENCES organization(id) ON DELETE CASCADE NOT VALID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'organization_member_user_id_fk') THEN
        ALTER TABLE organization_member ADD CONSTRAINT organization_member_user_id_fk FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE NOT VALID;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'asset_organization_id_fk') THEN
        ALTER TABLE asset ADD CONSTRAINT asset_organization_id_fk FOREIGN KEY (organization_id) REFERENCES organization(id) ON DELETE CASCADE NOT VALID;
    END IF;
END $$;
