# PLAN: Global Codebase Refactoring & Optimization

Review and optimize all modules against `.ai/INSTRUCTIONS.md` and `GOLANG_BEST_PRACTICES.md`.

## Review Levels
- **Level**: L2 (Cross-module Changes)
- **Impact**: Touches `Messaging`, `IAM`, `Apps`, `Notification`, `Workflow`.
- **Review Required**: Lead (Architectural change in Messaging decoupling).

## Proposed Changes

### [Messaging] Inter-module Decoupling
Introduce `AppReader` and `MemberReader` interfaces in `Messaging` domain to remove direct dependencies on `Apps` and `IAM` repositories.
- **New Files**:
    - `internal/messaging/domain/repository/app_reader.go`
    - `internal/messaging/domain/repository/member_reader.go`
    - `internal/messaging/infrastructure/adapter/local_app_adapter.go`
    - `internal/messaging/infrastructure/adapter/local_member_adapter.go`
- **Changes**:
    - Refactor `ConversationService` to use these domain interfaces.

### [Apps] Magic Number Centralization
Remove hardcoded HTTP status ranges in `metrics.repository.go`.
- **New Files**:
    - `internal/apps/domain/model/constants.go`
- **Changes**:
    - Define `SuccessStatusMin`, `SuccessStatusMax`.
    - Update SQL queries to use parameters instead of hardcoded values.

### [All Modules] Error Wrapping & Logging
Systematic sweep to fix bare error returns (`return err`) and legacy logging.
- **Target**: All `service/impl` and `infrastructure/persistence/repository` files.
- **Pattern**: `return fmt.Errorf("context: %w", err)` and structured `global.Logger` usage.

## Verification Plan
1. `go build ./...`
2. `go test -race ./...`
3. Manual log inspection.
