# Workflow: /create

**Trigger:** User types `/create [type] [name] "[description]"`
**Examples:**
- `/create module notification "send email, SMS, and push notifications"`
- `/create endpoint "POST /api/v1/notifications" "create and queue a notification"`
- `/create repository INotificationRepository "CRUD for notification entity"`

---

## Supported create types

| Type | What gets created |
|------|-------------------|
| `module` | Full DDD module (all 4 layers) |
| `endpoint` | Handler + DTO + service method for 1 API endpoint |
| `repository` | Domain interface + sqlc queries + infrastructure impl |
| `service` | Application service interface + implementation skeleton |
| `entity` | Domain entity + value objects |

---

## Create type: module (full DDD module)

### Step 1 — Clarify requirements

Ask BEFORE writing any code:

1. **Entities:** What are the main domain entities? List their key fields.
2. **Use cases:** What are the primary operations (create, read, update, delete, special business actions)?
3. **External dependencies:** Does this module call external APIs, SMTP, queues?
4. **Cross-module needs:** Does this module need data from another module?
5. **API surface:** What HTTP endpoints will be exposed?
6. **Database:** New tables needed? Relationships?

### Step 2 — Scaffold in strict order

Create files in this exact sequence:

```
Phase 1: Domain Layer (no dependencies)
├── internal/[module]/domain/model/entity/[entity].go
├── internal/[module]/domain/model/vo/[value_object].go   (if needed)
├── internal/[module]/domain/repository/i_[entity]_repository.go
└── internal/[module]/domain/errors.go

Phase 2: Database
├── internal/[module]/infrastructure/persistence/queries/[entity].sql
└── Run: sqlc generate  (instruct user to run this)

Phase 3: Infrastructure Layer
├── internal/[module]/infrastructure/persistence/repository/[entity]_repository.go
└── internal/[module]/infrastructure/cache/[entity]_cache.go  (if needed)

Phase 4: Application Layer
├── internal/[module]/application/service/i_[module]_service.go
└── internal/[module]/application/service/[module]_service.go

Phase 5: Interface Layer
├── internal/[module]/controller/dto/[action]_request.go
├── internal/[module]/controller/dto/[action]_response.go
└── internal/[module]/controller/http/[module]_handler.go

Phase 6: Module wiring
└── internal/[module]/[module].module.go
    (wire: newRepository → newService → newHandler → register routes)

Phase 7: Tests
└── internal/[module]/application/service/[module]_service_test.go
```

### Step 3 — File templates

**Domain Entity:**
```go
// internal/[module]/domain/model/entity/[entity].go
package entity

import "time"

type [Entity] struct {
    ID        string
    // ... fields
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Business methods belong here (not in service)
func (e *[Entity]) IsValid() bool { ... }
```

**Repository Interface:**
```go
// internal/[module]/domain/repository/i_[entity]_repository.go
package repository

import "context"

type I[Entity]Repository interface {
    FindByID(ctx context.Context, id string) (*entity.[Entity], error)
    Save(ctx context.Context, e *entity.[Entity]) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter [Entity]Filter) ([]*entity.[Entity], int64, error)
}

type [Entity]Filter struct {
    Page     int
    PageSize int
    // domain-specific filters
}
```

**Domain Errors:**
```go
// internal/[module]/domain/errors.go
package domain

import "errors"

var (
    Err[Entity]NotFound    = errors.New("[entity] not found")
    Err[Entity]Duplicate   = errors.New("[entity] already exists")
    Err[Entity]InvalidState = errors.New("[entity] is in invalid state")
)
```

**sqlc SQL file:**
```sql
-- internal/[module]/infrastructure/persistence/queries/[entity].sql
-- name: Get[Entity]ByID :one
SELECT * FROM [table] WHERE id = $1 AND deleted_at IS NULL;

-- name: Create[Entity] :one
INSERT INTO [table] (id, ..., created_at, updated_at)
VALUES ($1, ..., NOW(), NOW())
RETURNING *;

-- name: Update[Entity] :one
UPDATE [table]
SET ..., updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: SoftDelete[Entity] :exec
UPDATE [table] SET deleted_at = NOW() WHERE id = $1;

-- name: List[Entity]s :many
SELECT * FROM [table]
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: Count[Entity]s :one
SELECT COUNT(*) FROM [table] WHERE deleted_at IS NULL;
```

**Repository Implementation:**
```go
// internal/[module]/infrastructure/persistence/repository/[entity]_repository.go
package repository

import (
    "context"
    "errors"
    "fmt"
    
    "github.com/jackc/pgx/v5"
    "[module]/internal/[module]/domain"
    domainRepo "[module]/internal/[module]/domain/repository"
    domainEntity "[module]/internal/[module]/domain/model/entity"
    "[module]/internal/[module]/infrastructure/persistence/sqlcgen"
)

type [Entity]Repository struct {
    q *sqlcgen.Queries
}

func New[Entity]Repository(q *sqlcgen.Queries) domainRepo.I[Entity]Repository {
    return &[Entity]Repository{q: q}
}

func (r *[Entity]Repository) FindByID(ctx context.Context, id string) (*domainEntity.[Entity], error) {
    row, err := r.q.Get[Entity]ByID(ctx, id)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, domain.Err[Entity]NotFound
        }
        return nil, fmt.Errorf("[entity]Repo.FindByID id=%s: %w", id, err)
    }
    return mapRowToEntity(row), nil
}

func mapRowToEntity(row sqlcgen.Get[Entity]ByIDRow) *domainEntity.[Entity] {
    return &domainEntity.[Entity]{
        ID: row.ID.String(),
        // map other fields
    }
}
```

---

## Create type: endpoint (single API endpoint)

When user runs `/create endpoint`:

1. Ask: What HTTP method and path?
2. Ask: What does it receive in request body / query params?
3. Ask: What does it return on success?
4. Ask: What business logic does it trigger?

Create in order:
1. Request DTO + validation tags
2. Response DTO
3. Service method (add to interface + implementation)
4. Handler method
5. Route registration
6. Unit test for service method

---

## Verification after create

After scaffolding, remind user to:
```bash
# If new SQL queries were added:
sqlc generate

# Run linter
go vet ./internal/[module]/...

# Run module tests
go test ./internal/[module]/... -v
```
