-- +goose Up
-- US-RA-02.5: Add direction column to usage_metrics for inbound/outbound tracking
ALTER TABLE usage_metrics ADD COLUMN direction TEXT NOT NULL DEFAULT 'outbound';

-- Indexes for common query patterns
CREATE INDEX idx_usage_metrics_app_direction ON usage_metrics(app_id, direction, timestamp);
CREATE INDEX idx_usage_metrics_env_direction ON usage_metrics(environment_id, direction, timestamp);

-- +goose Down
DROP INDEX IF EXISTS idx_usage_metrics_env_direction;
DROP INDEX IF EXISTS idx_usage_metrics_app_direction;
ALTER TABLE usage_metrics DROP COLUMN IF EXISTS direction;
