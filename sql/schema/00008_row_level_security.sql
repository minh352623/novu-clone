-- +goose Up
-- 8. Row Level Security (RLS)

-- Helper function
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION current_app_tenant() RETURNS UUID AS $$
    SELECT NULLIF(current_setting('app.current_tenant', TRUE), '')::UUID;
$$ LANGUAGE SQL STABLE;
-- +goose StatementEnd

ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE apps ENABLE ROW LEVEL SECURITY;
ALTER TABLE environments ENABLE ROW LEVEL SECURITY;
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;

-- Tenants: Users can see their own tenant
CREATE POLICY tenant_isolation ON tenants
    USING (id = current_app_tenant());

-- Apps: Must belong to current tenant/env
CREATE POLICY app_isolation ON apps
    USING (tenant_id = current_app_tenant());

-- Messages: RLS filter by Tenant ID stored in the row
CREATE POLICY message_isolation ON messages
    USING (tenant_id = current_app_tenant());

-- +goose Down
DROP POLICY IF EXISTS message_isolation ON messages;
DROP POLICY IF EXISTS app_isolation ON apps;
DROP POLICY IF EXISTS tenant_isolation ON tenants;

ALTER TABLE messages DISABLE ROW LEVEL SECURITY;
ALTER TABLE environments DISABLE ROW LEVEL SECURITY;
ALTER TABLE apps DISABLE ROW LEVEL SECURITY;
ALTER TABLE tenants DISABLE ROW LEVEL SECURITY;

DROP FUNCTION IF EXISTS current_app_tenant;
