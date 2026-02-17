-- +goose Up
-- Workflow Engine: Execution tracking tables

CREATE TABLE workflow_executions (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    subscriber_key TEXT NOT NULL,
    trigger_payload JSONB DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'running',
    current_step_id UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_workflow_executions_workflow ON workflow_executions(workflow_id);
CREATE INDEX idx_workflow_executions_status ON workflow_executions(status);

CREATE TABLE step_executions (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    execution_id UUID NOT NULL REFERENCES workflow_executions(id) ON DELETE CASCADE,
    step_id UUID NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    scheduled_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    output JSONB DEFAULT '{}'
);

CREATE INDEX idx_step_executions_execution ON step_executions(execution_id);
CREATE INDEX idx_step_executions_status_scheduled ON step_executions(status, scheduled_at);

-- +goose Down
DROP TABLE IF EXISTS step_executions;
DROP TABLE IF EXISTS workflow_executions;
