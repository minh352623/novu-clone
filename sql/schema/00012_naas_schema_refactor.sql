-- +goose Up
-- +goose StatementBegin

-- 1. IAM Updates
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS features JSONB DEFAULT '{}';

-- 2. Webhook Logs
CREATE TABLE IF NOT EXISTS webhook_logs (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    event_type TEXT NOT NULL,
    request_payload JSONB,
    response_code INTEGER,
    response_body TEXT,
    duration_ms INTEGER,
    status TEXT NOT NULL,
    tenant_id UUID NOT NULL,
    app_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Retention Policy Function
CREATE OR REPLACE FUNCTION delete_old_webhook_logs()
RETURNS void AS $$
BEGIN
    DELETE FROM webhook_logs WHERE created_at < NOW() - INTERVAL '90 days';
END;
$$ LANGUAGE plpgsql;

-- 3. Provider Configurations
CREATE TABLE IF NOT EXISTS provider_configs (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    provider_id UUID NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    app_id UUID NOT NULL,
    configuration JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (provider_id, environment_id)
);

CREATE TRIGGER update_provider_configs_modtime 
BEFORE UPDATE ON provider_configs 
FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

-- Data migration: Move existing configurations to new table
INSERT INTO provider_configs (provider_id, environment_id, tenant_id, app_id, configuration, is_active, created_at, updated_at)
SELECT id, environment_id, tenant_id, app_id, configuration, is_active, created_at, updated_at FROM providers
ON CONFLICT DO NOTHING;

-- Remove old configuration column from providers
ALTER TABLE providers DROP COLUMN IF EXISTS configuration;

-- 4. Workflow Updates
ALTER TABLE workflow_steps ADD COLUMN IF NOT EXISTS provider_config_id UUID REFERENCES provider_configs(id) ON DELETE SET NULL;

-- 5. Security (RLS)
ALTER TABLE provider_configs ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS provider_config_isolation ON provider_configs;
CREATE POLICY provider_config_isolation ON provider_configs
    USING (tenant_id = current_app_tenant());

ALTER TABLE webhook_logs ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS webhook_log_isolation ON webhook_logs;
CREATE POLICY webhook_log_isolation ON webhook_logs
    USING (tenant_id = current_app_tenant());

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 1. Revert Security
ALTER TABLE webhook_logs DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS webhook_log_isolation ON webhook_logs;
ALTER TABLE provider_configs DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS provider_config_isolation ON provider_configs;

-- 2. Revert Workflow
ALTER TABLE workflow_steps DROP COLUMN IF EXISTS provider_config_id;

-- 3. Revert Providers (Restore configuration from provider_configs)
ALTER TABLE providers ADD COLUMN configuration JSONB;

-- Move data back (Warning: This might be complex if multiple configs exist, we take the active one or most recent)
UPDATE providers p
SET configuration = pc.configuration
FROM (
    SELECT DISTINCT ON (provider_id) provider_id, configuration 
    FROM provider_configs 
    ORDER BY provider_id, updated_at DESC
) pc
WHERE p.id = pc.provider_id;

-- Drop new tables
DROP TABLE IF EXISTS provider_configs CASCADE;
DROP TABLE IF EXISTS webhook_logs CASCADE;
DROP FUNCTION IF EXISTS delete_old_webhook_logs();

-- 4. Revert IAM
ALTER TABLE pricing_plans DROP COLUMN IF EXISTS features;

-- +goose StatementEnd
