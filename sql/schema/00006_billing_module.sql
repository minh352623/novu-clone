-- +goose Up
-- 5. Billing Configuration Module

CREATE TABLE tenant_credits (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    balance NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    currency TEXT DEFAULT 'USD',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE credit_ledger (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    amount NUMERIC(10, 2) NOT NULL,
    description TEXT NOT NULL,
    reference_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Transaction Trigger to update balance
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_tenant_balance()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO tenant_credits (tenant_id, balance)
    VALUES (NEW.tenant_id, NEW.amount)
    ON CONFLICT (tenant_id)
    DO UPDATE SET
        balance = tenant_credits.balance + NEW.amount,
        updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trigger_update_balance
    AFTER INSERT ON credit_ledger
    FOR EACH ROW EXECUTE PROCEDURE update_tenant_balance();

-- +goose Down
DROP TRIGGER IF EXISTS trigger_update_balance ON credit_ledger;
DROP FUNCTION IF EXISTS update_tenant_balance;
DROP TABLE IF EXISTS credit_ledger;
DROP TABLE IF EXISTS tenant_credits;
