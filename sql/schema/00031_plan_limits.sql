-- +goose Up
-- Add plan limit columns to pricing_plans for enforceable quotas

ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS max_apps      INTEGER NOT NULL DEFAULT 0;  -- 0 = unlimited
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS max_members   INTEGER NOT NULL DEFAULT 0;  -- 0 = unlimited
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS max_workflows INTEGER NOT NULL DEFAULT 0;  -- 0 = unlimited
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS max_messages_per_month BIGINT NOT NULL DEFAULT 0; -- 0 = unlimited
ALTER TABLE pricing_plans ADD COLUMN IF NOT EXISTS rate_limit_rpm INTEGER NOT NULL DEFAULT 0; -- tenant-level RPM override, 0 = use env setting

-- Update seed data: Free plan gets reasonable limits
UPDATE pricing_plans
SET max_apps = 3,
    max_members = 5,
    max_workflows = 10,
    max_messages_per_month = 10000,
    rate_limit_rpm = 60
WHERE slug = 'free';

-- +goose Down
ALTER TABLE pricing_plans DROP COLUMN IF EXISTS rate_limit_rpm;
ALTER TABLE pricing_plans DROP COLUMN IF EXISTS max_messages_per_month;
ALTER TABLE pricing_plans DROP COLUMN IF EXISTS max_workflows;
ALTER TABLE pricing_plans DROP COLUMN IF EXISTS max_members;
ALTER TABLE pricing_plans DROP COLUMN IF EXISTS max_apps;
