# AGENTS.md — Project Constitution

> **⚠️ AI: Read this file FIRST before doing anything.**
> This is the single source of truth for how to work on this project.
> All rules here are NON-NEGOTIABLE unless explicitly overridden in task context.

---

## 1. Project Overview

**Project:** [CONVERDA]
**Domain:** [e.g., Notification Platform, E-commerce, Fintech]
**Architecture:** Domain-Driven Design (DDD) — 4 layers
**Primary Language:** Go 1.21+
**Key Tech:** GIN, PostgreSQL, Redis, sqlc, gRPC (optional)

---

## 2. Language Rules

| Context | Language |
|---|---|
| Code (variables, functions, types, packages) | **English** |
| Code comments | **English** |
| Log messages | **English** |
| User-facing error messages | **English** |
| Architecture docs, ADRs, RFCs | **English** |
| Chat responses to user | **Vietnamese** (if user writes in Vietnamese) |
| Team onboarding guides | **Vietnamese** |

---

## 3. Architecture Rules (NON-NEGOTIABLE)

### 3.1 Layer Structure & Dependencies

```
┌─────────────────────────────────────┐
│   Interface Layer                   │  controller/http/, controller/grpc/
│   (HTTP/gRPC handlers, DTOs)        │
└──────────────┬──────────────────────┘
               │ calls
┌──────────────▼──────────────────────┐
│   Application Layer                 │  application/service/
│   (Use cases, orchestration)        │  application/schedule/
└──────────────┬──────────────────────┘
               │ calls interface from
┌──────────────▼──────────────────────┐
│   Domain Layer                      │  domain/model/, domain/repository/
│   (Entities, business rules,        │  domain/errors.go
│    repository interfaces)           │
└──────────────▲──────────────────────┘
               │ implemented by
┌──────────────┴──────────────────────┐
│   Infrastructure Layer              │  infrastructure/persistence/
│   (DB, cache, external APIs)        │  infrastructure/cache/
│                                     │  infrastructure/adapter/
└─────────────────────────────────────┘
```

**Dependency Rules:**
- Domain → nothing (no imports from other layers)
- Application → Domain only (interfaces, models, errors)
- Infrastructure → Domain (implements interfaces)
- Interface → Application (calls service interfaces)
- **NEVER**: Application → Infrastructure directly
- **NEVER**: Domain → any other layer

### 3.2 Database Rules

- **ALWAYS** use sqlc-generated queries
- **NEVER** use GORM — this project does not have GORM as dependency
- **NEVER** write raw SQL strings in application/service layer
- SQL files location: `internal/[module]/infrastructure/persistence/queries/*.sql`
- Generated code location: `internal/[module]/infrastructure/persistence/sqlcgen/`
- Repository interface: `internal/[module]/domain/repository/`
- Repository implementation: `internal/[module]/infrastructure/persistence/repository/`

### 3.3 Cross-Module Communication

When module A needs data from module B, use **Interface + Adapter pattern**:

```go
// 1. Define port in CONSUMER module (NOT in provider module)
// internal/[consumer]/domain/repository/[provider]_reader.go
type I[Provider]Reader interface {
    FindByID(ctx context.Context, id string) (*[Provider]Info, error)
}

// 2. Implement local adapter in CONSUMER module
// internal/[consumer]/infrastructure/adapter/local_[provider]_adapter.go
type Local[Provider]Adapter struct { svc [provider]service.I[Provider]Service }

// 3. Wire in initializer — swap LocalAdapter → HttpAdapter for microservice
```

**NEVER** import another module's repository or entity directly.

### 3.4 Error Handling

```go
// Repository layer: add location context
return nil, fmt.Errorf("[module]Repo.[Method] id=%s: %w", id, err)

// Application layer: add operation context  
return nil, fmt.Errorf("[module]Service.[Method]: %w", err)

// Interface layer: map to HTTP status — use response.HandleError(c, err)
// NEVER: c.JSON(500, gin.H{"error": err.Error()}) — leaks internal errors
```

Rules:
- Always use `%w` (not `%v`) to preserve error chain
- Sentinel errors defined in `domain/errors.go` per module
- Service layer NEVER throws HTTP exceptions or knows about HTTP status codes
- HTTP mapping happens ONLY in response middleware/helper

### 3.5 Context Rules

