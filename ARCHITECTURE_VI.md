# Tài Liệu Cấu Trúc Dự Án - DDD Architecture

> **Hướng dẫn chi tiết về ý nghĩa và mục đích của từng folder/file trong dự án**

---

## 📁 Cấu Trúc Tổng Quan

```
go-ddd/
├── cmd/                    # Điểm khởi đầu ứng dụng
├── internal/              # Code business logic (DDD layers)
├── pkg/                   # Packages công khai có thể tái sử dụng
├── utils/                 # Các utility functions
├── global/                # Biến global và constants
├── environment/           # Cấu hình môi trường
├── scripts/               # Scripts build và deployment
├── tests/                 # Integration và E2E tests
└── go.mod                 # Go module dependencies
```

---

## 1️⃣ Thư Mục Gốc (Root)

### `cmd/` - Điểm Khởi Đầu Ứng Dụng

**Mục đích:** Chứa các file `main.go` để khởi chạy ứng dụng

```
cmd/
├── drunk/          # Ứng dụng chính
│   └── main.go    # Khởi tạo server, database, dependencies
└── swag/          # Swagger documentation generator
```

**Chức năng:**
- Khởi tạo database connection
- Setup router và middleware
- Dependency injection (kết nối các modules)
- Khởi động server

**Ví dụ code trong `main.go`:**
```go
func main() {
    // 1. Load config
    config := loadConfig()
    
    // 2. Init database
    db := initDatabase(config)
    
    // 3. Wire up modules
    authModule := initAuthModule(db)
    
    // 4. Start server
    router.Run(":8080")
}
```

---

### `internal/` - Core Business Logic (DDD Layers)

**Mục đích:** Chứa toàn bộ code business logic, tổ chức theo DDD

**Quy tắc quan trọng:**
- ✅ Code trong `internal/` KHÔNG thể import từ bên ngoài project
- ✅ Tổ chức theo domain/module, KHÔNG theo technical layer
- ✅ Mỗi module là một bounded context độc lập

```
internal/
├── auth/              # Module xác thực (authentication)
├── user/              # Module quản lý user
├── common/            # Code dùng chung
├── initialize/        # Khởi tạo các modules
└── middleware/        # HTTP middleware (CORS, logging, etc.)
```

---

## 2️⃣ Cấu Trúc Module (Ví dụ: `auth/`)

Mỗi module tuân theo kiến trúc DDD 4 tầng:

```
auth/
├── domain/              # 💎 Tầng Domain (Business Logic)
├── application/         # 🔧 Tầng Application (Use Cases)
├── infrastructure/      # 🏗️ Tầng Infrastructure (Technical)
└── controller/          # 🌐 Tầng Interface (API/UI)
```

---

### 💎 `domain/` - Tầng Domain (Tầng Quan Trọng Nhất)

**Mục đích:** Chứa business logic thuần túy, KHÔNG phụ thuộc vào framework hay database

#### `domain/model/entity/` - Domain Entities

**File:** `account.go`

**Là gì:** 
- Đối tượng có identity (ID) và lifecycle
- Đại diện cho các khái niệm nghiệp vụ cốt lõi
- Chứa business methods

**Ví dụ:**
```go
type Account struct {
    AccountId     int64
    Username      valueobject.Username    // Value object
    Email         valueobject.Email       // Value object  
    Password      valueobject.Password    // Value object
    Status        valueobject.AccountStatus
    CreateTime    time.Time
}

// Business methods
func (a *Account) IsActive() bool {
    return a.Status.IsActive() && a.IsDeleted == 0
}

func (a *Account) CanLogin() bool {
    return a.IsActive() && a.Status.CanLogin()
}
```

**Đặc điểm:**
- ✅ KHÔNG có GORM tags (không có `gorm:"..."`)
- ✅ Sử dụng value objects thay vì primitive types
- ✅ Chứa business rules và validation
- ✅ Có thể thay đổi được (mutable)

---

#### `domain/model/valueobject/` - Value Objects

**Là gì:**
- Đối tượng KHÔNG có identity
- Định nghĩa bởi giá trị của nó
- Immutable (không thay đổi được)
- Tự validate khi tạo

**Các file:**

##### `email.go` - Email Value Object
```go
type Email struct {
    value string  // Private - không thể thay đổi từ bên ngoài
}

func NewEmail(email string) (Email, error) {
    // Validation
    if !isValidEmail(email) {
        return Email{}, ErrInvalidEmailFormat
    }
    // Normalization
    return Email{value: strings.ToLower(email)}, nil
}
```

**Lợi ích:**
- ✅ Không thể tạo email không hợp lệ
- ✅ Type safety (không thể nhầm email với string thường)
- ✅ Validation tự động
- ✅ Tự động lowercase

##### `password.go` - Password Value Object
```go
type Password struct {
    hash string  // Lưu hash, không lưu plaintext
}

func NewPassword(plaintext string) (Password, error) {
    // Validate độ dài
    if len(plaintext) < 8 {
        return Password{}, ErrPasswordTooShort
    }
    
    // Tự động hash
    hash, _ := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
    return Password{hash: string(hash)}, nil
}

func (p Password) Verify(plaintext string) bool {
    return bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plaintext)) == nil
}
```

**Lợi ích:**
- ✅ Tự động hash password
- ✅ Không thể tạo password quá ngắn
- ✅ Method Verify() để kiểm tra
- ✅ Bảo mật cao hơn

##### `username.go` - Username Value Object
```go
type Username struct {
    value string
}

func NewUsername(username string) (Username, error) {
    username = strings.ToLower(strings.TrimSpace(username))
    
    // Validate length
    if len(username) < 3 || len(username) > 20 {
        return Username{}, ErrInvalidUsernameLength
    }
    
    // Validate format
    if !usernameRegex.MatchString(username) {
        return Username{}, ErrInvalidUsernameFormat
    }
    
    return Username{value: username}, nil
}
```

**Lợi ích:**
- ✅ Validate độ dài (3-20 ký tự)
- ✅ Chỉ cho phép alphanumeric + underscore + hyphen
- ✅ Tự động lowercase và trim

##### `account_status.go` - Account Status Value Object
```go
type AccountStatus int

const (
    AccountStatusInactive  AccountStatus = 0
    AccountStatusActive    AccountStatus = 1
    AccountStatusSuspended AccountStatus = 2
)

func (s AccountStatus) CanTransitionTo(target AccountStatus) bool {
    // Business rules cho state transitions
    switch s {
    case AccountStatusInactive:
        return target == AccountStatusActive
    case AccountStatusActive:
        return target == AccountStatusInactive || target == AccountStatusSuspended
    // ...
    }
}
```

