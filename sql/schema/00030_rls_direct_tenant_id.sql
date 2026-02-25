-- +goose Up
-- Replace multi-JOIN RLS policies with direct tenant_id-based policies

-- 1. threads: replace old join-based policy with simpler one
DROP POLICY IF EXISTS thread_isolation ON threads;
CREATE POLICY thread_isolation ON threads
    USING (tenant_id = current_app_tenant());

-- 2. subscribers: enable RLS + add policy
ALTER TABLE subscribers ENABLE ROW LEVEL SECURITY;
CREATE POLICY subscriber_isolation ON subscribers
    USING (tenant_id = current_app_tenant());

-- 3. workflows: enable RLS + add policy
ALTER TABLE workflows ENABLE ROW LEVEL SECURITY;
CREATE POLICY workflow_isolation ON workflows
    USING (tenant_id = current_app_tenant());

-- 4. thread_participants: simplify policy using new tenant_id on threads
DROP POLICY IF EXISTS participant_isolation ON thread_participants;
CREATE POLICY participant_isolation ON thread_participants
    USING (thread_id IN (SELECT id FROM threads WHERE tenant_id = current_app_tenant()));

-- +goose Down
DROP POLICY IF EXISTS participant_isolation ON thread_participants;
DROP POLICY IF EXISTS workflow_isolation ON workflows;
DROP POLICY IF EXISTS subscriber_isolation ON subscribers;
DROP POLICY IF EXISTS thread_isolation ON threads;

ALTER TABLE workflows DISABLE ROW LEVEL SECURITY;
ALTER TABLE subscribers DISABLE ROW LEVEL SECURITY;

-- Restore original join-based policy for threads
CREATE POLICY thread_isolation ON threads
    USING (environment_id IN (
        SELECT e.id FROM environments e
        JOIN apps a ON e.app_id = a.id
        WHERE a.tenant_id = current_app_tenant()
    ));

-- Restore original join-based participant policy
CREATE POLICY participant_isolation ON thread_participants
    USING (thread_id IN (
        SELECT t.id FROM threads t
        WHERE t.environment_id IN (
            SELECT e.id FROM environments e
            JOIN apps a ON e.app_id = a.id
            WHERE a.tenant_id = current_app_tenant()
        )
    ));
