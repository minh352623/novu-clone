-- +goose Up
-- 3. Workflow Engine Module

-- Workflow resources are scoped to Environments
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    trigger_identifier TEXT NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (environment_id, trigger_identifier)
);

CREATE TABLE workflow_steps (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    parent_step_id UUID REFERENCES workflow_steps(id),
    step_type TEXT NOT NULL,
    config JSONB DEFAULT '{}',
    "order" INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE subscribers (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    subscriber_key TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    data JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (environment_id, subscriber_key)
);

-- +goose Down
DROP TABLE IF EXISTS subscribers;
DROP TABLE IF EXISTS workflow_steps;
DROP TABLE IF EXISTS workflows;
