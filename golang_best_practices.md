# Golang Technical Best Practices & Guidelines

Tài liệu này quy định các tiêu chuẩn kỹ thuật bắt buộc (Mandatory) cho toàn bộ đội ngũ Backend Engineer. Mục tiêu là đảm bảo chất lượng code đồng nhất, hiệu năng cao, an toàn bộ nhớ và dễ dàng bảo trì theo triết lý **Clean Code**.

> **Phiên bản:** v2.0 — Cập nhật 18/02/2026
> **Thay đổi chính so với v1:**
> - **[Fix Bug]** Section 2.5: Sửa context leak khi dùng goroutine trong HTTP handler
> - **[Fix Bug]** Section 3.3: Sửa race condition trong `AppError.Wrap()` và `WithDetails()`
> - **[Cập nhật]** Section 2.3: Làm rõ quy tắc kích thước Interface
> - **[Cập nhật]** Section 4.2: Thêm cảnh báo busy-loop trong goroutine worker
> - **[Mới]** Section 4.7: `sync.RWMutex` cho shared state
> - **[Mới]** Section 5.6: `sync.Pool` để giảm GC pressure
> - **[Mới]** Section 7.4: Benchmark Test
> - **[Mới]** Section 8.3: `context.TODO()` vs `context.Background()`
> - **[Mới]** Section 8.4: Hướng dẫn hạn chế `init()`
> - **[Cập nhật]** Section 11.2: Transaction qua Unit of Work pattern (DDD-compliant)
> - **[Mới]** Section 14: Tooling & Linting (`golangci-lint`, Makefile, pre-commit hook)

## 1. Cấu trúc Dự án & Đặt tên (Project Layout & Naming)

### 1.1. Quy tắc đặt tên Package
Package phải ngắn gọn, chữ thường, **một từ duy nhất**, danh từ số ít.

```go
// ❌ Incorrect
package user_repository // Dùng underscore
package Services        // Dùng chữ hoa, số nhiều

// ✅ Correct
package user
package auth
package order
```

### 1.2. Receiver Name
Viết tắt 1-3 ký tự của struct, nhất quán trong toàn bộ struct. Tuyệt đối không dùng `this`, `self`.

```go
type OrderService struct{}

// ❌ Incorrect
func (this *OrderService) Create() {}
func (self *OrderService) Update() {}

// ✅ Correct
func (s *OrderService) Create() {} // 's' viết tắt cho Service
func (s *OrderService) Update() {} // Nhất quán dùng 's' cho mọi method
```

### 1.3. Exported vs Unexported
Chỉ export (viết hoa chữ cái đầu) những gì **thực sự cần** được sử dụng từ bên ngoài package. Mọi thứ khác **phải unexported** (chữ thường).

```go
// ❌ Incorrect: Export quá nhiều, lộ implementation details
type UserService struct {
    DB       *gorm.DB    // ❌ Export field internal
    Cache    *redis.Client
}

func (s *UserService) ValidateEmail(email string) bool {} // ❌ Helper không cần export

// ✅ Correct: Chỉ export interface và public methods
type UserService struct {
    db    *gorm.DB       // unexported
    cache *redis.Client  // unexported
}

func (s *UserService) Create(ctx context.Context, req *CreateUserReq) (*User, error) {} // Export
func (s *UserService) validateEmail(email string) bool {}                                // unexported helper
```

### 1.4. Đặt tên biến & hàm
*   **Biến**: camelCase, ngắn gọn nhưng có nghĩa. Tránh viết tắt khó hiểu.
*   **Hàm**: Động từ + Danh từ, mô tả chính xác hành động.
*   **Boolean**: Bắt đầu bằng `is`, `has`, `can`, `should`.

```go
// ❌ Incorrect
var u = GetU(id)         // Tên biến quá ngắn, không rõ nghĩa
func Proc(d []byte) {}   // Tên hàm không rõ hành động
var flag bool            // Boolean không rõ ý nghĩa

// ✅ Correct
var user = GetUserByID(id)
func ProcessPayload(data []byte) {}
var isActive bool
var hasPermission bool
```

---

## 2. Clean Code & Architecture

### 2.1. Dependency Injection (DI)
Sử dụng **Constructor Injection** thay vì khởi tạo dependency bên trong hoặc dùng biến global.

```go
// ❌ Incorrect: Hard dependency, khó test
func NewUserService() *UserService {
    return &UserService{
        repo: &MySQLRepository{}, // Tự khởi tạo
    }
}

// ✅ Correct: Dependency Injection qua Interface
func NewUserService(repo user.Repository, cache cache.Store) *UserService {
    return &UserService{
        repo:  repo,
        cache: cache,
    }
}
```

### 2.2. Interface Design — "Accept Interfaces, Return Structs"
Hàm/method nên **nhận Interface** làm parameter (linh hoạt) và **trả về Struct cụ thể** (rõ ràng).

```go
// ❌ Incorrect: Nhận struct cụ thể → không thể mock khi test
func ProcessOrder(repo *MySQLOrderRepo) error {
    // Bị phụ thuộc cứng vào MySQL
}

// ❌ Incorrect: Trả về interface → mất type information
func NewUserService(repo UserRepository) UserService {
    return &userServiceImpl{repo: repo}
}

// ✅ Correct: Nhận interface, trả struct
type OrderProcessor interface {
    FindByID(ctx context.Context, id string) (*Order, error)
}

func ProcessOrder(repo OrderProcessor) error {
    // Có thể inject mock repo khi test
}

func NewUserService(repo UserRepository) *userServiceImpl {
    return &userServiceImpl{repo: repo}
}
```

### 2.3. Interface nên nhỏ gọn (Interface Segregation)
Interface chỉ nên chứa những method mà **consumer thực sự cần dùng**. Tránh "God Interface" với quá nhiều method vì dẫn đến mock phức tạp và coupling cao. Thông thường interface tốt có 1–3 methods, nhưng con số quan trọng hơn là **mức độ phù hợp với nhu cầu thực tế của caller**.

```go
// ❌ Incorrect: God interface, quá nhiều method
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
    ListAll(ctx context.Context) ([]*User, error)
    Count(ctx context.Context) (int64, error)
    Search(ctx context.Context, query string) ([]*User, error)
}

// ✅ Correct: Chia nhỏ theo nhu cầu sử dụng
type UserReader interface {
    FindByID(ctx context.Context, id string) (*User, error)
    FindByEmail(ctx context.Context, email string) (*User, error)
}

type UserWriter interface {
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}

// Service chỉ nhận interface mà nó thực sự cần
type AuthService struct {
    reader UserReader // Chỉ cần đọc, không cần write
}
```

