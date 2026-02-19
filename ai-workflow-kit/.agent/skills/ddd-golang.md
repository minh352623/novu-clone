# Skill: DDD in Go

**Load when:** Creating new modules, reviewing architecture, explaining layer structure.

---

## The 4 Layers and their Go packages

```
internal/[module]/
├── domain/                     # Layer 1: Domain (innermost, no dependencies)
│   ├── model/
│   │   ├── entity/             # Core business objects
│   │   └── vo/                 # Value Objects (immutable, comparable by value)
│   ├── repository/             # Repository INTERFACES (not implementations)
│   └── errors.go               # Domain-specific error variables
│
├── application/                # Layer 2: Application (depends on Domain)
│   ├── service/                # Use cases / application services
│   │   ├── i_[name]_service.go # Service interface
│   │   └── [name]_service.go   # Service implementation
│   └── schedule/               # Background jobs (optional)
│
├── infrastructure/             # Layer 3: Infrastructure (depends on Domain)
│   ├── persistence/
│   │   ├── queries/            # sqlc .sql files
│   │   ├── sqlcgen/            # sqlc generated code (do not edit)
│   │   └── repository/         # Repository implementations
│   ├── cache/                  # Redis / in-memory cache implementations
│   └── adapter/                # Adapters for cross-module communication
│
└── controller/                 # Layer 4: Interface (depends on Application)
    ├── dto/                    # Request / Response DTOs
    │   ├── [action]_request.go
    │   └── [action]_response.go
    ├── http/                   # GIN handlers
    └── grpc/                   # gRPC handlers (if applicable)
```

---

## Entity vs Value Object

**Entity:** Has identity (ID field), mutable over time, compared by ID.
```go
// Entity — has ID, changes over time
type Order struct {
    ID        string      // identity
    Status    OrderStatus // mutable
    Items     []OrderItem
    UpdatedAt time.Time
}

// Entities have behavior methods
func (o *Order) Cancel() error {
    if o.Status == OrderStatusCompleted {
        return ErrOrderAlreadyCompleted
    }
    o.Status = OrderStatusCancelled
    return nil
}
```

**Value Object:** No identity, immutable, compared by value.
```go
// Value Object — no ID, immutable, replaced not modified
type Money struct {
    Amount   int64  // in cents
    Currency string // "USD", "VND"
}

// Value Objects have no setters, return new instances
func (m Money) Add(other Money) (Money, error) {
    if m.Currency != other.Currency {
        return Money{}, ErrCurrencyMismatch
    }
    return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Compared by value, not reference
func (m Money) Equals(other Money) bool {
    return m.Amount == other.Amount && m.Currency == other.Currency
}
```

---

## Repository pattern

**Interface in domain (what the application needs):**
```go
// domain/repository/i_order_repository.go
package repository

type IOrderRepository interface {
    FindByID(ctx context.Context, id string) (*entity.Order, error)
    FindByUserID(ctx context.Context, userID string, filter OrderFilter) ([]*entity.Order, int64, error)
    Save(ctx context.Context, order *entity.Order) error
    Delete(ctx context.Context, id string) error
}

type OrderFilter struct {
    Status   *entity.OrderStatus
    Page     int
    PageSize int
}
```

**Implementation in infrastructure (how the application gets it):**
```go
// infrastructure/persistence/repository/order_repository.go
package repository

// Implements domain.IOrderRepository using sqlc-generated queries
type OrderRepository struct {
    q *sqlcgen.Queries
}

func NewOrderRepository(q *sqlcgen.Queries) domainRepo.IOrderRepository {
    return &OrderRepository{q: q}
}
```

---

## Application Service pattern

```go
// application/service/i_order_service.go
type IOrderService interface {
    CreateOrder(ctx context.Context, req *dto.CreateOrderReq) (*dto.OrderResp, error)
    CancelOrder(ctx context.Context, orderID string, userID string) error
}

// application/service/order_service.go
type OrderService struct {
    orderRepo  domainRepo.IOrderRepository   // interface, not concrete
    userReader domainRepo.IUserReader         // cross-module port
    eventBus   event.IEventBus               // interface
}

func (s *OrderService) CreateOrder(ctx context.Context, req *dto.CreateOrderReq) (*dto.OrderResp, error) {
    // 1. Validate business rules
    user, err := s.userReader.FindByID(ctx, req.UserID)
    if err != nil { return nil, fmt.Errorf("orderService.CreateOrder: %w", err) }
    
    // 2. Create domain entity (business logic in entity or service)
    order := &entity.Order{
        ID:     uuid.New().String(),
        UserID: user.ID,
        Status: entity.OrderStatusPending,
    }
    
    // 3. Persist
    if err := s.orderRepo.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("orderService.CreateOrder save: %w", err)
    }
    
    // 4. Emit event (non-blocking)
    s.eventBus.Emit("order.created", &event.OrderCreatedEvent{OrderID: order.ID})
    
    // 5. Return DTO (not entity)
    return &dto.OrderResp{ID: order.ID, Status: string(order.Status)}, nil
}
```

---

## Module wiring (Dependency injection by hand)

```go
// internal/order/order.module.go
package order

type Module struct {
    Handler *http.OrderHandler
}

func NewModule(db *pgx.Pool, userSvc userService.IUserService) *Module {
    // Build dependency graph bottom-up
    q := sqlcgen.New(db)
    
    orderRepo := repository.NewOrderRepository(q)      // infra
    userReader := adapter.NewLocalUserAdapter(userSvc)  // infra
    
    orderSvc := service.NewOrderService(orderRepo, userReader)  // app
    
    handler := httpHandler.NewOrderHandler(orderSvc)             // interface
    
    return &Module{Handler: handler}
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
    orders := rg.Group("/orders")
    orders.POST("", m.Handler.Create)
    orders.GET("/:id", m.Handler.GetByID)
    orders.DELETE("/:id", m.Handler.Cancel)
}
```
