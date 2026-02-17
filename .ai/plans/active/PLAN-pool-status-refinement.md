# [Pool Status Refinement]

Harden thread status management in the Messaging module: typed constants, valid transition matrix, reopen flow, and missing broadcast events.

## Current State

```
Thread.Status = "unassigned" | "assigned" | "resolved"   (raw strings, no validation)
```

**Issues found:**
1. Raw strings — no typed constants or domain-level validation
2. No transition guard — e.g. `resolved → assigned` is allowed (should require reopen first)
3. Missing `ReopenThread` — customer message after resolve creates NEW thread instead of reopening
4. `UnassignThread` has no broadcast event (assign + resolve do)
5. `ReceiveMessage` doesn't auto-reopen resolved threads

## Proposed Changes

### 1. `internal/messaging/domain/model/entity`

#### [MODIFY] [messaging.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/domain/model/entity/messaging.go)

Add typed status constants and transition validation:

```go
type ThreadStatus string

const (
    ThreadStatusUnassigned ThreadStatus = "unassigned"
    ThreadStatusAssigned   ThreadStatus = "assigned"
    ThreadStatusResolved   ThreadStatus = "resolved"
)

// ValidTransitions defines allowed status changes
var ValidTransitions = map[ThreadStatus][]ThreadStatus{
    ThreadStatusUnassigned: {ThreadStatusAssigned},
    ThreadStatusAssigned:   {ThreadStatusUnassigned, ThreadStatusResolved},
    ThreadStatusResolved:   {ThreadStatusUnassigned},  // reopen goes to unassigned
}

func (t *Thread) TransitionTo(target ThreadStatus) error { ... }
```

Change `Status string` → `Status ThreadStatus`.

---

### 2. `internal/messaging/application/service/impl`

#### [MODIFY] [conversation.service.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/impl/conversation.service.impl.go)

| Function | Change |
|----------|--------|
| `AssignThread` | Use `thread.TransitionTo(ThreadStatusAssigned)` instead of raw assignment |
| `UnassignThread` | Use `thread.TransitionTo(ThreadStatusUnassigned)` + add `broadcastEvent` |
| `ResolveThread` | Use `thread.TransitionTo(ThreadStatusResolved)` |
| `ReceiveMessage` | If existing thread is `resolved`, call `TransitionTo(ThreadStatusUnassigned)` to reopen |

#### [NEW] `ReopenThread` method (optional API, auto-triggered by `ReceiveMessage`)

---

### 3. `internal/messaging/controller`

#### [MODIFY] [conversation.controller.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/conversation.controller.go)

Update Swagger docs to reflect typed status values.

---

### 4. `internal/messaging/infrastructure/persistence`

#### [MODIFY] Thread model/mapper

Ensure `ThreadStatus` type is properly mapped for GORM (string underlying type — no schema change needed).

## Verification Plan

### Automated Tests
- `go build ./...` — compile check
- `go test ./internal/messaging/...` — if tests exist

### Manual Verification
1. Assign → Resolve → Send inbound message → verify thread reopens as `unassigned`
2. Try invalid transition (e.g. `unassigned → resolved`) → verify error returned
3. Unassign → check WebSocket broadcast event fires
