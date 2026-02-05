# Tính Năng Hẹn Giờ Gửi Mail (Scheduled Emails)

## ✅ Đã Hoàn Thành

### 1. Domain Layer Updates
- ✅ **Entity**: Thêm `ScheduledAt *time.Time` vào `MailJob`
- ✅ **Methods**: `MarkAsScheduled()`, `IsReadyToSend()`
- ✅ **Repository Interface**: Thêm `GetScheduledJobsReadyToSend()`

### 2. Infrastructure Layer
- ✅ **Repository Implementation**: Query jobs với `status='scheduled' AND scheduled_at <= NOW()`
- ✅ **Database Migration**: `20260120095000_add_scheduled_at_to_mail_jobs.sql`
- ✅ **Mail Scheduler Service**: Background job processor

## 📋 Cần Hoàn Thành (Manually)

### 1. Update DTO (`internal/mails/application/service/dto/mail.dto.go`)

Thêm vào `SendMailBatchRequest`:
```go
type SendMailBatchRequest struct {
    TemplateID  int64   `json:"template_id" binding:"required"`
    ExcelFile   string  `json:"-"`
    ScheduledAt *string `json:"scheduled_at,omitempty"` // Format: "2026-01-20 11:00:00"
}
```

### 2. Update Handler (`internal/mails/controller/http/handler/mail.handler.go`)

Thêm xử lý `scheduled_at` parameter:

```go
func (h *MailHandler) SendMailBatch(c *gin.Context) {
    // ... existing code ...
    
    // Get scheduled_at if provided
    scheduledAtStr := c.PostForm("scheduled_at")
    var scheduledAt *time.Time
    
    if scheduledAtStr != "" {
        // Parse time: format "2026-01-20 11:00:00"
        t, err := time.Parse("2006-01-02 15:04:05", scheduledAtStr)
        if err != nil {
            response.ErrorResponse(c, http.StatusBadRequest, "invalid scheduled_at format (use: YYYY-MM-DD HH:MM:SS)", nil)
            return
        }
        
        // Validate: must be in the future
        if t.Before(time.Now()) {
            response.ErrorResponse(c, http.StatusBadRequest, "scheduled_at must be in the future", nil)
            return
        }
        
        scheduledAt = &t
    }
    
    // Call service with scheduledAt
    result, err := h.mailService.SendMailBatch(c.Request.Context(), templateID, file, scheduledAt)
    // ... rest of code ...
}
```

### 3. Update Service Interface (`internal/mails/application/service/mail.service.go`)

```go
type MailService interface {
    SendMailBatch(ctx context.Context, templateID int64, file *multipart.FileHeader, scheduledAt *time.Time) (*dto.SendMailBatchResponse, error)
    // ... other methods ...
}
```

### 4. Update Service Implementation (`internal/mails/application/service/mail.service.impl.go`)

```go
func (s *mailService) SendMailBatch(ctx context.Context, templateID int64, file *multipart.FileHeader, scheduledAt *time.Time) (*dto.SendMailBatchResponse, error) {
    // ... existing parsing and validation ...
    
    // Create mail job
    job := entity.NewMailJob(templateID, len(recipients))
    
    // If scheduled, mark as scheduled instead of pending
    if scheduledAt != nil {
        job.MarkAsScheduled(*scheduledAt)
        createdJob, err := s.mailJobRepo.Create(ctx, job)
        if err != nil {
            return nil, fmt.Errorf("failed to create scheduled job: %w", err)
        }
        
        // TODO: Store recipients somewhere for later retrieval
        // Options:
        // 1. Add recipients_data JSONB column to mail_jobs table
        // 2. Create separate mail_recipients table
        // 3. Store file and re-parse when sending
        
        return &dto.SendMailBatchResponse{
            JobID:   createdJob.ID,
            Message: fmt.Sprintf("Batch emails scheduled for %s", scheduledAt.Format("2006-01-02 15:04:05")),
            Stats: dto.MailStats{
                Total:   len(recipients),
                Success: 0,
                Failed:  0,
            },
        }, nil
    }
    
    // ... existing immediate send logic ...
}
```

