# PLAN: US-RA-02.4 — Webhook Retry Policy Config

## Problem
Webhook retry policy is **hardcoded** in [webhook_dispatcher.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/notification/application/service/impl/webhook_dispatcher.impl.go):

```go
const (
    maxRetries       = 3
    backoffBase      = 5 * time.Second
    backoffMultipler = 6 // 5s → 30s → 180s
)
```

Partners cannot configure retry behavior per webhook. The backlog item US-RA-02.4 requires these to be **configurable per webhook**.

## Review Level: L2 (AI Review)
- No breaking changes to existing APIs
- Additive columns with sensible defaults (backward compatible)

---

## Proposed Changes

### Phase 1 — Migration + Entity

#### [NEW] `sql/schema/00023_webhook_retry_config.sql`

Add 3 columns to `webhooks` table with defaults matching current behavior:

```sql
ALTER TABLE webhooks
  ADD COLUMN max_retries INT NOT NULL DEFAULT 3,
  ADD COLUMN retry_backoff_seconds INT NOT NULL DEFAULT 5,
  ADD COLUMN retry_timeout_seconds INT NOT NULL DEFAULT 10;
```

#### [MODIFY] Apps Webhook Entity + GORM Model + Mapper

| File | Change |
|------|--------|
| [webhook.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/apps/domain/model/entity/webhook.go) | Add `MaxRetries`, `RetryBackoffSeconds`, `RetryTimeoutSeconds` fields |
| [webhook.model.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/apps/infrastructure/persistence/model/webhook.model.go) | Add GORM columns with `default` tags |
| Webhook mapper | Map new fields bidirectionally |

#### [MODIFY] Webhook Controller DTOs

Add optional fields to Create/Update webhook request:
```go
MaxRetries           *int `json:"max_retries,omitempty"`           // default: 3
RetryBackoffSeconds  *int `json:"retry_backoff_seconds,omitempty"` // default: 5
RetryTimeoutSeconds  *int `json:"retry_timeout_seconds,omitempty"` // default: 10
```

---

### Phase 2 — Dispatcher Integration

#### [MODIFY] Notification Webhook Entity

| File | Change |
|------|--------|
| [webhook.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/notification/domain/entity/webhook.go) | Add `MaxRetries`, `RetryBackoffSeconds` to `Webhook` struct |

#### [MODIFY] [webhook_dispatcher.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/notification/application/service/impl/webhook_dispatcher.impl.go)

- Remove hardcoded `maxRetries`, `backoffBase` constants
- Read from the `Webhook` entity instead:
  ```go
  logEntry.MaxRetries = webhook.MaxRetries  // from config
  backoff := time.Duration(webhook.RetryBackoffSeconds) * time.Second
  ```
- Keep constants as **fallback defaults** if webhook config is zero

---

### Phase 3 — State Management

#### [MODIFY] `AI_STATE_MINH.md`
Update with completed task and next items.

---

## Verification Plan

### Automated Tests
- `go build ./...`
- `go test -race ./internal/notification/...` — update existing test to use configurable retries
- `go test -race ./internal/apps/...`

### Manual Verification
- Apply migration `make upse`
- Create webhook with custom `max_retries: 5, retry_backoff_seconds: 10`
- Verify dispatch uses configured values instead of defaults

---

## Effort Estimate
~1–2 hours total (small task, builds on existing retry engine)
