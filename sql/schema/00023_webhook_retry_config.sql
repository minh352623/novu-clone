-- +goose Up
-- US-RA-02.4: Configurable webhook retry policy
-- Adds retry configuration columns to webhooks table with defaults matching current behavior

ALTER TABLE webhooks
  ADD COLUMN max_retries INT NOT NULL DEFAULT 3,
  ADD COLUMN retry_backoff_seconds INT NOT NULL DEFAULT 5,
  ADD COLUMN retry_timeout_seconds INT NOT NULL DEFAULT 10;

COMMENT ON COLUMN webhooks.max_retries IS 'Maximum number of retry attempts for failed webhooks';
COMMENT ON COLUMN webhooks.retry_backoff_seconds IS 'Base backoff duration in seconds (multiplied by 6x each retry)';
COMMENT ON COLUMN webhooks.retry_timeout_seconds IS 'HTTP request timeout in seconds for webhook dispatch';

-- +goose Down
ALTER TABLE webhooks
  DROP COLUMN IF EXISTS max_retries,
  DROP COLUMN IF EXISTS retry_backoff_seconds,
  DROP COLUMN IF EXISTS retry_timeout_seconds;
