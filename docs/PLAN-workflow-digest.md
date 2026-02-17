# Workflow Engine Phase 3: Digest Steps

## Problem

The Workflow Engine supports `channel` (send notification) and `delay` (wait N time) steps, but the third step type — `digest` — is not implemented. The constant `StepTypeDigest = "digest"` exists in [workflow.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/domain/model/entity/workflow.go#L13) but no handler exists.

**Digest** = collect (batch) multiple trigger events within a time window, then forward the aggregated batch to the next step. Use case: "Send one daily digest email with all 24 events from today" instead of 24 separate emails.

---

## User Review Required

> [!IMPORTANT]
> **New DB table required**: `digest_events` to buffer events during the digest window.

> [!WARNING]
> **Breaking change to `WorkflowExecutor`**: constructor gains a `digestHandler` param + a `digest` case in the step-type switch. The init wiring file must be updated.

---

## Proposed Changes

### Component 1: Database

#### [NEW] 00027_digest_events.sql

New migration creating `digest_events`:

```sql
CREATE TABLE digest_events (
    id UUID PRIMARY KEY DEFAULT generate_uuid_v7(),
    step_id UUID NOT NULL REFERENCES workflow_steps(id) ON DELETE CASCADE,
    execution_id UUID NOT NULL REFERENCES workflow_executions(id) ON DELETE CASCADE,
    subscriber_key TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_digest_events_step_exec ON digest_events(step_id, execution_id);
```

---

### Component 2: Domain Entity

#### [MODIFY] [workflow.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/domain/model/entity/workflow.go)

Add `DigestEvent` entity:

```go
type DigestEvent struct {
    ID             uuid.UUID
    StepID         uuid.UUID
    ExecutionID    uuid.UUID
    SubscriberKey  string
    Payload        map[string]interface{}
    CreatedAt      time.Time
}
```

Add `StepStatusDigesting = "digesting"` constant for steps waiting/collecting events.

---

### Component 3: Repository

#### [MODIFY] [execution.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/domain/repository/execution.repository.go)

Add 3 methods:

```go
BufferDigestEvent(ctx, event *DigestEvent) error
FlushDigestEvents(ctx, stepID, executionID uuid.UUID) ([]*DigestEvent, error)
GetDigestingSteps(ctx, limit int) ([]*StepExecution, error)
```

#### [NEW] digest.model.go — GORM model + mapper

---

### Component 4: Digest Handler

#### [NEW] digest_handler.go

`DigestHandler` implements `StepHandler`:

1. Parse config: `{ "window": "1h" }` or `{ "window_seconds": 3600 }`
2. Buffer the trigger payload as a `DigestEvent`
3. Schedule the step to `digesting` status with `scheduled_at = now + window`
4. On flush (when window expires): collect all buffered events, merge payloads into a single `[]map[string]interface{}` array, store in `stepExec.Output["events"]`

---

### Component 5: Digest Flusher (Worker Enhancement)

#### [MODIFY] [workflow_executor.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/application/worker/workflow_executor.go)

- Add `digestHandler service.StepHandler` field to `WorkflowExecutor`
- Add `case entity.StepTypeDigest:` in step-type switch (lines 166–175)
- Add separate poll: `GetDigestingSteps` → for each due step, flush events and advance to next step
- Digest step flush logic: collect all `DigestEvent` rows, merge into `exec.TriggerPayload["digest_events"]`, then continue to next step (usually a `channel` step that renders the digest template)

---

### Component 6: Init Wiring

#### [MODIFY] [workflow init](file:///Users/tekix/Documents/company/converda/converda-service/internal/initialize) (workflow init file)

Pass `digestHandler` when constructing `WorkflowExecutor`.

---

## Review Levels

### Cấp 1 — Local Impact
- **Workflow module only**: new handler follows same `StepHandler` interface pattern as `ChannelHandler` / `DelayHandler`
- No impact on `notification`, `messaging`, `apps`, or `iam` modules

### Cấp 2 — Breaking Changes
- `WorkflowExecutor` constructor signature changes (adds `digestHandler` param)
- Must update init wiring to pass the new handler

### Cấp 3 — Lead Approval
Awaiting `APPROVED` before implementation.

---

## Verification Plan

### Automated Tests
1. Unit test `DigestHandler.Execute` — buffers event, sets status to `digesting`
2. Unit test digest flush — collects N events, merges payloads, advances execution
3. `go build ./...` — clean compilation
4. `go test -race ./internal/... ./pkg/...` — all packages pass

### Manual Verification
- Create workflow: `delay → digest(window=5s) → channel`
- Trigger 3 times within window
- Verify single notification sent with all 3 events in payload

---

## Effort Estimate

| Task | Estimate |
|------|----------|
| DB migration | ~5 min |
| Entity + constants | ~5 min |
| Repository + model/mapper | ~15 min |
| DigestHandler | ~15 min |
| Executor integration | ~15 min |
| Init wiring | ~5 min |
| Tests | ~15 min |
| **Total** | **~1.5 hours** |

---

## AI State Update

After completion → update `AI_STATE_MINH.md`:
- Mark "Workflow Engine Phase 3: Digest Steps" as ✅
- Update `PROJECT_MAIN_BACKLOG.md`: move Workflow Engine progress to ~90%
