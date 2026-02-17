# [WebSocket Real-Time Updates]

Complete the WebSocket infrastructure: auth, structured events, typing indicators, online presence, and inbound message handling.

## Current State

**Already built (~80%):**
- ✅ Hub: register/unregister, `BroadcastToEnvironment()`, `BroadcastToUsers()`
- ✅ Client: `readPump` / `writePump` with ping/pong heartbeat
- ✅ `ServeWs()` handler, `ServeWS` controller, `/ws` route
- ✅ `broadcastEvent()` fires 7 event types across all service methods
- ✅ `WsEvent` DTO, `gorilla/websocket` dependency

**Gaps identified:**

| # | Gap | Impact |
|---|-----|--------|
| 1 | `/ws` route has **no auth middleware** | Anyone can connect |
| 2 | `ServeWS` token validation is incomplete | Query param `user_id` accepted as-is (dev hack) |
| 3 | `WsEvent` has no `event_id`, `timestamp` | Client can't deduplicate or detect missed events |
| 4 | `readPump` drops all inbound messages | No typing indicator or client→server events |
| 5 | No online presence tracking | Can't show who's online |
| 6 | No reconnection / missed-event protocol | Client loses events during disconnect |

## User Review Required

> [!IMPORTANT]
> **Scope question**: Items 1–4 are essential for production readiness. Items 5–6 (online presence, missed-event recovery) add complexity. Recommend implementing **1–4 in this task** and deferring 5–6 to a separate backlog item. Do you agree?

## Proposed Changes

### 1. WebSocket Auth (Fix Security Gap)

#### [MODIFY] [router.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/router.go)

Add auth middleware to `/ws` route:
```go
group.GET("/ws", authMiddleware, conversationController.ServeWS)
```

#### [MODIFY] [conversation.controller.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/conversation.controller.go)

Remove the dev-mode `user_id` query param hack. Rely on middleware-set `user_id` from JWT.

---

### 2. Structured Event Envelope

#### [MODIFY] [conversation.dto.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/dto/conversation.dto.go)

Enhance `WsEvent`:
```go
type WsEvent struct {
    ID        string      `json:"id"`        // UUID for idempotency
    Type      string      `json:"type"`
    Payload   interface{} `json:"payload"`
    Timestamp time.Time   `json:"timestamp"`
}
```

#### [MODIFY] [conversation.service.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/impl/conversation.service.impl.go)

Update `broadcastEvent()` to populate `ID` and `Timestamp`.

---

### 3. Inbound Client Events (Typing Indicators)

#### [MODIFY] [client.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/infrastructure/gateway/client.go)

Process inbound WS messages in `readPump`:
```go
// Parse inbound event
var event WsInboundEvent
json.Unmarshal(message, &event)
switch event.Type {
    case "typing_start": hub.handleTyping(client, event)
    case "typing_stop":  hub.handleTyping(client, event)
}
```

#### [MODIFY] [hub.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/infrastructure/gateway/hub.go)

Add `handleTyping()` — broadcasts `typing_start`/`typing_stop` to env clients (exclude sender).

#### [NEW] [ws_events.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/infrastructure/gateway/ws_events.go)

Inbound/outbound event type definitions:
```go
type WsInboundEvent struct {
    Type     string          `json:"type"`
    ThreadID uuid.UUID       `json:"thread_id"`
    Data     json.RawMessage `json:"data,omitempty"`
}
```

---

### 4. Tests

#### [MODIFY] [hub_test.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/infrastructure/gateway/hub_test.go)

Add tests: typing broadcast, environment scoping, structured event parsing.

## Verification Plan

### Automated Tests
- `go build ./...`
- `go test ./internal/messaging/...`

### Manual Verification
- Connect via `wscat -c ws://localhost:PORT/conversations/ws -H "Authorization: Bearer TOKEN"`
- Verify auth rejection without token
- Verify event structure has `id` + `timestamp`
- Send typing event and verify broadcast
