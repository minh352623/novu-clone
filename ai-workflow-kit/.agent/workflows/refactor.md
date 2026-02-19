# Workflow: /refactor

**Trigger:** User types `/refactor [target] "[goal]"`
**Examples:**
- `/refactor internal/auth/application/service/ "separate email sending into async event"`
- `/refactor internal/order/ "fix layer violations — service directly imports infrastructure"`
- `/refactor internal/user/infrastructure/persistence/repository/user_repository.go "migrate from raw SQL to sqlc"`

---

## Step 1 — Understand the goal

Ask the user:
1. What is the specific problem with the current code?
2. What is the desired outcome? (better testability / fix violation / improve performance / reduce coupling)
3. Are there any constraints? (must not change public API / must not break existing tests / time-boxed)

---

## Step 2 — Analyze current state

Before proposing any changes:
1. Read the target files carefully
2. Identify ALL callers of the code being refactored
3. Identify which tests exist for the current code
4. Map the current layer/dependency structure

Report to user:
```
Current state analysis:
- Pattern violations found: [list]
- Callers affected: [list files/functions]
- Existing tests: [list test files]
- Risk level: Low / Medium / High
```

---

## Step 3 — Refactor strategy

Choose the safest strategy:

**Strategy A — Parallel implementation (safest, for large refactors)**
1. Create new implementation alongside old
2. Run both in parallel with feature flag
3. Migrate callers one by one
4. Remove old implementation

**Strategy B — Extract and delegate (medium risk)**
1. Extract logic into new function/type
2. Old code delegates to new
3. Gradually move callers
4. Remove delegation wrapper

**Strategy C — Direct replacement (only for isolated, well-tested code)**
1. Ensure tests exist (write them if not)
2. Refactor in small, atomic steps
3. Run tests after each step

Present the chosen strategy and WHY. Wait for approval.

---

## Step 4 — Layer violation refactors (common case)

When fixing cross-layer violations:

**Application imports Infrastructure directly:**
```go
// Before (violation)
// application/service/order_service.go
import "project/internal/order/infrastructure/persistence/repository"
// ...
func (s *OrderService) CreateOrder(ctx context.Context, ...) {
    r := repository.NewOrderRepository(s.db)  // ← directly using infra
}

// After (compliant)
// 1. Create interface in domain
// domain/repository/i_order_repository.go
type IOrderRepository interface { Save(ctx, order) error; ... }

// 2. Service uses interface
// application/service/order_service.go
type OrderService struct {
    repo domain.IOrderRepository  // interface, not concrete
}

// 3. Wire in module init
// order.module.go → inject infrastructure.NewOrderRepository()
```

**Module imports another module's repository:**
```
See .ai/context/PATTERNS.md → Pattern 1: Cross-module Communication
Follow the Port + Adapter pattern exactly.
```

---

## Step 5 — Refactor rules

During refactor:
- Make ONE type of change at a time (don't mix renaming + logic change)
- Keep public API signatures identical unless explicitly agreed
- Add/update tests before deleting old code
- Each commit should leave the code in a working state

---

## Step 6 — Verification

After refactor:
```bash
# Run existing tests — none should break
go test ./internal/[module]/... -v

# Run race detector
go test -race ./internal/[module]/...

# Check no new layer violations
grep -r "infrastructure" internal/[module]/domain/ || echo "Clean"
grep -r "gorm.io" internal/ || echo "No GORM"

# Run linter
go vet ./internal/[module]/...
```

Report results to user.
