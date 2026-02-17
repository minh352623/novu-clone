# WebSocket Real-Time Updates

**Date**: 2026-02-17  
**Author**: Minh  
**Module**: Messaging  

## Changes
- Applied auth middleware to `/ws` route
- Removed dev hack (`user_id` query param) from `ServeWS`
- Enhanced `WsEvent` with `ID` (UUID) and `Timestamp` fields
- Implemented inbound event parsing in `readPump`
- Added `HandleInbound()` to Hub for typing indicators
- Added `BroadcastToEnvironmentExclude()` method
- Created `ws_events.go` with event type constants
- Wrote 8 Hub tests (was 1)

## Files Changed
- `internal/messaging/controller/router.go`
- `internal/messaging/controller/conversation.controller.go`
- `internal/messaging/controller/dto/conversation.dto.go`
- `internal/messaging/application/service/impl/conversation.service.impl.go`
- `internal/messaging/infrastructure/gateway/ws_events.go` [NEW]
- `internal/messaging/infrastructure/gateway/client.go`
- `internal/messaging/infrastructure/gateway/hub.go`
- `internal/messaging/infrastructure/gateway/hub_test.go`