### 2.4. Hạn chế Magic Numbers & Strings
Mọi giá trị cố định phải được khai báo thành **constants**.

```go
// ❌ Incorrect
if retryCount > 5 {
    time.Sleep(10 * time.Second)
}
if user.Status == 1 { /* ... */ }

// ✅ Correct
const (
    MaxRetries    = 5
    RetryInterval = 10 * time.Second
)

const (
    StatusActive   = 1
    StatusInactive = 2
    StatusBanned   = 3
)

if retryCount > MaxRetries {
    time.Sleep(RetryInterval)
}
if user.Status == StatusActive { /* ... */ }
```

### 2.5. Single Responsibility Principle (SRP) cho Service
Mỗi Service chỉ xử lý logic nghiệp vụ của **một domain duy nhất**.

```go
// ❌ Incorrect: OrderService xử lý cả payment, email, inventory
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderReq) error {
    order := &Order{...}
    if err := s.repo.Create(ctx, order); err != nil {
        return err
    }
    s.chargePayment(ctx, order.Amount)     // ❌ Logic payment
    s.sendEmail(ctx, order.UserEmail)       // ❌ Logic email
    s.reduceStock(ctx, order.ProductID)     // ❌ Logic inventory
    return nil
}

// ✅ Correct: Chia nhỏ, phối hợp qua DI
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderReq) error {
    order := &Order{...}
    if err := s.repo.Create(ctx, order); err != nil {
        return fmt.Errorf("failed to create order: %w", err)
    }
    if err := s.paymentSvc.Charge(ctx, order.Amount); err != nil {
        return fmt.Errorf("failed to charge payment: %w", err)
    }
    if err := s.inventorySvc.ReduceStock(ctx, order.ProductID); err != nil {
        return fmt.Errorf("failed to reduce stock: %w", err)
    }
    // Email gửi async (không block flow chính)
    // ⚠️ BẮT BUỘC dùng context tách biệt khỏi request context.
    // ctx của HTTP handler sẽ bị cancel ngay khi handler return,
    // dẫn đến goroutine bị kill trước khi gửi email xong.
    go func() {
        // context.WithoutCancel (Go 1.21+): giữ lại values (traceID, userID)
        // nhưng không bị cancel theo request context.
        detachedCtx := context.WithoutCancel(ctx)
        defer func() {
            if r := recover(); r != nil {
                slog.Error("panic in mail goroutine", "err", r, "stack", string(debug.Stack()))
            }
        }()
        s.mailSvc.SendOrderConfirmation(detachedCtx, order)
    }()
    return nil
}
```

---

## 3. Xử lý Lỗi (Error Handling)

### 3.1. Fail Fast & Guard Clauses
Tránh nesting (lồng nhau) quá sâu bằng cách return sớm.

```go
// ❌ Incorrect: Nesting sâu, khó đọc
func processUser(u *User) error {
    if u != nil {
        if u.IsActive {
            // logic...
            return nil
        } else {
            return errors.New("user inactive")
        }
    } else {
        return errors.New("user nil")
    }
}

// ✅ Correct: Flatten code, xử lý lỗi trước
func processUser(u *User) error {
    if u == nil {
        return errors.New("user nil")
    }
    if !u.IsActive {
        return errors.New("user inactive")
    }
    
    // logic main flow nằm ở indent thấp nhất
    return nil
}
```

### 3.2. Error Wrapping
Luôn wrap lỗi với `%w` để giữ lại chain và cho phép `errors.Is`/`errors.As` hoạt động.

```go
// ❌ Incorrect: Mất ngữ cảnh, mất error chain
if err != nil {
    return err
}

// ❌ Incorrect: Dùng %v → mất error chain, errors.Is sẽ không hoạt động
if err != nil {
    return fmt.Errorf("failed: %v", err)
}

// ✅ Correct: Dùng %w → giữ error chain
if err != nil {
    return fmt.Errorf("failed to fetch user by id %s: %w", userID, err)
}
```

### 3.3. Custom Error Type (Business Error)
Ngoài sentinel errors, sử dụng **custom error type** khi cần đính kèm thêm metadata (error code, details).

```go
// common/apperror/error.go
type AppError struct {
    Code    string                 // Machine-readable code (VD: USER_NOT_FOUND)
    Message string                 // Human-readable message
    Details map[string]interface{} // Chi tiết bổ sung (optional)
    Err     error                  // Original error (for wrapping)
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

// Constructor helpers
func NewAppError(code, message string) *AppError {
    return &AppError{Code: code, Message: message}
}

// ✅ QUAN TRỌNG: Wrap và WithDetails PHẢI trả về instance MỚI.
// Không được mutate trực tiếp biến receiver vì các sentinel errors (VD: ErrUserNotFound)
// là biến global — nhiều goroutine đồng thời gọi .Wrap(err) sẽ gây RACE CONDITION.
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
    return &AppError{ // Trả về bản copy mới, không thay đổi gốc
        Code:    e.Code,
        Message: e.Message,
        Details: details,
        Err:     e.Err,
    }
}

func (e *AppError) Wrap(err error) *AppError {
    return &AppError{ // Trả về bản copy mới, không thay đổi gốc
        Code:    e.Code,
        Message: e.Message,
        Details: e.Details,
        Err:     err,
    }
}
```

```go
// domain/errors.go — Khai báo tập trung tất cả mã lỗi
var (
    ErrUserNotFound     = NewAppError("USER_NOT_FOUND", "User does not exist")
    ErrEmailExists      = NewAppError("EMAIL_ALREADY_EXISTS", "Email is already registered")
    ErrBalanceNotEnough = NewAppError("BALANCE_NOT_ENOUGH", "Insufficient balance")
    ErrInvalidInput     = NewAppError("INVALID_INPUT", "Input validation failed")
)
```

```go
// ❌ Incorrect: Service chỉ trả text, Controller không nhận biết được lỗi gì
func (s *userService) GetByID(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, errors.New("không tìm thấy user")
    }
    return user, nil
}

// ✅ Correct: Trả về AppError với mã lỗi rõ ràng
func (s *userService) GetByID(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, ErrUserNotFound.Wrap(err)
    }
    return user, nil
}

// Controller/Handler — errors.As để extract mã lỗi
func MapErrorToStatus(err error) int {
    var appErr *AppError
    if errors.As(err, &appErr) {
        switch appErr.Code {
        case "USER_NOT_FOUND":
            return http.StatusNotFound
        case "EMAIL_ALREADY_EXISTS":
            return http.StatusConflict
        case "BALANCE_NOT_ENOUGH":
            return http.StatusUnprocessableEntity
        default:
            return http.StatusInternalServerError
        }
    }
    return http.StatusInternalServerError
}
```

