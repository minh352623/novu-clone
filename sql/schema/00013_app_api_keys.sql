-- +goose Up
-- +goose StatementBegin
CREATE TABLE app_api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    key_hash TEXT NOT NULL UNIQUE,
    key_prefix TEXT NOT NULL, -- e.g., sk_live_
    key_suffix TEXT NOT NULL, -- e.g., ...abcd
    name TEXT NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for fast lookup by environment
CREATE INDEX idx_app_api_keys_environment_id ON app_api_keys(environment_id);

-- Enable RLS
ALTER TABLE app_api_keys ENABLE ROW LEVEL SECURITY;

-- Create policy for RLS (matching existing iam policies logic)
-- Assuming we have an app_id in environments to join and check ownership if needed
-- For now, simple RLS policy if it follows the pattern
-- CREATE POLICY app_api_keys_tenant_isolation ON app_api_keys
--     USING (environment_id IN (SELECT id FROM environments WHERE app_id IN (SELECT id FROM apps WHERE tenant_id = current_setting('app.current_tenant_id')::uuid)));

-- Migration of existing keys from environments to app_api_keys (Optional/Grace period)
-- We'll keep environments.api_key for now but mark it as deprecated in code logic.
-- Actually, it's better to move them or keep them as the "Primary" legacy key.
-- Let's assume we start fresh or the app_builder will handle migration logic if needed.

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS app_api_keys;
-- +goose StatementEnd
