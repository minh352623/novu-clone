# Quick Reference: Scheduled Emails

## 🚀 Cách Sử Dụng

### Gửi Email Ngay Lập Tức (Hiện Tại)
```bash
curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=1' \
  -F 'file=@docs/sample_mail_recipients.csv'
```

### Hẹn Giờ Gửi Email
```bash
curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=1' \
  -F 'file=@docs/sample_mail_recipients.csv' \
  -F 'scheduled_at=2026-01-20 11:00:00'
```

## 📅 Format Thời Gian

**Format:** `YYYY-MM-DD HH:MM:SS`

**Examples:**
- `2026-01-20 11:00:00` - 11:00 AM, Jan 20, 2026
- `2026-01-25 14:30:00` - 2:30 PM, Jan 25, 2026
- `2026-02-01 09:15:00` - 9:15 AM, Feb 1, 2026

## ⏰ Tạo Scheduled Time Tự Động

### Hẹn giờ +2 phút
```bash
SCHEDULED=$(date -v+2M '+%Y-%m-%d %H:%M:%S')
curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=1' \
  -F 'file=@docs/sample_mail_recipients.csv' \
  -F "scheduled_at=$SCHEDULED"
```

### Hẹn giờ +1 giờ
```bash
SCHEDULED=$(date -v+1H '+%Y-%m-%d %H:%M:%S')
curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=1' \
  -F 'file=@docs/sample_mail_recipients.csv' \
  -F "scheduled_at=$SCHEDULED"
```

### Hẹn giờ +1 ngày
```bash
SCHEDULED=$(date -v+1d '+%Y-%m-%d %H:%M:%S')
curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=1' \
  -F 'file=@docs/sample_mail_recipients.csv' \
  -F "scheduled_at=$SCHEDULED"
```

## 📊 Response Examples

### Immediate Send
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "job_id": 123,
    "message": "Batch emails sent successfully",
    "stats": {
      "total": 3,
      "success": 3,
      "failed": 0
    }
  }
}
```

### Scheduled Send
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "job_id": 124,
    "message": "Batch emails scheduled for 2026-01-20 11:00:00",
    "stats": {
      "total": 3,
      "success": 0,
      "failed": 0
    }
  }
}
```

## 🔍 Check Job Status

```bash
curl -X GET 'http://localhost:8098/v1/api/mails/jobs/124'
```

### Status Values
- `scheduled` - Đang chờ thời gian gửi
- `pending` - Sẵn sàng gửi ngay
- `processing` - Đang gửi
- `completed` - Đã gửi xong
- `failed` - Gửi thất bại

## ⚙️ Scheduler Settings

Scheduler check mỗi **1 phút** để tìm jobs cần gửi.

Có thể thay đổi interval trong code:
```go
mailScheduler := scheduler.NewMailScheduler(
    mailJobRepo,
    mailService,
    30 * time.Second, // Check mỗi 30 giây
)
```

## ✅ Best Practices

1. **Validate Time**: Luôn đảm bảo `scheduled_at` ở tương lai
2. **Timezone**: Sử dụng server timezone (default: UTC)
3. **Buffer Time**: Đặt schedule ít nhất 2-3 phút trước để tránh miss
4. **File Size**: Limit file size cho scheduled jobs (recommend < 1000 recipients)

## 📋 Example Use Cases

### Marketing Campaign
```bash
# Schedule cho 9AM ngày mai
TOMORROW_9AM=$(date -v+1d '+%Y-%m-%d 09:00:00')
curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=2' \
  -F 'file=@campaign_list.csv' \
  -F "scheduled_at=$TOMORROW_9AM"
```

### Weekly Newsletter
```bash
# Schedule cho Thứ 2 tuần sau 10AM
NEXT_MONDAY=$(date -v+mon '+%Y-%m-%d 10:00:00')
curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=3' \
  -F 'file=@subscribers.csv' \
  -F "scheduled_at=$NEXT_MONDAY"
```

### Birthday Emails
```bash
# Schedule cho đúng ngày sinh nhật
curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=4' \
  -F 'file=@birthday_today.csv' \
  -F 'scheduled_at=2026-03-15 08:00:00'
```