### 3.4. Không bỏ qua Error
Tuyệt đối **không bỏ qua** error trả về. Nếu thực sự không cần handle, phải comment lý do.

```go
// ❌ Incorrect: Bỏ qua error → bug ẩn
json.Unmarshal(data, &result)
file.Close()

// ✅ Correct
if err := json.Unmarshal(data, &result); err != nil {
    return fmt.Errorf("failed to unmarshal: %w", err)
}

// Nếu thực sự không cần handle, comment lý do
_ = file.Close() // Best-effort close, error logged elsewhere
```

---

## 4. Xử lý Đồng thời (Concurrency)

### 4.1. Context Propagation
Context phải luôn là **tham số đầu tiên** của mọi hàm I/O (database, HTTP, gRPC, file).

```go
// ❌ Incorrect
func (r *Repo) GetUser(id string) (*User, error) {
    // Không có context, không thể cancel hoặc timeout
}

// ✅ Correct
func (r *Repo) GetUser(ctx context.Context, id string) (*User, error) {
    err := r.db.QueryRowContext(ctx, "SELECT ...", id).Scan(...)
}
```

### 4.2. Goroutine Leak Prevention
Luôn đảm bảo Goroutine sẽ dừng lại bằng cách lắng nghe `ctx.Done()` hoặc đóng channel.

```go
// ❌ Incorrect: Goroutine chạy mãi mãi nếu không có tín hiệu dừng
go func() {
    for {
        process()
    }
}()

// ❌ Cũng Incorrect: select với default gây busy-loop nếu process() trả về nhanh
// → CPU 100% do goroutine liên tục spin không có điểm blocking
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            process() // Nếu hàm này chạy rất nhanh → spin loop
        }
    }
}()

// ✅ Correct cho worker chạy liên tục — dùng channel blocking
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        case job := <-jobChan: // Blocking: chỉ chạy khi có job
            process(job)
        }
    }
}()

// ✅ Correct cho worker poll định kỳ — dùng ticker
go func() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C: // Blocking: chỉ chạy theo interval
            process()
        }
    }
}()
```

### 4.3. Goroutine Group (Errgroup)
Ưu tiên `errgroup` để quản lý nhiều goroutine song song có khả năng trả về lỗi.

```go
// ❌ Incorrect: Phức tạp, dễ sai sót khi handle lỗi và đồng bộ
var wg sync.WaitGroup
errChan := make(chan error, 2)

wg.Add(2)
go func() {
    defer wg.Done()
    if err := doTask1(); err != nil {
        errChan <- err
    }
}()
go func() {
    defer wg.Done()
    if err := doTask2(); err != nil {
        errChan <- err
    }
}()
wg.Wait()
close(errChan)

// ✅ Correct: errgroup tự động quản lý context cancel và error propagation
g, gCtx := errgroup.WithContext(ctx)

g.Go(func() error {
    return doTask1(gCtx)
})

g.Go(func() error {
    return doTask2(gCtx)
})

if err := g.Wait(); err != nil {
    return fmt.Errorf("group task failed: %w", err)
}
```

### 4.4. Panic Recovery
Mọi Goroutine chạy ngầm (background worker) **BẮT BUỘC** phải có cơ chế recover panic.

```go
// ❌ Incorrect: Nếu job panic, cả app sẽ chết
go func() {
    processJob()
}()

// ✅ Correct: Luôn recover trong background goroutine
go func() {
    defer func() {
        if r := recover(); r != nil {
            slog.Error("recovered from panic", "err", r, "stack", string(debug.Stack()))
        }
    }()
    processJob()
}()
```

### 4.5. Graceful Shutdown
Ứng dụng **BẮT BUỘC** phải handle OS signals để shutdown gracefully — đảm bảo hoàn thành requests đang xử lý, đóng database connections, flush logs.

```go
func main() {
    srv := &http.Server{Addr: ":8080", Handler: router}

    // Chạy server trong goroutine
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("listen: %s\n", err)
        }
    }()

    // Đợi signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    slog.Info("shutting down server...")

    // Cho phép 30s để hoàn thành requests đang xử lý
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        slog.Error("server forced to shutdown", "error", err)
    }

    // Cleanup: đóng DB, Redis, flush logs...
    db.Close()
    slog.Info("server exited gracefully")
}
```

### 4.6. sync.Once cho Initialization
Sử dụng `sync.Once` khi cần khởi tạo singleton (database connection, config) một cách thread-safe.

```go
// ❌ Incorrect: Race condition khi nhiều goroutine gọi đồng thời
var dbConn *gorm.DB

func GetDB() *gorm.DB {
    if dbConn == nil {
        dbConn, _ = gorm.Open(...) // ❌ Có thể khởi tạo nhiều lần
    }
    return dbConn
}

// ✅ Correct: sync.Once đảm bảo chỉ chạy 1 lần duy nhất
var (
    dbConn *gorm.DB
    dbOnce sync.Once
)

func GetDB() *gorm.DB {
    dbOnce.Do(func() {
        var err error
        dbConn, err = gorm.Open(...)
        if err != nil {
            log.Fatalf("failed to connect database: %v", err)
        }
    })
    return dbConn
}
```

### 4.7. sync.RWMutex cho Shared State
Sử dụng `sync.RWMutex` thay vì `sync.Mutex` khi shared state được **đọc nhiều, ghi ít** (VD: in-memory cache, config store). RWMutex cho phép nhiều goroutine đọc đồng thời, chỉ ghi mới độc quyền.

```go
// ❌ Incorrect: Không có Mutex → data race khi map được đọc/ghi đồng thời
var cache = map[string]string{}

func Get(key string) string { return cache[key] }
func Set(key, val string)   { cache[key] = val }

// ❌ Cũng Incorrect: sync.Mutex cho phép ghi độc quyền,
// nhưng chặn cả các goroutine chỉ muốn ĐỌC → throughput thấp
type Cache struct {
    mu   sync.Mutex
    data map[string]string
}
func (c *Cache) Get(key string) string {
    c.mu.Lock()         // ❌ Block tất cả goroutine kể cả khi chỉ đọc
    defer c.mu.Unlock()
    return c.data[key]
}

// ✅ Correct: sync.RWMutex — nhiều reader đồng thời, writer độc quyền
type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()         // Nhiều goroutine có thể RLock cùng một lúc
    defer c.mu.RUnlock()
    val, ok := c.data[key]
    return val, ok
}

func (c *Cache) Set(key, val string) {
    c.mu.Lock()          // Ghi thì Lock độc quyền — block cả reader lẫn writer
    defer c.mu.Unlock()
    c.data[key] = val
}

func (c *Cache) Delete(key string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.data, key)
}
```

