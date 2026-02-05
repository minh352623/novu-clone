# 🚀 Quick Reference - Mail Module Testing

## 📝 Bước 1: Tạo Email Template

```bash
curl -X POST 'http://localhost:8098/v1/api/settings/email-templates' \
  -H 'Content-Type: application/json' \
  -d '{
  "template_name": "Welcome to TekNix",
  "template_code": "WELCOME_TEKNIX",
  "subject": "Welcome to {{.Company}} - Your Account is Ready!",
  "body_html": "<!DOCTYPE html><html><body><h1>Welcome {{.Name}}!</h1><p>Thank you for joining <strong>{{.Company}}</strong>.</p><p>Email: {{.Email}}<br/>Member Type: {{.MemberType}}<br/>Join Date: {{.JoinDate}}</p><p>Support: {{.SupportEmail}}</p><p>{{.CompanyAddress}}</p></body></html>",
  "category": "transactional",
  "status": 1
}'
```

**→ Lưu lại `template_id` từ response**

## 📧 Bước 2: Send Batch Emails

```bash
curl -X POST 'http://localhost:8098/v1/api/mails/send-batch' \
  -F 'template_id=1' \
  -F 'file=@docs/sample_mail_recipients.csv'
```

**→ Lưu lại `job_id` từ response**

## 📊 Bước 3: Check Status

```bash
curl -X GET 'http://localhost:8098/v1/api/mails/jobs/123'
```

## 🎯 One-Liner Test

```bash
# Copy-paste command này để test ngay:
curl -X POST 'http://localhost:8098/v1/api/settings/email-templates' -H 'Content-Type: application/json' -d @docs/create_welcome_template.json && echo "\n✅ Template created! Now run:\ncurl -X POST 'http://localhost:8098/v1/api/mails/send-batch' -F 'template_id=1' -F 'file=@docs/sample_mail_recipients.csv'"
```

## 📋 CSV Variables Available

```
Name, Company, Email, MemberType, JoinDate, SupportEmail, CompanyAddress
```

Use as: `{{.Name}}`, `{{.Company}}`, etc.

## ✅ Expected Result

3 emails sent to:
- congminh352623@gmail.com
- john@example.com  
- jane@example.com
