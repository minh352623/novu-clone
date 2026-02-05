-- +goose Up
-- 1. IAM & Core Tenancy Module

-- Pricing Plans
CREATE TABLE pricing_plans (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    monthly_credits BIGINT NOT NULL DEFAULT 0,
    price NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    currency TEXT DEFAULT 'USD',
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Seed Default Plan
INSERT INTO pricing_plans (name, slug, monthly_credits, price, is_default) 
VALUES ('Free', 'free', 30000, 0, TRUE);

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    pricing_plan_id UUID REFERENCES pricing_plans(id),
    plan_start_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_tenants_modtime BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

-- Global Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT,
    is_root_admin BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_users_modtime BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

-- Tenant Members & Roles
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    permissions JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (name,slug)
);

CREATE TABLE tenant_members (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (tenant_id, user_id)
);

-- +goose Down
DROP TABLE IF EXISTS tenant_members;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;
DROP TABLE IF EXISTS pricing_plans;