> **Lưu ý:** `sync.Map` từ stdlib là lựa chọn khác cho concurrent map, nhưng phù hợp hơn với pattern **write-once, read-many** hoặc khi keys ít thay đổi. Với general-purpose cache, `RWMutex + map` thường cho phép kiểm soát tốt hơn.

---

## 5. Hiệu suất (Performance)

### 5.1. Slice Pre-allocation
Cấp phát trước bộ nhớ nếu biết kích thước (hoặc ước lượng được).

```go
// ❌ Incorrect: Gây ra nhiều lần cấp phát lại (re-allocation) khi append
var users []User
for _, row := range rows {
    users = append(users, row)
}

// ✅ Correct: Chỉ cấp phát 1 lần
users := make([]User, 0, len(rows))
for _, row := range rows {
    users = append(users, row)
}
```

### 5.2. String Concatenation
Sử dụng `strings.Builder` khi concatenate nhiều lần trong vòng lặp.

```go
// ❌ Incorrect: Tạo nhiều object string tạm, O(n²) memory
s := ""
for i := 0; i < 1000; i++ {
    s += "data"
}

// ✅ Correct: O(n) memory, chỉ 1 allocation
var sb strings.Builder
sb.Grow(1000 * 4) // Pre-allocate nếu ước lượng được tổng kích thước
for i := 0; i < 1000; i++ {
    sb.WriteString("data")
}
s := sb.String()

// ✅ Với danh sách có separator — dùng strings.Join thay vì Builder
items := []string{"a", "b", "c"}
result := strings.Join(items, ",") // Ngắn gọn, rõ ràng hơn
```

### 5.3. Pointer vs Value (Memory Optimization)
Chọn receiver type phù hợp để tối ưu GC và hiệu năng.

*   **Pointer Receiver (`*T`)**: Dùng khi struct lớn (> 64 bytes) HOẶC cần thay đổi state.
*   **Value Receiver (`T`)**: Dùng khi struct nhỏ, immutable, concurrency-safe.

```go
type Config struct {
    Timeout int
    Retries int
}

// ✅ Value receiver cho struct nhỏ, read-only
func (c Config) GetTimeout() int { return c.Timeout }

type LargeData struct {
    Data [1024]byte
}

// ✅ Pointer receiver tránh copy struct lớn
func (d *LargeData) Process() {}
```

### 5.4. Tránh N+1 Query
Tuyệt đối không query database bên trong vòng lặp.

```go
// ❌ Incorrect: N+1 Query → 1 + N queries
orders, _ := repo.GetAllOrders(ctx)
for _, order := range orders {
    user, _ := repo.GetUserByID(ctx, order.UserID) // ❌ Query trong loop
    order.UserName = user.Name
}

// ✅ Correct: 2 queries tổng cộng
orders, _ := repo.GetAllOrders(ctx)
userIDs := make([]string, 0, len(orders))
for _, o := range orders {
    userIDs = append(userIDs, o.UserID)
}

users, _ := repo.GetUsersByIDs(ctx, userIDs) // 1 query WHERE id IN (...)
userMap := make(map[string]*User, len(users))
for _, u := range users {
    userMap[u.ID] = u
}

for i := range orders {
    if u, ok := userMap[orders[i].UserID]; ok {
        orders[i].UserName = u.Name
    }
}
```

### 5.5. Pagination bắt buộc
Mọi API trả về danh sách **BẮT BUỘC** phải có pagination. Không bao giờ trả về toàn bộ bảng.

```go
// ❌ Incorrect: Trả toàn bộ → crash nếu bảng triệu records
func (r *Repo) GetAllUsers(ctx context.Context) ([]*User, error) {
    return r.db.Find(&users).Error
}

// ✅ Correct: Luôn có pagination
type PaginationReq struct {
    Page  int `form:"page" binding:"min=1"`
    Limit int `form:"limit" binding:"min=1,max=100"`
}

func (r *Repo) GetUsers(ctx context.Context, req PaginationReq) ([]*User, int64, error) {
    var users []*User
    var total int64

    offset := (req.Page - 1) * req.Limit
    err := r.db.WithContext(ctx).
        Model(&User{}).
        Count(&total).
        Offset(offset).
        Limit(req.Limit).
        Order("created_at DESC").
        Find(&users).Error

    return users, total, err
}
```

### 5.6. sync.Pool để Tái sử dụng Object (Giảm GC Pressure)
Với hệ thống high-throughput, việc liên tục allocate và GC các object tạm (buffer, encoder, v.v.) gây áp lực lớn lên GC. `sync.Pool` cho phép tái sử dụng object, giảm đáng kể số lần GC.

```go
// ❌ Incorrect: Mỗi request tạo buffer mới → GC liên tục với traffic cao
func processPayload(data []byte) []byte {
    buf := new(bytes.Buffer) // Alloc mới mỗi lần
    buf.Write(data)
    // xử lý...
    return buf.Bytes()
}

// ✅ Correct: sync.Pool tái sử dụng buffer, giảm GC pressure
var bufPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processPayload(data []byte) []byte {
    buf := bufPool.Get().(*bytes.Buffer) // Lấy từ pool, không alloc mới
    defer func() {
        buf.Reset()        // Reset trước khi trả lại
        bufPool.Put(buf)   // Trả lại pool để tái sử dụng
    }()

    buf.Write(data)
    result := make([]byte, buf.Len())
    copy(result, buf.Bytes()) // Copy ra ngoài trước khi Reset
    return result
}
```

> **Lưu ý:** Object lấy từ Pool có thể bị GC bất cứ lúc nào khi Pool bị thu hồi. Không dùng Pool để lưu trữ state lâu dài. Chỉ dùng cho các object tạm, tốn kém khởi tạo, và **không có state** sau khi Reset.

---

## 6. Logging (Structured Logging)

### 6.1. Sử dụng `slog` (Go 1.21+) hoặc Structured Logger
Tuyệt đối không dùng `fmt.Println` / `log.Println` trong production. Sử dụng structured logging với key-value pairs.

```go
// ❌ Incorrect: Khó parse, thiếu cấu trúc, không filter được
fmt.Printf("Error updating user %d: %v\n", userID, err)
log.Println("user created:", userID)

// ✅ Correct: Structured logging với slog
slog.Info("user created", "user_id", userID, "email", user.Email)

slog.Error("failed to update user",
    "user_id", userID,
    "error", err,
    "attempt", retryCount,
)
```

