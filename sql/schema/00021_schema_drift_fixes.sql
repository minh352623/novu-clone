-- +goose Up
-- +goose StatementBegin

-- C2: Rename notification_logs to notifications (matches Go GORM model)
ALTER TABLE notification_logs RENAME TO notifications;

-- C3: Add SLA threshold columns to apps and environments
ALTER TABLE apps ADD COLUMN IF NOT EXISTS sla_threshold_seconds INTEGER DEFAULT 0;
ALTER TABLE environments ADD COLUMN IF NOT EXISTS sla_threshold_seconds INTEGER DEFAULT 0;

-- M1: Add channel and is_overdue columns to threads
ALTER TABLE threads ADD COLUMN IF NOT EXISTS channel TEXT NOT NULL DEFAULT 'web';
ALTER TABLE threads ADD COLUMN IF NOT EXISTS is_overdue BOOLEAN NOT NULL DEFAULT FALSE;

-- M2: Add type column to messages (standard vs internal_note)
ALTER TABLE messages ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'standard';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE messages DROP COLUMN IF EXISTS type;
ALTER TABLE threads DROP COLUMN IF EXISTS is_overdue;
ALTER TABLE threads DROP COLUMN IF EXISTS channel;
ALTER TABLE environments DROP COLUMN IF EXISTS sla_threshold_seconds;
ALTER TABLE apps DROP COLUMN IF EXISTS sla_threshold_seconds;
ALTER TABLE notifications RENAME TO notification_logs;

-- +goose StatementEnd