**Lợi ích:**
- ✅ Type-safe (không thể dùng số bất kỳ)
- ✅ Business rules cho state transitions
- ✅ Clear và dễ đọc

---

#### `domain/repository/` - Repository Interfaces

**File:** `auth.repository.go`

**Là gì:** 
- Định nghĩa INTERFACE (contract) cho việc truy xuất dữ liệu
- Implementation ở tầng infrastructure
- Return domain entities

**Ví dụ:**
```go
type AuthRepository interface {
    GetById(ctx context.Context, id int64) (*entity.Account, error)
    FindByEmail(ctx context.Context, email string) (*entity.Account, error)
    Create(ctx context.Context, account *entity.Account) (int64, error)
    UsernameExists(ctx context.Context, username string) (bool, error)
}
```

**Đặc điểm:**
- ✅ Chỉ là interface, KHÔNG có implementation
- ✅ Sử dụng domain entities
- ✅ KHÔNG biết về database, GORM, hay SQL

---

#### `domain/service/` - Domain Services

**File:** `account_domain_service.go`

**Là gì:**
- Business logic không thuộc về một entity cụ thể
- Logic liên quan đến nhiều entities
- Stateless (không lưu trạng thái)

**Khi nào dùng:**
- Logic cần nhiều entities
- Kiểm tra uniqueness (username, email)
- Validation phức tạp

**Ví dụ:**
```go
type AccountDomainService struct {
    authRepo repository.AuthRepository
}

func (s *AccountDomainService) CanCreateAccount(
    ctx context.Context,
    username valueobject.Username,
    email valueobject.Email,
) error {
    // Business rule: Username phải unique
    exists, _ := s.authRepo.UsernameExists(ctx, username.String())
    if exists {
        return ErrUsernameAlreadyExists
    }
    
    // Business rule: Email phải unique
    exists, _ = s.authRepo.EmailExists(ctx, email.String())
    if exists {
        return ErrEmailAlreadyExists
    }
    
    return nil
}
```

---

#### `domain/errors.go` - Domain Errors

**Là gì:** 
- Định nghĩa các lỗi đặc thù của domain
- Type-safe error handling
- Dễ dàng translate tại boundaries

**Ví dụ:**
```go
var (
    ErrAccountNotFound        = errors.New("account not found")
    ErrEmailAlreadyExists     = errors.New("email already registered")
    ErrUsernameAlreadyExists  = errors.New("username already taken")
    ErrInvalidCredentials     = errors.New("invalid credentials")
)
```

**Lợi ích:**
- ✅ Type-safe error checking với `errors.Is()`
- ✅ Dễ dàng map sang HTTP status codes
- ✅ Clear và maintainable

---

### 🔧 `application/` - Tầng Application (Use Cases)

**Mục đích:** Điều phối domain objects để thực hiện use cases

#### `application/service/` - Application Services

**Files:**
- `auth.service.go` - Interface
- `auth.service.impl.go` - Implementation

**Là gì:**
- Orchestrator (điều phối viên)
- Thin layer - KHÔNG chứa business logic
- Sử dụng domain services và repositories
- Định nghĩa transaction boundaries

**Ví dụ:**
```go
func (s *authService) Create(ctx context.Context, dto AccountAppDTO) (int64, error) {
    // 1. Tạo value objects (validation tự động)
    username, err := valueobject.NewUsername(dto.Username)
    if err != nil {
        return 0, err
    }
    
    email, err := valueobject.NewEmail(dto.Email)
    if err != nil {
        return 0, err
    }
    
    password, err := valueobject.NewPassword(dto.Password)
    if err != nil {
        return 0, err
    }
    
    // 2. Kiểm tra business rules (dùng domain service)
    err = s.domainService.CanCreateAccount(ctx, username, email)
    if err != nil {
        return 0, err
    }
    
    // 3. Tạo entity
    account := &entity.Account{
        Username: username,
        Email:    email,
        Password: password,
        // ...
    }
    
    // 4. Save qua repository
    return s.authRepo.CreateUser(ctx, account)
}
```

**Trách nhiệm:**
- ✅ Convert DTOs → value objects
- ✅ Gọi domain services
- ✅ Gọi repositories
- ✅ Handle transactions
- ❌ KHÔNG chứa business logic
- ❌ KHÔNG validate (để value objects làm)

---

#### `application/service/dto/` - Application DTOs

**File:** `account.app.dto.go`

**Là gì:**
- Data Transfer Objects cho application layer
- Khác với controller DTOs
- Dùng cho communication giữa layers

**Ví dụ:**
```go
type AccountAppDTO struct {
    Username string
    Email    string
    Password string
    Lang     string
    Status   int
}
```

---

### 🏗️ `infrastructure/` - Tầng Infrastructure (Technical)

**Mục đích:** Implementation chi tiết kỹ thuật (database, cache, external APIs)

#### `infrastructure/persistence/model/` - Database Models

**File:** `account_model.go`

**Là gì:**
- Model cho database với GORM tags
- CHỈ dùng trong infrastructure layer
- KHÔNG được sử dụng ở domain layer

**Ví dụ:**
```go
type AccountModel struct {
    AccountId     int64      `gorm:"column:account_id;primaryKey;autoIncrement"`
    Username      string     `gorm:"column:username;uniqueIndex"`
    Password      string     `gorm:"column:password"`
    Email         string     `gorm:"column:email;uniqueIndex"`
    Status        int        `gorm:"column:status;default:1"`
    CreateTime    time.Time  `gorm:"column:create_time"`
    UpdateTime    time.Time  `gorm:"column:update_time"`
    IsDeleted     int        `gorm:"column:is_deleted;default:0"`
}

func (AccountModel) TableName() string {
    return "drunk_user"
}
```

**Tại sao cần tách:**
- ✅ Domain layer thuần túy (không phụ thuộc GORM)
- ✅ Dễ thay đổi database
- ✅ Dễ test domain layer
- ✅ Clean architecture

---

#### `infrastructure/persistence/mapper/` - Mappers

**File:** `account_mapper.go`

**Là gì:**
- Convert giữa domain entities và database models
- Bidirectional conversion (2 chiều)

