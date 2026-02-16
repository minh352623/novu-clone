-- +goose Up
-- 1. Seed Default Roles
INSERT INTO roles (id, name, slug, permissions) VALUES 
(generate_uuid_v7(), 'Tenant Admin', 'tenant_admin', '{
    "iam.tenants.read": true,
    "iam.tenants.update": true,
    "iam.members.read": true,
    "iam.members.create": true,
    "iam.members.delete": true,
    "iam.invitations.create": true,
    "iam.invitations.read": true,
    "iam.roles.read": true,
    "apps.create": true,
    "apps.read": true,
    "apps.update": true,
    "apps.delete": true,
    "apps.api_keys.manage": true,
    "metrics.read": true
}'),
(generate_uuid_v7(), 'Tenant Member', 'tenant_member', '{
    "iam.tenants.read": true,
    "iam.members.read": true,
    "apps.read": true,
    "metrics.read": true
}');

-- 2. Refactor Environments table to allow nullable api_key
-- We are moving towards app_api_keys table for security.
ALTER TABLE environments ALTER COLUMN api_key DROP NOT NULL;

-- +goose Down
-- Since we use generate_uuid_v7(), we delete by slug
DELETE FROM roles WHERE slug IN ('tenant_admin', 'tenant_member');

ALTER TABLE environments ALTER COLUMN api_key SET NOT NULL;
