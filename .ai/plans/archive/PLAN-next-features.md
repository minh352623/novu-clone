# PLAN: Remaining Features Sprint

Consolidates 3 tasks into a single plan:
1. **Task A — WebhookRetryWorker Init** (5 min)
2. **Task B — US-RA-01: System Health Dashboard** (new feature)
3. **Task C — US-CA-06: Personal Dashboard Enhancement** (fix stub)

---

## Task A — WebhookRetryWorker Init

### Scope
Wire `WebhookRetryWorker` into app startup so it actually runs. Currently `notification.go` creates the dispatcher but never starts the retry worker.

### Proposed Changes

#### [MODIFY] [notification.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/initialize/notification/notification.go)

```diff
+import "CONVERDA/internal/notification/application/worker"

 // After webhookDispatcher initialization:
-_ = webhookDispatcher
+// Start Webhook Retry Worker
+retryWorker := worker.NewWebhookRetryWorker(webhookDispatcher, webhookLogRepo)
+go retryWorker.Run(context.Background())
```

### Review Level: L1 (Auto-merge)
Single-line wiring, no design decisions.

---

## Task B — US-RA-01: System Health Dashboard

### Scope
New API endpoint providing operational health overview:
- **API Performance**: avg response time, error rate, requests/min  
- **Queue Health**: unassigned threads count, overdue count, avg wait time  
- **System Status**: DB connectivity, webhook success rate, notification delivery rate  
- **SLA Compliance**: % threads resolved within SLA threshold  

### Proposed Changes

#### [NEW] [health.dto.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/dto/health.dto.go)

DTOs for the health dashboard:
```go
type SystemHealthResponse struct {
    QueueHealth    QueueHealth    `json:"queue_health"`
    SLACompliance  SLACompliance  `json:"sla_compliance"`
    WebhookHealth  WebhookHealth  `json:"webhook_health"`
    GeneratedAt    time.Time      `json:"generated_at"`
}

type QueueHealth struct {
    UnassignedCount int     `json:"unassigned_count"`
    OverdueCount    int     `json:"overdue_count"`
    AvgWaitTimeSec  float64 `json:"avg_wait_time_seconds"`
    TotalActive     int     `json:"total_active"`
}

type SLACompliance struct {
    TotalThreads   int     `json:"total_threads"`
    WithinSLA      int     `json:"within_sla"`
    Breached       int     `json:"breached"`
    ComplianceRate float64 `json:"compliance_rate_percent"`
}

type WebhookHealth struct {
    TotalDispatched int     `json:"total_dispatched"`
    SuccessCount    int     `json:"success_count"`
    FailedCount     int     `json:"failed_count"`
    PendingRetries  int     `json:"pending_retries"`
    SuccessRate     float64 `json:"success_rate_percent"`
}
```

#### [MODIFY] [conversation.service.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/conversation.service.go)

Add `GetSystemHealth(ctx, envID, from, to) (*dto.SystemHealthResponse, error)` to interface.

#### [MODIFY] [conversation.service.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/impl/conversation.service.impl.go)

Implement `GetSystemHealth`:
- Query `ThreadRepository` for queue counts (unassigned, overdue, active)
- Query `AssignmentLogRepository` for SLA compliance (resolved within threshold vs. breached)
- Inject `WebhookLogRepository` to query webhook stats (success/failed/pending)

#### [MODIFY] [thread.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/domain/repository/thread.repository.go)

Add repository method:
```go
GetQueueHealth(ctx context.Context, envID uuid.UUID) (*QueueHealthData, error)
```

#### [MODIFY] [assignment_log.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/domain/repository/assignment_log.repository.go)

Add repository method:
```go
GetSLACompliance(ctx context.Context, envID uuid.UUID, slaThreshold int, from, to time.Time) (*SLAComplianceData, error)
```

#### [MODIFY] [webhook.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/notification/domain/repository/webhook.repository.go)

Add repository method:
```go
GetHealthStats(ctx context.Context, from, to time.Time) (*WebhookHealthData, error)
```

#### [MODIFY] [conversation.controller.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/conversation.controller.go)

Add `GetSystemHealth` handler + route.

> [!IMPORTANT]
> **Cross-module dependency**: This feature requires the messaging module to read from `webhook_logs` (notification module). Options:
> 1. Inject `WebhookLogRepository` into `ConversationService` ← simple but breaks module boundaries
> 2. Create a dedicated `HealthService` in a shared module ← cleaner, more work
> 3. Return webhook health as a separate endpoint from notification module ← two API calls
>
> **Recommendation**: Option 1 for now (pragmatic), refactor later if modules are split.

### Review Level: L2 (AI Review)

---

## Task C — US-CA-06: Personal Dashboard Enhancement

### Scope
Fix the **stubbed** `calculateTimeline` method that currently returns `Value: 0` for all data points. Also enhance with:
- **Real activity data**: resolved threads per time bucket from `assignment_log`
- **SLA Compliance trend**: hourly/daily SLA rate  
- **Performance comparison**: agent vs. team average

### Current State
```go
// conversation.service.impl.go:657-681
func (s *conversationServiceImpl) calculateTimeline(...) []dto.ActivityPoint {
    // Returns Value: 0 for all points — STUB
}
```

### Proposed Changes

#### [MODIFY] [assignment_log.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/domain/repository/assignment_log.repository.go)

Add repository method:
```go
GetActivityTimeline(ctx context.Context, envID, memberID uuid.UUID, from, to time.Time, bucketHours int) ([]ActivityBucket, error)
```

Uses SQL `date_trunc` + `GROUP BY` for efficient bucketing.

#### [MODIFY] GORM repository implementation

Implement `GetActivityTimeline` with:
```sql
SELECT date_trunc('hour', resolved_at) AS bucket,
       COUNT(*) AS resolved_count,
       AVG(response_time_seconds) AS avg_response_time
FROM assignment_logs 
WHERE environment_id = ? AND assigned_to_member_id = ?
  AND resolved_at BETWEEN ? AND ?
GROUP BY bucket ORDER BY bucket;
```

#### [MODIFY] [conversation.service.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/impl/conversation.service.impl.go)

Replace stub `calculateTimeline` with real repository call.

#### [MODIFY] [dto/dashboard.dto.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/dto)

Enhance `ActivityPoint` with additional fields:
```go
type ActivityPoint struct {
    Time             time.Time `json:"time"`
    ResolvedCount    int       `json:"resolved_count"`
    AvgResponseTime  float64   `json:"avg_response_time_seconds"`
}
```

### Review Level: L2 (AI Review)

---

## Execution Order

```
Task A (5 min)  →  Task C (1–2 hrs)  →  Task B (2–3 hrs)
   ↓                    ↓                     ↓
  Wire startup      Fix stub data        New endpoints
```

> [!NOTE]
> Task A is a prerequisite blocker (webhook retry isn't running without it).
> Task C before B because it fixes broken functionality rather than adding new.

---

## Verification Plan

### Automated Tests
- `go build ./...` after each task
- `go test -race ./internal/messaging/...`
- `go test -race ./internal/notification/...`
- New tests for `GetActivityTimeline`, `GetQueueHealth`, `GetSLACompliance`

### Manual
- Call `/dashboards/personal` and verify non-zero timeline values
- Call `/dashboards/health` and verify queue/SLA/webhook stats
- Confirm `WebhookRetryWorker` logs `"Webhook Retry Worker started"` on app boot