**Ví dụ:**
```go
type AccountMapper struct{}

// Database Model → Domain Entity
func (m *AccountMapper) ToDomain(dbModel *model.AccountModel) (*entity.Account, error) {
    // Tạo value objects từ string
    email, err := valueobject.NewEmail(dbModel.Email)
    username, err := valueobject.NewUsername(dbModel.Username)
    password, err := valueobject.NewPasswordFromHash(dbModel.Password)
    status, err := valueobject.NewAccountStatus(dbModel.Status)
    
    return &entity.Account{
        AccountId: dbModel.AccountId,
        Username:  username,
        Email:     email,
        Password:  password,
        Status:    status,
        // ...
    }, nil
}

// Domain Entity → Database Model
func (m *AccountMapper) ToModel(account *entity.Account) (*model.AccountModel, error) {
    return &model.AccountModel{
        AccountId: account.AccountId,
        Username:  account.Username.String(),
        Email:     account.Email.String(),
        Password:  account.Password.Hash(),
        Status:    account.Status.Int(),
        // ...
    }, nil
}
```

**Lợi ích:**
- ✅ Tách biệt domain và infrastructure
- ✅ Dễ test
- ✅ Có thể thay đổi database schema mà không ảnh hưởng domain
- ✅ Error handling tập trung

---

#### `infrastructure/persistence/repository/` - Repository Implementation

**File:** `auth.repository.go`

**Là gì:**
- Implementation cụ thể của repository interface
- Sử dụng GORM để query database
- Dùng mapper để convert

**Ví dụ:**
```go
type authRepository struct {
    db     *gorm.DB
    mapper *mapper.AccountMapper
}

func (r *authRepository) GetById(ctx context.Context, id int64) (*entity.Account, error) {
    // 1. Query database model
    var dbModel model.AccountModel
    err := r.db.WithContext(ctx).
        Where("account_id = ?", id).
        First(&dbModel).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("account not found")
        }
        return nil, err
    }
    
    // 2. Convert sang domain entity
    return r.mapper.ToDomain(&dbModel)
}

func (r *authRepository) Create(ctx context.Context, account *entity.Account) (*entity.Account, error) {
    // 1. Convert domain entity sang database model
    dbModel, err := r.mapper.ToModel(account)
    if err != nil {
        return nil, err
    }
    
    // 2. Save vào database
    err = r.db.WithContext(ctx).Create(dbModel).Error
    if err != nil {
        return nil, err
    }
    
    // 3. Update ID vào entity
    account.AccountId = dbModel.AccountId
    
    return account, nil
}
```

**Trách nhiệm:**
- ✅ Database queries (GORM)
- ✅ Error handling
- ✅ Sử dụng mapper
- ✅ Context propagation
- ❌ KHÔNG chứa business logic

---

#### `infrastructure/cache/` - Cache Implementation

**Mục đích:** Implement caching (Redis, Memcached, etc.)

**Ví dụ cấu trúc:**
```
cache/
└── auth_cache.go
```

---

#### `infrastructure/config/` - Module Configuration

**Mục đích:** Configuration riêng cho module

---

### 🌐 `controller/` - Tầng Interface (API)

**Mục đích:** Xử lý external interfaces (HTTP, gRPC, CLI)

#### `controller/http/` - HTTP Handlers

**File:** `auth.handler.go`

**Là gì:**
- HTTP request handlers
- Sử dụng Gin framework
- Convert HTTP requests → Application service calls
- Return HTTP responses

**Ví dụ:**
```go
type AuthHandler struct {
    service service.AuthService
}

func (h *AuthHandler) Register(c *gin.Context) {
    // 1. Parse request
    var req dto.UserRegisterReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
        return
    }
    
    // 2. Validate
    if err := utils.ValidateStruct(req, validator); err != nil {
        response.Error(c, http.StatusBadRequest, "Validation failed", err.Error())
        return
    }
    
    // 3. Call application service
    accountId, err := h.service.Create(c, appDto.AccountAppDTO{
        Username: req.Username,
        Email:    req.Email,
        Password: req.Password,
        Lang:     req.Lang,
    })
    
    // 4. Handle errors với proper HTTP status codes
    if err != nil {
        if strings.Contains(err.Error(), "already exists") {
            response.Error(c, http.StatusConflict, "Already exists", err.Error())
            return
        }
        if strings.Contains(err.Error(), "invalid") {
            response.Error(c, http.StatusBadRequest, "Invalid input", err.Error())
            return
        }
        response.Error(c, http.StatusInternalServerError, "Server error", "")
        return
    }
    
    // 5. Return success
    response.Success(c, http.StatusCreated, gin.H{"account_id": accountId})
}
```

**Trách nhiệm:**
- ✅ Parse và validate HTTP requests
- ✅ Call application services
- ✅ Error handling với proper HTTP status codes:
  - `200 OK` - Success
  - `201 Created` - Resource created
  - `400 Bad Request` - Validation errors
  - `404 Not Found` - Resource not found
  - `409 Conflict` - Duplicate resources
  - `500 Internal Server Error` - Server errors
- ✅ Format responses
- ❌ KHÔNG chứa business logic

---

#### `controller/dto/` - Request/Response DTOs

**File:** `account.dto.go`

**Là gì:**
- DTOs cho HTTP requests/responses
- Có validation tags
- CHỈ dùng ở controller layer

**Ví dụ:**
```go
type UserRegisterReq struct {
    Username string `json:"username" binding:"required,min=3,max=20"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Lang     string `json:"lang" binding:"required"`
}

type UserProfileReq struct {
    AccountID int64 `json:"account_id" binding:"required"`
}
```

**Validation tags:**
- `required` - Bắt buộc
- `min=3` - Tối thiểu 3 ký tự
- `email` - Validate email format
- `omitempty` - Có thể bỏ qua

---

#### `controller/http/router.go` - Route Registration

**Mục đích:** Đăng ký các routes

**Ví dụ:**
```go
func RegisterRoutes(router *gin.RouterGroup, handler *AuthHandler) {
    auth := router.Group("/auth")
    {
        auth.POST("/register", handler.Register)
        auth.POST("/login", handler.Login)
        auth.POST("/profile", handler.GetUserProfile)
    }
}
```

---

## 3️⃣ Thư Mục Hỗ Trợ (Supporting Directories)

### `pkg/` - Public Packages

**Mục đích:** Packages có thể import từ bên ngoài project

```
pkg/
└── response/
    ├── response.go
    └── error.go
```

**Ví dụ:**
```go
package response

func Success(c *gin.Context, statusCode int, data interface{}) {
    c.JSON(statusCode, APIResponse{
        Code:    statusCode,
        Message: "Success",
        Data:    data,
    })
}

