# Quick Guide: Tạo Email Template và Test Mail Module

## 📋 Các Biến Từ File CSV

File `sample_mail_recipients.csv` có các biến sau:
- **Name** - Tên người nhận
- **Company** - Tên công ty
- **Email** - Email người nhận
- **MemberType** - Loại membership (Premium, Standard)
- **JoinDate** - Ngày tham gia
- **SupportEmail** - Email support
- **CompanyAddress** - Địa chỉ công ty

## 🚀 Cách 1: Tạo Email Template Qua API

### Step 1: Tạo Template
```bash
curl -X POST 'http://localhost:8098/v1/api/settings/email-templates' \
  -H 'Content-Type: application/json' \
  -d @docs/create_welcome_template.json
```

**Response sẽ trả về template_id (ví dụ: 1)**

### Step 2: Send Batch Emails với CSV
```bash
curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=1' \
  -F 'file=@docs/sample_mail_recipients.csv'
```

### Step 3: Check Job Status
```bash
# Dùng job_id từ response của step 2 (ví dụ: 123)
curl -X GET 'http://localhost:8098/v1/api/mails/jobs/123'
```

## 💾 Cách 2: Tạo Email Template Qua SQL

### Step 1: Run SQL Script
```bash
# Kết nối database và chạy
psql -h 172.16.12.146 -p 11507 -U postgres -d xp_db -f docs/insert_welcome_template.sql
```

### Step 2: Lấy Template ID
```sql
SELECT template_id FROM email_templates WHERE template_code = 'WELCOME_TEKNIX';
```

### Step 3: Send Emails (same as API method)
```bash
curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=<ID_FROM_STEP_2>' \
  -F 'file=@docs/sample_mail_recipients.csv'
```

## 📊 Template HTML Preview

Template sẽ render các biến như sau:

**Subject:**
```
Welcome to TekNix - Your Account is Ready!
```

**Email Body (HTML):**
- Header: "Welcome Minh!" (từ {{.Name}})
- Company: "Thank you for joining **TekNix**" (từ {{.Company}})
- Account Details Box:
  - Email: congminh352623@gmail.com
  - Member Type: Premium
  - Join Date: January 19 2026
- Support Email: support@teknix.dev
- Footer: Company address

## 🧪 Full Test Flow

```bash
# 1. Tạo template
curl -X POST 'http://localhost:8098/v1/api/settings/email-templates' \
  -H 'Content-Type: application/json' \
  -d @docs/create_welcome_template.json

# Response example:
# {
#   "code": 200,
#   "message": "success",
#   "data": {
#     "template_id": 1,
#     "template_name": "Welcome to TekNix",
#     ...
#   }
# }

# 2. Send batch emails
curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=1' \
  -F 'file=@docs/sample_mail_recipients.csv'

# Response example:
# {
#   "code": 200,
#   "message": "success",
#   "data": {
#     "job_id": 123,
#     "message": "Batch emails sent successfully",
#     "stats": {
#       "total": 3,
#       "success": 3,
#       "failed": 0
#     }
#   }
# }

# 3. Check job status
curl -X GET 'http://localhost:8098/v1/api/mails/jobs/123'

# Response example:
# {
#   "code": 200,
#   "message": "success",
#   "data": {
#     "job_id": 123,
#     "message": "Job status: completed",
#     "stats": {
#       "total": 3,
#       "success": 3,
#       "failed": 0
#     }
#   }
# }
```

## 📝 Mapping Variables

| CSV Column      | Template Variable    | Example Value                          |
|-----------------|---------------------|----------------------------------------|
| Name            | {{.Name}}           | Minh, John Doe, Jane Smith            |
| Company         | {{.Company}}        | TekNix, TekNix Corp                   |
| Email           | {{.Email}}          | congminh352623@gmail.com              |
| MemberType      | {{.MemberType}}     | Premium, Standard                     |
| JoinDate        | {{.JoinDate}}       | January 19 2026                       |
| SupportEmail    | {{.SupportEmail}}   | support@teknix.dev                    |
| CompanyAddress  | {{.CompanyAddress}} | 123 Tech Street Innovation City...    |

## ✅ Expected Results

Sau khi chạy, 3 emails sẽ được gửi:
1. **congminh352623@gmail.com** - Minh (Premium member)
2. **john@example.com** - John Doe (Standard member)
3. **jane@example.com** - Jane Smith (Premium member)

Mỗi email sẽ có:
- ✅ Subject personalized với tên company
- ✅ Welcome message với tên người nhận
- ✅ Account details với email, membership type, join date
- ✅ Support contact information
- ✅ Company address footer

## 🎯 Files Created

- ✅ `docs/create_welcome_template.json` - JSON body cho API
- ✅ `docs/insert_welcome_template.sql` - SQL INSERT statement
- ✅ `docs/sample_mail_recipients.csv` - Sample data với 3 recipients
- ✅ `docs/TEMPLATE_USAGE_GUIDE.md` - Hướng dẫn này

## 🔍 Troubleshooting

**Template không tìm thấy?**
```bash
# List all templates
curl -X GET 'http://localhost:8098/v1/api/settings/email-templates'
```

**Mail service không hoạt động?**
- Check external mail service đang chạy tại `http://localhost:9098`
- Check credentials trong `.env_dev`:
  - `MAIL_SERVICE_URL=http://localhost:9098`
  - `MAIL_SERVICE_USERNAME=admin`
  - `MAIL_SERVICE_PASSWORD=SHlifHEpF1Mh2ShsS9SG`

**Job failed?**
```bash
# Check job status để xem error message
curl -X GET 'http://localhost:8098/v1/api/mails/jobs/{job_id}'
```
