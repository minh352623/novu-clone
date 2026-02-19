# Prompt Template: Viết unit test cho service method

## Cách dùng
Copy phần giữa `---START---` và `---END---`, điền thông tin, paste vào chat.

---START---
Viết unit test cho method sau:

**File cần test:** `internal/[module]/application/service/[module]_service.go`
**Method:** `[MethodName](ctx context.Context, [params]) ([returns], error)`

**Paste implementation code:**
```go
[paste code của function cần test tại đây]
```

**Dependencies (interfaces cần mock):**
- `[InterfaceName]` — [ghi rõ methods nào sẽ được gọi]

**Test cases cần cover:**
1. [Happy path: điều kiện thành công]
2. [Error case 1: ví dụ user not found]
3. [Error case 2: ví dụ repository error]
4. [Edge case nếu có]

**Yêu cầu đặc biệt:**
- [Table-driven test — bắt buộc]
- [Test file path: internal/[module]/application/service/[module]_service_test.go]
- [Dùng testify/assert và mockery mocks]

---END---

## Ví dụ đã điền

---START---
Viết unit test cho method sau:

**File cần test:** `internal/order/application/service/order_service.go`
**Method:** `CancelOrder(ctx context.Context, orderID string, requesterID string) error`

**Implementation:**
```go
func (s *OrderService) CancelOrder(ctx context.Context, orderID string, requesterID string) error {
    order, err := s.orderRepo.FindByID(ctx, orderID)
    if err != nil {
        return fmt.Errorf("orderService.CancelOrder: %w", err)
    }
    
    if order.UserID != requesterID {
        return domain.ErrOrderNotFound // security: don't reveal ownership
    }
    
    if err := order.Cancel(); err != nil { // domain entity method
        return fmt.Errorf("orderService.CancelOrder: %w", err)
    }
    
    if err := s.orderRepo.Save(ctx, order); err != nil {
        return fmt.Errorf("orderService.CancelOrder save: %w", err)
    }
    
    return nil
}
```

**Dependencies:**
- `IOrderRepository` — methods: `FindByID`, `Save`

**Test cases:**
1. Happy path: order exists, requester is owner, status is cancellable → success
2. Order not found: repo returns ErrOrderNotFound → propagated
3. Not owner: order.UserID != requesterID → ErrOrderNotFound (security)
4. Already cancelled: order.Cancel() returns domain error → propagated
5. Save fails: repo.Save returns error → wrapped error returned

**Yêu cầu:**
- Table-driven test
- Test file: `internal/order/application/service/order_service_test.go`
- Dùng testify/assert, require và mockery mocks
---END---
