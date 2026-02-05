# Demo: Sử Dụng Cache Trong DDD Architecture

> **Hướng dẫn chi tiết cách integrate Redis cache vào DDD với Repository Pattern**

---

## 📋 Mục Lục

1. [Tổng Quan](#tổng-quan)
2. [Kiến Trúc Cache Layer](#kiến-trúc-cache-layer)
3. [Implementation Chi Tiết](#implementation-chi-tiết)
4. [Sử Dụng Trong Code](#sử-dụng-trong-code)
5. [Best Practices](#best-practices)

---

## Tổng Quan

### Vị Trí Cache Trong DDD

```
┌─────────────────────────────────────────────────┐
│                Controller Layer                  │
│           (HTTP Handlers)                        │
└────────────────┬────────────────────────────────┘
                 │
┌────────────────▼────────────────────────────────┐
│             Application Layer                    │
│        (Services - Orchestration)                │
└────────────────┬────────────────────────────────┘
                 │
        ┌────────┴─────────┐
        │                  │
┌───────▼──────┐   ┌──────▼────────┐
│Domain Layer  │   │Infrastructure │
│(Interfaces)  │   │  (Cache Impl) │
└──────────────┘   └───────────────┘
```

### Nguyên Tắc

✅ **Cache interface** → Domain Layer  
✅ **Cache implementation** → Infrastructure Layer  
✅ **Cache decorator** cho Repository → Infrastructure Layer  
✅ Transparent caching (business logic không biết về cache)

---

## Kiến Trúc Cache Layer

### File Structure

```
internal/auth/
├── domain/
│   ├── cache/
│   │   └── auth_cache.go              ✅ Interface (Domain)
│   └── repository/
│       └── auth.repository.go
├── infrastructure/
│   ├── cache/
│   │   └── redis_auth_cache.go        ✅ Redis Implementation
│   └── persistence/
│       └── repository/
│           ├── auth.repository.go
│           └── cached_auth.repository.go  ✅ Repository Decorator
└── application/
    └── service/
        └── auth.service.impl.go       ✅ Sử dụng Cached Repository
```

---

## Implementation Chi Tiết

### 1. Cache Interface (Domain Layer)

**File:** `/internal/auth/domain/cache/auth_cache.go`

```go
package cache

import (
	"context"
	"time"

	"CONVERDA/internal/auth/domain/model/entity"
)

// AuthCache định nghĩa interface cho auth caching
// Interface này ở domain layer, implementation ở infrastructure
type AuthCache interface {
	// Get account from cache by ID
	GetAccount(ctx context.Context, accountId int64) (*entity.Account, error)
	
	// Set account to cache
	SetAccount(ctx context.Context, account *entity.Account, ttl time.Duration) error
	
	// Delete account from cache
	DeleteAccount(ctx context.Context, accountId int64) error
	
	// Get account by username
	GetAccountByUsername(ctx context.Context, username string) (*entity.Account, error)
	
	// Set account by username
	SetAccountByUsername(ctx context.Context, username string, account *entity.Account, ttl time.Duration) error
	
	// Delete account by username
	DeleteAccountByUsername(ctx context.Context, username string) error
	
	// Clear all auth cache
	ClearAll(ctx context.Context) error
}
```

**Giải thích:**
- ✅ Interface thuần túy (không phụ thuộc Redis)
- ✅ Trả về domain entities
- ✅ Có TTL (time to live) parameter
- ✅ Context cho cancellation

---

### 2. Redis Cache Implementation (Infrastructure Layer)

**File:** `/internal/auth/infrastructure/cache/redis_auth_cache.go`

```go
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	
	"CONVERDA/internal/auth/domain/cache"
	"CONVERDA/internal/auth/domain/model/entity"
	"CONVERDA/internal/auth/infrastructure/persistence/mapper"
	"CONVERDA/internal/auth/infrastructure/persistence/model"
)

// RedisAuthCache implements AuthCache interface using Redis
type RedisAuthCache struct {
	client *redis.Client
	mapper *mapper.AccountMapper
	prefix string // Key prefix để tránh conflict
}

// NewRedisAuthCache creates a new Redis cache instance
func NewRedisAuthCache(client *redis.Client) cache.AuthCache {
	return &RedisAuthCache{
		client: client,
		mapper: mapper.NewAccountMapper(),
		prefix: "auth:account:",
	}
}

// GetAccount retrieves account from cache by ID
func (c *RedisAuthCache) GetAccount(ctx context.Context, accountId int64) (*entity.Account, error) {
	key := c.accountKey(accountId)
	
	// Get from Redis
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss (không phải error)
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}
	
	// Deserialize
	var dbModel model.AccountModel
	if err := json.Unmarshal(data, &dbModel); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}
	
	// Convert to domain entity
	return c.mapper.ToDomain(&dbModel)
}

// SetAccount stores account in cache
func (c *RedisAuthCache) SetAccount(ctx context.Context, account *entity.Account, ttl time.Duration) error {
	key := c.accountKey(account.AccountId)
	
	// Convert to database model
	dbModel, err := c.mapper.ToModel(account)
	if err != nil {
		return fmt.Errorf("mapper error: %w", err)
	}
	
	// Serialize
	data, err := json.Marshal(dbModel)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}
	
	// Set to Redis with TTL
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	
	return nil
}

// DeleteAccount removes account from cache
func (c *RedisAuthCache) DeleteAccount(ctx context.Context, accountId int64) error {
	key := c.accountKey(accountId)
	return c.client.Del(ctx, key).Err()
}

// GetAccountByUsername retrieves account by username
func (c *RedisAuthCache) GetAccountByUsername(ctx context.Context, username string) (*entity.Account, error) {
	key := c.usernameKey(username)
	
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}
	
	var dbModel model.AccountModel
	if err := json.Unmarshal(data, &dbModel); err != nil {
		return nil, fmt.Errorf("json unmarshal error: %w", err)
	}
	
	return c.mapper.ToDomain(&dbModel)
}

// SetAccountByUsername stores account by username
func (c *RedisAuthCache) SetAccountByUsername(ctx context.Context, username string, account *entity.Account, ttl time.Duration) error {
	key := c.usernameKey(username)
	
	dbModel, err := c.mapper.ToModel(account)
	if err != nil {
		return err
	}
	
	data, err := json.Marshal(dbModel)
	if err != nil {
		return err
	}
	
	return c.client.Set(ctx, key, data, ttl).Err()
}

// DeleteAccountByUsername removes account by username
func (c *RedisAuthCache) DeleteAccountByUsername(ctx context.Context, username string) error {
	key := c.usernameKey(username)
	return c.client.Del(ctx, key).Err()
}

// ClearAll clears all auth cache
func (c *RedisAuthCache) ClearAll(ctx context.Context) error {
	pattern := c.prefix + "*"
	
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	
	return iter.Err()
}

// Helper methods for key generation
func (c *RedisAuthCache) accountKey(accountId int64) string {
	return fmt.Sprintf("%sid:%d", c.prefix, accountId)
}

func (c *RedisAuthCache) usernameKey(username string) string {
	return fmt.Sprintf("%susername:%s", c.prefix, username)
}
```

**Giải thích:**
- ✅ Implement interface từ domain
- ✅ Dùng JSON để serialize/deserialize
- ✅ Dùng mapper để convert domain ↔ model
- ✅ Key prefix để organize
- ✅ TTL cho cache expiration
- ✅ Handle cache miss (return nil, nil)

---

### 3. Cached Repository Decorator (Infrastructure Layer)

**File:** `/internal/auth/infrastructure/persistence/repository/cached_auth.repository.go`

```go
package repository

import (
	"context"
	"time"

	"CONVERDA/internal/auth/domain/cache"
	"CONVERDA/internal/auth/domain/model/entity"
	"CONVERDA/internal/auth/domain/repository"
)

const (
	// Cache TTL
	AccountCacheTTL = 15 * time.Minute
)

// CachedAuthRepository wraps AuthRepository with caching
// Sử dụng Decorator Pattern
type CachedAuthRepository struct {
	repo  repository.AuthRepository // Original repository
	cache cache.AuthCache            // Cache layer
}

// NewCachedAuthRepository creates a cached repository
func NewCachedAuthRepository(
	repo repository.AuthRepository,
	cache cache.AuthCache,
) repository.AuthRepository {
	return &CachedAuthRepository{
		repo:  repo,
		cache: cache,
	}
}

// GetById implements repository.AuthRepository with caching
func (r *CachedAuthRepository) GetById(ctx context.Context, accountId int64) (*entity.Account, error) {
	// 1. Try cache first (Cache-Aside pattern)
	account, err := r.cache.GetAccount(ctx, accountId)
	if err != nil {
		// Log error nhưng không fail, fallback to database
		// log.Printf("Cache error: %v", err)
	}
	
	// 2. Cache hit - return immediately
	if account != nil {
		return account, nil
	}
	
	// 3. Cache miss - query database
	account, err = r.repo.GetById(ctx, accountId)
	if err != nil {
		return nil, err
	}
	
	// 4. Update cache (fire and forget - không block)
	go func() {
		if err := r.cache.SetAccount(context.Background(), account, AccountCacheTTL); err != nil {
			// Log error
			// log.Printf("Failed to set cache: %v", err)
		}
	}()
	
	return account, nil
}

// GetByUsername implements repository.AuthRepository with caching
func (r *CachedAuthRepository) Login(ctx context.Context, username string) (*entity.Account, error) {
	// 1. Try cache
	account, err := r.cache.GetAccountByUsername(ctx, username)
	if err != nil {
		// Log but don't fail
	}
	
	if account != nil {
		return account, nil
	}
	
	// 2. Query database
	account, err = r.repo.Login(ctx, username)
	if err != nil {
		return nil, err
	}
	
	// 3. Update cache (async)
	go func() {
		bgCtx := context.Background()
		_ = r.cache.SetAccountByUsername(bgCtx, username, account, AccountCacheTTL)
		_ = r.cache.SetAccount(bgCtx, account, AccountCacheTTL)
	}()
	
	return account, nil
}

// CreateUser implements repository.AuthRepository
// Khi create, cần invalidate cache nếu cần
func (r *CachedAuthRepository) CreateUser(ctx context.Context, account *entity.Account) (int64, error) {
	// 1. Create in database
	accountId, err := r.repo.CreateUser(ctx, account)
	if err != nil {
		return 0, err
	}
	
	// 2. Update account ID
	account.AccountId = accountId
	
	// 3. Set cache (async)
	go func() {
		bgCtx := context.Background()
		_ = r.cache.SetAccount(bgCtx, account, AccountCacheTTL)
		if !account.Username.IsEmpty() {
			_ = r.cache.SetAccountByUsername(bgCtx, account.Username.String(), account, AccountCacheTTL)
		}
	}()
	
	return accountId, nil
}

// UsernameExists implements repository.AuthRepository
// Existence checks thường không cache vì cần real-time accuracy
func (r *CachedAuthRepository) UsernameExists(ctx context.Context, userName string) (bool, error) {
	return r.repo.UsernameExists(ctx, userName)
}

// EmailExists implements repository.AuthRepository
func (r *CachedAuthRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.repo.EmailExists(ctx, email)
}
```

**Giải thích Decorator Pattern:**
- ✅ Wrap original repository
- ✅ Implement cùng interface
- ✅ Cache-Aside pattern (check cache → miss → query DB → update cache)
- ✅ Async cache updates (không block response)
- ✅ Graceful degradation (cache error không làm fail request)

---

### 4. Redis Client Setup

**File:** `/internal/initialize/redis.go`

```go
package initialize

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func InitRedis(config RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}
```

---

### 5. Dependency Injection Setup

**File:** `/internal/initialize/auth/repository.go`

```go
package auth

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	authCache "CONVERDA/internal/auth/infrastructure/cache"
	authRepo "CONVERDA/internal/auth/infrastructure/persistence/repository"
	"CONVERDA/internal/auth/domain/repository"
)

// InitAuthRepository khởi tạo auth repository với cache
func InitAuthRepository(db *gorm.DB, redisClient *redis.Client) repository.AuthRepository {
	// 1. Tạo base repository
	baseRepo := authRepo.NewAuthRepository(db)
	
	// 2. Tạo cache layer
	cache := authCache.NewRedisAuthCache(redisClient)
	
	// 3. Wrap với cache decorator
	cachedRepo := authRepo.NewCachedAuthRepository(baseRepo, cache)
	
	return cachedRepo
}
```

---

### 6. Main Application Setup

**File:** `/cmd/drunk/main.go`

```go
package main

import (
	"log"

	"CONVERDA/internal/initialize"
	initAuth "CONVERDA/internal/initialize/auth"
)

func main() {
	// 1. Load config
	config := loadConfig()

	// 2. Initialize database
	db, err := initialize.InitDatabase(config.Database)
	if err != nil {
		log.Fatal("Failed to init database:", err)
	}

	// 3. Initialize Redis
	redisClient, err := initialize.InitRedis(initialize.RedisConfig{
		Host:     config.Redis.Host,
		Port:     config.Redis.Port,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})
	if err != nil {
		log.Fatal("Failed to init Redis:", err)
	}
	defer redisClient.Close()

	// 4. Initialize repositories với cache
	authRepo := initAuth.InitAuthRepository(db, redisClient)

	// 5. Initialize services
	authService := initAuth.InitAuthService(authRepo)

	// 6. Initialize handlers
	authHandler := initAuth.InitAuthHandler(authService)

	// 7. Setup router
	router := setupRouter(authHandler)

	// 8. Start server
	log.Fatal(router.Run(":8080"))
}
```

---

## Sử Dụng Trong Code

### Từ Góc Nhìn Application Service

**Application service không biết về cache!**

```go
// internal/auth/application/service/auth.service.impl.go

func (s *authService) GetAccountById(ctx context.Context, accountId int64) (*entity.Account, error) {
	// Repository tự động handle cache
	// Service không cần biết cache exist hay không
	account, err := s.authRepo.GetById(ctx, accountId)
	if err != nil {
		return nil, err
	}
	
	return account, nil
}
```

**Lợi ích:**
- ✅ Business logic sạch, không bị ô nhiễm bởi cache logic
- ✅ Dễ test (có thể inject non-cached repo cho testing)
- ✅ Cache là implementation detail

---

### Cache Invalidation Strategy

#### Strategy 1: Time-based (TTL)
```go
// Tự động expire sau 15 phút
const AccountCacheTTL = 15 * time.Minute
```

#### Strategy 2: Write-through
```go
// Khi update, invalidate cache
func (r *CachedAuthRepository) Update(ctx context.Context, account *entity.Account) error {
	// 1. Update database
	err := r.repo.Update(ctx, account)
	if err != nil {
		return err
	}
	
	// 2. Invalidate cache
	go func() {
		bgCtx := context.Background()
		_ = r.cache.DeleteAccount(bgCtx, account.AccountId)
		_ = r.cache.DeleteAccountByUsername(bgCtx, account.Username.String())
	}()
	
	return nil
}
```

#### Strategy 3: Write-behind (Update cache)
```go
// Khi update, update cache luôn
func (r *CachedAuthRepository) Update(ctx context.Context, account *entity.Account) error {
	err := r.repo.Update(ctx, account)
	if err != nil {
		return err
	}
	
	// Update cache với data mới
	go func() {
		bgCtx := context.Background()
		_ = r.cache.SetAccount(bgCtx, account, AccountCacheTTL)
	}()
	
	return nil
}
```

---

## Testing

### Unit Test cho Cache Implementation

```go
package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"

	"CONVERDA/internal/auth/infrastructure/cache"
)

func TestRedisAuthCache_GetAccount(t *testing.T) {
	// Setup
	db, mock := redismock.NewClientMock()
	authCache := cache.NewRedisAuthCache(db)
	ctx := context.Background()
	
	// Mock Redis response
	mock.ExpectGet("auth:account:id:1").SetVal(`{"account_id":1,"username":"john"}`)
	
	// Execute
	account, err := authCache.GetAccount(ctx, 1)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, account)
	assert.Equal(t, int64(1), account.AccountId)
}
```

### Integration Test cho Cached Repository

```go
func TestCachedRepository_GetById_CacheHit(t *testing.T) {
	// Setup
	mockRepo := &MockAuthRepository{}
	mockCache := &MockAuthCache{}
	cachedRepo := NewCachedAuthRepository(mockRepo, mockCache)
	
	// Mock cache hit
	expectedAccount := &entity.Account{AccountId: 1}
	mockCache.On("GetAccount", mock.Anything, int64(1)).Return(expectedAccount, nil)
	
	// Execute
	account, err := cachedRepo.GetById(context.Background(), 1)
	
	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedAccount, account)
	mockRepo.AssertNotCalled(t, "GetById") // DB không được gọi
}
```

---

## Best Practices

### ✅ DO

1. **Cache ở Infrastructure Layer**
   ```go
   // ✅ GOOD
   internal/auth/infrastructure/cache/
   ```

2. **Interface ở Domain Layer**
   ```go
   // ✅ GOOD
   internal/auth/domain/cache/
   ```

3. **Async Cache Updates**
   ```go
   // ✅ GOOD - không block response
   go func() {
       _ = cache.SetAccount(ctx, account, ttl)
   }()
   ```

4. **Graceful Degradation**
   ```go
   // ✅ GOOD
   account, err := cache.GetAccount(ctx, id)
   if err != nil {
       log.Error(err) // Log nhưng không fail
   }
   ```

5. **Key Naming Convention**
   ```go
   // ✅ GOOD
   "auth:account:id:123"
   "auth:account:username:john"
   ```

### ❌ DON'T

1. **Cache Logic trong Domain Entities**
   ```go
   // ❌ BAD
   func (a *Account) GetByIdFromCache(id int64) {...}
   ```

2. **Cache Logic trong Application Service**
   ```go
   // ❌ BAD
   func (s *AuthService) GetAccount(id int64) {
       // Check cache...
       // Query DB...
   }
   ```

3. **Synchronous Cache Failures Block Requests**
   ```go
   // ❌ BAD
   if err := cache.Set(account); err != nil {
       return err // Fail request vì cache error
   }
   ```

4. **Cache Everything**
   - ❌ Uniqueness checks (UsernameExists, EmailExists)
   - ❌ Real-time data
   - ❌ Frequently changing data

5. **Infinite TTL**
   ```go
   // ❌ BAD
   cache.Set(account, 0) // Không bao giờ expire
   ```

---

## Cache Patterns

### 1. Cache-Aside (Lazy Loading)
```go
// Read: Check cache → miss → query DB → update cache
account, _ := cache.Get(id)
if account == nil {
    account, _ = db.Get(id)
    cache.Set(account)
}
```

### 2. Write-Through
```go
// Write: Update DB → update cache
db.Update(account)
cache.Set(account)
```

### 3. Write-Behind (Write-Back)
```go
// Write: Update cache → async update DB
cache.Set(account)
go db.Update(account)
```

### 4. Refresh-Ahead
```go
// Proactively refresh cache trước khi expire
if time.Until(cache.ExpireTime(key)) < 2*time.Minute {
    go refreshCache(key)
}
```

---

## Monitoring và Metrics

### Cache Hit Rate

```go
type CacheMetrics struct {
	hits   int64
	misses int64
}

func (m *CacheMetrics) HitRate() float64 {
	total := m.hits + m.misses
	if total == 0 {
		return 0
	}
	return float64(m.hits) / float64(total)
}
```

### Log Cache Operations

```go
func (c *RedisAuthCache) GetAccount(ctx context.Context, id int64) (*entity.Account, error) {
	start := time.Now()
	
	account, err := c.client.Get(ctx, key).Result()
	
	duration := time.Since(start)
	log.Printf("Cache GET: key=%s, hit=%v, duration=%v", key, err == nil, duration)
	
	return account, err
}
```

---

## Kết Luận

### Lợi Ích của Cache trong DDD

1. **Performance** - Giảm database load
2. **Scalability** - Handle nhiều requests hơn
3. **Clean Architecture** - Cache là implementation detail
4. **Flexibility** - Dễ dàng enable/disable cache
5. **Testability** - Mock cache dễ dàng

### Key Takeaways

✅ Cache interface ở **Domain Layer**  
✅ Cache implementation ở **Infrastructure Layer**  
✅ Sử dụng **Decorator Pattern** cho repository  
✅ **Async** cache updates  
✅ **Graceful degradation** khi cache fail  
✅ Proper **key naming** và **TTL**  
✅ **Monitor** cache hit rate  

---

**Happy Caching! 🚀**
