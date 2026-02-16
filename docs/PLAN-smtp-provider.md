# PLAN: SMTP Provider Implementation

## Goal
Replace the mock SMTP implementation in `Dispatcher` with a real SMTP client that uses configurations stored in the `ProviderConfig` entity.

## Context
- **Module**: `internal/notification/infrastructure/provider`
- **Config**: Stored in `notification_provider_configs` table (JSONB field).
- **Library**: Use standard `net/smtp` or `github.com/jordan-wright/email` if already in project.

## Proposed Changes

### 1. Define SMTP Config Schema
The JSONB `config` field for SMTP should contain:
```json
{
  "host": "string",
  "port": "int",
  "username": "string",
  "password": "string",
  "from_email": "string",
  "from_name": "string",
  "encryption": "ssl|tls|none"
}
```

### 2. Update Dispatcher
- [ ] Create `smtp_provider.go` to encapsulate SMTP logic.
- [ ] Parse JSONB config into a struct.
- [ ] Implement `Send(to, subject, body)` logic.
- [ ] Handle auth (PlainAuth/LoginAuth).
- [ ] Support HTML and Plain Text (Multipart).

### 3. Integration
- [ ] Inject `SmtpProvider` into `Dispatcher`.
- [ ] Map `config.Type == "smtp"` to the new provider.

## Verification
- **Unit Test**: Stub the SMTP server using a mock or a local test server (e.g., MailHog).
- **Manual**: Configure a real SMTP (e.g., Gmail/SendGrid) in a sandbox environment and trigger a notification.

## Review Levels
- **Level 1 (Local)**: No impact on other modules beyond `notification`.
- **Level 2 (Proposal)**: This plan.
- **Level 3 (Lead Approval)**: Awaiting `APPROVED`.
