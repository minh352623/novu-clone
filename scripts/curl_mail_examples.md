# Mail Module - cURL Examples

Base URL: `http://localhost:8080/v1/api`

---

## 1. Get All Mail Jobs (Danh sách tất cả jobs)

```bash
curl --location --request GET 'http://localhost:8080/v1/api/mails/jobs?limit=20' \
--header 'accept: application/json' \
--header 'content-type: application/json'
```

### With status filter:
```bash
curl --location --request GET 'http://localhost:8080/v1/api/mails/jobs?limit=20&status=completed' \
--header 'accept: application/json' \
--header 'content-type: application/json'
```

### With cursor pagination:
```bash
curl --location --request GET 'http://localhost:8080/v1/api/mails/jobs?limit=20&cursor=YOUR_CURSOR_HERE' \
--header 'accept: application/json' \
--header 'content-type: application/json'
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "jobs": [
      {
        "id": 1,
        "template_id": 1,
        "status": "completed",
        "total_count": 100,
        "success_count": 70,
        "failed_count": 30,
        "success_rate": "70%",
        "created_at": "2026-01-27T10:00:00Z",
        "updated_at": "2026-01-27T10:05:00Z"
      }
    ],
    "next_cursor": "...",
    "has_more": true,
    "total_count": 50
  }
}
```

---

## 2. Get Overall Statistics (Thống kê tổng quan)

```bash
curl --location --request GET 'http://localhost:8080/v1/api/mails/stats' \
--header 'accept: application/json' \
--header 'content-type: application/json'
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total_jobs": 50,
    "total_sent": 5000,
    "total_success": 4750,
    "total_failed": 250,
    "avg_delivery_rate": "95%"
  }
}
```

---

## 3. Get Mail Job Status (Chi tiết 1 job)

```bash
curl --location --request GET 'http://localhost:8080/v1/api/mails/jobs/1' \
--header 'accept: application/json' \
--header 'content-type: application/json'
```

---

## 4. Get Mail Job Statistics (Thống kê chi tiết 1 job)

```bash
curl --location --request GET 'http://localhost:8080/v1/api/mails/jobs/1/stats' \
--header 'accept: application/json' \
--header 'content-type: application/json'
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "job_id": 1,
    "total_count": 100,
    "success_count": 70,
    "failed_count": 30,
    "retry_count": 5,
    "success_rate": "70%",
    "tracking_stats": {
      "total_sent": 70,
      "total_opened": 35,
      "open_rate": "50%"
    }
  }
}
```

---

## 5. Get Mail Job Send Logs (Danh sách logs gửi mail)

### All logs:
```bash
curl --location --request GET 'http://localhost:8080/v1/api/mails/jobs/1/logs?limit=20' \
--header 'accept: application/json' \
--header 'content-type: application/json'
```

### Filter by success:
```bash
curl --location --request GET 'http://localhost:8080/v1/api/mails/jobs/1/logs?limit=20&status=success' \
--header 'accept: application/json' \
--header 'content-type: application/json'
```

### Filter by failed:
```bash
curl --location --request GET 'http://localhost:8080/v1/api/mails/jobs/1/logs?limit=20&status=failed' \
--header 'accept: application/json' \
--header 'content-type: application/json'
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "logs": [
      {
        "id": 1,
        "recipient_email": "user@example.com",
        "status": "success",
        "attempt_number": 1,
        "attempt_history": [
          {
            "attempt": 1,
            "status": "success",
            "timestamp": "2026-01-27T10:05:00Z"
          }
        ],
        "sent_at": "2026-01-27T10:05:00Z",
        "created_at": "2026-01-27T10:00:00Z"
      },
      {
        "id": 2,
        "recipient_email": "failed@example.com",
        "status": "failed",
        "error_message": "SMTP connection refused",
        "attempt_number": 2,
        "attempt_history": [
          {
            "attempt": 1,
            "status": "failed",
            "error": "SMTP timeout",
            "timestamp": "2026-01-27T10:05:00Z"
          },
          {
            "attempt": 2,
            "status": "failed",
            "error": "SMTP connection refused",
            "timestamp": "2026-01-27T10:10:00Z"
          }
        ],
        "created_at": "2026-01-27T10:00:00Z"
      }
    ],
    "next_cursor": "...",
    "has_more": false
  }
}
```

---

## 6. Get Mail Job Tracking Stats (Thống kê tracking)

```bash
curl --location --request GET 'http://localhost:8080/v1/api/mails/jobs/1/tracking' \
--header 'accept: application/json' \
--header 'content-type: application/json'
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "totalSent": 70,
    "totalOpened": 35,
    "openRate": "50%"
  }
}
```

---

## 7. Retry Failed Emails (Gửi lại email thất bại)

```bash
curl --location --request POST 'http://localhost:8080/v1/api/mails/jobs/1/retry' \
--header 'accept: application/json' \
--header 'content-type: application/json'
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "job_id": 1,
    "message": "Retried 30 failed emails",
    "total_retried": 30,
    "success": 25,
    "failed": 5
  }
}
```

---

## 8. Send Mail Batch (Gửi mail hàng loạt)

### Gửi ngay:
```bash
curl --location --request POST 'http://localhost:8080/v1/api/mails/send-batch' \
--form 'template_id=1' \
--form 'file=@/path/to/recipients.xlsx'
```

### Lên lịch gửi:
```bash
curl --location --request POST 'http://localhost:8080/v1/api/mails/send-batch' \
--form 'template_id=1' \
--form 'file=@/path/to/recipients.xlsx' \
--form 'scheduled_at=2026-01-28 10:00:00'
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "job_id": 123,
    "message": "Batch emails sent successfully",
    "stats": {
      "total": 100,
      "success": 98,
      "failed": 2
    }
  }
}
```

---

## 9. Track Email Open (Tracking Pixel - Public)

```bash
# This returns a 1x1 transparent PNG image
curl --location --request GET 'http://localhost:8080/v1/api/track/YOUR_TRACKING_TOKEN' \
--output pixel.png
```

---

## Sample Recipients Excel/CSV Format

| to | Name | Company | Email | MemberType | JoinDate | SupportEmail | CompanyAddress |
|----|------|---------|-------|------------|----------|--------------|----------------|
| user1@example.com | John | TekNix | user1@example.com | Premium | January 27 2026 | support@teknix.dev | 123 Tech Street |
| user2@example.com | Jane | TekNix | user2@example.com | Basic | January 27 2026 | support@teknix.dev | 123 Tech Street |
