-- +goose Up
-- +goose StatementBegin

-- Add retry columns to webhook_logs for retry engine
ALTER TABLE webhook_logs ADD COLUMN IF NOT EXISTS retry_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE webhook_logs ADD COLUMN IF NOT EXISTS max_retries INTEGER NOT NULL DEFAULT 3;
ALTER TABLE webhook_logs ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMP WITH TIME ZONE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE webhook_logs DROP COLUMN IF EXISTS next_retry_at;
ALTER TABLE webhook_logs DROP COLUMN IF EXISTS max_retries;
ALTER TABLE webhook_logs DROP COLUMN IF EXISTS retry_count;

-- +goose StatementEnd
