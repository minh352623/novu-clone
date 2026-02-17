# [Audit Trail Pagination]

Add paginated, filterable audit trail API for thread assignment/resolution history.

## Current State

**Already built:**
- ✅ `AssignmentLog` entity (ID, ThreadID, AssignedToMemberID, AssignedAt, ResolvedAt, ResponseTimeSeconds)
- ✅ `AssignmentLogRepository` interface + GORM impl (Create, GetByThread, GetLastByThread, Update, stats)
- ✅ `GetThreadAuditTrail` service method with member name enrichment
- ✅ `GetThreadAuditTrail` controller with Swagger docs
- ✅ `AuditTrailResponse` + `AssignmentLogResponse` DTOs

**Gaps:**

| # | Gap | Impact |
|---|-----|--------|
| 1 | **Route not registered** in `router.go` | Endpoint unreachable |
| 2 | `GetByThread` repo returns **all logs, no pagination** | Scales poorly for long-lived threads |
| 3 | **No date-range filtering** | Can't query specific time periods |
| 4 | `AuditTrailResponse` has **no pagination metadata** | Client can't paginate |

## Proposed Changes

### 1. Register Route

#### [MODIFY] [router.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/router.go)

Add missing route:
```go
group.GET("/:id/audit-trail", response.Wrap(conversationController.GetThreadAuditTrail, http.StatusOK))
```

---

### 2. Paginated Repository

#### [MODIFY] [assignment_log.repository.go (domain)](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/domain/repository/assignment_log.repository.go)

Add `AuditTrailFilter` struct and `ListByThread` method:
```go
type AuditTrailFilter struct {
    ThreadID uuid.UUID
    From     *time.Time
    To       *time.Time
    Limit    int
    Offset   int
}

ListByThread(ctx context.Context, filter AuditTrailFilter) ([]*entity.AssignmentLog, int64, error)
```

#### [MODIFY] [assignment_log.repository.go (infra)](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/infrastructure/persistence/repository/assignment_log.repository.go)

Implement `ListByThread` with GORM: count + paginated query with optional date filter.

---

### 3. Update Service + Controller

#### [MODIFY] [conversation.service.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/conversation.service.go)

Update `GetThreadAuditTrail` signature to accept pagination/filter params.

#### [MODIFY] [conversation.service.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/impl/conversation.service.impl.go)

Use `ListByThread` instead of `GetByThread`.

#### [MODIFY] [conversation.controller.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/conversation.controller.go)

Parse query params: `page`, `page_size`, `from`, `to`. Update Swagger docs.

---

### 4. Response DTO with Pagination

#### [MODIFY] [conversation.dto.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/dto/conversation.dto.go)

Add pagination fields to `AuditTrailResponse`:
```go
type AuditTrailResponse struct {
    ThreadID uuid.UUID                `json:"thread_id"`
    Logs     []*AssignmentLogResponse `json:"logs"`
    Total    int64                    `json:"total"`
    Page     int                      `json:"page"`
    PageSize int                      `json:"page_size"`
}
```

## Verification Plan

### Automated Tests
- `go build ./...`
- `go test ./internal/messaging/...`
