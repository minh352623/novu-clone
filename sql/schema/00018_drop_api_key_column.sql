-- +goose Up
-- +goose StatementBegin
-- Drop the legacy api_key column from environments table
ALTER TABLE environments DROP COLUMN IF EXISTS api_key;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Restore the legacy api_key column (nullable)
ALTER TABLE environments ADD COLUMN IF NOT EXISTS api_key TEXT;
-- Note: Data restoration is not possible here as the data was deleted.
-- Users would need to restore from backup or rely on migrated data in app_api_keys.
ALTER TABLE environments ADD CONSTRAINT environments_api_key_key UNIQUE (api_key);
-- +goose StatementEnd
