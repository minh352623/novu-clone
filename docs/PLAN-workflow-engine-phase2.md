# PLAN: Workflow Engine — Phase 2 (Trigger + Execution Engine)

## Problem
Phase 1 (CRUD) is complete: Workflows and Steps can be created/managed via API. However, **no workflow can actually be triggered or executed**. This phase adds the runtime engine that:
1. Receives trigger events via API
2. Creates execution records
3. Processes steps sequentially (Channel → Delay → Channel)
4. Polls for scheduled steps (Delay) via background worker

## Review Level: L3 (Lead Approval)
- Cross-module dependency: calls `NotificationService.Send()`
- Background worker (polling `step_executions`)
- Execution state machine logic

---

## Existing Foundation (Phase 1)

| Layer | Status | Key Files |
|-------|--------|-----------|
| Entities | ✅ | `WorkflowExecution`, `StepExecution` in [workflow.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/domain/model/entity/workflow.go) |
| ExecutionRepo | ✅ | [execution.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/domain/repository/execution.repository.go) (interface + GORM impl) |
| WorkflowRepo | ✅ | `GetByTrigger(envID, triggerIdentifier)` already implemented |
| Migration | ✅ | `workflow_executions` + `step_executions` tables deployed |
| NotificationService | ✅ | Exported as `initializeNotification.NotificationService` |

---

## Proposed Changes

### Component 1: Trigger Service

#### [NEW] [trigger.service.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/application/service/trigger.service.go)

Interface:
```go
type TriggerService interface {
    Trigger(ctx context.Context, envID uuid.UUID, triggerIdentifier string,
            subscriberKey string, payload map[string]interface{}) error
}
```

#### [NEW] [trigger.service.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/application/service/impl/trigger.service.impl.go)

Logic:
1. `workflowRepo.GetByTrigger(envID, triggerIdentifier)` → find active workflow
2. Create `WorkflowExecution` (status=`running`, subscriber_key, trigger_payload)
3. Create `StepExecution` rows for all steps (status=`pending`)
4. Execute first step immediately (Channel → send notification, Delay → schedule)
5. If first step is Channel and succeeds → advance to next step (recursive until Delay or end)

---

### Component 2: Step Handlers

#### [NEW] [step_handler.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/application/service/step_handler.go)

```go
type StepHandler interface {
    Execute(ctx context.Context, step *entity.WorkflowStep, exec *entity.WorkflowExecution) error
}
```

#### [NEW] [channel_handler.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/application/service/impl/channel_handler.go)

- Reads `step.Config`: `{ "template_code": "welcome", "channel": "email" }`
- Calls `NotificationService.Send(SendRequest{...})`
- Maps `exec.SubscriberKey` → `SendRequest.Recipient`
- Maps `exec.TriggerPayload` → `SendRequest.Data`

#### [NEW] [delay_handler.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/application/service/impl/delay_handler.go)

- Reads `step.Config`: `{ "duration": "1h" }` or `{ "duration_seconds": 3600 }`
- Parses duration
- Sets `StepExecution.ScheduledAt = NOW() + duration`
- Sets `StepExecution.Status = "scheduled"`

---

### Component 3: Workflow Executor Worker

#### [NEW] [workflow_executor.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/application/worker/workflow_executor.go)

Polling-based worker (same pattern as `WebhookRetryWorker` / `SLAWorker`):
- Polls `step_executions WHERE status='scheduled' AND scheduled_at <= NOW()` every 30s
- For each due step: execute it, advance to next step
- On completion of last step: mark `WorkflowExecution` as `completed`

```go
type WorkflowExecutor struct {
    execRepo       repository.ExecutionRepository
    workflowRepo   repository.WorkflowRepository
    channelHandler StepHandler
    delayHandler   StepHandler
    interval       time.Duration  // 30s
}

func (w *WorkflowExecutor) Run(ctx context.Context)
func (w *WorkflowExecutor) processStep(ctx context.Context, stepExec *entity.StepExecution) error
func (w *WorkflowExecutor) advanceToNext(ctx context.Context, exec *entity.WorkflowExecution, currentStep *entity.WorkflowStep) error
```

---

### Component 4: Trigger API Endpoint

#### [MODIFY] [workflow.controller.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/controller/workflow.controller.go)

Add new endpoint:
```go
// POST /v1/api/workflows/trigger
func (c *WorkflowController) TriggerWorkflow(ctx *gin.Context) (interface{}, error)
```

#### [NEW] Trigger DTO (in [workflow.dto.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/workflow/controller/dto/workflow.dto.go))

```go
type TriggerWorkflowRequest struct {
    EnvironmentID     uuid.UUID              `json:"environment_id" binding:"required"`
    TriggerIdentifier string                 `json:"trigger_identifier" binding:"required"`
    SubscriberKey     string                 `json:"subscriber_key" binding:"required"`
    Payload           map[string]interface{} `json:"payload"`
}
```

---

### Component 5: Module Wiring

#### [MODIFY] [workflow.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/initialize/workflow/workflow.go)

- Create `ExecutionRepository`
- Create `ChannelHandler` (needs `NotificationService` — import from `initializeNotification`)
- Create `DelayHandler`
- Create `TriggerService` (needs `WorkflowRepo`, `ExecutionRepo`, step handlers)
- Create `WorkflowExecutor` worker → start as goroutine
- Register `POST /workflows/trigger` route

---

## User Review Required

> [!IMPORTANT]
> **Cross-module import**: The workflow init package will import `initializeNotification.NotificationService`. This creates a dependency chain: `workflow` → `notification`. Ensure notification module initializes **before** workflow module in `router.go` (currently it does).

> [!WARNING]
> **Step execution is fire-and-forget for Channel steps**: if `NotificationService.Send()` fails, the step is marked `failed` and the workflow stops. No automatic retry of workflow steps (separate from webhook retry). Confirm this is acceptable.

---

## Verification Plan

### Automated Tests
```bash
go build ./...
go test -race ./internal/workflow/...
```
- Unit tests for `TriggerService`: mock repos → verify execution + step creation
- Unit tests for `ChannelHandler`: mock `NotificationService` → verify `Send()` called with correct args
- Unit tests for `DelayHandler`: verify `ScheduledAt` calculation
- Unit tests for `WorkflowExecutor.processStep`: verify step advancement logic

### Manual Verification
1. Create a workflow with steps: `Channel(email) → Delay(30s) → Channel(email)`
2. Trigger via `POST /workflows/trigger`
3. Verify first email sent immediately
4. Wait 30s, verify executor picks up scheduled step
5. Verify second email sent

---

## Effort Estimate

| Component | Effort |
|-----------|--------|
| TriggerService (interface + impl) | ~1.5h |
| StepHandlers (Channel + Delay) | ~1h |
| WorkflowExecutor worker | ~1.5h |
| Controller + DTO + Wiring | ~1h |
| Unit tests | ~1.5h |
| **Total** | **~6–7 hours** |

---

## State Management
After completion, update:
- `AI_STATE_MINH.md` — mark Phase 2 complete
- `PROJECT_MAIN_BACKLOG.md` — update Workflow Engine to ~60%