func Error(c *gin.Context, statusCode int, message, detail string) {
    c.JSON(statusCode, APIError{
        Code:    statusCode,
        Message: message,
        Detail:  detail,
    })
}
```

---

### `utils/` - Utility Functions

**Mục đích:** Helper functions dùng chung

**Ví dụ:**
- String manipulation
- Date/time formatting
- Encryption/decryption
- Validation helpers

---

### `global/` - Global Variables

**Mục đích:** Biến global và configuration

**Ví dụ:**
```go
var (
    DB     *gorm.DB
    Cache  *redis.Client
    Logger *zap.Logger
)
```

**⚠️ Warning:** Nên dùng dependency injection thay vì global variables

---

### `environment/` - Environment Configurations

**Mục đích:** File cấu hình theo môi trường

```
environment/
├── dev.env
├── staging.env
└── production.env
```

---

### `scripts/` - Build và Deployment Scripts

**Mục đích:** Automation scripts

**Ví dụ:**
- Database migrations
- Docker compose files
- CI/CD scripts

---

### `tests/` - Integration Tests

**Mục đích:** Tests xuyên suốt nhiều modules

**Unit tests** nên đặt cạnh code (ví dụ: `auth_test.go`)

---

## 📊 Luồng Dữ Liệu (Data Flow)

### Request Flow (HTTP Request → Response)

```
1. HTTP Request
   ↓
2. Controller (auth.handler.go)
   - Parse request
   - Validate input
   ↓
3. Application Service (auth.service.impl.go)
   - Create value objects
   - Call domain service
   - Call repository
   ↓
4. Domain Service (account_domain_service.go)
   - Business logic
   - Check rules
   ↓
5. Repository (auth.repository.go)
   - Use mapper to convert domain → DB
   - Query database (GORM)
   - Use mapper to convert DB → domain
   ↓
6. Response ngược lại qua các layers
   ↓
7. HTTP Response
```

### Ví Dụ Cụ Thể: Đăng Ký User

```
POST /auth/register
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePass123"
}

1. ✅ auth.handler.Register() - Parse request
2. ✅ Validate với binding tags
3. ✅ auth.service.Create()
   - NewUsername("john_doe") → Username value object
   - NewEmail("john@example.com") → Email value object  
   - NewPassword("SecurePass123") → Password value object (auto-hash)
4. ✅ domainService.CanCreateAccount()
   - Check username exists?
   - Check email exists?
5. ✅ authRepo.Create()
   - mapper.ToModel() → AccountModel
   - db.Create(accountModel)
   - mapper.ToDomain() → Account entity
6. ✅ Return HTTP 201 Created

Response:
{
  "code": 201,
  "message": "Success",
  "data": {
    "account_id": 123
  }
}
```

---

## ✅ Checklist Kiểm Tra Cấu Trúc

### Domain Layer
- [ ] Entities KHÔNG có GORM tags
- [ ] Sử dụng value objects thay vì primitives
- [ ] Business logic trong entities/domain services
- [ ] Repository là interfaces

### Application Layer
- [ ] Services điều phối, KHÔNG chứa business logic
- [ ] Sử dụng domain services và repositories
- [ ] Tạo value objects từ DTOs

### Infrastructure Layer
- [ ] Database models có GORM tags
- [ ] Mapper convert domain ↔ database
- [ ] Repository implementation dùng mapper

### Controller Layer
- [ ] HTTP handlers chỉ xử lý HTTP concerns
- [ ] Proper HTTP status codes
- [ ] DTOs có validation tags
- [ ] KHÔNG chứa business logic

---

## 🎯 Tóm Tắt

### Nguyên Tắc Quan Trọng

1. **Dependency Rule:**
   ```
   Controller → Application → Domain ← Infrastructure
   ```
   - Các dependency chỉ đi từ ngoài vào trong
   - Domain KHÔNG phụ thuộc gì cả

2. **Separation of Concerns:**
   - Domain: Business logic
   - Application: Orchestration
   - Infrastructure: Technical details
   - Controller: HTTP handling

3. **Value Objects:**
   - Validation tự động
   - Type safety
   - Immutable

4. **Mapper Pattern:**
   - Tách domain và infrastructure
   - Clean conversion

### Files Quan Trọng Nhất

| File | Mục đích | Layer |
|------|----------|-------|
| `account.go` (entity) | Core business object | Domain |
| `email.go`, `password.go` (value objects) | Validation & type safety | Domain |
| `auth.repository.go` (interface) | Data access contract | Domain |
| `account_domain_service.go` | Complex business logic | Domain |
| `auth.service.impl.go` | Use case orchestration | Application |
| `account_model.go` | Database schema | Infrastructure |
| `account_mapper.go` | Domain ↔ DB conversion | Infrastructure |
| `auth.repository.go` (impl) | Database operations | Infrastructure |
| `auth.handler.go` | HTTP handling | Controller |

---

## � Hướng Dẫn Step-by-Step: Xây Dựng Module Mới

> **Checklist đầy đủ từ khi bắt đầu đến khi hoàn thành một module theo DDD**

---

### 📋 Tổng Quan Quy Trình

```
Planning → Domain → Infrastructure → Application → Interface → Testing
```

**Thời gian ước tính:** 1-3 ngày cho một module cơ bản

---

## Giai Đoạn 1: Lên Kế Hoạch (Planning) - 30 phút

### Bước 1.1: Xác Định Bounded Context

**Câu hỏi cần trả lời:**
- ❓ Module này quản lý gì? (VD: Product, Order, Payment)
- ❓ Trách nhiệm chính là gì?
- ❓ Có phụ thuộc vào module nào khác không?

**Ví dụ:** Module Product
- Quản lý danh sách sản phẩm
- CRUD operations
- Tìm kiếm và filter sản phẩm
- Phụ thuộc: Auth module (để check permissions)

---

### Bước 1.2: Thiết Kế Data Model

**Vẽ ERD đơn giản:**

```
Product
├── id (PK)
├── name
├── description
├── price
├── category_id (FK)
├── stock_quantity
├── created_at
└── updated_at
```

**Xác định:**
- ✅ Primary keys
- ✅ Foreign keys
- ✅ Unique constraints
- ✅ Indexes cần thiết

---

### Bước 1.3: Xác Định Use Cases

Liệt kê các use case chính:

```markdown
## Product Module Use Cases

