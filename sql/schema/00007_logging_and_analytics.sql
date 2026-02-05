-- +goose Up
-- 6. Logging & Analytics

CREATE TABLE workflow_logs (
    id UUID NOT NULL DEFAULT generate_uuid_v7(),
    environment_id UUID NOT NULL,
    workflow_id UUID,
    step_id UUID,
    status TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);

CREATE TABLE system_logs (
    id UUID NOT NULL DEFAULT generate_uuid_v7(),
    level TEXT NOT NULL,
    component TEXT NOT NULL,
    message TEXT,
    meta JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id)
);

-- View: Average Response Time per Member
CREATE VIEW view_member_response_stats AS
SELECT
    m.id AS member_id,
    u.full_name,
    COUNT(al.id) AS total_assignments,
    AVG(al.response_time_seconds) AS avg_response_time_sec
FROM tenant_members m
JOIN users u ON m.user_id = u.id
JOIN assignment_logs al ON al.assigned_to_member_id = m.id
WHERE al.response_time_seconds IS NOT NULL
GROUP BY m.id, u.full_name;

-- +goose Down
DROP VIEW IF EXISTS view_member_response_stats;
DROP TABLE IF EXISTS system_logs;
DROP TABLE IF EXISTS workflow_logs;
