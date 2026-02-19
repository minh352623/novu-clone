# Skill: Error Handling in Go DDD

**Load when:** Writing any function that can return error, reviewing error handling code.

---

## The error hierarchy

```
Infrastructure errors (pgx.ErrNoRows, network timeout)
    ↓ wrapped by
Repository layer (fmt.Errorf("orderRepo.FindByID id=%s: %w", id, err))
    ↓ wrapped by (or mapped to domain error)
Application layer (fmt.Errorf("orderService.CreateOrder: %w", err))
    ↓ read by
Handler layer (response.HandleError maps domain errors → HTTP status)
```

---

## Domain errors — sentinel variables

```go
// internal/[module]/domain/errors.go
package domain

import "errors"

var (
    // Not found errors
    ErrOrderNotFound       = errors.New("order not found")
    ErrUserNotFound        = errors.New("user not found")
    
    // Business rule violations
    ErrOrderAlreadyPaid    = errors.New("order already paid")
    ErrInsufficientBalance = errors.New("insufficient balance")
    ErrInvalidOrderStatus  = errors.New("invalid order status transition")
    
    // Input errors
    ErrInvalidAmount       = errors.New("amount must be positive")
    ErrDuplicateEmail      = errors.New("email already registered")
)
```

---

## Error wrapping rules

### Repository layer
```go
// Pattern: "[module]Repo.[Method] [key context]: %w"
func (r *OrderRepository) FindByID(ctx context.Context, id string) (*entity.Order, error) {
    row, err := r.q.GetOrderByID(ctx, uid)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, domain.ErrOrderNotFound  // map to domain error, NO wrapping
        }
        return nil, fmt.Errorf("orderRepo.FindByID id=%s: %w", id, err)  // wrap infra error
    }
    return r.mapToEntity(row), nil
}
```

### Application layer
```go
// Pattern: "[module]Service.[Method]: %w"
func (s *OrderService) GetOrder(ctx context.Context, id string, requesterID string) (*dto.OrderResp, error) {
    order, err := s.orderRepo.FindByID(ctx, id)
    if err != nil {
        // Don't add more context if domain error — it's already descriptive
        return nil, fmt.Errorf("orderService.GetOrder: %w", err)
    }
    
    if order.UserID != requesterID {
        return nil, domain.ErrOrderNotFound  // security: return NotFound not Forbidden
    }
    
    return toResponse(order), nil
}
```

### Interface layer (handler)
```go
// Handler maps domain errors to HTTP status — does NOT add more error context
func (h *OrderHandler) GetByID(c *gin.Context) {
    order, err := h.orderSvc.GetOrder(c.Request.Context(), c.Param("id"), currentUserID(c))
    if err != nil {
        response.HandleError(c, err)  // central mapping function
        return
    }
    c.JSON(http.StatusOK, response.Success(order))
}

// response/handler.go — central error mapper
func HandleError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, domain.ErrOrderNotFound),
         errors.Is(err, domain.ErrUserNotFound):
        c.JSON(http.StatusNotFound, Error(err))
        
    case errors.Is(err, domain.ErrOrderAlreadyPaid),
         errors.Is(err, domain.ErrInvalidOrderStatus):
        c.JSON(http.StatusConflict, Error(err))
        
    case errors.Is(err, domain.ErrInsufficientBalance),
         errors.Is(err, domain.ErrInvalidAmount):
        c.JSON(http.StatusUnprocessableEntity, Error(err))
        
    default:
        // Log the full error chain (internal, not exposed to client)
        slog.Error("unhandled error", "err", err, "path", c.Request.URL.Path)
        c.JSON(http.StatusInternalServerError, Error(errors.New("internal server error")))
    }
}
```

---

## Error checking rules

### Always use errors.Is for sentinel errors
```go
// ❌ Wrong: string comparison breaks error wrapping
if err.Error() == "order not found" { ... }

// ❌ Wrong: type assertion misses wrapped errors
if _, ok := err.(*OrderNotFoundError); ok { ... }

// ✅ Correct: works through any number of wrapping layers
if errors.Is(err, domain.ErrOrderNotFound) { ... }
```

### Always check error before using result
```go
// ❌ Wrong: might panic if err != nil and result is nil
result, err := someFunc()
process(result.Field)  // nil pointer if err != nil

// ✅ Correct: always check error first
result, err := someFunc()
if err != nil {
    return nil, fmt.Errorf("context: %w", err)
}
process(result.Field)
```

### Never swallow errors
```go
// ❌ Never do this
result, _ := someFunc()  // error silently ignored

// ❌ Never do this in a goroutine
go func() {
    err := doWork()
    // err ignored — bug silently fails
}()

// ✅ If you truly can't handle it, log it
go func() {
    if err := doWork(); err != nil {
        slog.Error("background work failed", "err", err)
    }
}()
```
