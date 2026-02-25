-- +goose Up
-- Add tenant_id to threads, subscribers, workflows for direct tenant isolation
-- This avoids expensive multi-table JOINs in RLS policies

-- Step 1: Add nullable columns
ALTER TABLE threads ADD COLUMN tenant_id UUID REFERENCES tenants(id);
ALTER TABLE subscribers ADD COLUMN tenant_id UUID REFERENCES tenants(id);
ALTER TABLE workflows ADD COLUMN tenant_id UUID REFERENCES tenants(id);

-- Step 2: Backfill from environments → apps → tenant_id
UPDATE threads t SET tenant_id = a.tenant_id
FROM environments e JOIN apps a ON e.app_id = a.id
WHERE t.environment_id = e.id AND t.tenant_id IS NULL;

UPDATE subscribers s SET tenant_id = a.tenant_id
FROM environments e JOIN apps a ON e.app_id = a.id
WHERE s.environment_id = e.id AND s.tenant_id IS NULL;

UPDATE workflows w SET tenant_id = a.tenant_id
FROM environments e JOIN apps a ON e.app_id = a.id
WHERE w.environment_id = e.id AND w.tenant_id IS NULL;

-- Step 3: Set NOT NULL after backfill
ALTER TABLE threads ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE subscribers ALTER COLUMN tenant_id SET NOT NULL;
ALTER TABLE workflows ALTER COLUMN tenant_id SET NOT NULL;

-- Step 4: Indexes for fast lookups
CREATE INDEX idx_threads_tenant_id ON threads(tenant_id);
CREATE INDEX idx_subscribers_tenant_id ON subscribers(tenant_id);
CREATE INDEX idx_workflows_tenant_id ON workflows(tenant_id);

-- +goose Down
DROP INDEX IF EXISTS idx_workflows_tenant_id;
DROP INDEX IF EXISTS idx_subscribers_tenant_id;
DROP INDEX IF EXISTS idx_threads_tenant_id;

ALTER TABLE workflows DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE subscribers DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE threads DROP COLUMN IF EXISTS tenant_id;
