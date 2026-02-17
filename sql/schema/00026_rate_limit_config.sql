-- +goose Up
-- Rate limiting configuration per environment
ALTER TABLE environments ADD COLUMN rate_limit_rpm INTEGER NOT NULL DEFAULT 0;
ALTER TABLE environments ADD COLUMN rate_limit_daily INTEGER NOT NULL DEFAULT 0;
-- 0 = unlimited (no rate limit enforced)

-- +goose Down
ALTER TABLE environments DROP COLUMN IF EXISTS rate_limit_daily;
ALTER TABLE environments DROP COLUMN IF EXISTS rate_limit_rpm;