```go
// ctx is ALWAYS the first parameter of any I/O function
func (s *Service) DoSomething(ctx context.Context, ...) error {}

// Background goroutines detached from request — use WithoutCancel (Go 1.21+)
go func() {
    detachedCtx := context.WithoutCancel(ctx) // keeps values, no cancel propagation
    defer func() {
        if r := recover(); r != nil {
            slog.Error("goroutine panic", "err", r, "stack", string(debug.Stack()))
        }
    }()
    // work...
}()

// NEVER store ctx in struct fields
// NEVER use context.Background() inside a function that receives ctx
```

---

## 4. Naming Conventions

### 4.1 Go Naming

| Element | Convention | Example |
|---|---|---|
| Package | lowercase, single word | `auth`, `notification`, `user` |
| Receiver | 1-3 char abbreviation | `(s *AuthService)`, `(h *UserHandler)` |
| Interface | `I` prefix | `IAuthRepository`, `IUserService` |
| Error vars | `Err` prefix | `ErrUserNotFound`, `ErrInvalidToken` |
| Constants | PascalCase or SCREAMING_SNAKE | `MaxRetries`, `StatusActive` |
| File names | snake_case with suffix | `auth_service.go`, `user_repository.go` |

### 4.2 API Naming

- Endpoints: `kebab-case` (`/api/v1/notification-templates`)
- JSON fields: `snake_case` (`user_id`, `created_at`)
- Query params: `snake_case` (`page_size`, `sort_by`)

---

## 5. Code Quality Gates

Before finalizing any code, verify ALL of these:

### Architecture
- [ ] No dependency rule violations (run: `grep -r "infrastructure" internal/*/domain/`)
- [ ] No direct cross-module repository imports
- [ ] No GORM import (`grep -r "gorm.io" internal/`)

### Go Quality
- [ ] `ctx` is first parameter in all I/O functions
- [ ] Errors wrapped with `%w` (not `%v`)
- [ ] No magic numbers — use named constants
- [ ] Goroutines have exit conditions (ctx.Done() or channel)
- [ ] Background goroutines use context.WithoutCancel()
- [ ] Goroutines have recover() for panics
- [ ] No exported fields/functions that should be unexported

### Database
- [ ] Only sqlc-generated methods used in repositories
- [ ] Multi-table writes use transaction (UnitOfWork interface)
- [ ] No SQL string concatenation

### Security
- [ ] No sensitive data in logs (password, token, card number)
- [ ] External HTTP calls have timeout
- [ ] No internal errors exposed in HTTP responses

### Testing
- [ ] New business logic has unit tests
- [ ] Tests use mock interfaces (not concrete implementations)
- [ ] Table-driven tests for multiple input cases

---

## 6. Workflow Rules

### 6.1 Before writing any code
1. Read `.ai/context/ARCHITECTURE.md` for system overview
2. Check `.ai/context/MODULES.md` for the right module to modify
3. Check `.ai/context/ADR.md` for constraints affecting your task
4. Follow patterns in `.ai/context/PATTERNS.md`

### 6.2 Task types and commands

| Task type | Command to use |
|---|---|
| Create new module | `/create` — NEVER freehand a new module |
| Fix a bug | `/debug` — systematic RCA before fixing |
| Add endpoint to existing module | Use `/create endpoint` |
| Refactor existing code | `/refactor` — explain what pattern to follow |
| Code review | `/review` — specify what to check |
| Plan complex feature | `/plan` — always plan before implementing |

### 6.3 Clarify before coding

For any non-trivial task (> 20 lines of new code):
1. Ask at least 2 clarifying questions before writing code
2. Create an implementation plan and present it to the user
3. Wait for explicit approval before implementing
4. If requirements change mid-implementation, stop and re-plan

### 6.4 Never assume

If uncertain about:
- Which module a piece of code belongs to → ask
- Whether to use existing pattern or create new → ask
- Database schema or field names → ask (don't guess)
- Business rules or edge cases → ask

---

## 7. Module Catalog

> Update this section when adding new modules

| Module | Location | Responsibility | Key Entities |
|---|---|---|---|
| auth | `internal/auth/` | Authentication, JWT, refresh tokens | Account, Token |
| user | `internal/user/` | User profile management | User |
| notification | `internal/notification/` | Multi-channel notification delivery | Notification, Template, Channel |
| [add more...] | | | |

---

## 8. References

- Architecture details: `.ai/context/ARCHITECTURE.md`
- All ADRs: `.ai/context/ADR.md`
- Code patterns: `.ai/context/PATTERNS.md`
- Golang best practices: `golang_best_practices.md`
- Workflow commands: `.agent/workflows/`
- Prompt templates: `.ai/prompts/`

---

*Last updated: 2026-02-18 | Maintainer: [Team Lead name]*
*⚠️ Every ADR approval MUST result in an update to this file.*
