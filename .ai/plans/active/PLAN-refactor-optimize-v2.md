# PLAN: Comprehensive Codebase Refactoring & Optimization (Round 2)

This plan addresses remaining technical debt identified during a systematic audit of all modules, ensuring 100% compliance with `.ai/INSTRUCTIONS.md` and `GOLANG_BEST_PRACTICES.md`.

## Review Levels
- **Level**: L2 (Cross-module Changes)
- **Impact**: All core modules (`IAM`, `Apps`, `Messaging`, `Notification`, `Workflow`) and `Infrastructure`.
- **Review Required**: Lead (Standardization follow-up).

## Proposed Changes

### Phase 1: Systematic Error Wrapping (BP 3.2)
Wrap all bare error returns with context-rich messages using `%w`.
- **[IAM]**: `role.service.impl.go`, `tenant.service.impl.go`, all user repositories.
- **[Messaging]**: `subscriberRepository` and `threadRepository` in `common.repository.go`, `message.repository.go`.
- **[Workflow]**: `workflow.repository.go`, `execution.repository.go`.
- **[Infrastructure]**: `persistence/plugin/rls.go`.

### Phase 2: Performance Optimization (BP 5.1)
Implement slice pre-allocation in all DTO mapping and collection loops.
- **[IAM]**: Audit all `To...ResponseList` functions in `iam.dto.go`.
- **[Messaging]**: Audit repository `List` and `Get` methods for pre-allocation.
- **[Apps]**: Correct any remaining `append` loops in metrics repositories.
- **[Notification]**: Check `job.controller.go` and mapper files.

### Phase 3: Global Logger & Panic Safety (BP 4.4, 6.1)
- Ensure all `go func()` blocks have a `recover()` with structured `global.Logger` usage.
- Standardize log keys (use `zap.Error(err)` instead of string concatenation).

### Phase 4: Controller Standardisation (BP 10)
Systematic check to ensure all controllers return `(interface{}, error)` signature to work with the global `response.Wrap` pattern.

## Impact Analysis
- **Breaking Changes**: None. This is a pure refactoring and optimization.
- **Observability**: Improved error context in logs and consistent structured fields.
- **Performance**: Reduced GC pressure due to efficient slice allocations.

## Verification Plan
1. **Build**: `go build ./...`
2. **Race Test**: `go test -race ./...`
3. **Manual Check**: Verify log output for a few sample flows to ensure structured fields are present.
