# Golang Technical Best Practices & Guidelines

Tài liệu này quy định các tiêu chuẩn kỹ thuật bắt buộc (Mandatory) cho toàn bộ đội ngũ Backend Engineer. Mục tiêu là đảm bảo chất lượng code đồng nhất, hiệu năng cao, an toàn bộ nhớ và dễ dàng bảo trì theo triết lý **Clean Code**.

## 1. Cấu trúc Dự án & Đặt tên (Project Layout & Naming)

### 1.1. Quy tắc đặt tên Package
Package phải ngắn gọn, chữ thường, **một từ duy nhất**, danh từ số ít.

```go
// ❌ Incorrect
package user_repository // Dùng underscore
package Services // Dùng chữ hoa, số nhiều

// ✅ Correct
package user
package auth
package order
```

### 1.2. Receiver Name
Viết tắt 1-3 ký tự của struct, nhất quán. Tuyệt đối không dùng `this`, `self`.

```go
type OrderService struct{}

// ❌ Incorrect
func (this *OrderService) Create() {}
func (self *OrderService) Update() {}

// ✅ Correct
func (s *OrderService) Create() {} // 's' viết tắt cho Service
func (os *OrderService) Update() {} // Hoặc 'os'
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
func NewUserService(repo user.Repository) *UserService {
    return &UserService{
        repo: repo,
    }
}
```

### 2.2. Hạn chế Magic Numbers

```go
// ❌ Incorrect
if retryCount > 5 {
    time.Sleep(10 * time.Second)
}

// ✅ Correct
const (
    MaxRetries    = 5
    RetryInterval = 10 * time.Second
)

if retryCount > MaxRetries {
    time.Sleep(RetryInterval)
}
```

---

## 3. Xử lý Lỗi (Error Handling)

### 3.1. Fail Fast & Guard Clauses
Tránh nesting (lồng nhau) quá sâu bằng cách return sớm.

```go
// ❌ Incorrect: Nesting sâu, khó đọc
func procesUser(u *User) error {
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
func procesUser(u *User) error {
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
Luôn wrap lỗi để giữ lại stack trace hoặc ngữ cảnh nơi xảy ra lỗi.

```go
// ❌ Incorrect: Mất ngữ cảnh, chỉ biết query lỗi
if err != nil {
    return err
}

// ✅ Correct: Biết rõ lỗi xảy ra khi nào, ở đâu
if err != nil {
    return fmt.Errorf("failed to fetch user by id %s: %w", userID, err)
}
```

---

## 4. Xử lý Đồng thời (Concurrency)

### 4.1. Context Propagation
Context phải luôn là tham số đầu tiên của hàm I/O.

```go
// ❌ Incorrect
func (r *Repo) GetUser(id string) (*User, error) {
    // Không có context, không thể cancel hoặc timeout
}

// ✅ Correct
func (r *Repo) GetUser(ctx context.Context, id string) (*User, error) {
    // Truyền ctx xuống database driver
    err := r.db.QueryRowContext(ctx, "SELECT ...", id).Scan(...)
}
```

### 4.2. Goroutine Leak Prevention
Luôn đảm bảo Goroutine sẽ dừng lại.

```go
// ❌ Incorrect: Goroutine chạy mãi mãi nếu không có tín hiệu dừng
go func() {
    for {
        process()
    }
}()

