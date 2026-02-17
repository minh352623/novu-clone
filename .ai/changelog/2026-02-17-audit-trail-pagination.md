# Audit Trail Pagination

**Date**: 2026-02-17  
**Author**: Minh  
**Module**: Messaging  

## Changes
- Registered `GET /:id/audit-trail` route in auth-protected group
- Added `AuditTrailFilter` struct and `ListByThread()` to domain repo interface
- Implemented `ListByThread()` in GORM repo with date-range filter + count
- Updated `AuditTrailResponse` DTO with `total`, `page`, `page_size`
- Updated service interface and impl to accept pagination params
- Updated controller to parse `page`, `page_size`, `from`, `to` query params

## Files Changed
- `internal/messaging/controller/router.go`
- `internal/messaging/domain/repository/assignment_log.repository.go`
- `internal/messaging/infrastructure/persistence/repository/assignment_log.repository.go`
- `internal/messaging/application/service/conversation.service.go`
- `internal/messaging/application/service/impl/conversation.service.impl.go`
- `internal/messaging/controller/conversation.controller.go`
- `internal/messaging/controller/dto/conversation.dto.go`
- `internal/messaging/application/service/impl/conversation_service_test.go`
- `internal/messaging/application/worker/sla_worker_test.go`
