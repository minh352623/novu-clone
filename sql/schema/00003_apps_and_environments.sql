-- +goose Up
-- 2. Apps & Environments Module (Refactored)

CREATE TABLE apps (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_apps_modtime BEFORE UPDATE ON apps FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

-- System Defined Environments (Global Lookup)
CREATE TABLE system_environments (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed Default Environments
INSERT INTO system_environments (code, name, description) VALUES
('development', 'Development', 'Sandbox environment for testing'),
('production', 'Production', 'Live environment for real users'),
('staging', 'Staging', 'Pre-production environment');

CREATE TABLE environments (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    environment_code TEXT NOT NULL REFERENCES system_environments(code),
    api_key TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (app_id, environment_code)
);

CREATE TRIGGER update_environments_modtime BEFORE UPDATE ON environments FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

-- Webhooks: Dedicated table for event subscriptions
CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    secret TEXT NOT NULL,
    description TEXT,
    events JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_webhooks_modtime BEFORE UPDATE ON webhooks FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

CREATE TABLE providers (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    app_id UUID NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    provider_type TEXT NOT NULL,
    provider_name TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    configuration JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS providers;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS environments;
DROP TABLE IF EXISTS system_environments;
DROP TABLE IF EXISTS apps;
