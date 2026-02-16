-- +goose Up
-- App-Level Permissions

-- Add app_id column to tenant_members
ALTER TABLE tenant_members ADD COLUMN app_id UUID REFERENCES apps(id) ON DELETE CASCADE;

-- Drop old unique constraint
ALTER TABLE tenant_members DROP CONSTRAINT IF EXISTS tenant_members_tenant_id_user_id_key;

-- Add new unique constraint (tenant_id, user_id, app_id)
-- Note: app_id can be NULL (which means tenant-level permission)
-- In PostgreSQL, UNIQUE constraints allow multiple NULLs, so (tenant_id=1, user_id=1, app_id=NULL) can exist alongside (tenant_id=1, user_id=1, app_id=2).
-- However, we only want ONE tenant-level permission per user per tenant.
-- So we need a partial unique index for the case where app_id IS NULL.
CREATE UNIQUE INDEX idx_tenant_members_unique_tenant_app 
ON tenant_members (tenant_id, user_id, app_id) 
WHERE app_id IS NOT NULL;

CREATE UNIQUE INDEX idx_tenant_members_unique_tenant_only 
ON tenant_members (tenant_id, user_id) 
WHERE app_id IS NULL;


-- +goose Down
DROP INDEX IF EXISTS idx_tenant_members_unique_tenant_only;
DROP INDEX IF EXISTS idx_tenant_members_unique_tenant_app;
ALTER TABLE tenant_members ADD CONSTRAINT tenant_members_tenant_id_user_id_key UNIQUE (tenant_id, user_id);
ALTER TABLE tenant_members DROP COLUMN app_id;
