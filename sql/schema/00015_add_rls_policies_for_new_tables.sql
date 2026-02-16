-- +goose Up
-- +goose StatementBegin

-- Row Level Security (RLS) for app_api_keys
-- Only allows access if the app belongs to the current tenant
CREATE POLICY api_key_isolation ON app_api_keys
    USING (app_id IN (SELECT id FROM apps WHERE tenant_id = current_app_tenant()));

-- Row Level Security (RLS) for usage_metrics
-- Only allows access if the app belongs to the current tenant
CREATE POLICY usage_metrics_isolation ON usage_metrics
    USING (app_id IN (SELECT id FROM apps WHERE tenant_id = current_app_tenant()));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP POLICY IF EXISTS usage_metrics_isolation ON usage_metrics;
DROP POLICY IF EXISTS api_key_isolation ON app_api_keys;
-- +goose StatementEnd
