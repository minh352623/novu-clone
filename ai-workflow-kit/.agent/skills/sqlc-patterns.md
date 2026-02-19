# Skill: sqlc Patterns

**Load when:** Creating or modifying database queries, setting up repository, writing SQL.

---

## What is sqlc and why we use it

sqlc generates type-safe Go code from SQL queries. You write SQL, sqlc generates Go functions.
This project uses sqlc EXCLUSIVELY — no GORM, no raw pgx in service layer.

---

## File structure for sqlc

```
internal/[module]/
└── infrastructure/
    └── persistence/
        ├── queries/
        │   └── [entity].sql          ← You write SQL here
        ├── sqlcgen/                  ← sqlc generates this (DO NOT EDIT)
        │   ├── db.go
        │   ├── models.go
        │   └── [entity].sql.go
        └── repository/
            └── [entity]_repository.go  ← You use sqlcgen here
```

---

## sqlc.yaml configuration

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/*/infrastructure/persistence/queries"
    schema: "sql/schema"
    gen:
      go:
        package: "sqlcgen"
        out: "internal/[module]/infrastructure/persistence/sqlcgen"
        emit_json_tags: true
        emit_db_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
        emit_empty_slices: true
        emit_exported_queries: false
        emit_result_struct_pointers: true
        emit_params_struct_pointers: true
        overrides:
          - db_type: "uuid"
            go_type: "github.com/google/uuid.UUID"
          - db_type: "timestamptz"
            go_type: "time.Time"
```

---

## SQL query patterns

### Basic CRUD
```sql
-- name: GetOrderByID :one
SELECT id, user_id, status, total_amount, created_at, updated_at
FROM orders
WHERE id = $1 AND deleted_at IS NULL;

-- name: CreateOrder :one
INSERT INTO orders (id, user_id, status, total_amount, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING *;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2, updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteOrder :exec
UPDATE orders
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1;
```

### Pagination pattern
```sql
-- name: ListOrdersByUserID :many
SELECT id, user_id, status, total_amount, created_at
FROM orders
WHERE user_id = $1
  AND deleted_at IS NULL
  AND ($3::text IS NULL OR status = $3::order_status)  -- optional filter
ORDER BY created_at DESC
LIMIT $2 OFFSET ($4::int * $2);

-- name: CountOrdersByUserID :one
SELECT COUNT(*)
FROM orders
WHERE user_id = $1
  AND deleted_at IS NULL
  AND ($2::text IS NULL OR status = $2::order_status);
```

### Batch operations
```sql
-- name: GetOrdersByIDs :many
SELECT * FROM orders
WHERE id = ANY($1::uuid[])
  AND deleted_at IS NULL;

-- name: BulkUpdateOrderStatus :exec
UPDATE orders
SET status = $2, updated_at = NOW()
WHERE id = ANY($1::uuid[]);
```

### Join query
```sql
-- name: GetOrderWithItems :many
SELECT
    o.id AS order_id,
    o.status,
    oi.id AS item_id,
    oi.product_id,
    oi.quantity,
    oi.unit_price
FROM orders o
JOIN order_items oi ON oi.order_id = o.id
WHERE o.id = $1 AND o.deleted_at IS NULL;
```

---

## Repository implementation patterns

### Using sqlc in repository
```go
package repository

import (
    "context"
    "errors"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/google/uuid"
    
    "project/internal/order/domain"
    domainRepo "project/internal/order/domain/repository"
    "project/internal/order/domain/model/entity"
    "project/internal/order/infrastructure/persistence/sqlcgen"
)

type OrderRepository struct {
    q *sqlcgen.Queries
}

func NewOrderRepository(q *sqlcgen.Queries) domainRepo.IOrderRepository {
    return &OrderRepository{q: q}
}

func (r *OrderRepository) FindByID(ctx context.Context, id string) (*entity.Order, error) {
    uid, err := uuid.Parse(id)
    if err != nil {
        return nil, domain.ErrOrderNotFound // invalid UUID = not found
    }
    
    row, err := r.q.GetOrderByID(ctx, uid)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, domain.ErrOrderNotFound
        }
        return nil, fmt.Errorf("orderRepo.FindByID id=%s: %w", id, err)
    }
    
    return r.mapToEntity(row), nil
}

func (r *OrderRepository) List(ctx context.Context, filter domainRepo.OrderFilter) ([]*entity.Order, int64, error) {
    // Build sqlc params from domain filter
    params := sqlcgen.ListOrdersByUserIDParams{
        UserID:  uuid.MustParse(filter.UserID),
        Limit:   int32(filter.PageSize),
        Column3: toNullString(filter.Status),  // optional filter
        Column4: int32(filter.Page - 1),        // 0-indexed offset
    }
    
    rows, err := r.q.ListOrdersByUserID(ctx, params)
    if err != nil {
        return nil, 0, fmt.Errorf("orderRepo.List: %w", err)
    }
    
    count, err := r.q.CountOrdersByUserID(ctx, sqlcgen.CountOrdersByUserIDParams{
        UserID:  uuid.MustParse(filter.UserID),
        Column2: toNullString(filter.Status),
    })
    if err != nil {
        return nil, 0, fmt.Errorf("orderRepo.List count: %w", err)
    }
    
    orders := make([]*entity.Order, len(rows))
    for i, row := range rows {
        orders[i] = r.mapListRowToEntity(row)
    }
    
    return orders, count, nil
}

// mapToEntity: private, converts sqlc row → domain entity
// Never expose sqlc types outside this file
func (r *OrderRepository) mapToEntity(row *sqlcgen.GetOrderByIDRow) *entity.Order {
    return &entity.Order{
        ID:          row.ID.String(),
        UserID:      row.UserID.String(),
        Status:      entity.OrderStatus(row.Status),
        TotalAmount: row.TotalAmount,
        CreatedAt:   row.CreatedAt.Time,
        UpdatedAt:   row.UpdatedAt.Time,
    }
}
```

### Transaction pattern
```go
// When a use case needs to write to multiple tables atomically
// Use the UnitOfWork pattern — see PATTERNS.md

// Inside a transaction, pass the *sqlcgen.Queries created from the tx:
func (u *TypeOrmUnitOfWork) Execute(ctx context.Context, fn func(q *sqlcgen.Queries) error) error {
    tx, err := u.pool.Begin(ctx)
    if err != nil { return fmt.Errorf("begin tx: %w", err) }
    defer tx.Rollback(ctx) // no-op if already committed
    
    qtx := u.q.WithTx(tx)
    if err := fn(qtx); err != nil {
        return err
    }
    
    return tx.Commit(ctx)
}
```

---

## After modifying .sql files

Always run:
```bash
sqlc generate
go build ./...  # verify generated code compiles
```