### 6.2. Log Levels chuẩn
Sử dụng đúng log level theo ngữ nghĩa:

*   `Debug`: Thông tin chi tiết cho development (SQL queries, request payload).
*   `Info`: Sự kiện nghiệp vụ quan trọng (user created, order placed).
*   `Warn`: Tình huống bất thường nhưng hệ thống vẫn hoạt động (retry, fallback).
*   `Error`: Lỗi cần xử lý nhưng hệ thống tiếp tục chạy (query failed, API call timeout).

### 6.3. Không Log Dữ liệu Nhạy cảm
Tuyệt đối không log password, token, credit card, PII.

```go
// ❌ Incorrect
slog.Info("login attempt", "email", email, "password", password)
slog.Info("payment", "card_number", cardNumber)

// ✅ Correct
slog.Info("login attempt", "email", email)
slog.Info("payment", "card_last4", cardNumber[len(cardNumber)-4:])
```

---

## 7. Testing

### 7.1. Table-Driven Tests
Mẫu chuẩn cho unit test với nhiều test case.

```go
func TestCalculateDiscount(t *testing.T) {
    tests := []struct {
        name     string
        price    float64
        discount float64
        want     float64
        wantErr  bool
    }{
        {"normal discount", 100, 10, 90, false},
        {"zero discount", 100, 0, 100, false},
        {"full discount", 100, 100, 0, false},
        {"negative price", -100, 10, 0, true},
        {"over 100% discount", 100, 150, 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := CalculateDiscount(tt.price, tt.discount)
            if (err != nil) != tt.wantErr {
                t.Errorf("CalculateDiscount() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("CalculateDiscount() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 7.2. Mock với Interface
Sử dụng interface để mock dependency trong unit test. Khuyến khích dùng `testify/mock` hoặc `gomock`.

```go
// Mock repository cho unit test
type MockUserRepo struct{}

func (m *MockUserRepo) FindByID(ctx context.Context, id string) (*User, error) {
    if id == "not-found" {
        return nil, nil
    }
    return &User{ID: id, Name: "Test User", IsActive: true}, nil
}

func TestUserService_GetByID(t *testing.T) {
    mockRepo := &MockUserRepo{}
    service := NewUserService(mockRepo)

    t.Run("user found", func(t *testing.T) {
        user, err := service.GetByID(context.Background(), "123")
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if user.Name != "Test User" {
            t.Errorf("got name %q, want %q", user.Name, "Test User")
        }
    })

    t.Run("user not found", func(t *testing.T) {
        _, err := service.GetByID(context.Background(), "not-found")
        if err == nil {
            t.Fatal("expected error, got nil")
        }
    })
}
```

### 7.3. Test File Organization
*   File test cùng package: `user_service_test.go`
*   Tuyệt đối không test trực tiếp database/API trong unit test (dùng mock).
*   Integration test đặt trong folder `_test/` hoặc dùng build tag `//go:build integration`.

### 7.4. Benchmark Test
Viết benchmark để đo lường và theo dõi hiệu năng theo thời gian, đặc biệt quan trọng trước khi tối ưu hóa ("đo trước, tối ưu sau").

```go
// Benchmark cơ bản
func BenchmarkProcessPayload(b *testing.B) {
    data := generateTestData(1000)

    b.ResetTimer()   // Không tính thời gian setup vào kết quả
    b.ReportAllocs() // Hiển thị số lần allocation

    for i := 0; i < b.N; i++ {
        ProcessPayload(data)
    }
}
// Chạy: go test -bench=BenchmarkProcessPayload -benchmem -count=5
// Output: BenchmarkProcessPayload-8   500000   2341 ns/op   128 B/op   3 allocs/op

// Benchmark so sánh nhiều implementation
func BenchmarkStringConcat(b *testing.B) {
    items := make([]string, 100)
    for i := range items {
        items[i] = "item"
    }

    b.Run("plus_operator", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            s := ""
            for _, item := range items {
                s += item
            }
        }
    })

    b.Run("strings_builder", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            var sb strings.Builder
            sb.Grow(len(items) * 4)
            for _, item := range items {
                sb.WriteString(item)
            }
            _ = sb.String()
        }
    })

    b.Run("strings_join", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            _ = strings.Join(items, "")
        }
    })
}

// Benchmark concurrent để phát hiện bottleneck dưới tải
func BenchmarkCacheGet(b *testing.B) {
    cache := NewCache()
    cache.Set("key", "value")

    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            cache.Get("key")
        }
    })
}
```

> **Quy tắc:** Không tối ưu hoá code khi chưa có benchmark chứng minh đó là bottleneck. Chạy benchmark tối thiểu `-count=5` để kết quả ổn định.

---

## 8. Configuration (12-Factor App)

### 8.1. Environment Variables với Struct Validation
Cấu hình phải được load từ **Environment Variables** và validate khi khởi động. App phải **crash ngay** nếu thiếu config bắt buộc.

```go
type Config struct {
    Port        int    `env:"PORT" env-default:"8080"`
    DatabaseURL string `env:"DATABASE_URL" env-required:"true"`
    JWTSecret   string `env:"JWT_SECRET" env-required:"true"`
    RedisURL    string `env:"REDIS_URL" env-required:"true"`
    Environment string `env:"APP_ENV" env-default:"development"`
}

func LoadConfig() (*Config, error) {
    var cfg Config
    if err := cleanenv.ReadEnv(&cfg); err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    return &cfg, nil
}

func main() {
    cfg, err := LoadConfig()
    if err != nil {
        log.Fatalf("config error: %v", err) // Crash ngay nếu thiếu config
    }
    // ...
}
```

### 8.2. Không Hardcode Credentials
Tuyệt đối không hardcode bất kỳ secret/credential/API key nào trong code.

```go
// ❌ Incorrect
const jwtSecret = "my-super-secret-key-123"
db, _ := gorm.Open(mysql.Open("root:password123@tcp(localhost:3306)/mydb"))

// ✅ Correct
jwtSecret := cfg.JWTSecret
db, _ := gorm.Open(mysql.Open(cfg.DatabaseURL))
```

### 8.3. context.TODO() vs context.Background()
Hai hàm này đều trả về context rỗng nhưng có **ngữ nghĩa khác nhau** — dùng đúng giúp team hiểu ý định của code.

