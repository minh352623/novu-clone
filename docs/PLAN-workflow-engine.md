# PLAN: Workflow Engine — Flow D

## Problem
The Workflow Engine is at **~5% completion**. The DB schema exists ([00004_workflow_engine.sql](file:///Users/tekix/Documents/company/converda/converda-service/sql/schema/00004_workflow_engine.sql)) but **zero Go code** has been written. This module enables automation: trigger events → execute a sequence of steps (Channel/Delay/Digest).

## Review Level: L3 (Lead Approval)
- New module with cross-module dependencies (Notification Pipeline)
- Introduces background job processing (delay steps, digest collection)
- Significant architectural decisions needed

---

## Existing Schema

```sql
-- workflows: scoped to environments, trigger_identifier-based
CREATE TABLE workflows (
    id UUID PRIMARY KEY,
    environment_id UUID NOT NULL REFERENCES environments(id),
    name TEXT NOT NULL,
    trigger_identifier TEXT NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    UNIQUE (environment_id, trigger_identifier)
);

-- workflow_steps: ordered steps with parent hierarchy, JSONB config
CREATE TABLE workflow_steps (
    id UUID PRIMARY KEY,
    workflow_id UUID NOT NULL REFERENCES workflows(id),
    parent_step_id UUID REFERENCES workflow_steps(id),
    step_type TEXT NOT NULL,        -- 'channel', 'delay', 'digest'
    config JSONB DEFAULT '{}',
    "order" INTEGER NOT NULL DEFAULT 0
);

-- subscribers: already used by notification pipeline
CREATE TABLE subscribers (...);
```

---

## Architecture Decisions Required

> [!IMPORTANT]
> These decisions need Lead approval before coding begins.

### Decision 1: Execution Model
| Option | Pros | Cons |
|--------|------|------|
| **A. Synchronous in-process** | Simple, no infra needed | Delay steps block goroutines, no crash recovery |
| **B. Job queue (in-DB)** | Crash recovery, audit trail, scalable | More tables, polling overhead |
| **C. External queue (Redis/NATS)** | Best performance, true async | New dependency, operational complexity |

**Recommendation:** Option B — DB-backed job queue. Fits the current stack (PostgreSQL) and enables crash recovery without new infra.

### Decision 2: Delay Step Implementation
| Option | Pros | Cons |
|--------|------|------|
| **A. `time.AfterFunc` goroutine** | Simple | Lost on restart |
| **B. DB row + polling worker** | Durable, survives restarts | 30s polling granularity |
| **C. Scheduled job table** | Precise, durable | More complex |

**Recommendation:** Option B — consistent with SLAWorker / WebhookRetryWorker patterns already in codebase.

### Decision 3: Digest Step Scope
Digest steps batch multiple events into a single notification. This is the most complex step type.

**Recommendation:** Defer digest to Phase 2. Implement Channel + Delay first.

---

## Proposed Phases

### Phase 1 — Foundation (DDD Structure + CRUD)

```
internal/workflow/
├── domain/
│   ├── model/entity/
│   │   ├── workflow.go          # Workflow, WorkflowStep entities
│   │   └── execution.go         # WorkflowExecution, StepExecution entities
│   └── repository/
│       ├── workflow.repository.go
│       └── execution.repository.go
├── application/
│   ├── service/
│   │   ├── workflow.service.go       # Interface
│   │   └── impl/
│   │       └── workflow.service.impl.go
│   └── worker/
│       └── workflow_executor.go       # Background step executor
├── controller/
│   ├── workflow.controller.go
│   ├── dto/
│   │   └── workflow.dto.go
│   └── router.go
└── infrastructure/
    └── persistence/
        ├── model/
        │   ├── workflow.model.go
        │   └── execution.model.go
        ├── mapper/
        │   ├── workflow.mapper.go
        │   └── execution.mapper.go
        └── repository/
            ├── workflow.repository.go
            └── execution.repository.go
```

#### New Migration: `00024_workflow_execution.sql`

```sql
CREATE TABLE workflow_executions (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    workflow_id UUID NOT NULL REFERENCES workflows(id),
    subscriber_key TEXT NOT NULL,
    trigger_payload JSONB DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'running', -- running, completed, failed, cancelled
    current_step_id UUID REFERENCES workflow_steps(id),
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE step_executions (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    execution_id UUID NOT NULL REFERENCES workflow_executions(id),
    step_id UUID NOT NULL REFERENCES workflow_steps(id),
    status TEXT NOT NULL DEFAULT 'pending', -- pending, running, completed, failed, scheduled
    scheduled_at TIMESTAMP WITH TIME ZONE,  -- for delay steps
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    output JSONB DEFAULT '{}'
);
```

#### API Endpoints

| Method | Route | Description |
|--------|-------|-------------|
| POST | `/workflows` | Create workflow |
| GET | `/workflows` | List workflows (by environment) |
| GET | `/workflows/:id` | Get workflow + steps |
| PATCH | `/workflows/:id` | Update workflow |
| DELETE | `/workflows/:id` | Delete workflow |
| POST | `/workflows/:id/steps` | Add step |
| PATCH | `/workflows/:id/steps/:stepId` | Update step |
| DELETE | `/workflows/:id/steps/:stepId` | Delete step |
| POST | `/workflows/:id/toggle` | Activate/Deactivate |

---

### Phase 2 — Trigger + Execution Engine

#### Trigger Service
```go
type TriggerService interface {
    // Called by external apps or internal events
    Trigger(ctx context.Context, envID uuid.UUID, triggerIdentifier string, subscriberKey string, payload map[string]interface{}) error
}
```

- Looks up active `Workflow` by `(environment_id, trigger_identifier)`
- Creates `WorkflowExecution` + `StepExecution` rows
- Starts execution of the first step

#### Step Executor (Worker)
```go
type WorkflowExecutor struct {
    // Polls step_executions where status='scheduled' AND scheduled_at <= NOW()
    // Executes the step, then advances to the next step
}
```

**Step type handlers:**

| Step Type | Config Example | Behavior |
|-----------|---------------|----------|
| `channel` | `{"template_code": "welcome", "channel": "email"}` | Calls `NotificationService.Send()` |
| `delay` | `{"duration": "1h"}` | Schedules next step for `NOW() + duration` |

#### Cross-Module Integration
- `WorkflowExecutor` depends on `NotificationService` (already exported as module-level var)
- Trigger API is a new endpoint: `POST /v1/api/workflows/trigger`

---

### Phase 3 — Digest Steps (Deferred)
- Requires batch collection window
- More complex scheduling
- Can be added as incremental improvement later

---

## Verification Plan

### Automated Tests
- `go build ./...`
- `go test -race ./internal/workflow/...`
- Unit tests for: Workflow CRUD, Step ordering, Trigger → Execution creation
- Integration test: Trigger → Channel step → Notification sent

### Manual
- Create workflow with 2 steps: Channel → Delay → Channel
- Trigger it via API
- Verify first notification sent immediately
- Verify second notification sent after delay

---

## Effort Estimate

| Phase | Effort | Dependencies |
|-------|--------|--------------|
| Phase 1 (CRUD) | ~4–6 hours | None |
| Phase 2 (Engine) | ~6–8 hours | Notification Pipeline |
| Phase 3 (Digest) | ~4–6 hours | Phase 2 |

**Total: ~14–20 hours across 2–3 sprints**

---

## State Management
After each phase, update:
- `AI_STATE_MINH.md` — mark phase complete
- `PROJECT_MAIN_BACKLOG.md` — update workflow progress %