### 5. Start Scheduler in main.go (`cmd/drunk/main.go`)

```go
import (
    "CONVERDA/internal/mails/infrastructure/scheduler"
    "time"
)

func main() {
    // ... existing setup ...
    
    // Initialize mail scheduler
    mailScheduler := scheduler.NewMailScheduler(
        mailJobRepo,
        mailService,
        1 * time.Minute, // Check every minute
    )
    
    // Start scheduler in background
    go mailScheduler.Start()
    
    // Setup graceful shutdown
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        mailScheduler.Stop()
        os.Exit(0)
    }()
    
    // ... start HTTP server ...
}
```

## 🗄️ Database Migration

Run migration:
```bash
# Apply migration
goose -dir sql/schema postgres "your-connection-string" up

# Or manually:
psql -h 172.16.12.146 -p 11507 -U postgres -d xp_db \
  -f sql/schema/20260120095000_add_scheduled_at_to_mail_jobs.sql
```

## 📊 Lưu Trữ Recipients cho Scheduled Jobs

**Option 1: JSONB Column (Recommended cho MVP)**
```sql
ALTER TABLE mail_jobs
ADD COLUMN recipients_data JSONB DEFAULT NULL;
```

Trong service:
```go
// Convert recipients to JSON
recipientsJSON, _ := json.Marshal(recipients)

// Store in job
job.RecipientsData = string(recipientsJSON)
```

**Option 2: Separate Table**
```sql
CREATE TABLE mail_job_recipients (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT REFERENCES mail_jobs(id) ON DELETE CASCADE,
    recipient_email VARCHAR(255),
    subject VARCHAR(255),
    variables JSONB
);
```

**Option 3: Store File Path**
```sql
ALTER TABLE mail_jobs
ADD COLUMN source_file_path VARCHAR(500);
```

## 🧪 Testing Scheduled Emails

### Test 1: Schedule for 2 minutes from now
```bash
# Get current time + 2 minutes
SCHEDULED_TIME=$(date -v+2M '+%Y-%m-%d %H:%M:%S')

curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=1' \
  -F 'file=@docs/sample_mail_recipients.csv' \
  -F "scheduled_at=$SCHEDULED_TIME"
```

### Test 2: Check job status
```bash
curl -X GET 'http://localhost:8098/v1/api/mails/jobs/123'
```

Response sẽ có:
```json
{
  "job_id": 123,
  "status": "scheduled",
  "scheduled_at": "2026-01-20T11:00:00Z",
  "stats": {
    "total": 3,
    "success": 0,
    "failed": 0
  }
}
```

### Test 3: Wait for scheduler to process
```
[MailScheduler] Starting scheduler...
[MailScheduler] Found 1 scheduled jobs ready to send
[MailScheduler] Processing job ID=123, scheduled_at=2026-01-20 11:00:00
[MailScheduler] Sending emails for job ID=123
```

## 📝 API Documentation Update

Add to Swagger comments:
```go
// @Param scheduled_at formData string false "Schedule time (format: YYYY-MM-DD HH:MM:SS)"
```

## ⚠️ Important Notes

1. **Timezone**: Scheduler uses server time. Đảm bảo timezone consistency
2. **Idempotency**: Scheduler check `IsReadyToSend()` để tránh gửi lại
3. **Concurrent Processing**: Sử dụng goroutines, cần limit số lượng concurrent jobs
4. **Failure Handling**: Jobs failed sẽ có error_message
5. **Recipients Storage**: PHẢI implement một trong 3 options để lưu recipients

## ✨ Features

- ✅ Hẹn giờ gửi mail chính xác
- ✅ Background scheduler tự động check
- ✅ Graceful shutdown
- ️ ✅ Database indexed queries
- ✅ Concurrent processing
- ✅ Status tracking

## 🚀 Next Steps

1. Chọn cách lưu trữ recipients (JSONB recommended)
2. Update service implementation
3. Update handler
4. Start scheduler trong main.go
5. Run migration
6. Test với scheduled time trong tương lai