```go
// context.Background() — điểm gốc thực sự của context tree
// Dùng khi: main(), top-level server setup, background worker tồn tại
// suốt vòng đời ứng dụng.
func main() {
    ctx := context.Background()
    server.Run(ctx)
}

func startBackgroundWorker() {
    ctx := context.Background() // Worker này sống cùng app, không có parent
    go worker.Run(ctx)
}

// context.TODO() — placeholder khi CHƯA biết context nào sẽ dùng
// Dùng khi: đang refactor code cũ chưa có context, cần đánh dấu "xem lại sau".
// Đây là TÍN HIỆU CHO TEAM — "đoạn code này cần được cập nhật".
func legacyFunction() error {
    // TODO: nhận ctx từ caller sau khi refactor xong
    user, err := repo.FindByID(context.TODO(), "123")
    return err
}
```

> **Quy tắc:** `context.TODO()` không được tồn tại lâu dài trong code. Mỗi `context.TODO()` nên có issue tracking tương ứng để xử lý.

### 8.4. Hạn chế sử dụng `init()`
Hàm `init()` chạy tự động và khó kiểm soát thứ tự thực thi khi có nhiều file. Hạn chế dùng `init()` để tránh side effect ẩn.

```go
// ❌ Incorrect: init() với side effects — khó debug, thứ tự không rõ ràng
func init() {
    db, _ = connectDB()    // Side effect: kết nối DB ngay khi import package
    loadConfig()           // Thứ tự chạy so với init() ở file khác không đoán được
    initLogger()           // Nếu lỗi ở đây → panic khó trace
}

// ✅ Correct: Chỉ dùng init() cho việc đăng ký (register) — không có side effect
func init() {
    // OK: Đăng ký driver — đây là convention chính thống của Go
    sql.Register("my-driver", &MyDriver{})
    // OK: Đăng ký codec, plugin — không có I/O, không panic
    encoding.Register("msgpack", &MsgpackCodec{})
}

// ✅ Correct: Mọi initialization có thể lỗi → explicit trong main()
func main() {
    cfg, err := config.Load()  // Rõ ràng, có thể handle lỗi
    if err != nil {
        log.Fatalf("config error: %v", err)
    }
    db, err := database.New(cfg.DatabaseURL) // Thứ tự rõ ràng
    if err != nil {
        log.Fatalf("db error: %v", err)
    }
    // ...
}
```

---

## 9. Data Structures & JSON Tags

### 9.1. Struct Tags Consistency
*   Mọi API struct phải có `json` tag rõ ràng (ưu tiên **snake_case**).
*   Ẩn field nhạy cảm bằng `json:"-"`.
*   Sử dụng `omitempty` khi field có thể empty/nil.

```go
type User struct {
    ID        string     `json:"id"`
    FirstName string     `json:"first_name"`
    Email     string     `json:"email"`
    AvatarURL *string    `json:"avatar_url,omitempty"` // Nullable field
    Password  string     `json:"-"`                     // Tuyệt đối ẩn
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty"`
}
```

### 9.2. Tách Request/Response DTO khỏi Entity
Tuyệt đối **không** dùng Entity trực tiếp làm API request/response. Luôn tạo DTO riêng.

```go
// ❌ Incorrect: Dùng entity trực tiếp → lộ internal fields, risk mass assignment
func (h *Handler) CreateUser(c *gin.Context) {
    var user models.User
    c.ShouldBindJSON(&user) // ❌ Client có thể set ID, Role, CreatedAt...
    h.repo.Create(&user)
}

// ✅ Correct: DTO riêng cho request
type CreateUserReq struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

type UserResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (h *Handler) CreateUser(c *gin.Context) {
    var req CreateUserReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, err)
        return
    }
    user, err := h.service.Create(c, req)
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Created(c, UserResponse{ID: user.ID, Name: user.Name, Email: user.Email})
}
```

### 9.3. Time Format chuẩn
Sử dụng `time.Time` cho datetime. Khi serialize JSON, nếu cần custom format thì dùng custom type.

```go
// ❌ Incorrect: Lưu time dạng string
type Order struct {
    CreatedAt string `json:"created_at"` // ❌ "2024-01-15 10:30:00" — mất timezone
}

// ✅ Correct: Luôn dùng time.Time → ISO 8601 (RFC 3339) tự động
type Order struct {
    CreatedAt time.Time `json:"created_at"` // → "2024-01-15T10:30:00Z"
}
```

---

## 10. API Response Standard (RESTful & Clean Code)

### 10.1. Standard Response Format
Thống nhất cấu trúc JSON trả về cho toàn bộ API.

**Success Response:**
```go
type Response[T any] struct {
    Data    T           `json:"data,omitempty"`
    Meta    interface{} `json:"meta,omitempty"`
    Message string      `json:"message,omitempty"`
}

// Helper functions
func Success[T any](c *gin.Context, data T) {
    c.JSON(http.StatusOK, Response[T]{Data: data})
}

func Created[T any](c *gin.Context, data T) {
    c.JSON(http.StatusCreated, Response[T]{Data: data, Message: "Created successfully"})
}
```

**Error Response:**
```go
type ErrorResponse struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

func Error(c *gin.Context, err error) {
    var appErr *AppError
    if errors.As(err, &appErr) {
        status := MapCodeToHTTPStatus(appErr.Code)
        c.JSON(status, ErrorResponse{
            Code:    appErr.Code,
            Message: appErr.Message,
            Details: appErr.Details,
        })
        return
    }
    // Lỗi không xác định → 500, KHÔNG lộ internal error
    c.JSON(http.StatusInternalServerError, ErrorResponse{
        Code:    "INTERNAL_ERROR",
        Message: "An unexpected error occurred",
    })
}
```

### 10.2. HTTP Status Codes
Sử dụng đúng HTTP Status Code theo ngữ nghĩa RESTful. **Không** trả `200 OK` kèm error code bên trong body.

*   **2xx**: `200 OK`, `201 Created`, `204 No Content`
*   **4xx**: `400 Bad Request`, `401 Unauthorized`, `403 Forbidden`, `404 Not Found`, `409 Conflict`, `422 Unprocessable Entity`, `429 Too Many Requests`
*   **5xx**: `500 Internal Server Error`

### 10.3. Centralized Response Handling
Controller/Handler không construct JSON thủ công. Sử dụng package `response` helper.

```go
// ❌ Incorrect: Duplicate logic, magic numbers
func (h *UserHandler) GetByID(c *gin.Context) {
    user, err := h.service.GetUser(c, id)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()}) // ❌ Lộ internal error
        return
    }
    c.JSON(200, user)
}

// ✅ Correct
func (h *UserHandler) GetByID(c *gin.Context) {
    user, err := h.service.GetUser(c, id)
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Success(c, user)
}
```

---

## 11. Database & Repository

### 11.1. Luôn dùng Parameterized Queries
Tuyệt đối **không** concat string vào SQL query. Luôn dùng placeholder để chống SQL Injection.

```go
// ❌ Incorrect: SQL Injection
query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name)
db.Raw(query).Scan(&users)

