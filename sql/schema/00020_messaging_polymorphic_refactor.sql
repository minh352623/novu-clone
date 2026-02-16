-- +goose Up
-- +goose StatementBegin

-- 1. Rename conversation_pools to threads
ALTER TABLE conversation_pools RENAME TO threads;

-- 2. Update Column: pool_id -> thread_id in related tables
-- Messages table
ALTER TABLE messages RENAME COLUMN pool_id TO thread_id;
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_pool_id_fkey;
ALTER TABLE messages ADD CONSTRAINT messages_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE;

-- Assignment Logs table
ALTER TABLE assignment_logs RENAME COLUMN pool_id TO thread_id;
ALTER TABLE assignment_logs DROP CONSTRAINT IF EXISTS assignment_logs_pool_id_fkey;
ALTER TABLE assignment_logs ADD CONSTRAINT assignment_logs_thread_id_fkey FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE;

-- 3. Add Hardening Columns to threads
ALTER TABLE threads ADD COLUMN IF NOT EXISTS reference_hash TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_threads_reference_hash ON threads(reference_hash);

-- 4. Create thread_participants table
CREATE TABLE IF NOT EXISTS thread_participants (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    thread_id UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    entity_type TEXT NOT NULL, -- 'user', 'subscriber'
    entity_id UUID NOT NULL,
    last_read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (thread_id, entity_type, entity_id)
);

-- 5. Data Migration: Move existing relations to thread_participants
-- Migrate Subscribers
INSERT INTO thread_participants (thread_id, entity_type, entity_id, created_at)
SELECT id, 'subscriber', subscriber_id, created_at FROM threads 
WHERE subscriber_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- Migrate Assigned Agents
INSERT INTO thread_participants (thread_id, entity_type, entity_id, created_at)
SELECT id, 'user', assigned_to_member_id, created_at FROM threads 
WHERE assigned_to_member_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- 6. Cleanup old columns from threads (Wait until participants are migrated)
ALTER TABLE threads DROP COLUMN IF EXISTS subscriber_id;
ALTER TABLE threads DROP COLUMN IF EXISTS assigned_to_member_id;

-- 7. Security: Enable RLS for new tables
ALTER TABLE threads ENABLE ROW LEVEL SECURITY;
ALTER TABLE thread_participants ENABLE ROW LEVEL SECURITY;

-- Drop old policies if any (conversation_pools policies might still exist under old name if renamed, but usually better to re-apply)
DROP POLICY IF EXISTS thread_isolation ON threads;
CREATE POLICY thread_isolation ON threads
    USING (environment_id IN (SELECT e.id FROM environments e JOIN apps a ON e.app_id = a.id WHERE a.tenant_id = current_app_tenant()));

DROP POLICY IF EXISTS participant_isolation ON thread_participants;
CREATE POLICY participant_isolation ON thread_participants
    USING (thread_id IN (SELECT t.id FROM threads t WHERE t.environment_id IN (SELECT e.id FROM environments e JOIN apps a ON e.app_id = a.id WHERE a.tenant_id = current_app_tenant())));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- 1. Restore columns to threads
ALTER TABLE threads ADD COLUMN subscriber_id UUID REFERENCES subscribers(id);
ALTER TABLE threads ADD COLUMN assigned_to_member_id UUID REFERENCES tenant_members(id);

-- 2. Reverse Data Migration (Subscribers)
UPDATE threads t
SET subscriber_id = tp.entity_id
FROM thread_participants tp
WHERE t.id = tp.thread_id AND tp.entity_type = 'subscriber';

-- 3. Reverse Data Migration (Agents)
UPDATE threads t
SET assigned_to_member_id = tp.entity_id
FROM thread_participants tp
WHERE t.id = tp.thread_id AND tp.entity_type = 'user';

-- 4. Revert Structure
DROP TABLE IF EXISTS thread_participants;
ALTER TABLE threads DROP COLUMN IF EXISTS reference_hash;

ALTER TABLE assignment_logs RENAME COLUMN thread_id TO pool_id;
ALTER TABLE messages RENAME COLUMN thread_id TO pool_id;

ALTER TABLE threads RENAME TO conversation_pools;

-- +goose StatementEnd
