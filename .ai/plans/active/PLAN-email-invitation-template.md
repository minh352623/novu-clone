# [Email Invitation Template]

Use the Notification Pipeline's template engine to send rich HTML invitation emails instead of hardcoded strings.

## User Review Required
> [!IMPORTANT]
> **Initialization Order Change**: `InitNotificationModule` will now be called BEFORE `InitIAMModule` in `router.go`.
> **Dependency Injection**: `InitIAMModule` signature will change to accept `EmailService`.

## Proposed Changes

### 1. `internal/iam`
#### [NEW] [notification_email_adapter.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/iam/infrastructure/email/notification_email_adapter.go)
- Implements `iam.EmailService` interface.
- Wraps `notification.NotificationService`.
- Adapts `SendInvitation` -> `SendRequest` with template `invitation`.

#### [MODIFY] [iam.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/initialize/iam/iam.go)
- Change signature: `func InitIAMModule(..., emailService service.EmailService)`
- Remove internal `SESEmailService` creation.
- Use passed `emailService`.

### 2. `internal/notification`
#### [MODIFY] [notification.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/initialize/notification/notification.go)
- Return `service.NotificationService`: `func InitNotificationModule(...) service.NotificationService`

### 3. `internal/initialize`
#### [MODIFY] [router.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/initialize/router.go)
- Reorder: Notification first -> IAM second.
- Wire: `notifService` -> `NewNotificationEmailAdapter` -> `InitIAMModule`.

### 4. `sql/schema`
#### [NEW] [00028_insert_invitation_template.sql](file:///Users/tekix/Documents/company/converda/converda-service/sql/schema/00028_insert_invitation_template.sql)
- Insert `templates` with code `invitation` (group: system).
- Insert `template_contents` (English) with HTML body.

## Verification Plan

### Automated Tests
- `go test ./internal/iam/infrastructure/email/...` (Test adapter)
- `go build ./...` (Verify wiring)

### Manual Verification
1. Run `make dev`.
2. Call `POST /v1/api/tenants/{id}/invitations` (invite a user).
3. Check `notifications` table for new record with status `sent` (or `pending` if provider fails).
4. Verify variables like `{{.tenant_name}}` are replaced correctly.
