-- +goose Up
-- 9. Notification Module (Enhanced / CMS-Ready)

-- Enums
CREATE TYPE notification_channel AS ENUM ('email', 'push', 'sms', 'in_app');
CREATE TYPE notification_job_status AS ENUM ('pending', 'scheduled', 'processing', 'completed', 'failed', 'cancelled');
CREATE TYPE notification_send_status AS ENUM ('pending', 'sent', 'delivered', 'failed', 'read');

-- [NEW] Notification Groups (Topics / Categories)
CREATE TABLE notification_groups (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    key TEXT NOT NULL,
    description TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(environment_id, key)
);

-- [NEW] Notification Layouts (Common Wrappers)
CREATE TABLE notification_layouts (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    content_html TEXT NOT NULL,
    variables_schema JSONB DEFAULT '{}',
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Notification Templates (CMS Structure: Logic + Metadata)
CREATE TABLE notification_templates (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    
    group_id UUID REFERENCES notification_groups(id) ON DELETE SET NULL,
    layout_id UUID REFERENCES notification_layouts(id) ON DELETE SET NULL,
    
    channel notification_channel NOT NULL,
    template_code TEXT NOT NULL,
    template_name TEXT NOT NULL,
    description TEXT,
    
    active_version INTEGER DEFAULT 1,
    variables_schema JSONB DEFAULT '{}',
    
    status TEXT NOT NULL DEFAULT 'active',
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (environment_id, template_code)
);

CREATE TRIGGER update_notification_templates_modtime BEFORE UPDATE ON notification_templates FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

-- [NEW] Template Content Translations & Versions
CREATE TABLE notification_template_contents (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    template_id UUID NOT NULL REFERENCES notification_templates(id) ON DELETE CASCADE,
    
    version INTEGER NOT NULL,
    language_code VARCHAR(10) NOT NULL DEFAULT 'en',
    
    subject TEXT,
    body_text TEXT,
    body_html TEXT,
    body_push JSONB DEFAULT '{}',
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(template_id, version, language_code)
);

-- Notification Jobs (Batches)
CREATE TABLE notification_jobs (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    environment_id UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    
    channel notification_channel NOT NULL,
    template_id UUID REFERENCES notification_templates(id) ON DELETE SET NULL,
    
    status notification_job_status NOT NULL DEFAULT 'pending',
    
    total_count INTEGER NOT NULL DEFAULT 0,
    success_count INTEGER NOT NULL DEFAULT 0,
    failed_count INTEGER NOT NULL DEFAULT 0,
    
    scheduled_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    metadata JSONB DEFAULT '{}',
    recipients_data JSONB, 
    
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_notification_jobs_modtime BEFORE UPDATE ON notification_jobs FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

-- Notification Logs (Individual Sends)
CREATE TABLE notification_logs (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    job_id UUID, 
    environment_id UUID NOT NULL, 
    channel notification_channel NOT NULL,
    
    subscriber_id UUID REFERENCES subscribers(id) ON DELETE SET NULL,
    recipient TEXT NOT NULL,            
    recipient_type TEXT NOT NULL,       
    
    status notification_send_status NOT NULL DEFAULT 'pending',
    error_message TEXT,
    attempt_number INTEGER NOT NULL DEFAULT 0,
    attempt_history JSONB NOT NULL DEFAULT '[]',
    
    recipient_data JSONB DEFAULT '{}',
    
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    read_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Notification Tracking (Opens/Clicks)
CREATE TABLE notification_trackings (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    tracking_token TEXT UNIQUE NOT NULL,
    tracking_token TEXT UNIQUE NOT NULL,
    
    opened_at TIMESTAMP WITH TIME ZONE,
    open_count INTEGER NOT NULL DEFAULT 0,
    first_opened_ip TEXT,
    first_opened_user_agent TEXT,
    
    click_count INTEGER NOT NULL DEFAULT 0,
    last_clicked_at TIMESTAMP WITH TIME ZONE,
    click_data JSONB DEFAULT '[]',
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Subscriber Preferences (Updated to use Groups)
CREATE TABLE subscriber_notification_preferences (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    subscriber_id UUID NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
    
    channel notification_channel,
    group_id UUID REFERENCES notification_groups(id) ON DELETE CASCADE,
    
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    settings JSONB DEFAULT '{}', 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TRIGGER update_subscriber_prefs_modtime BEFORE UPDATE ON subscriber_notification_preferences FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

-- Subscriber Devices (for Push)
CREATE TABLE subscriber_devices (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    subscriber_id UUID NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
    device_token TEXT NOT NULL,
    platform TEXT NOT NULL,
    device_name TEXT,
    device_model TEXT,
    os_version TEXT,
    app_version TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(device_token)
);

CREATE TRIGGER update_subscriber_devices_modtime BEFORE UPDATE ON subscriber_devices FOR EACH ROW EXECUTE PROCEDURE update_modified_column();

-- Enable RLS
ALTER TABLE notification_groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_layouts ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriber_devices ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscriber_notification_preferences ENABLE ROW LEVEL SECURITY;

-- Simple RLS Policies
CREATE POLICY template_isolation ON notification_templates USING (tenant_id = current_app_tenant());
CREATE POLICY job_isolation ON notification_jobs USING (tenant_id = current_app_tenant());

-- +goose Down
DROP POLICY IF EXISTS job_isolation ON notification_jobs;
DROP POLICY IF EXISTS template_isolation ON notification_templates;

ALTER TABLE subscriber_notification_preferences DISABLE ROW LEVEL SECURITY;
ALTER TABLE subscriber_devices DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_jobs DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_templates DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_layouts DISABLE ROW LEVEL SECURITY;
ALTER TABLE notification_groups DISABLE ROW LEVEL SECURITY;

DROP TABLE IF EXISTS subscriber_devices;
DROP TABLE IF EXISTS subscriber_notification_preferences;
DROP TABLE IF EXISTS notification_trackings;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS notification_jobs;
DROP TABLE IF EXISTS notification_template_contents;
DROP TABLE IF EXISTS notification_templates;
DROP TABLE IF EXISTS notification_layouts;
DROP TABLE IF EXISTS notification_groups;

DROP TYPE IF EXISTS notification_send_status;
DROP TYPE IF EXISTS notification_job_status;
DROP TYPE IF EXISTS notification_channel;