1. [CREATE] Tạo sản phẩm mới
2. [READ] Xem chi tiết sản phẩm
3. [READ] Danh sách sản phẩm (với pagination)
4. [UPDATE] Cập nhật thông tin sản phẩm
5. [DELETE] Xóa sản phẩm (soft delete)
6. [SEARCH] Tìm kiếm sản phẩm theo tên/category
```

---

## Giai Đoạn 2: Xây Dựng Domain Layer - 2-4 giờ

> **Nguyên tắc:** Domain layer KHÔNG phụ thuộc vào bất kỳ layer nào khác

### Bước 2.1: Tạo Folder Structure

```bash
mkdir -p internal/product/domain/model/{entity,valueobject}
mkdir -p internal/product/domain/repository
mkdir -p internal/product/domain/service
touch internal/product/domain/errors.go
```

---

### Bước 2.2: Tạo Value Objects

**Thứ tự ưu tiên:**
1. Các trường cần validation phức tạp
2. Các trường dùng nhiều lần
3. Các business rules quan trọng

**File:** `internal/product/domain/model/valueobject/price.go`

```go
package valueobject

import "errors"

type Price struct {
    amount   float64
    currency string
}

var (
    ErrInvalidPrice = errors.New("price must be positive")
    ErrInvalidCurrency = errors.New("invalid currency code")
)

func NewPrice(amount float64, currency string) (Price, error) {
    // Validation
    if amount < 0 {
        return Price{}, ErrInvalidPrice
    }
    
    if currency != "USD" && currency != "VND" {
        return Price{}, ErrInvalidCurrency
    }
    
    return Price{
        amount:   amount,
        currency: currency,
    }, nil
}

func (p Price) Amount() float64 {
    return p.amount
}

func (p Price) Currency() string {
    return p.currency
}
```

**Các Value Objects thường dùng:**
- ✅ Email, Phone, Address
- ✅ Money/Price, SKU, ProductCode
- ✅ Status (enum-like types)

---

### Bước 2.3: Tạo Domain Entity

**File:** `internal/product/domain/model/entity/product.go`

```go
package entity

import (
    "time"
    "errors"
    
    "your-project/internal/product/domain/model/valueobject"
)

type Product struct {
    ID            int64
    Name          string
    Description   string
    Price         valueobject.Price
    CategoryID    int64
    StockQuantity int
    IsDeleted     bool
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

// Constructor
func NewProduct(
    name string,
    description string,
    price valueobject.Price,
    categoryID int64,
    stockQuantity int,
) (*Product, error) {
    // Business validations
    if name == "" {
        return nil, errors.New("product name is required")
    }
    
    if stockQuantity < 0 {
        return nil, errors.New("stock quantity cannot be negative")
    }
    
    now := time.Now()
    return &Product{
        Name:          name,
        Description:   description,
        Price:         price,
        CategoryID:    categoryID,
        StockQuantity: stockQuantity,
        IsDeleted:     false,
        CreatedAt:     now,
        UpdatedAt:     now,
    }, nil
}

// Business methods
func (p *Product) IsAvailable() bool {
    return !p.IsDeleted && p.StockQuantity > 0
}

func (p *Product) DecreaseStock(quantity int) error {
    if quantity > p.StockQuantity {
        return errors.New("insufficient stock")
    }
    p.StockQuantity -= quantity
    p.UpdatedAt = time.Now()
    return nil
}

func (p *Product) IncreaseStock(quantity int) {
    p.StockQuantity += quantity
    p.UpdatedAt = time.Now()
}
```

**Checklist Entity:**
- ✅ NO GORM tags
- ✅ Sử dụng value objects
- ✅ Constructor validation
- ✅ Business methods
- ✅ Immutability cho data quan trọng

---

### Bước 2.4: Tạo Repository Interface

**File:** `internal/product/domain/repository/product.repository.go`

```go
package repository

import (
    "context"
    "your-project/internal/product/domain/model/entity"
)

type ProductRepository interface {
    // CRUD
    Create(ctx context.Context, product *entity.Product) (int64, error)
    GetById(ctx context.Context, id int64) (*entity.Product, error)
    Update(ctx context.Context, product *entity.Product) error
    Delete(ctx context.Context, id int64) error
    
    // Queries
    List(ctx context.Context, offset, limit int) ([]*entity.Product, error)
    Search(ctx context.Context, query string) ([]*entity.Product, error)
    FindByCategory(ctx context.Context, categoryID int64) ([]*entity.Product, error)
    
    // Existence checks
    ExistsByName(ctx context.Context, name string) (bool, error)
}
```

**Tips:**
- ✅ Chỉ interface, KHÔNG implementation
- ✅ Context parameter đầu tiên
- ✅ Return domain entities
- ✅ Error handling

---

### Bước 2.5: Tạo Domain Service (Nếu Cần)

**Khi nào cần Domain Service:**
- Logic liên quan đến nhiều entities
- Business rules phức tạp không thuộc về một entity
- Tính toán cần nhiều data

**File:** `internal/product/domain/service/product_domain_service.go`

```go
package service

import (
    "context"
    "errors"
    
    "your-project/internal/product/domain/repository"
    "your-project/internal/product/domain/model/entity"
)

type ProductDomainService struct {
    productRepo repository.ProductRepository
}

func NewProductDomainService(repo repository.ProductRepository) *ProductDomainService {
    return &ProductDomainService{
        productRepo: repo,
    }
}

// Complex business logic
func (s *ProductDomainService) CanCreateProduct(ctx context.Context, name string) error {
    // Business rule: Product name must be unique
    exists, err := s.productRepo.ExistsByName(ctx, name)
    if err != nil {
        return err
    }
    
    if exists {
        return errors.New("product name already exists")
    }
    
    return nil
}
```

---

### Bước 2.6: Định Nghĩa Domain Errors

**File:** `internal/product/domain/errors.go`

```go
package domain

import "errors"

var (
    // Entity errors
    ErrProductNotFound     = errors.New("product not found")
    ErrProductNameExists   = errors.New("product name already exists")
    
    // Business rule errors
    ErrInsufficientStock   = errors.New("insufficient stock")
    ErrInvalidPrice        = errors.New("invalid price")
    
    // Validation errors
    ErrInvalidInput        = errors.New("invalid input")
)
```

---

## Giai Đoạn 3: Xây Dựng Infrastructure Layer - 2-3 giờ

> **Nguyên tắc:** Implementation chi tiết kỹ thuật (database, cache, external APIs)

### Bước 3.1: Tạo Folder Structure

```bash
mkdir -p internal/product/infrastructure/persistence/{model,mapper,repository}
mkdir -p internal/product/infrastructure/cache
```

---

### Bước 3.2: Tạo Database Model

**File:** `internal/product/infrastructure/persistence/model/product_model.go`

```go
package model

import "time"

type ProductModel struct {
    ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
    Name          string    `gorm:"column:name;uniqueIndex;not null"`
    Description   string    `gorm:"column:description"`
    Price         float64   `gorm:"column:price;not null"`
    Currency      string    `gorm:"column:currency;default:'USD'"`
    CategoryID    int64     `gorm:"column:category_id;index"`
    StockQuantity int       `gorm:"column:stock_quantity;default:0"`
    IsDeleted     bool      `gorm:"column:is_deleted;default:false;index"`
    CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
    UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ProductModel) TableName() string {
    return "products"
}
```

---

### Bước 3.3: Tạo Mapper

**File:** `internal/product/infrastructure/persistence/mapper/product_mapper.go`

```go
package mapper

import (
    "fmt"
    
    "your-project/internal/product/domain/model/entity"
    "your-project/internal/product/domain/model/valueobject"
    "your-project/internal/product/infrastructure/persistence/model"
)

type ProductMapper struct{}

func NewProductMapper() *ProductMapper {
    return &ProductMapper{}
}

// DB Model → Domain Entity
func (m *ProductMapper) ToDomain(dbModel *model.ProductModel) (*entity.Product, error) {
    price, err := valueobject.NewPrice(dbModel.Price, dbModel.Currency)
    if err != nil {
        return nil, fmt.Errorf("invalid price in database: %w", err)
    }
    
    return &entity.Product{
        ID:            dbModel.ID,
        Name:          dbModel.Name,
        Description:   dbModel.Description,
        Price:         price,
        CategoryID:    dbModel.CategoryID,
        StockQuantity: dbModel.StockQuantity,
        IsDeleted:     dbModel.IsDeleted,
        CreatedAt:     dbModel.CreatedAt,
        UpdatedAt:     dbModel.UpdatedAt,
    }, nil
}

// Domain Entity → DB Model
func (m *ProductMapper) ToModel(product *entity.Product) (*model.ProductModel, error) {
    return &model.ProductModel{
        ID:            product.ID,
        Name:          product.Name,
        Description:   product.Description,
        Price:         product.Price.Amount(),
        Currency:      product.Price.Currency(),
        CategoryID:    product.CategoryID,
        StockQuantity: product.StockQuantity,
        IsDeleted:     product.IsDeleted,
        CreatedAt:     product.CreatedAt,
        UpdatedAt:     product.UpdatedAt,
    }, nil
}
```

---

### Bước 3.4: Implement Repository

**File:** `internal/product/infrastructure/persistence/repository/product.repository.go`

```go
package repository

import (
    "context"
    "errors"
    "fmt"
    
    "gorm.io/gorm"
    
    "your-project/internal/product/domain/model/entity"
    "your-project/internal/product/domain/repository"
    "your-project/internal/product/infrastructure/persistence/mapper"
    "your-project/internal/product/infrastructure/persistence/model"
)

type productRepository struct {
    db     *gorm.DB
    mapper *mapper.ProductMapper
}

func NewProductRepository(db *gorm.DB) repository.ProductRepository {
    return &productRepository{
        db:     db,
        mapper: mapper.NewProductMapper(),
    }
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) (int64, error) {
    dbModel, err := r.mapper.ToModel(product)
    if err != nil {
        return 0, err
    }
    
    if err := r.db.WithContext(ctx).Create(dbModel).Error; err != nil {
        return 0, fmt.Errorf("failed to create product: %w", err)
    }
    
    return dbModel.ID, nil
}

func (r *productRepository) GetById(ctx context.Context, id int64) (*entity.Product, error) {
    var dbModel model.ProductModel
    
    err := r.db.WithContext(ctx).
        Where("id = ? AND is_deleted = ?", id, false).
        First(&dbModel).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errors.New("product not found")
        }
        return nil, err
    }
    
    return r.mapper.ToDomain(&dbModel)
}

