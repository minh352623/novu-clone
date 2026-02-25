# Multi-Tenant Project — Toàn Bộ Những Gì Cần Chú Ý

> **Multi-tenant không khó ở kỹ thuật — khó ở discipline:** mọi người trong team đều phải nhớ *"tôi đang làm trong context của tenant nào?"* ở mọi dòng code.

---

## Mục Lục

1. [Chọn Chiến Lược Isolation](#1-trước-tiên-chọn-chiến-lược-isolation)
2. [Data Layer](#2-data-layer--nguy-hiểm-nhất)
3. [Authentication & Authorization](#3-authentication--authorization)
4. [API Design — Tenant Routing](#4-api-design--tenant-routing)
5. [Tenant Lifecycle Management](#5-tenant-lifecycle-management)
6. [Feature Flags & Plan Limits](#6-feature-flags--plan-limits--billing-aware)
7. [Caching](#7-caching--cực-kỳ-dễ-leak-data)
8. [Async Jobs & Queue](#8-async-jobs--queue--tenant-context-phải-đi-theo)
9. [Observability](#9-observability--monitor-per-tenant)
10. [Security Checklist](#10-security--checklist-không-được-bỏ-qua)
11. [Tenant Isolation Test](#11-điều-tối-quan-trọng--tenant-isolation-test)
12. [Checklist Tổng Hợp](#tóm-tắt--checklist-khi-bắt-đầu-dự-án)

---

## 1. Trước Tiên: Chọn Chiến Lược Isolation

Đây là quyết định **quan trọng nhất**, ảnh hưởng đến mọi thứ phía sau.

### 3 Mô Hình Chính

| Mô hình | Mô tả | Pros | Cons | Dùng khi |
|---------|-------|------|------|----------|
| **Silo** (Database per tenant) | Mỗi tenant 1 DB riêng | Isolation hoàn toàn, dễ backup/restore từng tenant | Chi phí cao, khó manage nhiều DB | Yêu cầu compliance cao (healthcare, finance) |
| **Bridge** (Schema per tenant) | Chung DB, riêng schema | Cân bằng isolation vs chi phí | Phức tạp migration, giới hạn số schema | SaaS vừa, vài trăm tenants |
| **Pool** (Shared schema) | Chung DB, dùng `tenant_id` column | Chi phí thấp, dễ scale | Data leak risk nếu miss filter | SaaS lớn, hàng nghìn tenants |

> **Thực tế production:** Hầu hết SaaS hiện đại dùng **Pool** hoặc **Hybrid** — Pool cho standard plan, Silo cho enterprise plan.

### So Sánh Chi Phí & Độ Phức Tạp

```
Silo    │████████████████████│ Isolation cao nhất   │ Chi phí cao nhất
Bridge  │████████████        │ Isolation tốt        │ Chi phí trung bình
Pool    │████                │ Isolation cơ bản     │ Chi phí thấp nhất
```

---

## 2. Data Layer — Nguy Hiểm Nhất

### 2.1 Tenant ID Phải Có Mặt Ở Khắp Nơi

**Rule tuyệt đối:** Mọi bảng đều phải có `tenant_id`. Không có ngoại lệ.

```sql
CREATE TABLE orders (
    id          UUID PRIMARY KEY,
    tenant_id   UUID NOT NULL,          -- ← KHÔNG ĐƯỢC THIẾU
    user_id     UUID NOT NULL,
    amount      DECIMAL(15, 2),
    status      VARCHAR(50),
    created_at  TIMESTAMPTZ DEFAULT NOW(),

    -- Index PHẢI bắt đầu bằng tenant_id
    INDEX idx_orders_tenant      (tenant_id, created_at DESC),
    INDEX idx_orders_tenant_user (tenant_id, user_id),

    FOREIGN KEY (tenant_id) REFERENCES tenants(id)
);
```

> **Lý do index phải bắt đầu bằng `tenant_id`:** Nếu không có nó, query của tenant A sẽ full scan qua data của tất cả các tenant — cực kỳ chậm và nguy hiểm về bảo mật.

### 2.2 Row Level Security (RLS) — Lớp Bảo Vệ Tại DB Level

Với PostgreSQL, enforce isolation ngay tại database, không phụ thuộc vào application code:

```sql
-- Enable RLS trên bảng
ALTER TABLE orders ENABLE ROW LEVEL SECURITY;

-- Policy: mỗi query chỉ thấy data của tenant hiện tại
CREATE POLICY tenant_isolation ON orders
    USING (tenant_id = current_setting('app.current_tenant_id')::UUID);

-- Application set context trước khi query:
SET app.current_tenant_id = 'tenant-uuid-here';
SELECT * FROM orders;
-- → Tự động chỉ trả về data của tenant đó, dù không có WHERE clause
```

RLS là **lưới an toàn thứ 2** — dù application quên filter, DB vẫn chặn data leak.

### 2.3 Migration — Nỗi Đau Lớn Nhất

```
Vấn đề thực tế:
  ALTER TABLE trên bảng 100M rows của 1000 tenants
  → LOCK toàn bộ bảng
  → Downtime cho tất cả tenants

Giải pháp đúng:
  Bước 1: ADD COLUMN ... DEFAULT NULL        (không lock)
  Bước 2: Backfill data theo batch nhỏ       (1000 rows/batch, có sleep delay)
  Bước 3: ADD CONSTRAINT NOT NULL            (sau khi backfill xong)
  Bước 4: Blue/Green migration cho thay đổi schema lớn
```

```sql
-- Sai: Lock cả bảng
ALTER TABLE orders ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}';

-- Đúng: Chia 3 bước
-- Bước 1: Thêm column nullable (không lock)
ALTER TABLE orders ADD COLUMN metadata JSONB;

-- Bước 2: Backfill (chạy batch job, không lock)
UPDATE orders SET metadata = '{}' WHERE metadata IS NULL LIMIT 1000;
-- Lặp lại cho đến hết

-- Bước 3: Sau khi backfill xong, thêm constraint
ALTER TABLE orders ALTER COLUMN metadata SET NOT NULL;
ALTER TABLE orders ALTER COLUMN metadata SET DEFAULT '{}';
```

---

## 3. Authentication & Authorization

### 3.1 JWT Phải Mang Tenant Context

```json
{
  "sub": "user-uuid",
  "tenant_id": "tenant-uuid",
  "tenant_slug": "acme-corp",
  "roles": ["admin"],
  "permissions": ["orders:read", "orders:write", "reports:read"],
  "plan": "enterprise",
  "iat": 1700000000,
  "exp": 1700003600
}
```

### 3.2 Tenant Middleware — Extract & Inject Context

```go
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract tenant từ JWT
        claims, err := parseJWT(r.Header.Get("Authorization"))
        if err != nil {
            http.Error(w, "unauthorized", 401)
            return
        }

        tenantID := claims.TenantID
        if tenantID == "" {
            http.Error(w, "tenant not found", 401)
            return
        }

        // Validate tenant state (không serve nếu tenant bị suspended)
        tenant, err := tenantRepo.FindByID(r.Context(), tenantID)
        if err != nil || tenant.Status != "active" {
            http.Error(w, "tenant unavailable", 403)
            return
        }

        // Inject vào context — truyền xuống mọi layer
        ctx := context.WithValue(r.Context(), TenantKey, tenant)

        // Set cho DB nếu dùng RLS
        db.ExecContext(ctx, "SET LOCAL app.current_tenant_id = $1", tenantID)

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 3.3 RBAC Phải Scoped Theo Tenant

```
Nguyên tắc quan trọng:
  User A là ADMIN trong Tenant X  ≠  User A là ADMIN trong Tenant Y

Một user có thể có vai trò khác nhau ở các tenant khác nhau.
```

```sql
-- Schema RBAC multi-tenant
CREATE TABLE tenant_user_roles (
    tenant_id   UUID NOT NULL,
    user_id     UUID NOT NULL,
    role        VARCHAR(50) NOT NULL,  -- 'admin', 'editor', 'viewer'
    created_at  TIMESTAMPTZ DEFAULT NOW(),

    PRIMARY KEY (tenant_id, user_id, role),
    FOREIGN KEY (tenant_id) REFERENCES tenants(id),
    FOREIGN KEY (user_id)   REFERENCES users(id)
);
```

```go
// Check permission luôn phải kèm tenant context
func (s *AuthService) HasPermission(ctx context.Context, permission string) bool {
    tenantID := TenantFromCtx(ctx).ID
    userID   := UserFromCtx(ctx).ID

    roles := s.repo.GetUserRoles(ctx, tenantID, userID)  // scoped by tenant
    return roleHasPermission(roles, permission)
}
```

---

## 4. API Design — Tenant Routing

### Cách 1: Subdomain (Phổ Biến Nhất — Recommended cho B2B)

```
acme.yourapp.com      → tenant_slug = "acme"
globex.yourapp.com    → tenant_slug = "globex"
```

Ưu điểm: Professional, isolate cookie, dễ config CORS.
Nhược điểm: Cần wildcard SSL certificate (`*.yourapp.com`).

### Cách 2: Path Prefix

```
yourapp.com/acme/api/v1/orders
yourapp.com/globex/api/v1/orders
```

Ưu điểm: Không cần wildcard SSL, dễ implement.
Nhược điểm: URL dài, không professional.

### Cách 3: Header

```http
GET /api/v1/orders
Authorization: Bearer <jwt>
X-Tenant-ID: acme-uuid
```

Ưu điểm: Linh hoạt, tốt cho mobile app.
Nhược điểm: Dễ bị giả mạo nếu không validate kỹ với JWT.

> **Recommendation:** Dùng **Subdomain** cho B2B SaaS cần branding riêng. Dùng **Header** cho mobile-first hoặc internal API. Kết hợp JWT claim + Header để double-validate.

---

## 5. Tenant Lifecycle Management

Đây là phần **hay bị bỏ quên nhất** trong thiết kế ban đầu.

### State Machine

```
PENDING
  → Vừa đăng ký, chưa verify email/payment
  → Không thể login, chỉ thấy onboarding page
      ↓
ACTIVE
  → Hoạt động bình thường
  → Full access theo plan
      ↓
SUSPENDED
  → Vi phạm ToS hoặc quá hạn thanh toán
  → Có thể login để xem billing, không được dùng feature
      ↓
DEACTIVATED
  → Chủ động tắt tài khoản
  → Data vẫn còn, có thể reactivate trong 30 ngày
      ↓
DELETED (Soft Delete)
  → Đã quá 30 ngày sau deactivate
  → Data vẫn còn theo retention policy
      ↓
PURGED (Hard Delete)
  → Xóa hoàn toàn mọi data (GDPR Right to Erasure)
  → Không thể khôi phục
```

```go
type TenantStatus string

const (
    TenantPending     TenantStatus = "pending"
    TenantActive      TenantStatus = "active"
    TenantSuspended   TenantStatus = "suspended"
    TenantDeactivated TenantStatus = "deactivated"
    TenantDeleted     TenantStatus = "deleted"
    TenantPurged      TenantStatus = "purged"
)

// Middleware check state — dù JWT còn hạn vẫn phải check
func validateTenantStatus(tenant *Tenant) error {
    switch tenant.Status {
    case TenantActive:
        return nil
    case TenantSuspended:
        return ErrTenantSuspended   // 402 Payment Required
    case TenantPending:
        return ErrTenantPending     // 403 Forbidden
    default:
        return ErrTenantUnavailable // 403 Forbidden
    }
}
```

---

## 6. Feature Flags & Plan Limits — Billing-Aware

### Cấu Hình Plan per Tenant

```go
type TenantConfig struct {
    Plan              string          // "free", "pro", "enterprise"
    MaxUsers          int             // 5 / 50 / unlimited (-1)
    MaxStorageBytes   int64           // 1GB / 10GB / unlimited
    MaxAPICallsPerDay int             // 1000 / 10000 / unlimited
    Features          map[string]bool // feature flags
    RateLimitPerMin   int             // requests/minute
    DataRetentionDays int             // 30 / 90 / 365
}

var Plans = map[string]TenantConfig{
    "free": {
        MaxUsers: 5, MaxStorageBytes: 1 << 30,
        RateLimitPerMin: 60, DataRetentionDays: 30,
        Features: map[string]bool{"export": false, "api_access": false},
    },
    "pro": {
        MaxUsers: 50, MaxStorageBytes: 10 << 30,
        RateLimitPerMin: 300, DataRetentionDays: 90,
        Features: map[string]bool{"export": true, "api_access": true},
    },
    "enterprise": {
        MaxUsers: -1, MaxStorageBytes: -1,   // unlimited
        RateLimitPerMin: 3000, DataRetentionDays: 365,
        Features: map[string]bool{"export": true, "api_access": true, "sso": true},
    },
}
```

### Enforce Plan Limits Trong Service Layer

```go
func (s *UserService) CreateUser(ctx context.Context, req CreateUserReq) (*User, error) {
    tenant := TenantFromCtx(ctx)
    config := Plans[tenant.Plan]

    // Check limit
    if config.MaxUsers != -1 {
        currentCount, _ := s.repo.CountUsers(ctx, tenant.ID)
        if currentCount >= config.MaxUsers {
            return nil, ErrPlanLimitExceeded // 402 Payment Required
        }
    }

    // Check feature flag
    if req.Role == "sso_user" && !config.Features["sso"] {
        return nil, ErrFeatureNotAvailable // 403 Forbidden
    }

    return s.repo.CreateUser(ctx, tenant.ID, req)
}
```

---

## 7. Caching — Cực Kỳ Dễ Leak Data

### Sai Lầm Phổ Biến Nhất

```go
// ❌ NGUY HIỂM: cache key không có tenant_id
// Tenant A và Tenant B có thể có cùng userID (UUID collision rất hiếm
// nhưng business logic collision thì không)
cacheKey := fmt.Sprintf("user:%s", userID)

// ✅ ĐÚNG: luôn prefix bằng tenant_id
cacheKey := fmt.Sprintf("t:%s:user:%s", tenantID, userID)

// ✅ ĐÚNG: dùng helper để tránh quên
func CacheKey(tenantID, resource, id string) string {
    return fmt.Sprintf("t:%s:%s:%s", tenantID, resource, id)
}
```

### Invalidation Khi Tenant Bị Xóa

```go
// Xóa toàn bộ cache của 1 tenant (dùng SCAN + DEL, không dùng KEYS)
func (c *Cache) InvalidateTenant(ctx context.Context, tenantID string) error {
    pattern := fmt.Sprintf("t:%s:*", tenantID)
    var cursor uint64
    for {
        keys, nextCursor, err := c.redis.Scan(ctx, cursor, pattern, 100).Result()
        if err != nil {
            return err
        }
        if len(keys) > 0 {
            c.redis.Del(ctx, keys...)
        }
        cursor = nextCursor
        if cursor == 0 {
            break
        }
    }
    return nil
}
```

### Cache Tenant Config

```go
// Cache tenant config để không phải query DB mỗi request
func (r *TenantRepo) FindByID(ctx context.Context, id string) (*Tenant, error) {
    cacheKey := fmt.Sprintf("tenant:config:%s", id)

    // Try cache first
    cached, err := r.cache.Get(ctx, cacheKey)
    if err == nil {
        var tenant Tenant
        json.Unmarshal([]byte(cached), &tenant)
        return &tenant, nil
    }

    // Fallback to DB
    tenant, err := r.db.QueryTenant(ctx, id)
    if err != nil {
        return nil, err
    }

    // Cache 5 phút (ngắn để cập nhật plan change nhanh)
    r.cache.Set(ctx, cacheKey, tenant, 5*time.Minute)
    return tenant, nil
}
```

---

## 8. Async Jobs & Queue — Tenant Context Phải Đi Theo

### Vấn Đề Thường Gặp

Khi push job vào queue, nếu không mang theo `tenant_id`, worker sẽ chạy **không có context** — mọi DB write sẽ không có tenant, hoặc tệ hơn là write vào tenant sai.

### Đúng Cách: Tenant Context Trong Payload

```go
type JobPayload struct {
    TenantID  string          `json:"tenant_id"`   // ← bắt buộc, không được thiếu
    UserID    string          `json:"user_id"`
    JobType   string          `json:"job_type"`
    Data      json.RawMessage `json:"data"`
    CreatedAt time.Time       `json:"created_at"`
}

// Producer: luôn lấy tenant từ context hiện tại
func (s *ReportService) EnqueueReport(ctx context.Context, params ReportParams) error {
    tenant := TenantFromCtx(ctx)

    payload := JobPayload{
        TenantID: tenant.ID,   // ← inject từ current context
        UserID:   UserFromCtx(ctx).ID,
        JobType:  "generate_report",
        Data:     mustMarshal(params),
    }

    return s.queue.Push(ctx, payload)
}

// Consumer: restore tenant context trước khi xử lý
func (w *Worker) Process(payload JobPayload) error {
    // Tạo fresh context với tenant từ payload
    ctx := context.WithValue(context.Background(), TenantKey, payload.TenantID)

    // Tất cả DB calls trong handler đều dùng context này
    return w.reportHandler.Generate(ctx, payload.Data)
}
```

### Cron Jobs Per Tenant

```go
// Đừng chạy 1 cron job rồi loop qua tenants — scale kém
// Thay vào đó, mỗi tenant có job riêng với priority theo plan
func ScheduleTenantJobs() {
    tenants, _ := tenantRepo.FindAllActive(ctx)
    for _, tenant := range tenants {
        queue.ScheduleAt(
            ctx,
            JobPayload{TenantID: tenant.ID, JobType: "daily_digest"},
            nextScheduleTime(tenant),
            Priority(tenant.Plan),  // enterprise jobs được ưu tiên hơn
        )
    }
}
```

---

## 9. Observability — Monitor per Tenant

### Structured Logging

```go
// Mọi log line phải có tenant_id
logger.Info("order created",
    zap.String("tenant_id", tenant.ID),
    zap.String("tenant_slug", tenant.Slug),
    zap.String("order_id", order.ID),
    zap.String("user_id", user.ID),
    zap.Float64("amount", order.Amount),
)
```

### Metrics Scoped by Tenant

```go
// Prometheus metrics với tenant label
var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{Name: "http_request_duration_seconds"},
        []string{"method", "path", "tenant_plan"},  // dùng plan thay vì tenant_id (cardinality)
    )

    tenantActiveUsers = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{Name: "tenant_active_users_total"},
        []string{"tenant_id"},
    )
)
```

> **Lưu ý cardinality:** Không dùng `tenant_id` làm label cho high-frequency metrics (request duration) vì hàng nghìn tenants sẽ làm Prometheus OOM. Dùng `tenant_plan` hoặc `tenant_tier` thay thế.

### Noisy Neighbor Detection

```go
// Alert khi 1 tenant dùng quá nhiều tài nguyên
// và ảnh hưởng đến các tenant khác
tenantResourceUsage := metrics.GetTopTenants(ctx, TopN(10), Window(5*time.Minute))
for _, usage := range tenantResourceUsage {
    if usage.CPUPercent > 20 || usage.RequestRate > tenant.RateLimit*2 {
        alerting.Send(Alert{
            Level:   "warning",
            Message: fmt.Sprintf("Noisy neighbor detected: tenant %s", usage.TenantID),
            Action:  "Consider throttling or moving to dedicated tier",
        })
    }
}
```

### Alerts Quan Trọng Cần Setup

```
[ ] Error rate tăng đột biến ở 1 tenant cụ thể
[ ] 1 tenant đang chiếm > 20% tổng CPU/memory
[ ] Tenant sắp hết quota (80%, 95%, 100%)
[ ] DB connection pool exhausted do 1 tenant
[ ] Job queue backup lớn cho specific tenant
[ ] Tenant config cache miss rate cao bất thường
```

---

## 10. Security — Checklist Không Được Bỏ Qua

### Cross-Tenant Data Leak Prevention

```go
// Mọi repository method phải enforce tenant filter
func (r *OrderRepo) FindByID(ctx context.Context, id string) (*Order, error) {
    tenantID := TenantIDFromCtx(ctx)

    // ✅ Luôn filter bằng tenant_id
    var order Order
    err := r.db.QueryRowContext(ctx,
        "SELECT * FROM orders WHERE id = $1 AND tenant_id = $2",
        id, tenantID,
    ).Scan(&order)

    if err == sql.ErrNoRows {
        return nil, ErrNotFound // Không được reveal "order exists but belongs to another tenant"
    }
    return &order, err
}
```

### GDPR Compliance

```go
// Right to Erasure — xóa TOÀN BỘ data của tenant
func (s *TenantService) PurgeTenant(ctx context.Context, tenantID string) error {
    // Transaction để đảm bảo atomic
    return s.db.Transaction(ctx, func(tx *sql.Tx) error {
        tables := []string{
            "audit_logs", "notifications", "attachments",
            "comments", "orders", "users", "tenants",  // thứ tự quan trọng (FK)
        }
        for _, table := range tables {
            _, err := tx.ExecContext(ctx,
                fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1", table),
                tenantID,
            )
            if err != nil {
                return err
            }
        }

        // Xóa cache
        s.cache.InvalidateTenant(ctx, tenantID)

        // Xóa files trên S3
        s.storage.DeleteTenantBucket(ctx, tenantID)

        // Ghi audit log vào separate immutable store
        s.auditLog.RecordPurge(ctx, tenantID, time.Now())

        return nil
    })
}
```

### Audit Log — Immutable

```go
type AuditEvent struct {
    ID         UUID      `json:"id"`
    TenantID   string    `json:"tenant_id"`
    UserID     string    `json:"user_id"`
    Action     string    `json:"action"`     // "order.created", "user.deleted"
    ResourceID string    `json:"resource_id"`
    IPAddress  string    `json:"ip_address"`
    UserAgent  string    `json:"user_agent"`
    OldValue   any       `json:"old_value,omitempty"`
    NewValue   any       `json:"new_value,omitempty"`
    OccurredAt time.Time `json:"occurred_at"`
}

// Audit log phải write-only — không cho phép update hay delete
// Lưu vào append-only storage (S3, BigQuery, hoặc DB với RLS no-delete policy)
```

### Security Checklist

```
Data:
  [ ] tenant_id có mặt trên mọi bảng
  [ ] Row Level Security enabled trên PostgreSQL
  [ ] Mọi query đều có tenant_id trong WHERE clause
  [ ] Encryption at rest — mỗi tenant có KMS key riêng (enterprise)

Auth:
  [ ] tenant_id trong JWT claim
  [ ] Tenant state check tại middleware (không chỉ token expiry)
  [ ] RBAC scoped per tenant
  [ ] Session invalidation khi tenant bị suspended

API:
  [ ] Admin API endpoint được bảo vệ riêng biệt (separate auth)
  [ ] Rate limiting per tenant
  [ ] API key rotation per tenant

Compliance:
  [ ] GDPR: API xóa toàn bộ data của tenant (Right to Erasure)
  [ ] GDPR: API export toàn bộ data của tenant (Right to Portability)
  [ ] Audit log immutable cho mọi action quan trọng
  [ ] Data residency: biết data của tenant đang nằm ở region nào

Testing:
  [ ] Penetration test: dùng token tenant A để access data tenant B
  [ ] Tenant isolation unit test trên mọi repository method
```

---

## 11. Điều Tối Quan Trọng — Tenant Isolation Test

Đây là test **quan trọng nhất** trong toàn bộ hệ thống. Phải pass 100% trước khi deploy production.

```go
func TestTenantIsolation_OrderRepository(t *testing.T) {
    // Setup: 2 tenant riêng biệt
    tenantA := fixtures.CreateTenant(t, "tenant-a")
    tenantB := fixtures.CreateTenant(t, "tenant-b")

    ctxA := fixtures.ContextWithTenant(tenantA)
    ctxB := fixtures.ContextWithTenant(tenantB)

    // Tạo order cho tenant A
    orderA, err := repo.CreateOrder(ctxA, OrderData{Amount: 100})
    require.NoError(t, err)

    // --- Test 1: Tenant B không thể lấy order của Tenant A ---
    result, err := repo.FindByID(ctxB, orderA.ID)
    assert.ErrorIs(t, err, ErrNotFound, "Tenant B should not see Tenant A's order")
    assert.Nil(t, result)

    // --- Test 2: List của Tenant B không chứa order của Tenant A ---
    orders, err := repo.ListOrders(ctxB, ListParams{})
    require.NoError(t, err)
    for _, o := range orders {
        assert.Equal(t, tenantB.ID, o.TenantID, "List should only return Tenant B's orders")
    }

    // --- Test 3: Tenant A vẫn thấy order của mình ---
    result, err = repo.FindByID(ctxA, orderA.ID)
    assert.NoError(t, err)
    assert.Equal(t, orderA.ID, result.ID)
}

// Chạy isolation test cho mọi repository
func TestTenantIsolation_All(t *testing.T) {
    t.Run("OrderRepository", TestTenantIsolation_OrderRepository)
    t.Run("UserRepository",  TestTenantIsolation_UserRepository)
    t.Run("ReportRepository", TestTenantIsolation_ReportRepository)
    // ... tất cả repositories
}
```

### Tạo Custom Linter

```go
// Tự động phát hiện query không có tenant_id filter trong code review
// Dùng go/analysis để scan AST
func checkMissingTenantFilter(pass *analysis.Pass) {
    // Flag bất kỳ sql.QueryContext nào không có "tenant_id" trong query string
    // và không có RLS annotation
}
```

---

## Tóm Tắt — Checklist Khi Bắt Đầu Dự Án

### Phase 1: Architecture (Sprint 1)

```
[ ] Chọn isolation model: Silo / Bridge / Pool / Hybrid
[ ] Chọn tenant routing: Subdomain / Path / Header
[ ] Thiết kế tenant state machine
[ ] Quyết định multi-region strategy (nếu cần data residency)
```

### Phase 2: Data & Auth (Sprint 1-2)

```
[ ] Schema: tenant_id trên mọi bảng, index đúng
[ ] Row Level Security (PostgreSQL)
[ ] Migration strategy không lock production
[ ] tenant_id trong JWT claims
[ ] Tenant middleware: extract, validate, inject context
[ ] RBAC scoped per tenant
```

### Phase 3: Application (Sprint 2-3)

```
[ ] TenantFromCtx() helper — dùng khắp nơi
[ ] Feature flags & plan limits trong service layer
[ ] Rate limiting per tenant
[ ] Cache key prefix bằng tenant_id
[ ] Async job payload mang tenant_id
[ ] Bulk cache invalidation khi tenant bị xóa
```

### Phase 4: Operations (Sprint 3-4)

```
[ ] Structured logging với tenant_id field
[ ] Metrics với tenant_plan label
[ ] Noisy neighbor detection & alerting
[ ] Tenant isolation unit tests (toàn bộ repositories)
[ ] GDPR: Right to Erasure API
[ ] GDPR: Right to Portability (data export) API
[ ] Audit log immutable
[ ] Penetration test cross-tenant access
```

---

## Quick Reference

### Helper Functions Nên Có Sẵn

```go
// Lấy tenant từ context — panic nếu không có (fail fast)
func MustTenantFromCtx(ctx context.Context) *Tenant { ... }

// Lấy tenant từ context — return nil nếu không có
func TenantFromCtx(ctx context.Context) *Tenant { ... }

// Build cache key an toàn
func TenantCacheKey(tenantID, resource, id string) string {
    return fmt.Sprintf("t:%s:%s:%s", tenantID, resource, id)
}

// Check feature flag
func (t *Tenant) HasFeature(feature string) bool {
    config := Plans[t.Plan]
    return config.Features[feature]
}

// Check plan limit
func (t *Tenant) CanAddUser(currentCount int) bool {
    config := Plans[t.Plan]
    return config.MaxUsers == -1 || currentCount < config.MaxUsers
}
```

### Error Types Chuẩn

```go
var (
    ErrTenantNotFound      = errors.New("tenant not found")
    ErrTenantSuspended     = errors.New("tenant is suspended")        // 402
    ErrTenantPending       = errors.New("tenant is pending approval") // 403
    ErrTenantUnavailable   = errors.New("tenant is unavailable")      // 503
    ErrPlanLimitExceeded   = errors.New("plan limit exceeded")        // 402
    ErrFeatureNotAvailable = errors.New("feature not available on current plan") // 403
    ErrCrossTenantAccess   = errors.New("cross-tenant access denied") // 403
)
```

---

*Tài liệu này nên được review và cập nhật mỗi khi có thay đổi lớn về architecture hoặc khi onboard team member mới.*
