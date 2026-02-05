# Inter-Module Communication Pattern (Interface + Adapter)

Tài liệu này mô tả cách implement pattern **Interface + Adapter** để module Mails có thể lấy dữ liệu từ module Settings một cách linh hoạt.

## 📊 Tổng Quan

Pattern này cho phép:
- **Loose coupling**: Module Mails không phụ thuộc trực tiếp vào Settings
- **Easy testing**: Mock interface dễ dàng cho unit tests
- **Microservice-ready**: Chỉ cần thêm HTTP/gRPC adapter khi tách service

## 🏗️ Cấu Trúc Files

```
internal/mails/
├── domain/
│   └── repository/
│       ├── template_reader.go          # Interface định nghĩa data cần
│       ├── mail_job.repository.go
│       ├── mail_tracking.repository.go
│       └── mail_send_log.repository.go
├── infrastructure/
│   └── adapter/
│       ├── local_template_adapter.go   # Adapter cho monolith (direct DB)
│       ├── http_template_adapter.go    # Adapter cho microservice (REST API)
│       └── cached_template_adapter.go  # Wrapper thêm caching layer
├── application/
│   └── service/
│       └── mail.service.impl.go        # Sử dụng TemplateReader interface
└── ...

internal/initialize/mails/
└── mail.go                             # Khởi tạo và chọn adapter phù hợp
```

## 💻 Interface Definition

```go
// internal/mails/domain/repository/template_reader.go

// TemplateReader defines the interface for reading email templates
// This interface is owned by mails module
type TemplateReader interface {
    GetByID(ctx context.Context, id int64) (*TemplateInfo, error)
    GetByCode(ctx context.Context, code string) (*TemplateInfo, error)
    List(ctx context.Context, page, pageSize int, search *string) ([]*TemplateInfo, int64, error)
    Exists(ctx context.Context, id int64) (bool, error)
}
```

## 🔧 Adapters

### 1. LocalTemplateAdapter (Monolith Mode)

Dùng khi cả hai modules chạy trong cùng một process:

```go
// internal/mails/infrastructure/adapter/local_template_adapter.go

type LocalTemplateAdapter struct {
    repo settingsRepo.EmailTemplateRepository
}

func (a *LocalTemplateAdapter) GetByID(ctx context.Context, id int64) (*TemplateInfo, error) {
    template, err := a.repo.GetById(ctx, id)
    if err != nil {
        return nil, err
    }
    return mapToTemplateInfo(template), nil
}
```

### 2. HTTPTemplateAdapter (Microservice Mode)

Dùng khi Settings đã được tách thành microservice riêng:

```go
// internal/mails/infrastructure/adapter/http_template_adapter.go

type HTTPTemplateAdapter struct {
    baseURL    string
    apiKey     string
    httpClient *http.Client
}

func (a *HTTPTemplateAdapter) GetByID(ctx context.Context, id int64) (*TemplateInfo, error) {
    url := fmt.Sprintf("%s/v1/api/settings/email-templates/%d", a.baseURL, id)
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+a.apiKey)
    
    resp, _ := a.httpClient.Do(req)
    // Parse response and return TemplateInfo
}
```

### 3. CachedTemplateAdapter (Performance Wrapper)

Wrap bất kỳ adapter nào để thêm caching:

```go
// internal/mails/infrastructure/adapter/cached_template_adapter.go

type CachedTemplateAdapter struct {
    underlying TemplateReader
    cache      *templateCache
    ttl        time.Duration
}

func (a *CachedTemplateAdapter) GetByID(ctx context.Context, id int64) (*TemplateInfo, error) {
    cacheKey := fmt.Sprintf("template:id:%d", id)
    
    // Try cache first
    if cached := a.getFromCache(cacheKey); cached != nil {
        return cached, nil
    }
    
    // Call underlying adapter
    template, err := a.underlying.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Store in cache
    a.setCache(cacheKey, template)
    return template, nil
}
```

## 🔀 Chuyển Đổi Sang Microservice

### Bước 1: Cập nhật Config

```yaml
# config/environment.yml
services:
  settings_mode: "http"  # "local" | "http" | "grpc"
  settings_url: "http://settings-service:8080"
  settings_api_key: "your-api-key"
```

### Bước 2: Sửa initTemplateReader()

```go
// internal/initialize/mails/mail.go

func initTemplateReader() mailsRepo.TemplateReader {
    switch global.Config.Services.SettingsMode {
    case "local":
        // Monolith mode
        settingsRepo := templatePersistence.NewEmailTemplateRepository(global.GormDB)
        return adapter.NewLocalTemplateAdapter(settingsRepo)
        
    case "http":
        // Microservice mode (HTTP)
        return adapter.NewHTTPTemplateAdapter(adapter.HTTPTemplateAdapterConfig{
            BaseURL: global.Config.Services.SettingsURL,
            APIKey:  global.Config.Services.SettingsAPIKey,
            Timeout: 10 * time.Second,
        })
        
    case "grpc":
        // Microservice mode (gRPC)
        conn, _ := grpc.Dial(global.Config.Services.SettingsGRPCURL)
        return adapter.NewGRPCTemplateAdapter(conn)
        
    default:
        panic("Invalid settings_mode configuration")
    }
}
```

### Bước 3: (Optional) Thêm Caching

```go
func initTemplateReader() mailsRepo.TemplateReader {
    var baseAdapter mailsRepo.TemplateReader
    
    // Choose base adapter based on config
    if global.Config.Services.SettingsMode == "http" {
        baseAdapter = adapter.NewHTTPTemplateAdapter(...)
    } else {
        baseAdapter = adapter.NewLocalTemplateAdapter(...)
    }
    
    // Wrap with caching
    return adapter.NewCachedTemplateAdapter(baseAdapter, adapter.CachedTemplateAdapterConfig{
        TTL:             5 * time.Minute,
        CleanupInterval: 10 * time.Minute,
    })
}
```

## ✅ Những Thành Phần Đã Được Cập Nhật

| File | Thay Đổi |
|------|----------|
| `mails/domain/repository/template_reader.go` | Tạo mới - Interface definition |
| `mails/infrastructure/adapter/local_template_adapter.go` | Tạo mới - Local adapter |
| `mails/infrastructure/adapter/http_template_adapter.go` | Tạo mới - HTTP adapter |
| `mails/infrastructure/adapter/cached_template_adapter.go` | Tạo mới - Cached wrapper |
| `mails/application/service/mail.service.impl.go` | Cập nhật - Dùng TemplateReader |
| `mails/application/service/dto/mail.dto.go` | Cập nhật - Dùng TemplateInfo |
| `mails/infrastructure/scheduler/mail_scheduler.go` | Cập nhật - Dùng TemplateReader |
| `initialize/mails/mail.go` | Cập nhật - Khởi tạo adapter |

## 🎯 Lợi Ích

1. **Zero code change trong business logic** khi tách microservice
2. **Gradual migration** - Có thể test HTTP adapter trước khi deploy
3. **Easy rollback** - Chỉ cần đổi config về "local"
4. **Better testability** - Mock TemplateReader cho unit tests
5. **Performance optimization** - Thêm caching layer dễ dàng

## 📝 Notes

- Adapter pattern cho phép di chuyển phần lớn complexity vào infra layer
- Business logic trong service và domain layer không cần biết data đến từ đâu
- Khi Settings API thay đổi, chỉ cần sửa adapter, không cần sửa service

---

**Tạo bởi:** Antigravity AI  
**Ngày:** 2026-02-02
