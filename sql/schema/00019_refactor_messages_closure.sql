-- +goose Up
-- +goose StatementBegin
-- Refactor Messages table from Nested Set to Closure Table

-- 1. Drop Nested Set Columns and Function
DROP FUNCTION IF EXISTS add_message_node(UUID, UUID, UUID, UUID, TEXT, UUID, JSONB);

ALTER TABLE messages DROP COLUMN IF EXISTS lft;
ALTER TABLE messages DROP COLUMN IF EXISTS rgt;
ALTER TABLE messages DROP COLUMN IF EXISTS depth;

-- 2. Create Closure Table
CREATE TABLE message_closure (
    ancestor_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    descendant_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    depth INTEGER NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id)
);

-- 3. Indexes for performance
CREATE INDEX idx_message_closure_descendant ON message_closure(descendant_id);
CREATE INDEX idx_message_closure_ancestor ON message_closure(ancestor_id);

-- 4. Enable RLS on Closure Table (inherit from messages implicitly via joins, but explicit if needed)
ALTER TABLE message_closure ENABLE ROW LEVEL SECURITY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS message_closure;
-- +goose StatementEnd

-- Optional: If we want strict RLS on closure, we need tenant_id denormalized or rely on join
-- For now, we assume application logic handles tenant isolation via the main 'messages' table join.
