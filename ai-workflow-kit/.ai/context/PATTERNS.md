# Code Patterns Reference

> **AI instruction:** These are APPROVED patterns. When writing code, match these patterns exactly.
> Do not invent new patterns — if you think a new pattern is needed, flag it and ask the user.

---

## Pattern 1: Cross-Module Communication (Port + Adapter)

Use when: Module A needs data from Module B.

### Step 1: Define Port in CONSUMER module
```go
// internal/ORDER/domain/repository/i_user_reader.go
// This interface lives in ORDER, not in USER
package repository

import "context"

// UserBasicInfo — only fields ORDER needs from USER
// Do NOT expose User entity with password, internal fields
type UserBasicInfo struct {
    ID    string
    Name  string
    Email string
}

type IUserReader interface {
    FindByID(ctx context.Context, id string) (*UserBasicInfo, error)
}
```

### Step 2: Service depends on Port (interface only)
```go
// internal/order/application/service/order_service.go
type OrderService struct {
    userReader domainRepo.IUserReader // interface, knows nothing about UserModule internals
}
```

### Step 3: Implement Local Adapter in CONSUMER module
```go
// internal/order/infrastructure/adapter/local_user_adapter.go
package adapter

import (
    "context"
    "fmt"
    
    userService "project/internal/user/application/service"
    domainRepo  "project/internal/order/domain/repository"
)

type LocalUserAdapter struct {
    userSvc userService.IUserService  // import SERVICE, not repository
}

func NewLocalUserAdapter(svc userService.IUserService) domainRepo.IUserReader {
    return &LocalUserAdapter{userSvc: svc}
}

func (a *LocalUserAdapter) FindByID(ctx context.Context, id string) (*domainRepo.UserBasicInfo, error) {
    user, err := a.userSvc.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("localUserAdapter.FindByID: %w", err)
    }
    // Map to minimal DTO — hide internal fields
    return &domainRepo.UserBasicInfo{ID: user.ID, Name: user.Name, Email: user.Email}, nil
}
```

### Step 4: Wire in module init
```go
// internal/order/order.module.go
func NewModule(pool *pgx.Pool, userSvc userService.IUserService) *Module {
    userReader := adapter.NewLocalUserAdapter(userSvc)  // ← monolith
    // userReader := adapter.NewHttpUserAdapter(cfg.UserServiceURL)  // ← microservice swap
    
    orderRepo := repository.NewOrderRepository(sqlcgen.New(pool))
    orderSvc := service.NewOrderService(orderRepo, userReader)
    // ...
}
```

**When going microservice:** Only change the adapter wiring in `order.module.go`. OrderService is untouched.

---

## Pattern 2: Background Worker (Goroutine)

Use when: Need a long-running goroutine started at app boot.

```go
// internal/notification/application/service/notification_service.go

func (s *NotificationService) StartDeliveryWorker(ctx context.Context) {
    go func() {
        // Detach from request context — keeps trace values, no cancellation propagation
        workerCtx := context.WithoutCancel(ctx)  // Go 1.21+
        
        defer func() {
            if r := recover(); r != nil {
                slog.Error("delivery worker panic",
                    "err", r,
                    "stack", string(debug.Stack()),
                )
            }
        }()
        
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        
        slog.Info("notification delivery worker started")
        for {
            select {
            case <-ctx.Done():  // original ctx for shutdown signal
                slog.Info("delivery worker shutting down")
                return
            case <-ticker.C:
                if err := s.processDeliveryQueue(workerCtx); err != nil {
                    slog.Error("delivery queue processing failed", "err", err)
                    // Don't return — keep retrying
                }
            }
        }
    }()
}
```

---

## Pattern 3: Event-Driven Side Effects

Use when: Service action triggers async side effects (email, notification, analytics).

```go
// Define event (in shared event package or in emitting module's domain)
// internal/order/domain/event/order_events.go
type OrderCreatedEvent struct {
    OrderID   string
    UserID    string
    Amount    int64
    CreatedAt time.Time
}

// Emit in service (non-blocking)
// internal/order/application/service/order_service.go
func (s *OrderService) CreateOrder(ctx context.Context, req *dto.CreateOrderReq) (*dto.OrderResp, error) {
    // ... create and save order
    
    // Emit — does not block, does not wait for listeners
    s.eventBus.Emit("order.created", &event.OrderCreatedEvent{
        OrderID:   order.ID,
        UserID:    order.UserID,
        Amount:    order.TotalAmount,
        CreatedAt: order.CreatedAt,
    })
    
    return toResp(order), nil
}

// Handle in listener (in notification module or email module)
// internal/notification/application/service/notification_listener.go
type NotificationListener struct {
    notifSvc INotificationService
}

func (l *NotificationListener) Register(bus event.IBus) {
    bus.On("order.created", l.handleOrderCreated)
}

func (l *NotificationListener) handleOrderCreated(payload any) {
    e, ok := payload.(*orderEvent.OrderCreatedEvent)
    if !ok { return }
    
    ctx := context.Background()
    if err := l.notifSvc.SendOrderConfirmation(ctx, e.OrderID, e.UserID); err != nil {
        slog.Error("failed to send order confirmation",
            "order_id", e.OrderID,
            "err", err,
        )
    }
}
```

---

## Pattern 4: Pagination

Use for: All list endpoints. No endpoint returns unbounded data.

