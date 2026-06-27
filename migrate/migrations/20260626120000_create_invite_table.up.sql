-- Invitations that grant scoped dashboard access to a Discord user without sharing
-- credentials. An invite is accepted when the invited person logs in with the
-- matching Discord account; acceptance creates an organization_member using the
-- role + permissions captured here.
CREATE TABLE IF NOT EXISTS invite (
    id               TEXT PRIMARY KEY,
    organization_id  TEXT NOT NULL,
    role             TEXT NOT NULL DEFAULT 'VIEWER',
    permissions      JSONB NOT NULL DEFAULT '{}'::jsonb,
    discord_id       TEXT NOT NULL DEFAULT '',
    discord_username TEXT NOT NULL DEFAULT '',
    email            TEXT NOT NULL DEFAULT '',
    created_by       BIGINT,
    accepted_by      BIGINT,
    expires_at       TIMESTAMPTZ,
    accepted_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_invite_organization_id ON invite (organization_id);
CREATE INDEX IF NOT EXISTS idx_invite_discord_id ON invite (discord_id);
