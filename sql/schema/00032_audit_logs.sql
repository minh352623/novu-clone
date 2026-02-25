-- +goose Up
-- Audit log table for compliance and security tracking

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    actor_id UUID,           -- user/system who performed the action
    actor_type TEXT NOT NULL, -- user, system, api_key
    action TEXT NOT NULL,     -- create, update, delete, login, export, etc.
    resource_type TEXT NOT NULL, -- tenant, user, app, message, etc.
    resource_id TEXT,         -- ID of affected resource
    changes JSONB DEFAULT '{}', -- before/after diff
    metadata JSONB DEFAULT '{}', -- IP, user-agent, etc.
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);

-- RLS
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
CREATE POLICY audit_log_isolation ON audit_logs
    USING (tenant_id = current_app_tenant());

-- Retention: partition-ready index for cleanup
CREATE INDEX idx_audit_logs_retention ON audit_logs(tenant_id, created_at);

-- +goose Down
DROP POLICY IF EXISTS audit_log_isolation ON audit_logs;
ALTER TABLE audit_logs DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS audit_logs;