```go
// Domain filter
type OrderFilter struct {
    UserID   string
    Status   *entity.OrderStatus  // nil = all statuses
    Page     int                  // 1-based
    PageSize int                  // max 100
}

// Application service
func (s *OrderService) List(ctx context.Context, f *dto.ListOrdersFilter) (*dto.PaginatedResp[dto.OrderResp], error) {
    if f.PageSize <= 0 || f.PageSize > 100 {
        f.PageSize = 20  // default
    }
    if f.Page <= 0 {
        f.Page = 1
    }
    
    orders, total, err := s.repo.List(ctx, domainRepo.OrderFilter{
        UserID:   f.UserID,
        Page:     f.Page,
        PageSize: f.PageSize,
    })
    if err != nil {
        return nil, fmt.Errorf("orderService.List: %w", err)
    }
    
    resp := make([]dto.OrderResp, len(orders))
    for i, o := range orders {
        resp[i] = *toResp(o)
    }
    
    return &dto.PaginatedResp[dto.OrderResp]{
        Data:  resp,
        Total: total,
        Page:  f.Page,
        Limit: f.PageSize,
    }, nil
}
```

---

## Pattern 5: Unit of Work (Multi-table Transaction)

Use when: A use case must write to multiple tables atomically.

```go
// Domain interface — application layer knows only this
// internal/common/uow/i_unit_of_work.go
type TransactionalRepos struct {
    OrderRepo     order.IOrderRepository
    InventoryRepo inventory.IInventoryRepository
}

type IUnitOfWork interface {
    Execute(ctx context.Context, fn func(repos *TransactionalRepos) error) error
}

// Infrastructure implementation
// internal/common/uow/pgx_unit_of_work.go
type PgxUnitOfWork struct {
    pool    *pgx.Pool
    newRepos func(q *sqlcgen.Queries) *TransactionalRepos
}

func (u *PgxUnitOfWork) Execute(ctx context.Context, fn func(*TransactionalRepos) error) error {
    tx, err := u.pool.Begin(ctx)
    if err != nil { return fmt.Errorf("begin tx: %w", err) }
    defer tx.Rollback(ctx)
    
    repos := u.newRepos(sqlcgen.New(tx))
    if err := fn(repos); err != nil {
        return err // tx.Rollback called by defer
    }
    
    return tx.Commit(ctx)
}

// Usage in service
func (s *OrderService) PlaceOrder(ctx context.Context, req *dto.PlaceOrderReq) error {
    return s.uow.Execute(ctx, func(repos *uow.TransactionalRepos) error {
        if err := repos.InventoryRepo.Reserve(ctx, req.ProductID, req.Quantity); err != nil {
            return err // rollback triggered
        }
        if err := repos.OrderRepo.Save(ctx, order); err != nil {
            return err // rollback triggered
        }
        return nil // commit
    })
}
```

---

## Pattern 6: Table-Driven Unit Test

Use for: All service method tests.

```go
func TestOrderService_CreateOrder(t *testing.T) {
    tests := []struct {
        name      string
        setup     func(*mocks.IOrderRepository, *mocks.IUserReader)
        input     *dto.CreateOrderReq
        wantErr   bool
        wantErrIs error         // specific domain error expected
        validate  func(*testing.T, *dto.OrderResp)
    }{
        {
            name: "success — creates order and returns response",
            setup: func(repo *mocks.IOrderRepository, userReader *mocks.IUserReader) {
                userReader.On("FindByID", mock.Anything, "user-1").
                    Return(&domainRepo.UserBasicInfo{ID: "user-1", Name: "Alice"}, nil)
                repo.On("Save", mock.Anything, mock.AnythingOfType("*entity.Order")).
                    Return(nil)
            },
            input:   &dto.CreateOrderReq{UserID: "user-1", ProductID: "prod-1", Quantity: 2},
            wantErr: false,
            validate: func(t *testing.T, resp *dto.OrderResp) {
                assert.NotEmpty(t, resp.ID)
                assert.Equal(t, "pending", resp.Status)
            },
        },
        {
            name: "user not found — returns ErrUserNotFound",
            setup: func(repo *mocks.IOrderRepository, userReader *mocks.IUserReader) {
                userReader.On("FindByID", mock.Anything, "unknown").
                    Return(nil, domain.ErrUserNotFound)
            },
            input:     &dto.CreateOrderReq{UserID: "unknown", ProductID: "prod-1", Quantity: 1},
            wantErr:   true,
            wantErrIs: domain.ErrUserNotFound,
        },
        {
            name: "repository save fails — returns wrapped error",
            setup: func(repo *mocks.IOrderRepository, userReader *mocks.IUserReader) {
                userReader.On("FindByID", mock.Anything, "user-1").
                    Return(&domainRepo.UserBasicInfo{ID: "user-1"}, nil)
                repo.On("Save", mock.Anything, mock.Anything).
                    Return(errors.New("connection refused"))
            },
            input:   &dto.CreateOrderReq{UserID: "user-1", ProductID: "prod-1", Quantity: 1},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo   := mocks.NewIOrderRepository(t)
            mockReader := mocks.NewIUserReader(t)
            tt.setup(mockRepo, mockReader)
            
            svc := service.NewOrderService(mockRepo, mockReader)
            
            resp, err := svc.CreateOrder(context.Background(), tt.input)
            
            if tt.wantErr {
                require.Error(t, err)
                if tt.wantErrIs != nil {
                    assert.ErrorIs(t, err, tt.wantErrIs)
                }
                return
            }
            
            require.NoError(t, err)
            require.NotNil(t, resp)
            if tt.validate != nil {
                tt.validate(t, resp)
            }
            mockRepo.AssertExpectations(t)
            mockReader.AssertExpectations(t)
        })
    }
}
```
