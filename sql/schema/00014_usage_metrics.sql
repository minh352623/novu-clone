-- +goose Up
-- +goose StatementBegin
CREATE TABLE usage_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    provider_type TEXT NOT NULL, -- e.g., 'email', 'sms', 'push'
    request_count BIGINT DEFAULT 1,
    status_code INTEGER,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index for querying metrics by app and time
CREATE INDEX idx_usage_metrics_app_id_timestamp ON usage_metrics(app_id, timestamp);
CREATE INDEX idx_usage_metrics_env_id_timestamp ON usage_metrics(environment_id, timestamp);

-- Enable RLS
ALTER TABLE usage_metrics ENABLE ROW LEVEL SECURITY;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS usage_metrics;
-- +goose StatementEnd
