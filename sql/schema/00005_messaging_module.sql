-- +goose Up
-- 4. Messaging Module (Nested Set Model)

CREATE TABLE conversation_pools (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    subscriber_id UUID REFERENCES subscribers(id),
    status TEXT NOT NULL DEFAULT 'unassigned',
    assigned_to_member_id UUID REFERENCES tenant_members(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE assignment_logs (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    pool_id UUID NOT NULL REFERENCES conversation_pools(id) ON DELETE CASCADE,
    assigned_to_member_id UUID REFERENCES tenant_members(id),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    response_time_seconds INTEGER
);

-- PARTITIONED Messages Table
-- Note: 'messages' should NOT be partitioned if we rely on global foreign keys easily,
-- but for scale, we partition. Nested set queries within a partition are fine.
-- COMPROMISE: We partition by created_at.
CREATE TABLE messages (
    id UUID NOT NULL DEFAULT generate_uuid_v7(),
    tenant_id UUID NOT NULL,
    environment_id UUID NOT NULL,
    pool_id UUID NOT NULL,
    sender_type TEXT NOT NULL,
    sender_id UUID,
    content JSONB NOT NULL,
    -- Nested Set Columns
    parent_id UUID,
    lft INTEGER NOT NULL,
    rgt INTEGER NOT NULL,
    depth INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);

-- Indexes for Tree Traversal
CREATE INDEX idx_messages_pool_lft ON messages (pool_id, lft);
CREATE INDEX idx_messages_pool_rgt ON messages (pool_id, rgt);
CREATE INDEX idx_messages_parent ON messages (parent_id);

-- Stored Procedure to Insert Message (Nested Set Logic)
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION add_message_node(
    p_tenant_id UUID,
    p_env_id UUID,
    p_pool_id UUID,
    p_parent_id UUID,
    p_sender_type TEXT,
    p_sender_id UUID,
    p_content JSONB
) RETURNS UUID AS $$
DECLARE
    v_rgt INTEGER;
    v_lft INTEGER;
    v_depth INTEGER;
    v_new_id UUID;
BEGIN
    -- If root node (first message in pool)
    IF p_parent_id IS NULL THEN
        v_lft := 1;
        v_rgt := 2;
        v_depth := 0;
    ELSE
        -- Get parent info
        SELECT rgt, depth INTO v_rgt, v_depth
        FROM messages
        WHERE id = p_parent_id AND pool_id = p_pool_id
        LIMIT 1;

        IF NOT FOUND THEN
             RAISE EXCEPTION 'Parent message not found';
        END IF;

        -- Update existing nodes to make space
        UPDATE messages SET rgt = rgt + 2 WHERE pool_id = p_pool_id AND rgt >= v_rgt;
        UPDATE messages SET lft = lft + 2 WHERE pool_id = p_pool_id AND lft > v_rgt;

        v_lft := v_rgt;
        v_rgt := v_rgt + 1;
        v_depth := v_depth + 1;
    END IF;

    v_new_id := generate_uuid_v7();

    INSERT INTO messages (
        id, tenant_id, environment_id, pool_id, sender_type, sender_id, content,
        parent_id, lft, rgt, depth, created_at
    ) VALUES (
        v_new_id, p_tenant_id, p_env_id, p_pool_id, p_sender_type, p_sender_id, p_content,
        p_parent_id, v_lft, v_rgt, v_depth, NOW()
    );

    RETURN v_new_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS add_message_node;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS assignment_logs;
DROP TABLE IF EXISTS conversation_pools;