// ✅ Correct: Parameterized query
db.Where("name = ?", name).Find(&users)
// Hoặc
db.Raw("SELECT * FROM users WHERE name = ?", name).Scan(&users)
```

### 11.2. Transaction cho Multi-step Operations
Mọi thao tác ghi liên quan đến **nhiều bảng** phải nằm trong **Transaction**.

> **Lưu ý về DDD:** Theo kiến trúc DDD (ADR-001), tầng Application **không được biết đến** GORM hay bất kỳ infrastructure chi tiết nào. Do đó, transaction không được thực hiện trực tiếp qua `gorm.DB` trong service. Thay vào đó, dùng pattern **Unit of Work** — định nghĩa interface ở Domain, implement ở Infrastructure.

```go
// ❌ Incorrect: Application layer bị phụ thuộc trực tiếp vào GORM → vi phạm DDD
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderReq) error {
    return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { // ❌ GORM lọt vào Application
        // ...
    })
}

// ✅ Correct — Bước 1: Domain layer định nghĩa interface UnitOfWork
// internal/order/domain/repository/unit_of_work.go
type OrderUnitOfWork interface {
    // Execute chạy fn trong một transaction. Nếu fn trả error → rollback tự động.
    Execute(ctx context.Context, fn func(repo OrderTxRepository) error) error
}

// Các repo operations có thể dùng trong transaction
type OrderTxRepository interface {
    CreateOrder(ctx context.Context, order *Order) error
    ReduceStock(ctx context.Context, productID string, qty int) error
}

// ✅ Correct — Bước 2: Infrastructure layer implement
// internal/order/infrastructure/persistence/unit_of_work.go
type gormOrderUnitOfWork struct {
    db *gorm.DB
}

func (u *gormOrderUnitOfWork) Execute(ctx context.Context, fn func(OrderTxRepository) error) error {
    return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        txRepo := &gormOrderTxRepo{db: tx}
        return fn(txRepo) // Application chỉ biết interface, không biết GORM
    })
}

// ✅ Correct — Bước 3: Application service hoàn toàn sạch, không có GORM
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderReq) error {
    return s.uow.Execute(ctx, func(repo OrderTxRepository) error {
        order := &Order{UserID: req.UserID, ProductID: req.ProductID, Qty: req.Quantity}

        if err := repo.CreateOrder(ctx, order); err != nil {
            return fmt.Errorf("failed to create order: %w", err)
        }
        if err := repo.ReduceStock(ctx, req.ProductID, req.Quantity); err != nil {
            return fmt.Errorf("failed to reduce stock: %w", err)
        }
        return nil // → Commit tự động
    })
    // Nếu fn trả error → Rollback tự động
}
```

### 11.3. Luôn `defer Close()` cho Resources
Mọi resource (file, database rows, HTTP response body) **BẮT BUỘC** phải `defer Close()` ngay sau khi mở.

```go
// ❌ Incorrect: Nếu hàm return sớm vì lỗi → resource leak
func ReadFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    data, err := io.ReadAll(f)
    if err != nil {
        return nil, err // ❌ file chưa được close!
    }
    f.Close()
    return data, nil
}

// ✅ Correct: defer Close() ngay sau khi mở thành công
func ReadFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer f.Close() // ✅ Luôn close dù có lỗi hay không

    data, err := io.ReadAll(f)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    return data, nil
}
```

---

## 12. Security

### 12.1. Timeout cho mọi External Call
Mọi HTTP/gRPC call ra bên ngoài **BẮT BUỘC** phải có timeout.

```go
// ❌ Incorrect: Không timeout → request có thể treo vĩnh viễn
resp, err := http.Get("https://external-api.com/data")

// ✅ Correct: Luôn dùng context với timeout
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://external-api.com/data", nil)
resp, err := http.DefaultClient.Do(req)
```

### 12.2. Không lộ Internal Error ra API Response
Khi xảy ra lỗi 500, **KHÔNG** trả error message chi tiết cho client. Log chi tiết ở server side.

```go
// ❌ Incorrect: Lộ stack trace, tên bảng DB, SQL query
c.JSON(500, gin.H{"error": err.Error()})
// Response: {"error": "Error 1062: Duplicate entry 'john@email.com' for key 'users.email'"}

// ✅ Correct: Trả message generic, log chi tiết ở server
slog.Error("failed to create user", "error", err, "email", req.Email)
c.JSON(500, ErrorResponse{
    Code:    "INTERNAL_ERROR",
    Message: "An unexpected error occurred",
})
```

---

## 13. Quy Chuẩn Giao Tiếp Liên Module (Inter-Module Communication)

Tài liệu này hướng dẫn cách truy xuất dữ liệu từ module khác trong kiến trúc Monolith, đảm bảo tính **Loosely Coupled** và sẵn sàng để tách thành **Microservice** bất cứ lúc nào.

### 13.1. Vấn Đề (The Problem)

Import trực tiếp Repository hoặc Service từ module khác tạo ra coupling chặt:

```go
// ❌ Phụ thuộc trực tiếp
import settingsRepo "github.com/.../internal/settings/domain/repository"

type mailService struct {
    templateRepo settingsRepo.EmailTemplateRepository // ❌ Cross-module dependency
}
```

**Hệ quả:**
1. **High Coupling:** Module Mails bị buộc chặt vào module Settings.
2. **Khó Scale:** Nếu Settings tách thành Microservice, code của Mails bị lỗi.
3. **Khó Test:** Unit test phải mock cả thành phần của Settings.

### 13.2. Giải Pháp: Interface + Adapter Pattern

Áp dụng **Dependency Inversion Principle (DIP)**: "Phụ thuộc vào trừu tượng, không phụ thuộc vào cụ thể."

#### Kiến trúc 3 lớp:
1. **Consumer (Mails Module):** Định nghĩa Interface mô tả nhu cầu.
2. **Implementation (Adapter):** Thực thi Interface bằng kỹ thuật cụ thể.
3. **Initializer:** Inject bản thực thi phù hợp vào Service.

### 13.3. Cấu Trúc Thư Mục Chuẩn

```text
internal/mails/
├── domain/
│   └── repository/
│       └── template_reader.go    # 🟢 [Interface] Mails cần đọc template
├── infrastructure/
│   └── adapter/
│       ├── local_adapter.go      # 🔵 [Impl] Lấy từ module Settings local
│       ├── http_adapter.go       # 🟠 [Impl] Lấy qua REST API (Microservice)
│       └── cached_adapter.go     # 🟡 [Optional] Thêm cache layer
```

### 13.4. Ví Dụ Thực Tế (Mails & Settings)

#### Bước 1: Định nghĩa Interface tại Domain Layer (của Mails)

```go
// internal/mails/domain/repository/template_reader.go
type TemplateInfo struct {
    ID      int64
    Subject string
    Content string
}