func (r *productRepository) List(ctx context.Context, offset, limit int) ([]*entity.Product, error) {
    var dbModels []model.ProductModel
    
    err := r.db.WithContext(ctx).
        Where("is_deleted = ?", false).
        Offset(offset).
        Limit(limit).
        Order("created_at DESC").
        Find(&dbModels).Error
    
    if err != nil {
        return nil, err
    }
    
    products := make([]*entity.Product, 0, len(dbModels))
    for _, dbModel := range dbModels {
        product, err := r.mapper.ToDomain(&dbModel)
        if err != nil {
            continue // Skip invalid records
        }
        products = append(products, product)
    }
    
    return products, nil
}

// Implement other methods...
```

---

## Giai Đoạn 4: Xây Dựng Application Layer - 1-2 giờ

> **Nguyên tắc:** Orchestration, không có business logic

### Bước 4.1: Tạo Application DTOs

**File:** `internal/product/application/dto/product.dto.go`

```go
package dto

type CreateProductDTO struct {
    Name          string  `json:"name"`
    Description   string  `json:"description"`
    Price         float64 `json:"price"`
    Currency      string  `json:"currency"`
    CategoryID    int64   `json:"category_id"`
    StockQuantity int     `json:"stock_quantity"`
}

type UpdateProductDTO struct {
    ID            int64   `json:"id"`
    Name          string  `json:"name"`
    Description   string  `json:"description"`
    Price         float64 `json:"price"`
    StockQuantity int     `json:"stock_quantity"`
}

type ProductResponseDTO struct {
    ID            int64   `json:"id"`
    Name          string  `json:"name"`
    Description   string  `json:"description"`
    Price         float64 `json:"price"`
    Currency      string  `json:"currency"`
    StockQuantity int     `json:"stock_quantity"`
    IsAvailable   bool    `json:"is_available"`
}
```

---

### Bước 4.2: Tạo Application Service

**File:** `internal/product/application/service/product.service.go`

```go
package service

import (
    "context"
    
    "your-project/internal/product/application/dto"
    "your-project/internal/product/domain/model/entity"
    "your-project/internal/product/domain/model/valueobject"
    "your-project/internal/product/domain/repository"
    "your-project/internal/product/domain/service"
)

type ProductService interface {
    Create(ctx context.Context, dto dto.CreateProductDTO) (int64, error)
    GetById(ctx context.Context, id int64) (*dto.ProductResponseDTO, error)
    List(ctx context.Context, page, pageSize int) ([]*dto.ProductResponseDTO, error)
    Update(ctx context.Context, dto dto.UpdateProductDTO) error
    Delete(ctx context.Context, id int64) error
}

