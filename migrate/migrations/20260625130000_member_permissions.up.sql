ALTER TABLE organization_member
    ADD COLUMN IF NOT EXISTS permissions JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE organization_member
SET role = UPPER(role)
WHERE role IS NOT NULL;

UPDATE organization_member
SET role = 'VIEWER'
WHERE role IS NULL OR role = '' OR role NOT IN ('ADMIN', 'EDITOR', 'VIEWER');

UPDATE organization_member
SET permissions = CASE
    WHEN role = 'ADMIN' THEN
        '{"overview":{"read":true,"write":true},"logs":{"read":true,"write":true},"storage":{"read":true,"write":true},"tokens":{"read":true,"write":true},"team":{"read":true,"write":true}}'::jsonb
    WHEN role = 'EDITOR' THEN
        '{"overview":{"read":true,"write":false},"logs":{"read":true,"write":true},"storage":{"read":true,"write":true},"tokens":{"read":false,"write":false},"team":{"read":false,"write":false}}'::jsonb
    ELSE
        '{"overview":{"read":true,"write":false},"logs":{"read":true,"write":false},"storage":{"read":true,"write":false},"tokens":{"read":false,"write":false},"team":{"read":false,"write":false}}'::jsonb
END
WHERE permissions = '{}'::jsonb;
