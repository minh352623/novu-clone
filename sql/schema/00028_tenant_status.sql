-- +goose Up
-- Add tenant lifecycle status (multi-tenant-guide §5)
ALTER TABLE tenants ADD COLUMN status TEXT NOT NULL DEFAULT 'active';

-- Index for middleware validation queries
CREATE INDEX idx_tenants_status ON tenants(status);

-- +goose Down
DROP INDEX IF EXISTS idx_tenants_status;
ALTER TABLE tenants DROP COLUMN IF EXISTS status;