type productService struct {
    productRepo    repository.ProductRepository
    domainService  *service.ProductDomainService
}

func NewProductService(
    repo repository.ProductRepository,
    domainService *service.ProductDomainService,
) ProductService {
    return &productService{
        productRepo:   repo,
        domainService: domainService,
    }
}

func (s *productService) Create(ctx context.Context, dto dto.CreateProductDTO) (int64, error) {
    // 1. Validate với domain service
    if err := s.domainService.CanCreateProduct(ctx, dto.Name); err != nil {
        return 0, err
    }
    
    // 2. Tạo value objects
    price, err := valueobject.NewPrice(dto.Price, dto.Currency)
    if err != nil {
        return 0, err
    }
    
    // 3. Tạo entity
    product, err := entity.NewProduct(
        dto.Name,
        dto.Description,
        price,
        dto.CategoryID,
        dto.StockQuantity,
    )
    if err != nil {
        return 0, err
    }
    
    // 4. Save qua repository
    return s.productRepo.Create(ctx, product)
}

func (s *productService) GetById(ctx context.Context, id int64) (*dto.ProductResponseDTO, error) {
    product, err := s.productRepo.GetById(ctx, id)
    if err != nil {
        return nil, err
    }
    
    return &dto.ProductResponseDTO{
        ID:            product.ID,
        Name:          product.Name,
        Description:   product.Description,
        Price:         product.Price.Amount(),
        Currency:      product.Price.Currency(),
        StockQuantity: product.StockQuantity,
        IsAvailable:   product.IsAvailable(),
    }, nil
}
```

---

## Giai Đoạn 5: Xây Dựng Interface Layer - 1-2 giờ

> **Nguyên tắc:** HTTP handling, không có business logic

### Bước 5.1: Tạo Controller DTOs

**File:** `internal/product/controller/dto/product.dto.go`

```go
package dto

type CreateProductRequest struct {
    Name          string  `json:"name" binding:"required,min=3,max=100"`
    Description   string  `json:"description"`
    Price         float64 `json:"price" binding:"required,gt=0"`
    Currency      string  `json:"currency" binding:"required,oneof=USD VND"`
    CategoryID    int64   `json:"category_id" binding:"required"`
    StockQuantity int     `json:"stock_quantity" binding:"gte=0"`
}

type UpdateProductRequest struct {
    Name          string  `json:"name" binding:"required,min=3,max=100"`
    Description   string  `json:"description"`
    Price         float64 `json:"price" binding:"required,gt=0"`
    StockQuantity int     `json:"stock_quantity" binding:"gte=0"`
}

type ListProductRequest struct {
    Page     int `form:"page" binding:"gte=1"`
    PageSize int `form:"page_size" binding:"gte=1,lte=100"`
}
```

---

### Bước 5.2: Tạo HTTP Handler

**File:** `internal/product/controller/http/product.handler.go`

```go
package http

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    
    "your-project/internal/product/application/dto"
    "your-project/internal/product/application/service"
    ctlDto "your-project/internal/product/controller/dto"
    "your-project/pkg/response"
)

type ProductHandler struct {
    service service.ProductService
}

func NewProductHandler(service service.ProductService) *ProductHandler {
    return &ProductHandler{service: service}
}

// Create Product
func (h *ProductHandler) Create(c *gin.Context) {
    var req ctlDto.CreateProductRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid request", err.Error())
        return
    }
    
    productID, err := h.service.Create(c.Request.Context(), dto.CreateProductDTO{
        Name:          req.Name,
        Description:   req.Description,
        Price:         req.Price,
        Currency:      req.Currency,
        CategoryID:    req.CategoryID,
        StockQuantity: req.StockQuantity,
    })
    
    if err != nil {
        if strings.Contains(err.Error(), "already exists") {
            response.Error(c, http.StatusConflict, "Product exists", err.Error())
            return
        }
        response.Error(c, http.StatusInternalServerError, "Failed to create", err.Error())
        return
    }
    
    response.Success(c, http.StatusCreated, gin.H{"product_id": productID})
}

// Get Product by ID
func (h *ProductHandler) GetById(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid ID", err.Error())
        return
    }
    
    product, err := h.service.GetById(c.Request.Context(), id)
    if err != nil {
        response.Error(c, http.StatusNotFound, "Product not found", err.Error())
        return
    }
    
    response.Success(c, http.StatusOK, product)
}

// List Products
func (h *ProductHandler) List(c *gin.Context) {
    var req ctlDto.ListProductRequest
    
    if err := c.ShouldBindQuery(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid query", err.Error())
        return
    }
    
    products, err := h.service.List(c.Request.Context(), req.Page, req.PageSize)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to list", err.Error())
        return
    }
    
    response.Success(c, http.StatusOK, gin.H{
        "products": products,
        "page":     req.Page,
        "size":     len(products),
    })
}

// Implement Update, Delete...
```

---

### Bước 5.3: Đăng Ký Routes

**File:** `internal/product/controller/http/routes.go`

```go
package http

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, handler *ProductHandler) {
    products := router.Group("/products")
    {
        products.POST("", handler.Create)
        products.GET("/:id", handler.GetById)
        products.GET("", handler.List)
        products.PUT("/:id", handler.Update)
        products.DELETE("/:id", handler.Delete)
    }
}
```

---

## Giai Đoạn 6: Initialization & Dependency Injection - 30 phút

### Bước 6.1: Repository Initialization

**File:** `internal/initialize/product/repository.go`

```go
package product

import (
    "gorm.io/gorm"
    
    "your-project/internal/product/domain/repository"
    productRepo "your-project/internal/product/infrastructure/persistence/repository"
)

func InitProductRepository(db *gorm.DB) repository.ProductRepository {
    return productRepo.NewProductRepository(db)
}
```

---

### Bước 6.2: Service Initialization

**File:** `internal/initialize/product/service.go`

```go
package product

import (
    "your-project/internal/product/application/service"
    "your-project/internal/product/domain/repository"
    domainService "your-project/internal/product/domain/service"
)

func InitProductService(repo repository.ProductRepository) service.ProductService {
    domainSvc := domainService.NewProductDomainService(repo)
    return service.NewProductService(repo, domainSvc)
}
```

---

### Bước 6.3: Complete Module Initialization

**File:** `internal/initialize/product/module.go`

```go
package product

