# PLAN: Rate Limiting per Partner

## Problem
Partners (Apps) call Converda APIs using API Keys. Currently there is **no throttling** — any partner can flood the system with unlimited requests. We need:
- **Per-environment rate limits** (max requests/minute, max messages/day)
- **Configurable via API** (each partner can have different limits)
- **Enforcement at middleware level** (before business logic runs)
- **Clear error responses** when limits are exceeded (HTTP 429)

### Existing Infrastructure
| Component | State |
|-----------|-------|
| `APIKeyMiddleware` | ✅ Validates key, sets appID/envID in context |
| `Environment` entity | Has `SLAThresholdSeconds` — config-per-env pattern established |
| Redis | ❌ Not available (no dependency exists) |
| Rate limiting | ❌ Nothing exists |

### Design Decision: In-Memory vs Redis
Since no Redis exists in the project, we use an **in-memory sliding window counter** with `sync.Map`:
- ✅ Zero external dependencies
- ✅ Fast (~50ns per check)
- ⚠️ Not shared across multiple instances (acceptable for current scale)
- ⚠️ Resets on restart (acceptable — rate limits are short windows)

> [!NOTE]
> When the project scales to multiple instances, migrate to Redis-based counters. The `RateLimiter` interface makes this a drop-in replacement.

---

## Review Level: L2 (Proposal)
- Adds new middleware + config fields to existing entity
- New files follow established patterns
- No breaking changes

---

## Proposed Changes

### Component 1: Database Migration

#### [NEW] [00026_rate_limit_config.sql](file:///Users/tekix/Documents/company/converda/converda-service/sql/schema/00026_rate_limit_config.sql)

```sql
-- Add rate limit config to environments table
ALTER TABLE environments ADD COLUMN rate_limit_rpm INTEGER NOT NULL DEFAULT 0;  -- 0 = unlimited
ALTER TABLE environments ADD COLUMN rate_limit_daily INTEGER NOT NULL DEFAULT 0; -- 0 = unlimited
```

---

### Component 2: Domain Layer

#### [MODIFY] [apps.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/apps/domain/model/entity/apps.go)

Add to `Environment` struct:
```go
RateLimitRPM   int `json:"rate_limit_rpm"`   // max requests per minute, 0 = unlimited
RateLimitDaily int `json:"rate_limit_daily"` // max messages per day, 0 = unlimited
```

#### [MODIFY] GORM model for Environment

Add corresponding columns to the GORM model.

---

### Component 3: Rate Limiter Core (`pkg/ratelimit`)

#### [NEW] [ratelimit.go](file:///Users/tekix/Documents/company/converda/converda-service/pkg/ratelimit/ratelimit.go)

```go
// RateLimiter checks and records requests against configured limits.
type RateLimiter interface {
    Allow(key string, limit int, window time.Duration) bool
}
```

#### [NEW] [memory.go](file:///Users/tekix/Documents/company/converda/converda-service/pkg/ratelimit/memory.go)

In-memory sliding window implementation using `sync.Map`:
- Stores timestamped counters per key
- Automatically evicts expired entries (background cleanup every 5 min)
- Thread-safe via `sync.Mutex` per bucket

---

### Component 4: Rate Limit Middleware

#### [NEW] [rate_limit.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/middleware/rate_limit.go)

Gin middleware that:
1. Reads `env_id` from context (set by `APIKeyMiddleware`)
2. Looks up rate limit config for the environment
3. Checks RPM limit: `Allow(envID+":rpm", rpmLimit, 1*time.Minute)`
4. Checks daily limit: `Allow(envID+":daily", dailyLimit, 24*time.Hour)`
5. Returns `429 Too Many Requests` with `Retry-After` header on violation
6. Sets `X-RateLimit-Limit` and `X-RateLimit-Remaining` headers

Dependencies: `RateLimiter`, `EnvironmentService` (or a config cache).

---

### Component 5: Environment Config API

#### [MODIFY] [app.controller.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/apps/controller/app.controller.go)

Extend `CreateEnvironment` and add `UpdateEnvironment` to accept rate limit config:
```go
type UpdateEnvironmentRequest struct {
    RateLimitRPM   *int `json:"rate_limit_rpm"`
    RateLimitDaily *int `json:"rate_limit_daily"`
}
```

#### [MODIFY] Router

Add `PUT /:app_id/environments/:env_id` route for updating environment config.

---

### Component 6: Wiring

#### [MODIFY] App initialization

- Create `RateLimiter` instance (in-memory)
- Wire `RateLimitMiddleware` into API-key-authenticated routes
- Ensure middleware runs **after** `APIKeyMiddleware` (needs envID in context)

---

## Verification Plan

### Automated Tests
```bash
go build ./...
go test -race ./pkg/ratelimit/...
go test -race ./internal/apps/...
```
- Unit test `MemoryRateLimiter`: concurrent `Allow()` calls, window expiry, unlimited (limit=0)
- Unit test middleware: mock limiter, verify 429 response + headers

### Manual Verification
1. Set environment rate limit to 5 RPM via API
2. Fire 6 requests in quick succession → 6th returns 429
3. Wait 1 minute → requests succeed again
4. Set daily limit to 10 → fire 11 requests → 11th returns 429

---

## Effort Estimate

| Component | Effort |
|-----------|--------|
| Migration + Entity + Model | ~20 min |
| RateLimiter pkg (interface + in-memory) | ~1.5h |
| Middleware | ~1h |
| Environment config API | ~45 min |
| Wiring | ~20 min |
| Tests | ~1h |
| **Total** | **~5 hours** |

---

## State Management
After completion, update:
- `AI_STATE_MINH.md` — mark Rate Limiting complete
- `PROJECT_MAIN_BACKLOG.md` — check off Rate Limiting, update Apps progress to ~95%