// ✅ Correct: Lắng nghe ctx.Done()
go func() {
    for {
        select {
        case <-ctx.Done(): // Thoát khi context bị cancel
            return
        default:
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
    if err := doTask2(); err != nil { // task 2 vẫn chạy dù task 1 lỗi
        errChan <- err
    }
}()
wg.Wait()
close(errChan)
// Phải loop read errChan cực khổ...

// ✅ Correct: errgroup tự động quản lý context cancel và error propagation
g, gCtx := errgroup.WithContext(ctx)

g.Go(func() error {
    // Nếu task này lỗi, gCtx sẽ bị cancel, các task khác nhận được tín hiệu
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
Mọi Goroutine chạy ngầm (background worker) **BẮT BUỘC** phải có cơ chế recover panic để tránh làm crash cả ứng dụng.

```go
// ❌ Incorrect: Nếu job panic, cả app sẽ chết
go func() {
    processJob()
}()

// ✅ Correct: Luôn recover trong background goroutine
go func() {
    defer func() {
        if r := recover(); r != nil {
            slog.Error("recovered from panic", "err", r)
        }
    }()
    processJob()
}()
```


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

```go
// ❌ Incorrect: Tạo nhiều object string tạm, chậm
s := ""
for i := 0; i < 1000; i++ {
    s += "data"
}

// ✅ Correct: Nhanh và tối ưu bộ nhớ
var sb strings.Builder
sb.Grow(1000 * 4) // (Optional) Pre-allocate nếu ước lượng được
for i := 0; i < 1000; i++ {
    sb.WriteString("data")
}
s := sb.String()
```

### 5.3. Pointer vs Value (Memory Optimization)
Chọn receiver type phù hợp để tối ưu GC và hiệu năng.

*   **Pointer Receiver (`*T`)**: Dùng khi struct lớn (> 64 bytes) HOẶC cần thay đổi giá trị state bên trong.
*   **Value Receiver (`T`)**: Dùng khi struct nhỏ (tọa độ, config), immutable, concurrency-safe (do copy).

```go
type Config struct {
    Timeout int
    Retries int
}

// ❌ Incorrect: Struct nhỏ nhưng dùng pointer (gây áp lực GC)
func (c *Config) GetTimeout() int { return c.Timeout }

// ✅ Correct: Value receiver cho struct nhỏ, read-only
func (c Config) GetTimeout() int { return c.Timeout }

type LargeData struct {
    Data [1024]byte
}

// ❌ Incorrect: Copy cả mảng lớn mỗi khi gọi hàm
func (d LargeData) Process() {}

// ✅ Correct: Pointer receiver tránh việc copy
func (d *LargeData) Process() {}
```

---

## 6. Logging (Structured Logging)

### 6.1. Contextual Logging

```go
// ❌ Incorrect: Khó parse, thiếu cấu trúc
fmt.Printf("Error updating user %d: %v\n", userID, err)

// ✅ Correct: Dễ dàng filter theo key-value trên hệ thống log
slog.Error("failed to update user",
    "user_id", userID,
    "error", err,
    "attempt", retryCount,
)
```

---

## 7. Testing

### 7.1. Table-Driven Tests

```go
// ✅ Correct
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Add(tt.a, tt.b); got != tt.want {
                t.Errorf("Add() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## 8. Configuration (12-Factor App)

### 8.1. Environment Variables
*   Cấu hình phải được load từ **Environment Variables**.
*   Sử dụng thư viện quản lý config (như `viper`, `koanf` hoặc `cleanenv`) để map env vars vào struct strongly-typed.
*   Tuyệt đối không hardcode credentials trong code.

---

## 9. Data Structures & JSON Tags

### 9.1. Struct Tags Consistency
*   Mọi API trả về JSON phải định nghĩa `json` tag rõ ràng (ưu tiên **snake_case**).
*   Tuyệt đối ẩn các trường nhạy cảm bằng `json:"-"`.

```go
type User struct {
    // ✅ Luôn có json tag, snake_case
    ID        string `json:"id"`
    FirstName string `json:"first_name"`
    
    // ❌ Incorrect: Lộ password ra API response
    // Password string `json:"password"`
    
    // ✅ Correct: Không bao giờ serialize field này
    Password  string `json:"-"`
}
```

---

## 10. API Response Standard (RESTful & Clean Code)

### 10.1. Standard Response Format
Thống nhất cấu trúc JSON trả về cho toàn bộ API để Frontend dễ dàng tích hợp và xử lý thống nhất.

**Success Response Structure:**
```go
type Response struct {
    Data    interface{} `json:"data,omitempty"`    // Dữ liệu chính trả về
    Meta    interface{} `json:"meta,omitempty"`    // Metadata (pagination info...)
    Message string      `json:"message,omitempty"` // Thông báo thành công (nếu cần)
}
```

**Error Response Structure:**
```go
type ErrorResponse struct {
    Code    string      `json:"code"`              // Mã lỗi machine-readable (VD: USER_NOT_FOUND)
    Message string      `json:"message"`           // Thông báo lỗi readable cho user/dev
    Details interface{} `json:"details,omitempty"` // Chi tiết lỗi (VD: danh sách field validation fail)
}
```

### 10.2. HTTP Status Codes
Sử dụng đúng HTTP Status Code theo ngữ nghĩa RESTful. **Tuyệt đối không** trả về `200 OK` kèm `error code` bên trong body cho các lỗi client/server thực sự.

*   **2xx Success**
    *   `200 OK`: Request thành công (GET, PUT, PATCH).
    *   `201 Created`: Tạo mới resource thành công (POST).
    *   `204 No Content`: Xử lý thành công nhưng không trả về dữ liệu (DELETE hoặc POST async).

*   **4xx Client Error**
    *   `400 Bad Request`: Input sai định dạng, thiếu param bắt buộc, sai kiểu dữ liệu.
    *   `401 Unauthorized`: Chưa xác thực (thiếu token, token hết hạn/không hợp lệ).
    *   `403 Forbidden`: Đã xác thực nhưng không có quyền truy cập resource.
    *   `404 Not Found`: Resource không tồn tại.
    *   `422 Unprocessable Entity`: Input đúng định dạng nhưng sai logic nghiệp vụ (VD: email trùng, số dư không đủ).

*   **5xx Server Error**
    *   `500 Internal Server Error`: Lỗi không mong muốn từ phía server/database.

### 10.3. Centralized Response Handling
Hạn chế việc construct JSON thủ công trong Controller/Handler. Nên xây dựng package `response` riêng.

```go
// ❌ Incorrect: Duplicate logic, magic numbers, format không đồng nhất
func (h *UserHandler) GetByID(c *gin.Context) {
    user, err := h.service.GetUser(c, id)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()}) // Lộ internal error message
        return
    }
    c.JSON(200, user)
}

// ✅ Correct: Code sạch, thống nhất format và status code
func (h *UserHandler) GetByID(c *gin.Context) {
    user, err := h.service.GetUser(c, id)
    if err != nil {
        response.Error(c, err) // Tự động map error -> HTTP status & Standard Error JSON
        return
    }
    response.Success(c, user)
}
```

### 10.4. Error Mapping Layer
Logic map từ lỗi nội bộ (Service/Domain) sang HTTP Status Code phải nằm ở tầng **Transport (HTTP Handler/Middleware)**.

*   **Service Layer**: Chỉ trả về Go `error` chuẩn hoặc sentinel errors (VD: `domain.ErrNotFound`, `domain.ErrEmailExists`). **Không** phụ thuộc vào package `net/http` hay `gin`.
*   **Helper/Handler**: Switch case trên error của service để quyết định status code.

```go
// Trong package response/error handling
func MapErrorToStatus(err error) int {
    switch {
    case errors.Is(err, domain.ErrNotFound):
        return http.StatusNotFound
    case errors.Is(err, domain.ErrEmailExists):
        return http.StatusConflict
    case errors.Is(err, domain.ErrInvalidInput):
        return http.StatusBadRequest
    default:
        return http.StatusInternalServerError
    }
}
```
---

## 11. Quy Chuẩn Giao Tiếp Liên Module (Inter-Module Communication)

Tài liệu này hướng dẫn cách truy xuất dữ liệu từ module khác trong kiến trúc Monolith của Converda, đảm bảo tính **Loosely Coupled** (rời rạc) và sẵn sàng để tách thành **Microservice** bất cứ lúc nào mà không phải sửa đổi Logic nghiệp vụ.

### 11.1. Vấn Đề (The Problem)

Trong kiến trúc Monolith, chúng ta thường có xu hướng import trực tiếp Repository hoặc Service từ module khác:

```go
// internal/mails/application/service/mail.service.impl.go

import settingsRepo "github.com/.../internal/settings/domain/repository" // ❌ Phụ thuộc trực tiếp

type mailService struct {
    templateRepo settingsRepo.EmailTemplateRepository
}
```

**Hệ quả:**
1. **Chặt chẽ (High Coupling):** Module Mails bị buộc chặt vào module Settings.
2. **Khó Scale:** Nếu Settings tách thành Microservice, code của Mails sẽ bị lỗi (vì Repo không còn database local).
3. **Khó Test:** Unit test của Mails phải mock cả thành phần của Settings.

### 11.2. Giải Pháp: Interface + Adapter Pattern

Chúng ta áp dụng nguyên lý **Dependency Inversion Principle (DIP)**: "Phụ thuộc vào trừu tượng, không phụ thuộc vào cụ thể".

#### Kiến trúc 3 lớp:
1. **Consumer (Mails Module):** Định nghĩa một `Interface` mô tả những gì nó CẦN.
2. **Implementation (Adapter):** Thực thi Interface đó bằng kỹ thuật cụ thể (DB query, API call).
3. **Initializer:** "Tiêm" (Inject) bản thực thi phù hợp vào Service.

### 11.3. Cấu Trúc Thư Mục Chuẩn

Mỗi module khi cần dữ liệu từ bên ngoài nên tổ chức như sau:

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

### 11.4. Ví Dụ Thực Tế (Mails & Settings)

#### Bước 1: Định nghĩa Interface tại Domain Layer (của Mails)

Interface này mô tả nhu cầu của Mails module, hoàn toàn độc lập với module Settings.

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

**Local Adapter (Dùng cho Monolith)**
Adapter này gọi trực tiếp Repository của Settings.

```go
// internal/mails/infrastructure/adapter/local_adapter.go
type LocalTemplateAdapter struct {
    settingsRepo settingsRepo.EmailTemplateRepository
}

func (a *LocalTemplateAdapter) GetByID(...) (*TemplateInfo, error) {
    t, _ := a.settingsRepo.GetById(ctx, id)
    return &TemplateInfo{ID: t.Id, ...}, nil // Map data sang DTO của Mails
}
```

**HTTP Adapter (Sẵn sàng cho Microservice)**
Khi tách service, chỉ cần viết thêm Adapter này, KHÔNG sửa MailService.

```go
// internal/mails/infrastructure/adapter/http_template_adapter.go
type HTTPTemplateAdapter struct {
    baseURL string
}

func (a *HTTPTemplateAdapter) GetByID(...) (*TemplateInfo, error) {
    resp, _ := http.Get(a.baseURL + "/templates/" + id)
    // Parse JSON sang TemplateInfo
}
```

### 11.5. Chiến Lược Dịch Chuyển (Microservice Readiness)

Khi module Settings được tách ra:
1. **Bước 1:** Viết `HTTPTemplateAdapter` gọi API của Settings Service.
2. **Bước 2:** Cập nhật file `initialize/mails/mail.go` để chuyển từ `LocalTemplateAdapter` sang `HTTPTemplateAdapter`.
3. **Kết quả:** Code logic trong `mail.service.impl.go` vẫn giữ nguyên 100%.

### 11.6. 5 Quy Tắc Vàng cho Team Leaders

> [!IMPORTANT]
> 1. **Consumer sở hữu Interface:** Interface `TemplateReader` phải nằm trong package của Mails, không phải Settings.
> 2. **Không Leak Entity:** Tránh trả về Entity của Settings qua Interface. Hãy dùng DTO đơn giản (như `TemplateInfo`).
> 3. **Adapter nằm ở Infra Layer:** Mọi logic về cách lấy dữ liệu (DB, API, gRPC) phải đóng gói trong `infrastructure/adapter/`.
> 4. **Dependency Injection:** Service chỉ nhận Interface qua Constructor.
> 5. **Mapping:** Luôn luôn có bước mapping dữ liệu từ nguồn (Settings) sang định dạng module hiện tại (Mails) cần.