import (
    "gorm.io/gorm"
    
    "your-project/internal/product/controller/http"
)

func InitProductModule(db *gorm.DB) *http.ProductHandler {
    // 1. Init repository
    repo := InitProductRepository(db)
    
    // 2. Init service
    service := InitProductService(repo)
    
    // 3. Init handler
    handler := http.NewProductHandler(service)
    
    return handler
}
```

---

### Bước 6.4: Register Routes in Main Router

**File:** `internal/initialize/router.go`

```go
import (
    productHTTP "your-project/internal/product/controller/http"
    initProduct "your-project/internal/initialize/product"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
    router := gin.Default()
    
    api := router.Group("/api/v1")
    {
        // Product module
        productHandler := initProduct.InitProductModule(db)
        productHTTP.RegisterRoutes(api, productHandler)
        
        // Other modules...
    }
    
    return router
}
```

---

## Giai Đoạn 7: Database Migration - 15 phút

### Bước 7.1: Tạo Migration File

**File:** `scripts/migrations/001_create_products_table.sql`

```sql
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    category_id BIGINT,
    stock_quantity INT NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0),
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_deleted ON products(is_deleted);
CREATE INDEX idx_products_created ON products(created_at DESC);
```

### Bước 7.2: Run Migration

```bash
# Using GORM AutoMigrate
go run cmd/migrate/main.go

# Or using migration tool
migrate -path scripts/migrations -database "postgres://..." up
```

---

## Giai Đoạn 8: Testing - 2-3 giờ

### Bước 8.1: Unit Test Value Objects

**File:** `internal/product/domain/model/valueobject/price_test.go`

```go
package valueobject_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "your-project/internal/product/domain/model/valueobject"
)

func TestNewPrice_Valid(t *testing.T) {
    price, err := valueobject.NewPrice(100.50, "USD")
    
    assert.NoError(t, err)
    assert.Equal(t, 100.50, price.Amount())
    assert.Equal(t, "USD", price.Currency())
}

func TestNewPrice_NegativeAmount(t *testing.T) {
    _, err := valueobject.NewPrice(-10, "USD")
    
    assert.Error(t, err)
    assert.Equal(t, valueobject.ErrInvalidPrice, err)
}
```

---

### Bước 8.2: Integration Test Repository

```go
func TestProductRepository_Create(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    repo := repository.NewProductRepository(db)
    
    // Create test product
    price, _ := valueobject.NewPrice(99.99, "USD")
    product, _ := entity.NewProduct("Test Product", "Description", price, 1, 10)
    
    // Execute
    id, err := repo.Create(context.Background(), product)
    
    // Assert
    assert.NoError(t, err)
    assert.Greater(t, id, int64(0))
}
```

---

### Bước 8.3: Manual API Testing

```bash
# Create Product
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPhone 15",
    "description": "Latest iPhone",
    "price": 999.99,
    "currency": "USD",
    "category_id": 1,
    "stock_quantity": 50
  }'

# Get Product
curl http://localhost:8080/api/v1/products/1

# List Products
curl "http://localhost:8080/api/v1/products?page=1&page_size=10"
```

---

## ✅ Checklist Hoàn Thành

### Domain Layer
- [ ] Value objects với validation
- [ ] Domain entity thuần túy (no GORM tags)
- [ ] Repository interface
- [ ] Domain service (nếu cần)
- [ ] Domain errors

### Infrastructure Layer
- [ ] Database model với GORM tags
- [ ] Mapper (domain ↔ database)
- [ ] Repository implementation

### Application Layer
- [ ] Application DTOs
- [ ] Application service (orchestration)

### Interface Layer
- [ ] Controller DTOs với validation tags
- [ ] HTTP handlers
- [ ] Routes registration

### Initialization
- [ ] Module initialization
- [ ] Dependency injection
- [ ] Router integration

### Database
- [ ] Migration file
- [ ] Indexes created

### Testing
- [ ] Unit tests (value objects, entities)
- [ ] Integration tests (repository)
- [ ] API manual testing

---

## 🎯 Tips Quan Trọng

### DO ✅

1. **Bắt đầu từ Domain** - Core business logic trước
2. **Test từng layer** - Đừng chờ đến cuối
3. **Validation ở nhiều layer** - Defense in depth
4. **Sử dụng value objects** - Type safety
5. **Async cache updates** - Don't block requests

### DON'T ❌

1. **GORM tags trong domain** - Tách riêng infrastructure
2. **Business logic ở controller** - Đặt trong domain
3. **Skip validation** - Validate ở mọi boundary
4. **Tạo giant services** - Chia nhỏ responsibilities
5. **Ignore errors** - Handle đúng cách

---

## 📊 Thời Gian Ước Tính

| Giai đoạn | Thời gian | Độ ưu tiên |
|-----------|-----------|------------|
| Planning | 30 phút | 🔴 High |
| Domain Layer | 2-4 giờ | 🔴 High |
| Infrastructure | 2-3 giờ | 🔴 High |
| Application | 1-2 giờ | 🟡 Medium |
| Interface | 1-2 giờ | 🟡 Medium |
| Initialization | 30 phút | 🟢 Low |
| Migration | 15 phút | 🟢 Low |
| Testing | 2-3 giờ | 🔴 High |
| **TOTAL** | **1-2 ngày** | - |

---

## 🚀 Quick Start Template

Copy folder structure này để bắt đầu nhanh:

```bash
# Script tạo module mới
./scripts/create-module.sh product

# Tạo structure:
internal/product/
├── domain/
│   ├── model/
│   │   ├── entity/
│   │   └── valueobject/
│   ├── repository/
│   ├── service/
│   └── errors.go
├── application/
│   ├── dto/
│   └── service/
├── infrastructure/
│   ├── persistence/
│   │   ├── model/
│   │   ├── mapper/
│   │   └── repository/
│   └── cache/
└── controller/
    ├── dto/
    └── http/
```

---

## �📚 Đọc Thêm

- [DDD Evaluation Report](./ddd_evaluation.md) - Đánh giá chi tiết
- [Folder Structure Docs](./folder_structure_docs.md) - Tài liệu đầy đủ (English)
- [DDD Examples](./ddd_examples.md) - Ví dụ thực tế
- [Implementation Summary](./implementation_summary.md) - Tóm tắt changes
- [Cache Demo](./CACHE_DEMO.md) - Redis caching guide

---

**Tạo bởi:** Antigravity AI  
**Ngày:** 2025-12-07  
**Phiên bản:** 2.0
