# PLAN: Comprehensive Refactoring & Optimization Phase 2-4

This plan follows `.ai/INSTRUCTIONS.md` and `GOLANG_BEST_PRACTICES.md` to complete the codebase optimization, focusing on database integrity, performance, and standardizing shared components.

## Review Levels
- **Level**: L2 (Cross-module Changes)
- **Impact**: Touches `IAM`, `Apps`, `Notification`, `Messaging`, `Workflow`, `Health`, `R2`, and `pkg/`.
- **Review Required**: Lead (Schema changes)

## Proposed Changes

### Phase 1: Database & Persistence Integrity
Address remaining schema and model mismatches from `AUDIT_REPORT.md`:
- **[Notification]**: Create migration `00021_rename_notification_logs.sql` to rename `notification_logs` → `notifications`. Update `NotificationModel.TableName()`.
- **[Apps]**: Create migration to add `sla_threshold_seconds` to `apps` and `environments` tables.
- **[Messaging]**: Create migration to add `type` column to `messages` and `channel` column to `threads`.
- **[Notification]**: Add `app_id` to `ProviderConfigModel`.

### Phase 2: Performance & Reliability (BP 4.4, 5.1, 2.2)
- **Panic Recovery Logging**: 
    - Fix missing logging in `internal/middleware/api_key.go`.
    - Fix missing logging in `internal/notification/controller/job.controller.go`.
- **Magic Numbers**:
    - Centralize `SLAThresholdSeconds` default (900s) and bucket intervals in `internal/messaging/domain/constants.go`.
    - Centralize success status ranges (200-299) in `internal/apps/domain/model/constants.go` for metrics.
- **Slice Pre-allocation**:
    - Audit and fix `calculateTimeline` in `internal/messaging` and other analytical loops.

### Phase 3: Module Standardisation (Phases 1-4 for remaining modules)
- **[Health] & [R2]**: Apply Phase 1 standards (Structured Logging, Error Wrapping).
- **[Middleware]**: Clean up legacy comments and standardize logging.

### Phase 4: Standard Response Format (BP 10)
- **Controllers**: systematic sweep to replace direct `c.JSON(http.StatusOK, gin.H{...})` with `pkg/response` helpers.

## Verification Plan
1. **Migrations**: Run `goose up` in a fresh local DB.
2. **Build**: `go build ./...`
3. **Tests**: `go test -race ./internal/... ./pkg/...`
4. **Logs**: Verify structured keys (e.g. `error`, `app_id`) in container logs.
