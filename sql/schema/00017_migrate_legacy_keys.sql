-- Migrate legacy environment API keys to app_api_keys table
-- This migration ensures that keys stored in the legacy 'api_key' column (plaintext) are hashed
-- and moved to the new 'app_api_keys' table so they work with the new Authenticator.

-- +goose Up
-- +goose StatementBegin
-- Ensure pgcrypto extension is checking (it's usually available standard or via extension)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO app_api_keys (
    id,
    app_id,
    environment_id,
    key_hash,
    key_prefix,
    key_suffix,
    name,
    created_at,
    updated_at
)
SELECT
    generate_uuid_v7(),
    app_id,
    id, -- environment_id
    encode(digest(api_key, 'sha256'), 'hex'), -- Hash the legacy plaintext key
    -- Extract prefix (first 7 chars if available, else 'legacy_')
    CASE 
        WHEN length(api_key) >= 7 THEN substring(api_key from 1 for 7)
        ELSE 'legacy_'
    END,
    -- Extract suffix (last 4 chars if available, else 'key')
    CASE 
        WHEN length(api_key) >= 4 THEN substring(api_key from length(api_key)-3 for 4)
        ELSE 'key'
    END,
    'Legacy Migration Key', -- Name to identify these keys
    created_at,
    updated_at
FROM 
    environments 
WHERE 
    api_key IS NOT NULL 
    AND api_key != '' 
    -- Avoid migrating if this environment already has keys? 
    -- No, safer to migrate and have duplicates than to lose access.
    -- But ensure we don't violate UNIQUE constraints if we run this multiple times?
    -- The source is unique (api_key is unique in environments).
    -- But if we run this twice, we might insert duplicates.
    -- Let's use ON CONFLICT DO NOTHING based on hash if possible?
    -- app_api_keys has unique(key_hash). So conflict on key_hash is perfect.
ON CONFLICT (key_hash) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- No rollback logic needed as this is additive migration data
-- +goose StatementEnd
