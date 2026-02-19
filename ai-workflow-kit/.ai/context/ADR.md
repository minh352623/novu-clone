# Architecture Decision Records (ADR)

> **Format:** ADR-NNN | Title | Status | Date
> **AI instruction:** Check this file before implementing anything that touches database, auth, error handling, or module structure. If your task would violate an ADR, flag it and ask the user before proceeding.

---

## ADR-001 | Use sqlc instead of GORM | Accepted | 2025-01-10

**Context:** Need a database layer that is type-safe, performant, and doesn't hide SQL complexity.

**Decision:** Use sqlc to generate Go code from SQL queries. No ORM (GORM) allowed.

**Consequences:**
- ✅ Type-safe queries caught at compile time
- ✅ SQL is visible and reviewable in version control
- ✅ No N+1 queries hidden by magic lazy loading
- ❌ More verbose than GORM for simple CRUD
- ❌ Schema changes require sqlc regeneration

**Enforcement:** `grep -r "gorm.io" internal/` in CI must return empty.

---

## ADR-002 | DDD 4-layer architecture per module | Accepted | 2025-01-10

**Context:** Need a structure that is testable, maintainable, and ready for microservice extraction.

**Decision:** Every feature module follows: Domain → Application → Infrastructure → Interface, with strict dependency rules (each layer can only import inward).

**Consequences:**
- ✅ Easy to unit test domain and application logic
- ✅ Can swap database, cache, HTTP framework without touching business logic
- ✅ Clear ownership of code
- ❌ More boilerplate than simple CRUD layering
- ❌ New developers need onboarding

**Enforcement:** Architecture review in every PR. Layer violation = PR blocked.

---

## ADR-003 | Cross-module communication via Port+Adapter | Accepted | 2025-01-15

**Context:** Modules need to share data without tight coupling. Must be extractable to microservices.

**Decision:** When module A needs data from module B, module A defines an abstract interface (Port) in its own domain layer. Module A also provides a local adapter implementation. When splitting to microservices, swap LocalAdapter → HttpAdapter without touching business logic.

**Consequences:**
- ✅ Zero logic change when going from monolith to microservice
- ✅ Easy to mock in tests
- ✅ No circular imports
- ❌ More files per cross-module interaction

**Enforcement:** Direct imports of another module's repository = PR blocked immediately.

---

## ADR-004 | Service layer must not throw HTTP exceptions | Accepted | 2025-01-15

**Context:** Services are reused across HTTP handlers and gRPC handlers. HTTP status codes are a transport concern.

**Decision:** Services return domain errors (sentinel `var Err... = errors.New(...)`) only. HTTP mapping happens in `response.HandleError()` at the interface layer.

**Consequences:**
- ✅ Services fully reusable across any transport (HTTP, gRPC, CLI, tests)
- ✅ HTTP mapping logic in one place (easy to change)
- ❌ Developers must remember to add new domain errors to the HTTP mapping switch

**Enforcement:** Code review. `grep -r "net/http" internal/*/application/` must return empty.

---

## ADR-005 | Async side effects via in-process event bus | Accepted | 2025-02-01

**Context:** Sending emails, notifications, analytics events should not block the request-response cycle.

**Decision:** Use in-process event bus (EventEmitter pattern) for async side effects within the monolith. Do not `await`/`go func()` directly from service — emit named events with typed payload.

**Rationale:** Avoids context leak (request ctx cancelled before email sends). Provides clean separation. Easily swapped for external message broker (RabbitMQ, Kafka) when scaling.

**Consequences:**
- ✅ Response time unaffected by email/notification delivery
- ✅ Event handlers are independently testable
- ✅ Adding new side effects = new listener, no change to core service
- ❌ Events are fire-and-forget in-process (no persistence). For critical events, add a dedicated outbox table.

---

## ADR-006 | Use slog for structured logging | Accepted | 2025-02-01

**Context:** Need structured logging for log aggregation tools (Datadog, Loki).

**Decision:** Use Go 1.21's `log/slog` stdlib package. No third-party logging library (zap, logrus) unless benchmarks prove insufficient.

**Pattern:**
```go
slog.Info("order created", "order_id", order.ID, "user_id", userID, "amount", order.Amount)
slog.Error("payment failed", "order_id", order.ID, "err", err)
```

**Consequences:**
- ✅ Zero dependency, standard library
- ✅ JSON output in production, human-readable in development
- ✅ Structured fields (not string interpolation)

---

## ADR-007 | Soft delete pattern for all user-facing entities | Accepted | 2025-02-10

**Context:** Business requires audit trail and ability to restore deleted records.

**Decision:** All user-facing entities use soft delete: `deleted_at TIMESTAMPTZ NULL`. Physical deletion only for compliance (GDPR right to erasure — handled separately).

**SQL pattern:** All queries must include `AND deleted_at IS NULL`.
**sqlc:** Use a shared base query or view that filters deleted records.

**Enforcement:** Code review. Any query missing `deleted_at IS NULL` filter = PR blocked.

---

## ADR Template (for new decisions)

```markdown
## ADR-NNN | [Title] | [Proposed/Accepted/Deprecated/Superseded] | [Date]

**Context:** [What problem are we solving? What forces are at play?]

**Decision:** [What did we decide to do?]

**Consequences:**
- ✅ [Positive consequence]
- ❌ [Negative consequence / trade-off]

**Enforcement:** [How do we prevent violations?]
```
