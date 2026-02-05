# 🐛 Job ID=4 Issue & Solution

## Problem Analysis

Job ID=4 was created **before** recipients storage was implemented.

**Evidence:**
```
[MailScheduler] Found 1 scheduled jobs ready to send
[MailScheduler] Processing job ID=4, scheduled_at=2026-01-20 10:40:00
```

But NO follow-up logs! This means:
- Job has `recipients_data = NULL`
- Scheduler fails silently at check

## Root Cause

Job created before this commit:
```go
// This code was TODO before
if scheduledAt != nil {
    recipientsJSON, err := json.Marshal(recipients)  // ← This wasn't implemented
    recipientsStr := string(recipientsJSON)
    job.RecipientsData = &recipientsStr               // ← Not stored
}
```

## Solution

### Option 1: Create NEW Scheduled Job ✅ (Recommended)

```bash
# Schedule for 2 minutes from now
SCHEDULED=$(date -v+2M '+%Y-%m-%d %H:%M:%S')

curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=1' \
  -F 'file=@docs/sample_mail_recipients.csv' \
  -F "scheduled_at=$SCHEDULED"
```

This NEW job will have recipients_data! ✅

### Option 2: Delete Old Job

```sql
-- Delete old job without recipients
DELETE FROM mail_jobs WHERE id = 4;
```

### Option 3: Manual Fix (Advanced)

```sql
-- Add recipients data manually (not recommended)
UPDATE mail_jobs
SET recipients_data = '[
  {
    "to": ["congminh352623@gmail.com"],
    "subject": "Welcome to TekNix",
    "variables": {
      "Name": "Minh",
      "Company": "TekNix"
    }
  }
]'::jsonb
WHERE id = 4;
```

## Improvements Made

### 1. Added Panic Recovery
**File:** `mail_scheduler.go`

```go
go func(j *entity.MailJob) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("[MailScheduler] PANIC in job %d: %v\n", j.ID, r)
            j.MarkAsFailed(fmt.Sprintf("panic: %v", r))
            s.mailJobRepo.Update(context.Background(), j)
        }
    }()
    s.sendScheduledJob(j)
}(job)
```

Now panics will be logged! 🔍

## Testing New Scheduled Job

### 1. Restart Server
```bash
# Ctrl+C to stop
make dev
```

### 2. Create NEW Scheduled Job
```bash
# Schedule 2 minutes from now
SCHEDULED=$(date -v+2M '+%Y-%m-%d %H:%M:%S')
echo "Scheduling for: $SCHEDULED"

curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=1' \
  -F 'file=@docs/sample_mail_recipients.csv' \
  -F "scheduled_at=$SCHEDULED"
```

**Expected Response:**
```json
{
  "code": 200,
  "data": {
    "job_id": 5,
    "message": "Batch emails scheduled for 2026-01-20 10:54:00",
    "stats": {
      "total": 3,
      "success": 0,
      "failed": 0
    }
  }
}
```

### 3. Wait and Watch Logs

After 2 minutes, you'll see:
```
[MailScheduler] Found 1 scheduled jobs ready to send
[MailScheduler] Processing job ID=5, scheduled_at=...
[MailScheduler] Sending emails for job ID=5
[MailScheduler] Job 5 completed: 3 success, 0 failed
```

### 4. Verify Job Status
```bash
curl -X GET 'http://localhost:8098/v1/api/mails/jobs/5'
```

**Expected:**
```json
{
  "code": 200,
  "data": {
    "job_id": 5,
    "message": "Job status: completed",
    "stats": {
      "total": 3,
      "success": 3,
      "failed": 0
    }
  }
}
```

## Database Check

To verify recipients data:
```sql
-- Check if job has recipients
SELECT 
    id,
    scheduled_at,
    recipients_data IS NOT NULL as has_data,
    jsonb_array_length(recipients_data::jsonb) as recipient_count
FROM mail_jobs
WHERE id IN (4, 5);
```

**Expected:**
```
id | scheduled_at | has_data | recipient_count
---|--------------|----------|----------------
 4 | 10:40:00     | false    | NULL           ← Old job (won't work)
 5 | 10:54:00     | true     | 3              ← New job (will work!)
```

## Summary

- ❌ Job ID=4: No recipients_data (created before feature)
- ✅ New jobs: Have recipients_data (will work!)
- ✅ Added panic recovery for better debugging
- ✅ Build successful

**Action:** Create a NEW scheduled job to test! 🎉
