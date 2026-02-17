# Pool Status Refinement

**Date**: 2026-02-17  
**Author**: Minh  
**Module**: Messaging  

## Changes
- Introduced typed `ThreadStatus` constants (`unassigned`, `assigned`, `resolved`)
- Added `TransitionTo()` method with validated status transitions
- Auto-reopen resolved threads on new customer message
- Added missing broadcast events for `UnassignThread` and `thread_reopened`

## Files Changed
- `internal/messaging/domain/model/entity/messaging.go`
- `internal/messaging/application/service/impl/conversation.service.impl.go`
- `internal/messaging/infrastructure/persistence/repository/common.repository.go`
- `internal/messaging/controller/dto/conversation.dto.go`
- `internal/messaging/application/worker/sla_worker.go`
- `internal/messaging/application/worker/sla_worker_test.go`
