# PLAN — SLA Worker Verification + SLA Config + Webhook Retry Engine

> Created: **2026-02-17** | Author: Minh (AI Senior Backend Developer)

---

## Phase 0 — Context Check

| Component | Current Status | Gap |
|---|---|---|
| SLA Worker | ✅ Runs every 1m, 2 unit tests | Missing: edge-case tests, error resilience, pagination, reset-on-resolve |
| SLA Config | ✅ Fully implemented (entity/GORM/controller/DTO) | Missing: API-level verification (E2E), docs |
| Webhook Dispatch | ✅ Core dispatch + HMAC signing + logging | Missing: retry with backoff, max attempts, dead letter |

---

## Phase 1 — SLA Worker Hardening

### Scope
Harden the existing `sla_worker.go` with additional test coverage and reliability improvements.

### Files to Modify

#### [MODIFY] `internal/messaging/application/worker/sla_worker.go`
- Add pagination loop (current `Limit: 500` misses threads beyond 500)
- Clear `IsOverdue` when thread is resolved (currently only marks, never unmarks)
- Add structured logging via `global.Logger` instead of `fmt.Printf` remnants

#### [MODIFY] `internal/messaging/application/worker/sla_worker_test.go`
Add test cases:
- Thread **within** SLA threshold → should NOT mark as overdue
- Thread with **nil environment** → should use default 900s
- Thread with **Env SLA = 0, App SLA > 0** → should fallback to App level
- Empty thread list → should not panic
- Concurrent `checkSLA` → no race condition (use `-race` flag)

### Review Level
- **Cấp 1 (Local):** Changes only affect `messaging/application/worker`.
- **Breaking Changes:** None.

### State Update
After completion, update `AI_STATE_MINH.md`:
- `[x] Verify background worker reliability and status transitions`

---

## Phase 2 — SLA Configuration Verification

### Scope
SLA Config is **already fully implemented**. This phase verifies correctness end-to-end.

### Verification Checklist
1. `POST /apps` with `sla_threshold_seconds` → persists to DB ✅
2. `PUT /apps/:id` with `sla_threshold_seconds` → updates App threshold ✅
3. `POST /apps/:id/environments` with `sla_threshold_seconds` → persists ✅
4. SLA Worker reads `env.SLAThresholdSeconds` → falls back to `app.SLAThresholdSeconds` → default 900 ✅
5. Dashboard APIs use `getSLAThreshold()` correctly ✅

### Files to Verify (Read-Only)
- `internal/apps/controller/app.controller.go` — CreateApp, UpdateApp, CreateEnvironment
- `internal/apps/application/service/impl/app.service.impl.go` — SLA param handling
- `internal/apps/application/service/impl/environment.service.impl.go` — SLA param handling
- `internal/apps/infrastructure/persistence/mapper/apps.mapper.go` — SLA field mapping

### Review Level
- **Cấp 1 (Local):** No code changes, only verification.

### State Update
After completion, update `AI_STATE_MINH.md`:
- `[x] Implement SLA Configuration (fallback App/Env level)` → mark as **VERIFIED**

---

## Phase 3 — Webhook Retry Engine

### Scope
Add retry logic with exponential backoff to the existing `webhook_dispatcher.impl.go`.

### Design

```
Dispatch → triggerWebhook → if FAILED:
  ├── attempt < maxRetries?
  │     YES → schedule retry (exponential backoff: 5s, 30s, 5m)
  │     NO  → mark as "dead_letter" in webhook_logs
  └── update webhook_log with attempt count + next_retry_at
```

**Config (hardcoded constants, configurable later):**
- `MaxRetries = 3`
- `BackoffBase = 5 * time.Second`
- `BackoffMultiplier = 6` (5s → 30s → 180s)

### Files to Modify

#### [MODIFY] `internal/notification/domain/entity/webhook.go`
- Add `RetryCount`, `MaxRetries`, `NextRetryAt` fields to `WebhookLog`

#### [NEW] `sql/schema/00022_webhook_retry_columns.sql`
```sql
ALTER TABLE webhook_logs ADD COLUMN IF NOT EXISTS retry_count INTEGER DEFAULT 0;
ALTER TABLE webhook_logs ADD COLUMN IF NOT EXISTS max_retries INTEGER DEFAULT 3;
ALTER TABLE webhook_logs ADD COLUMN IF NOT EXISTS next_retry_at TIMESTAMP WITH TIME ZONE;
```

#### [MODIFY] `internal/notification/infrastructure/persistence/model/webhook.model.go`
- Add `RetryCount`, `MaxRetries`, `NextRetryAt` to GORM model

#### [MODIFY] `internal/notification/application/service/impl/webhook_dispatcher.impl.go`
- Add `retryWebhook()` method with exponential backoff
- Replace `fmt.Printf` with `global.Logger`
- Add response body reader with size limit (max 1KB)
- On failure: schedule retry or mark dead letter

#### [NEW] `internal/notification/application/worker/webhook_retry_worker.go`
- Background worker (runs every 30s)
- Queries `webhook_logs WHERE status = 'failed' AND retry_count < max_retries AND next_retry_at <= NOW()`
- Re-dispatches failed webhooks

#### [MODIFY] `internal/notification/domain/repository/webhook.repository.go`
- Add `GetPendingRetries(ctx, limit int)` to `WebhookLogRepository`

#### [MODIFY] `internal/notification/infrastructure/persistence/repository/webhook.repository.go`
- Implement `GetPendingRetries`

### Review Level
- **Cấp 1 (Local):** Changes affect `notification` module only.
- **Cấp 2 (Proposal):** New background worker needs initialization in `main.go` or `initialize/`.
- **Breaking Changes:** None. `webhook_logs` schema is additive only.

### State Update
After completion, update `AI_STATE_MINH.md`:
- `[x] Webhook dispatch engine + retry logic`

---

## Verification Plan

### Automated Tests
```bash
# Phase 1
go test -v -race ./internal/messaging/application/worker/...

# Phase 3
go test -v ./internal/notification/application/service/impl/...

# Full build
go build ./...
```

### Manual Verification
- Trigger webhook dispatch with unreachable URL → verify retry attempts in `webhook_logs`
- Confirm SLA configuration persists via API calls
- Confirm SLA Worker pagination works with >500 threads

---

## Execution Order

| Order | Phase | Est. Time | Dependencies |
|---|---|---|---|
| 1 | Phase 1: SLA Worker Hardening | ~30 min | None |
| 2 | Phase 2: SLA Config Verification | ~10 min | Phase 1 |
| 3 | Phase 3: Webhook Retry Engine | ~60 min | None |
