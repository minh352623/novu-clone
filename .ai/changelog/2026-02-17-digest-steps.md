# [2026-02-17] Workflow Engine Phase 3: Digest Steps
**Author**: Minh  
**Module**: `internal/workflow/`

---

## What Changed

Implemented digest step support for the Workflow Engine. A **digest step** batches multiple trigger events within a configurable time window, then forwards the aggregated batch to the next step (typically a channel step that sends a digest notification).

### Config Format
```json
{ "step_type": "digest", "config": { "window": "1h" } }
```

---

## New Files
| File | Purpose |
|------|---------|
| `sql/schema/00027_digest_events.sql` | Migration for `digest_events` table |
| `internal/workflow/application/service/impl/digest_handler.go` | `DigestHandler` implementing `StepHandler` |

## Modified Files
| File | Change |
|------|--------|
| `internal/workflow/domain/model/entity/workflow.go` | `DigestEvent` entity + `StepStatusDigesting` constant |
| `internal/workflow/domain/repository/execution.repository.go` | +3 methods: Buffer, Flush, GetDigesting |
| `internal/workflow/infrastructure/persistence/model/execution.model.go` | `DigestEventModel` GORM model |
| `internal/workflow/infrastructure/persistence/mapper/execution.mapper.go` | Digest event mappers |
| `internal/workflow/infrastructure/persistence/repository/execution.repository.go` | Repo implementations |
| `internal/workflow/application/worker/workflow_executor.go` | `digestHandler` + `pollDigest()` + `flushDigestStep()` |
| `internal/initialize/workflow/workflow.go` | Wired `DigestHandler` |

---

## Impact on Other Modules
**None** — fully contained within `internal/workflow/`. No breaking changes to external APIs.

## Test Results
```
go build ./...     → clean
go test -race      → 14 packages, 0 failures
```
