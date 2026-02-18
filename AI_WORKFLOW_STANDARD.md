# AI Workflow Standard — Hướng dẫn làm việc với AI cho toàn team

> **Phiên bản:** v1.0 — 18/02/2026
> **Áp dụng cho:** Mọi dự án backend Go/NestJS sử dụng AI Coding Assistant (Claude, Cursor, Windsurf, OpenCode)
> **Ngôn ngữ:** Chat với AI bằng **Tiếng Việt** · Code comments & docs bằng **Tiếng Anh**

---

## Mục lục

1. [Đánh giá hiện trạng](#1-đánh-giá-hiện-trạng-novu-clone)
2. [Tổng quan kiến trúc AI Workflow](#2-tổng-quan-kiến-trúc-ai-workflow)
3. [Cấu trúc thư mục chuẩn](#3-cấu-trúc-thư-mục-chuẩn)
4. [AGENTS.md — Hiến pháp của AI](#4-agentsmd--hiến-pháp-của-ai)
5. [Folder `.agent/` — Thư viện AI](#5-folder-agent--thư-viện-ai)
6. [Folder `.ai/` — Context dự án](#6-folder-ai--context-dự-án)
7. [Quy trình làm việc hàng ngày](#7-quy-trình-làm-việc-hàng-ngày)
8. [Slash Commands chuẩn](#8-slash-commands-chuẩn)
9. [Onboarding cho thành viên mới](#9-onboarding-cho-thành-viên-mới)
10. [Anti-patterns cần tránh](#10-anti-patterns-cần-tránh)

---

## 1. Đánh giá hiện trạng (novu-clone)

### ✅ Điểm mạnh hiện tại

| Hạng mục | Trạng thái | Nhận xét |
|---|---|---|
| Antigravity Kit (`.agent/`) | ✅ Có | Framework tốt, 20 agents + 37 skills + 11 workflows |
| Best practices docs | ✅ Có | `golang_best_practices.md` tại root |
| Architecture docs | ✅ Có | `ARCHITECTURE_VI.md` |
| Project flows | ✅ Có | `project_flows.md` |
| Functions catalog | ✅ Có | `functions.md` |

### ❌ Vấn đề cần cải thiện

**Vấn đề 1 — Không có AGENTS.md (Hiến pháp)**
- Antigravity Kit có sẵn workflows nhưng thiếu file `AGENTS.md` ở root định nghĩa **rules cụ thể cho dự án này**
- AI không biết: dự án dùng DDD, sqlc (không phải GORM), Go 1.21+, convention naming của team

**Vấn đề 2 — Context documents không được cấu trúc để AI đọc**
- `golang_best_practices.md` tốt nhưng đặt ở root, AI không biết khi nào nên đọc file này
- `ARCHITECTURE_VI.md` viết bằng Tiếng Việt — AI xử lý kém hơn khi generate code từ Tiếng Việt
- `functions.md` và `project_flows.md` chưa có format chuẩn để AI tham chiếu

**Vấn đề 3 — Folder `.ai/` thiếu cấu trúc**
- Cursor rules (`.ai/`) thường chỉ là 1 file dài, không modular
- Không có phân tách: rules chung vs rules riêng cho từng loại file (handler, service, repository)

**Vấn đề 4 — Không có onboarding guide cho team**
- Team member mới không biết: dùng tool gì, gõ lệnh gì, khi nào dùng `/plan` vs gõ thẳng
- Không có ví dụ prompt mẫu cho các task phổ biến (tạo module mới, debug, review)

**Vấn đề 5 — Không có quy tắc kiểm tra output của AI**
- AI có thể tạo code đúng pattern nhưng dùng GORM thay vì sqlc → vi phạm ADR-003
- Không có checklist để developer verify trước khi commit

---

## 2. Tổng quan kiến trúc AI Workflow

```
┌─────────────────────────────────────────────────────────────┐
│                    AI WORKFLOW SYSTEM                        │
│                                                             │
│  AGENTS.md ────────────────────────────────────────────────►│  "Hiến pháp" — Rules & conventions bắt buộc
│                                                             │
│  .agent/                                                    │
│  ├── workflows/   ──────────────────────────────────────────►│  Slash commands: /plan, /debug, /create
│  ├── agents/      ──────────────────────────────────────────►│  Specialist personas: backend, security, qa
│  └── skills/      ──────────────────────────────────────────►│  Domain knowledge modules
│                                                             │
│  .ai/                                                       │
│  ├── rules/       ──────────────────────────────────────────►│  Cursor/Windsurf rules per file type
│  ├── context/     ──────────────────────────────────────────►│  Project architecture, ADRs, patterns
│  └── prompts/     ──────────────────────────────────────────►│  Reusable prompt templates
│                                                             │
└─────────────────────────────────────────────────────────────┘

Resolution Priority (cao → thấp):
AGENTS.md > .ai/rules/ > .agent/workflows/ > .agent/skills/
```

**Nguyên tắc cốt lõi:**
- AI đọc context **một lần**, apply **mọi lúc** — không cần nhắc lại mỗi chat
- Developer **mô tả ý định**, AI **thực thi theo pattern đã định** trong docs
- Mọi output của AI **phải qua checklist** trước khi commit

---

## 3. Cấu trúc thư mục chuẩn

```
project-root/
│
├── AGENTS.md                    # ⭐ Hiến pháp — AI đọc đầu tiên
│
├── .agent/                      # Antigravity Kit (không commit nếu dùng ag-kit)
│   ├── workflows/               # Slash command definitions
│   │   ├── plan.md              # /plan — Lập kế hoạch task
│   │   ├── create.md            # /create — Tạo module/feature mới
│   │   ├── debug.md             # /debug — Debug có hệ thống
│   │   ├── review.md            # /review — Code review
│   │   └── refactor.md          # /refactor — Refactor theo best practices
│   ├── agents/                  # Specialist personas
│   │   ├── backend-go.md        # Go backend specialist (DDD, sqlc)
│   │   ├── security.md          # Security auditor
│   │   └── qa.md                # QA & testing specialist
│   └── skills/                  # Knowledge modules
│       ├── ddd-golang.md        # DDD patterns in Go
│       ├── sqlc-patterns.md     # sqlc query patterns
│       └── error-handling.md    # Error handling patterns
│
├── .ai/                         # Project-specific AI context
│   ├── rules/                   # Cursor/Windsurf per-file rules
│   │   ├── handler.mdc          # Rules cho controller/handler files
│   │   ├── service.mdc          # Rules cho application service files
│   │   ├── repository.mdc       # Rules cho repository/persistence files
│   │   └── dto.mdc              # Rules cho DTO/request/response files
│   ├── context/                 # Project context cho AI
│   │   ├── ARCHITECTURE.md      # System architecture (English)
│   │   ├── ADR.md               # All Architecture Decision Records
│   │   ├── PATTERNS.md          # Code patterns & examples
│   │   └── MODULES.md           # Module catalog & responsibilities
│   └── prompts/                 # Reusable prompt templates
│       ├── new-module.md        # Prompt tạo module mới
│       ├── add-endpoint.md      # Prompt thêm API endpoint
│       └── write-test.md        # Prompt viết unit test
│
├── docs/                        # Team documentation (human-readable)
│   ├── adr/                     # Architecture Decision Records
│   ├── rfc/                     # Request for Comments
│   └── onboarding/              # Onboarding guides
│       └── AI_GUIDE.md          # ← Hướng dẫn này
│
└── [project source code...]
```

---

## 4. AGENTS.md — Hiến pháp của AI

File `AGENTS.md` ở root là file **AI đọc đầu tiên** khi khởi động session. Đây là "constitution" — định nghĩa tất cả quy tắc bắt buộc.

### Template AGENTS.md cho Go DDD Project

```markdown
# AGENTS.md — Project Constitution

## 1. Identity & Role
You are a senior Go backend engineer with deep expertise in:
- Domain-Driven Design (DDD) with 4-layer architecture
- Go 1.21+ (use context.WithoutCancel, slog, etc.)
- sqlc for database queries (NOT gorm, NOT raw pgx directly in services)
- GIN web framework
- PostgreSQL, Redis

## 2. Language Rules
- **Code & Comments**: English only
- **Chat responses**: Vietnamese (reply in Vietnamese when asked in Vietnamese)
- **Error messages (user-facing)**: English
- **Log messages**: English

## 3. Architecture Rules (NON-NEGOTIABLE)

### 3.1 Layer Dependencies
```
Interface Layer (controller/http/)
    ↓ calls
Application Layer (application/service/)
    ↓ calls interface from
Domain Layer (domain/repository/, domain/model/)
    ↑ implemented by
Infrastructure Layer (infrastructure/persistence/)
```

**Rules:**
- Domain layer MUST NOT import any other layer
- Application layer MUST NOT import infrastructure directly
- Infrastructure implements interfaces defined in Domain
- Cross-module communication MUST use Interface + Adapter pattern (see .ai/context/PATTERNS.md)

### 3.2 Database Rules
- ALWAYS use sqlc-generated queries — NEVER write raw SQL in service/application code
- NEVER use GORM — this project does not use GORM
- Queries live in: `internal/[module]/infrastructure/persistence/queries/`
- Repository interface defined in: `internal/[module]/domain/repository/`

### 3.3 Error Handling Rules
- Services return `error` — NEVER return HTTP status codes from service layer
- Use `fmt.Errorf("context: %w", err)` for wrapping
- Sentinel errors defined in: `internal/[module]/domain/errors.go`
- HTTP mapping happens ONLY in handler layer via global error middleware

### 3.4 Context Rules
- ctx is ALWAYS the first parameter of any I/O function
- Background goroutines MUST use `context.WithoutCancel(ctx)` (Go 1.21+)
- NEVER store ctx in struct fields

## 4. Naming Conventions
- Package names: single lowercase word (no underscores)
- Receiver names: 1-3 char abbreviation, consistent per struct (never `this`/`self`)
- Interface names: `I` prefix for repository interfaces (e.g., `IAuthRepository`)
- File names: snake_case (e.g., `auth_service.go`, `auth_repository.go`)

## 5. Workflow Rules

### Before writing any code:
1. Read `.ai/context/ARCHITECTURE.md` to understand the system
2. Check `.ai/context/MODULES.md` to find the right module
3. Check `.ai/context/ADR.md` for decisions that affect your task
4. Follow the pattern in `.ai/context/PATTERNS.md`

### When creating a new module:
Use the `/create` workflow command — do NOT freehand module creation

### When fixing a bug:
Use the `/debug` workflow command for systematic RCA

### When unsure about approach:
Ask for clarification BEFORE writing code — never assume

## 6. Code Quality Gates
Before finalizing any code, verify:
- [ ] No GORM imports
- [ ] ctx is first parameter in all I/O functions
- [ ] Errors are wrapped with %w
- [ ] No magic numbers/strings (use constants)
- [ ] No exported fields that should be unexported
- [ ] Unit tests exist for new business logic
- [ ] No goroutine without ctx.Done() or channel blocking
```

---

## 5. Folder `.agent/` — Thư viện AI

### 5.1 Workflow: `/create` — Tạo module DDD mới

File: `.agent/workflows/create.md`

```markdown
# Workflow: /create

## Trigger
User invokes: `/create [module-name] [description]`
Example: `/create notification "send email, SMS, and push notifications"`

## Steps

### Step 1 — Clarify (always ask before coding)
Ask the user:
1. What are the main entities in this module?
2. What external services does it need? (SMTP, Twilio, FCM?)
3. Does it need to communicate with other modules? Which ones?
4. Any special performance requirements?

### Step 2 — Plan
Create an implementation_plan.md with:
- Module structure (4 layers)
- Interfaces to define
- Dependencies to inject
- Database tables/queries needed
- API endpoints

Present the plan and wait for approval before writing code.

### Step 3 — Scaffold in order
Create files in this exact order:
1. `domain/model/entity/` — entities
2. `domain/repository/` — repository interfaces  
3. `domain/errors.go` — domain errors
4. `infrastructure/persistence/queries/` — sqlc SQL files
5. `infrastructure/persistence/repository/` — repository implementations
6. `application/service/` — service interface + implementation
7. `controller/dto/` — request/response DTOs
8. `controller/http/` — HTTP handler
9. `[module].module.go` — wire everything together

### Step 4 — Verify
Run the checklist:
- [ ] No cross-layer violations
- [ ] All interfaces properly injected
- [ ] ctx propagated through all I/O functions
- [ ] Errors wrapped with context
- [ ] README updated in module folder
```

### 5.2 Workflow: `/debug` — Debug có hệ thống

File: `.agent/workflows/debug.md`

```markdown
# Workflow: /debug

## Trigger
User invokes: `/debug [problem description]`

## Systematic RCA Process

### Step 1 — Gather Information
Ask for:
- Exact error message / stack trace
- Steps to reproduce
- Expected vs actual behavior
- Recent changes (last commit / PR)
- Environment (local/staging/production)

### Step 2 — Hypothesize (Top 3)
List top 3 most likely root causes based on information.
Rank by probability. Do NOT jump to conclusions.

### Step 3 — Investigate
For each hypothesis, suggest a specific diagnostic step:
- Log point to add
- Unit test to write
- DB query to run
- Code section to inspect

### Step 4 — Fix
Once root cause confirmed:
1. Explain WHY this caused the bug
2. Propose fix
3. Explain how to prevent recurrence
4. Suggest test case to add

### Step 5 — Document
If the bug reveals a missing best practice → suggest adding to golang_best_practices.md
```

### 5.3 Agent: Go Backend Specialist

File: `.agent/agents/backend-go.md`

```markdown
# Agent: @backend-go

## Persona
Senior Go engineer, 5+ years DDD, obsessed with correctness over cleverness.
Believes: "If it's not tested, it's broken."

## Auto-trigger keywords
- "tạo module", "add endpoint", "create service"
- "repository", "handler", "middleware"  
- "goroutine", "channel", "concurrency"
- "sqlc", "query", "migration"

## Mindset
- Always ask: "Which layer does this belong to?"
- Always ask: "Is this testable without infrastructure?"
- Always ask: "Will this goroutine leak?"
- Prefers boring, explicit code over clever abstractions

## Output format
- Show complete file, not snippets
- Include package declaration
- Include all imports
- Add TODO comments where follow-up is needed
```

---

## 6. Folder `.ai/` — Context dự án

### 6.1 Cursor Rules per file type

File: `.ai/rules/handler.mdc`

```markdown
---
description: Rules for HTTP handler files (controller/http/*.go)
globs: ["internal/*/controller/http/*.go"]
---

# Handler Rules

You are writing an HTTP handler in a Go DDD project.

## Responsibilities (what handler SHOULD do)
- Parse and validate HTTP request → DTO
- Call application service
- Map result to HTTP response
- Map errors to HTTP status codes

## What handler MUST NOT do
- Contain business logic
- Directly call repository
- Access database
- Know about domain entities (use DTOs only)

## Pattern to follow
```go
func (h *[Name]Handler) [Action](c *gin.Context) {
    var req dto.[Action]Req
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, response.Error(err))
        return
    }
    
    result, err := h.[service].[Action](c.Request.Context(), req)
    if err != nil {
        response.HandleError(c, err) // Maps domain errors to HTTP status
        return
    }
    
    c.JSON(http.StatusOK, response.Success(result))
}
```

## Never use
- `c.JSON(500, ...)` with hardcoded status code
- `gin.H{"error": err.Error()}` — leaks internal errors
- Direct database calls
```

File: `.ai/rules/service.mdc`

```markdown
---
description: Rules for application service files
globs: ["internal/*/application/service/*.go"]
---

# Application Service Rules

## Responsibilities
- Orchestrate domain logic
- Call repository interfaces (NEVER implementations)
- Handle transaction boundaries (via UnitOfWork interface)
- Return domain objects or errors — never HTTP concerns

## NEVER import
- `net/http` package
- `github.com/gin-gonic/gin`
- Any infrastructure package directly
- `gorm.io/gorm`

## Pattern to follow
```go
type [Name]ServiceImpl struct {
    repo    domain.[I][Name]Repository  // interface, not concrete
    cache   cache.Store                  // interface
    // other injected dependencies
}

func (s *[Name]ServiceImpl) [Action](ctx context.Context, req *dto.[Action]Req) (*domain.[Entity], error) {
    // 1. Validate business rules
    // 2. Load existing state if needed
    // 3. Execute domain logic
    // 4. Persist changes
    // 5. Return result or wrapped error
}
```
```

File: `.ai/rules/repository.mdc`

```markdown
---
description: Rules for repository implementation files
globs: ["internal/*/infrastructure/persistence/repository/*.go"]
---

# Repository Implementation Rules

## This file implements a domain repository interface using sqlc

## Pattern
```go
// ALWAYS embed the sqlc Queries struct
type [Name]Repository struct {
    q *sqlcgen.Queries  // generated by sqlc
}

// ALWAYS use sqlc-generated methods
// NEVER write raw SQL strings here
func (r *[Name]Repository) FindByID(ctx context.Context, id string) (*domain.[Entity], error) {
    row, err := r.q.Get[Name]ByID(ctx, id)  // sqlc generated
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, domain.Err[Name]NotFound
        }
        return nil, fmt.Errorf("[name]repo.FindByID: %w", err)
    }
    return mapRowToEntity(row), nil
}

// mapRowToEntity converts sqlc row to domain entity (private function)
func mapRowToEntity(row sqlcgen.Get[Name]ByIDRow) *domain.[Entity] {
    return &domain.[Entity]{
        ID:    row.ID.String(),
        // ...
    }
}
```
```

### 6.2 Context: PATTERNS.md

File: `.ai/context/PATTERNS.md`

```markdown
# Code Patterns Reference

This file contains approved patterns for common scenarios.
AI MUST follow these patterns exactly — do not improvise.

## Pattern 1: Cross-module Communication

When module A needs data from module B:

### Step 1: Define Port (in consumer module)
```go
// internal/[consumer]/domain/repository/[provider]_reader.go
type [Provider]Info struct {
    ID    string
    Name  string
    // Only fields this module actually needs
}

type I[Provider]Reader interface {
    FindByID(ctx context.Context, id string) (*[Provider]Info, error)
}
```

### Step 2: Implement Local Adapter
```go
// internal/[consumer]/infrastructure/adapter/local_[provider]_adapter.go
type Local[Provider]Adapter struct {
    svc [provider]service.I[Provider]Service
}

func (a *Local[Provider]Adapter) FindByID(ctx context.Context, id string) (*repository.[Provider]Info, error) {
    entity, err := a.svc.FindByID(ctx, id)
    if err != nil { return nil, fmt.Errorf("adapter: %w", err) }
    return &repository.[Provider]Info{ID: entity.ID, Name: entity.Name}, nil
}
```

### Step 3: Wire in Module Init
```go
// Monolith: inject LocalAdapter
container.Provide(func(svc *userService) repository.IUserReader {
    return adapter.NewLocalUserAdapter(svc)
})
// Microservice: swap to HttpAdapter — zero changes to service layer
```

## Pattern 2: Background Worker (Goroutine)

```go
// ALWAYS: use detached context + recover + done channel
func (s *NotificationService) StartWorker(ctx context.Context) {
    go func() {
        detachedCtx := context.WithoutCancel(ctx) // Keep values, no cancel propagation
        defer func() {
            if r := recover(); r != nil {
                slog.Error("panic in notification worker", "err", r, "stack", string(debug.Stack()))
            }
        }()
        
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():    // Original ctx for shutdown signal
                return
            case <-ticker.C:
                s.processQueue(detachedCtx)
            }
        }
    }()
}
```

## Pattern 3: Error Wrapping Chain

```go
// Repository layer — add context
return nil, fmt.Errorf("authRepo.FindByEmail [%s]: %w", email, err)

// Application layer — add operation context
return nil, fmt.Errorf("authService.Login: %w", err)

// Handler layer — map to HTTP (DON'T add more context here)
// Use response.HandleError(c, err) which reads error chain
```

## Pattern 4: Unit Test Structure

```go
func Test[Service]_[Method](t *testing.T) {
    tests := []struct {
        name    string
        setup   func(*mocks.I[Repo])  // inject mock behaviors
        input   [InputType]
        wantErr bool
        wantVal [OutputType]
    }{
        {
            name: "success case",
            setup: func(m *mocks.I[Repo]) {
                m.On("FindByID", mock.Anything, "123").Return(&domain.Entity{}, nil)
            },
            input:   "[InputType]{ID: "123"}",
            wantErr: false,
        },
        {
            name: "not found",
            setup: func(m *mocks.I[Repo]) {
                m.On("FindByID", mock.Anything, "999").Return(nil, domain.ErrNotFound)
            },
            input:   "[InputType]{ID: "999"}",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := mocks.NewI[Repo](t)
            tt.setup(mockRepo)
            svc := New[Service](mockRepo)
            
            result, err := svc.[Method](context.Background(), tt.input)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.wantVal, result)
        })
    }
}
```
```

### 6.3 Prompts library: Tạo module mới

File: `.ai/prompts/new-module.md`

```markdown
# Prompt Template: Tạo module mới

## Cách dùng
Copy đoạn dưới, điền vào [BRACKETS], paste vào chat với AI.

---

/create

**Module name:** [tên module, ví dụ: notification]
**Description:** [mô tả chức năng, 1-2 câu]

**Entities:**
- [Entity 1]: [thuộc tính chính]
- [Entity 2]: [thuộc tính chính]

**Operations (use cases):**
- [Operation 1, ví dụ: Send notification to user]
- [Operation 2]

**External dependencies:**
- [ví dụ: SMTP server, Firebase, Twilio, hoặc None]

**Cross-module dependencies:**
- Cần dữ liệu từ module [tên module] để [mục đích]
- [Hoặc: None]

**API endpoints:**
- [METHOD] /api/v1/[path] — [mô tả]

**Non-functional requirements:**
- [ví dụ: high throughput, idempotent, retry on failure, hoặc None]

---

**Lưu ý cho AI:** 
Đọc `.ai/context/ARCHITECTURE.md` và `.ai/context/PATTERNS.md` trước khi tạo code.
Hỏi clarifying questions nếu có gì chưa rõ.
Tạo implementation_plan.md trước, chờ approve rồi mới code.
```

---

## 7. Quy trình làm việc hàng ngày

### 7.1 Developer bắt đầu một task mới

```
1. Đọc task (Jira/GitHub Issue)
         ↓
2. Xác định loại task:
   - Tạo feature mới → dùng /create prompt
   - Fix bug → dùng /debug prompt  
   - Refactor → dùng /refactor prompt
   - Review → dùng /review prompt
         ↓
3. Copy prompt template từ .ai/prompts/
   Điền thông tin cụ thể của task
         ↓
4. Paste vào AI, chờ clarifying questions
         ↓
5. AI tạo implementation_plan.md
   → Developer review và approve plan
         ↓
6. AI implement theo plan
         ↓
7. Developer chạy AI Checklist (Mục 7.3)
         ↓
8. Commit & tạo PR
```

### 7.2 Quy tắc ngôn ngữ

| Ngữ cảnh | Ngôn ngữ |
|---|---|
| Chat với AI | Tiếng Việt ✅ |
| Code (variables, functions, types) | English ✅ |
| Code comments | English ✅ |
| Commit messages | English ✅ |
| PR description | English (summary) + Vietnamese (context) ✅ |
| Log messages | English ✅ |
| User-facing error messages | English ✅ |
| Internal documentation (ADR, RFC) | English ✅ |
| Team documentation (onboarding, guides) | Vietnamese ✅ |

### 7.3 AI Output Checklist — Chạy trước khi commit

```bash
# Checklist này dán vào cuối mỗi AI session khi review code

AI OUTPUT CHECKLIST — [Module/Feature name]
Date: [YYYY-MM-DD]
Reviewer: [your name]

## Architecture
- [ ] Không vi phạm dependency rule (Domain không import Infrastructure)
- [ ] Cross-module communication dùng Interface + Adapter pattern
- [ ] Không import GORM (`gorm.io/gorm` không được phép)

## Go Best Practices
- [ ] ctx là tham số đầu tiên của mọi I/O function
- [ ] Goroutine dùng context.WithoutCancel() nếu detached từ request
- [ ] Goroutine có recover() panic
- [ ] Không có goroutine leak (có exit condition)
- [ ] Errors được wrap với %w (không dùng %v)
- [ ] Không có magic numbers/strings (dùng constants)

## Database
- [ ] Chỉ dùng sqlc-generated queries
- [ ] Không có raw SQL string trong service/application layer
- [ ] Transactions qua UnitOfWork interface (nếu multi-table write)

## Security
- [ ] Không log password, token, PII
- [ ] Không lộ internal error ra HTTP response
- [ ] External HTTP calls có timeout
- [ ] Parameterized queries (không concat string vào SQL)

## Testing
- [ ] Unit tests cho business logic mới
- [ ] Mock interfaces (không mock concrete types)
- [ ] Table-driven tests cho các case input/output

## General
- [ ] Không có TODO/FIXME mà không có tracking issue
- [ ] Không có commented-out code
- [ ] Export chỉ những gì cần thiết
```

---

## 8. Slash Commands chuẩn

Đây là các lệnh hay dùng nhất. Gõ trong chat với AI:

### Lệnh tạo mới

```
/create module notification "Gửi thông báo qua email, SMS, push"
/create endpoint "POST /api/v1/notifications" "Gửi notification theo channel"
/create repository "INotificationRepository" với methods FindByID, Save, ListByUserID
```

### Lệnh debug

```
/debug "panic: nil pointer dereference ở notification_service.go:45"
/debug "query chạy chậm 3s khi list notifications của user có 10k records"
/debug "goroutine leak: số goroutine tăng liên tục không giảm"
```

### Lệnh review

```
/review .agent/workflows/create.md — xem code tôi vừa tạo đúng pattern chưa
/review "check cross-module dependency violations trong module order"
/review "security audit notification_handler.go"
```

### Lệnh refactor

```
/refactor "notification_service.go vi phạm SRP, tách logic delivery ra"
/refactor "convert raw SQL queries trong auth_repository sang sqlc"
```

### Lệnh plan

```
/plan "implement retry mechanism cho notification delivery với exponential backoff"
/plan "migrate từ single binary sang microservice cho notification module"
```

---

## 9. Onboarding cho thành viên mới

### Ngày 1: Setup

```bash
# 1. Clone repo
git clone [repo-url]
cd [project]

# 2. Cài Antigravity Kit (nếu chưa có .agent folder)
npx @vudovn/ag-kit init

# 3. Đọc AGENTS.md (bắt buộc, ~10 phút)
cat AGENTS.md

# 4. Đọc Architecture overview (~15 phút)
cat .ai/context/ARCHITECTURE.md

# 5. Đọc Patterns reference (~15 phút)
cat .ai/context/PATTERNS.md

# 6. Setup AI tool (chọn 1):
# - Cursor: settings đã có trong .ai/rules/ (tự động load)
# - Claude Code: AGENTS.md tự động được đọc
# - Windsurf: copy rules từ .ai/rules/ vào Windsurf settings
```

### Ngày 2-3: Task đầu tiên với AI

Làm một task nhỏ (thêm endpoint đơn giản) và follow full quy trình:
1. Dùng prompt template từ `.ai/prompts/add-endpoint.md`
2. Chờ AI hỏi clarifying questions (nếu không hỏi → nhắc AI đọc AGENTS.md)
3. Review implementation_plan.md trước khi AI code
4. Chạy checklist ở Mục 7.3 sau khi nhận code

### FAQ cho thành viên mới

**Q: AI tạo code dùng GORM, tôi phải làm gì?**
A: Dừng lại, nói với AI: "Project này không dùng GORM, dùng sqlc. Đọc lại .ai/context/ARCHITECTURE.md và tạo lại theo pattern."

**Q: AI không hỏi clarifying questions, code ngay?**
A: Nói: "Khoan, tôi muốn bạn hỏi clarifying questions trước. Đọc .ai/rules/ và AGENTS.md rồi follow quy trình."

**Q: AI tạo code không theo DDD layer?**
A: Nói: "Check dependency rules trong AGENTS.md. [Tên layer] không được import [tên layer kia]."

**Q: Tôi dùng Cursor hay Claude Code?**
A: Cả hai đều được. Cursor tốt hơn cho inline suggestions. Claude Code tốt hơn cho tasks lớn + planning.

---

## 10. Anti-patterns cần tránh

### ❌ Anti-pattern 1: Paste code thẳng không qua checklist

```
# SAI
User: "Tạo giúp tôi notification module"
AI: [tạo code]
User: [paste thẳng vào codebase, commit]

# ĐÚNG
User: "/create module notification [description]"
AI: [hỏi clarifying questions] → [tạo plan] → [chờ approve] → [tạo code]
User: [chạy checklist] → [review] → [commit]
```

### ❌ Anti-pattern 2: Nhờ AI fix bug không cung cấp context

```
# SAI
User: "Fix lỗi này" [paste stack trace]

# ĐÚNG
User: "/debug
Error: [stack trace]
File: [filepath:line]
Context: [mô tả ngắn đang làm gì]
Recent changes: [commit gần nhất]"
```

### ❌ Anti-pattern 3: Dùng AI như Google Search

```
# SAI (lãng phí, AI trả lời generic)
User: "goroutine leak là gì"

# ĐÚNG (AI trả lời trong context dự án)
User: "Trong service này [paste code], có goroutine leak không? 
Nếu có, fix theo pattern trong .ai/context/PATTERNS.md"
```

### ❌ Anti-pattern 4: Không đọc AGENTS.md khi onboard

Nhiều team member bỏ qua AGENTS.md và bắt đầu coding ngay. Kết quả: AI tạo code không theo convention → technical debt.

**Rule:** Mỗi session mới với AI trên project này, nói: *"Đọc AGENTS.md trước khi làm bất cứ điều gì."*

### ❌ Anti-pattern 5: AGENTS.md không được cập nhật

AGENTS.md là living document. Mỗi khi team ra quyết định kiến trúc mới (ADR), PHẢI cập nhật AGENTS.md.

**Rule:** Mọi ADR được approve → update AGENTS.md trong cùng PR.

---

## Changelog

| Version | Date | Changes |
|---|---|---|
| v1.0 | 2026-02-18 | Initial document — based on novu-clone review |
