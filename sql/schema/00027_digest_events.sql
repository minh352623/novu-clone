-- +goose Up
-- Digest events: buffer trigger payloads during a digest window

CREATE TABLE digest_events (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    step_id UUID NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    execution_id UUID NOT NULL REFERENCES workflow_executions(id) ON DELETE CASCADE,
    subscriber_key TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_digest_events_step_exec ON digest_events(step_id, execution_id);

-- +goose Down
DROP INDEX IF EXISTS idx_digest_events_step_exec;
DROP TABLE IF EXISTS digest_events;