type TemplateReader interface {
    GetByID(ctx context.Context, id int64) (*TemplateInfo, error)
}
```

#### Bước 2: Tạo Adapter tại Infrastructure Layer

**Local Adapter (Monolith):**
```go
// internal/mails/infrastructure/adapter/local_adapter.go
type LocalTemplateAdapter struct {
    settingsRepo settingsRepo.EmailTemplateRepository
}

func (a *LocalTemplateAdapter) GetByID(ctx context.Context, id int64) (*TemplateInfo, error) {
    t, err := a.settingsRepo.GetById(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get template %d: %w", id, err)
    }
    return &TemplateInfo{ID: t.Id, Subject: t.Subject, Content: t.Content}, nil
}
```

**HTTP Adapter (Microservice):**
```go
// internal/mails/infrastructure/adapter/http_template_adapter.go
type HTTPTemplateAdapter struct {
    baseURL    string
    httpClient *http.Client
}

func (a *HTTPTemplateAdapter) GetByID(ctx context.Context, id int64) (*TemplateInfo, error) {
    url := fmt.Sprintf("%s/templates/%d", a.baseURL, id)
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    resp, err := a.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to call settings service: %w", err)
    }
    defer resp.Body.Close()
    
    var info TemplateInfo
    if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    return &info, nil
}
```

### 13.5. Chiến Lược Dịch Chuyển (Microservice Readiness)

Khi module Settings được tách ra:
1. Viết `HTTPTemplateAdapter` gọi API của Settings Service.
2. Cập nhật file initializer để chuyển từ `LocalTemplateAdapter` sang `HTTPTemplateAdapter`.
3. Code logic trong `mail.service.impl.go` vẫn giữ nguyên **100%**.

### 13.6. 5 Quy Tắc Vàng cho Team Leaders

> [!IMPORTANT]
> 1. **Consumer sở hữu Interface:** Interface `TemplateReader` phải nằm trong package của Mails, không phải Settings.
> 2. **Không Leak Entity:** Tránh trả về Entity của Settings qua Interface. Hãy dùng DTO đơn giản (như `TemplateInfo`).
> 3. **Adapter nằm ở Infra Layer:** Mọi logic về cách lấy dữ liệu (DB, API, gRPC) phải đóng gói trong `infrastructure/adapter/`.
> 4. **Dependency Injection:** Service chỉ nhận Interface qua Constructor.
> 5. **Mapping:** Luôn luôn có bước mapping dữ liệu từ nguồn (Settings) sang định dạng module hiện tại (Mails) cần.

---

## 14. Tooling & Linting

Công cụ tự động hóa kiểm tra giúp đảm bảo toàn bộ team tuân thủ best practices mà không cần review thủ công mọi PR.

### 14.1. golangci-lint — Bắt buộc trong CI/CD

`golangci-lint` chạy nhiều linter song song, nhanh hơn chạy từng linter riêng lẻ. **Bắt buộc** pass trước khi merge PR.

```bash
# Cài đặt
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Chạy
golangci-lint run ./...

# Chạy và tự fix (áp dụng được)
golangci-lint run --fix ./...
```

### 14.2. Cấu hình `.golangci.yml` chuẩn

```yaml
# .golangci.yml — đặt ở root project
run:
  timeout: 5m
  go: "1.21"

linters:
  enable:
    - errcheck       # Phát hiện error bị bỏ qua (liên quan Section 3.4)
    - gosec          # Security: SQL injection, hardcoded secrets, weak crypto
    - govet          # go vet chính thức — phát hiện lỗi ngữ nghĩa
    - staticcheck    # Phân tích static nâng cao, phát hiện dead code, bugs
    - unused         # Phát hiện function/variable không được dùng
    - gofumpt        # Formatter nghiêm ngặt hơn gofmt
    - noctx          # Phát hiện HTTP call không có context (liên quan Section 12.1)
    - bodyclose      # Phát hiện HTTP response body không được close
    - gocritic       # Nhiều kiểm tra code style và performance
    - exhaustive     # Đảm bảo switch trên enum xử lý hết các case
    - misspell       # Phát hiện lỗi chính tả trong comment và string
    - prealloc        # Gợi ý pre-allocate slice (liên quan Section 5.1)
    - revive         # Thay thế golint với nhiều rule hơn
    - whitespace     # Phát hiện whitespace thừa

linters-settings:
  gosec:
    excludes:
      - G104  # Cho phép bỏ qua lỗi trong defer (trường hợp đã có comment lý do)
  govet:
    enable-all: true
  errcheck:
    check-type-assertions: true   # Kiểm tra cả type assertion
    check-blank: true             # Cảnh báo khi dùng _ để bỏ qua error

issues:
  exclude-rules:
    # Cho phép bỏ qua errcheck trong file test
    - path: _test\.go
      linters:
        - errcheck
  max-issues-per-linter: 0
  max-same-issues: 0
```

### 14.3. Makefile Commands chuẩn

Mỗi project nên có `Makefile` với các lệnh nhất quán để giảm friction cho developer:

```makefile
# Makefile

.PHONY: lint test test-race build tidy

## Chạy linter — bắt buộc pass trước khi commit
lint:
	golangci-lint run ./...

## Chạy toàn bộ unit test
test:
	go test ./... -v -count=1

## Chạy test với race detector — phát hiện data race (Section 4)
test-race:
	go test -race ./... -count=1

## Chạy benchmark
bench:
	go test -bench=. -benchmem -count=5 ./...

## Build binary
build:
	go build -o bin/app ./cmd/...

## Dọn dẹp dependencies
tidy:
	go mod tidy
	go mod verify

## Chạy tất cả kiểm tra trước khi push
check: tidy lint test-race
```

### 14.4. Pre-commit Hook

Tích hợp linter vào git hook để bắt lỗi trước khi commit:

```bash
# .git/hooks/pre-commit
#!/bin/sh
set -e

echo "Running linter..."
golangci-lint run ./...

echo "Running tests..."
go test -race ./... -count=1

echo "All checks passed!"
```

> **Lưu ý:** Khuyến khích dùng [pre-commit](https://pre-commit.com/) framework để quản lý hooks tập trung và dễ chia sẻ trong team hơn là script thủ công.