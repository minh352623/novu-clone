# AGENTS.md — Project Constitution

> **⚠️ AI: Read this file FIRST before doing ANYTHING.**
> This is the single source of truth for how to work on this project.
> Rules here are NON-NEGOTIABLE unless explicitly overridden in the task prompt.

---

## 1. Project Overview

**Project:** novu-clone — Multi-channel Notification Platform
**Domain:** Notification delivery (Email, SMS, Push, In-App, Webhook)
**Architecture:** Domain-Driven Design (DDD) — 4 strict layers
**Language:** Go 1.21+
**Tech Stack:** GIN · PostgreSQL · Redis · GORM · gRPC · Docker
**Status:** Monolith with microservice migration path

---

## 2. Language Rules

| Context | Language |
|---|---|
| Code — variables, functions, types, packages | **English only** |
| Code comments | **English only** |
| Log messages | **English only** |
| User-facing error messages | **English only** |
| Architecture docs, ADRs, RFCs | **English only** |
| Chat responses | **Vietnamese** (match user's language) |
| Team guides, onboarding | **Vietnamese** |

---

## 3. Architecture — NON-NEGOTIABLE Rules

### 3.1 Layer Dependency Graph

```
[Interface Layer]          controller/http/     controller/grpc/
       │ calls ↓
[Application Layer]        application/service/ application/schedule/
       │ calls interface ↓
[Domain Layer]             domain/model/        domain/repository/       domain/errors.go
       ↑ implemented by
[Infrastructure Layer]     infrastructure/persistence/  infrastructure/cache/  infrastructure/adapter/
```

**Hard Rules:**
- Domain imports → **nothing** (zero external dependencies)
- Application imports → Domain only (interfaces + models)
- Infrastructure imports → Domain (to implement interfaces)
- Interface imports → Application only (calls service interfaces)
- **❌ FORBIDDEN:** Application → Infrastructure (direct DB/cache access in service)
- **❌ FORBIDDEN:** Domain → any other layer
- **❌ FORBIDDEN:** Direct cross-module repo/entity import (use Adapter pattern)

### 3.2 Database — GORM Only

```
✅ ALWAYS:  r.q.GetUserByID(ctx, id)           // GORM-generated method
❌ NEVER:   db.Where("id = ?", id).Find(&user) // sqlc — not in this project
❌ NEVER:   db.QueryRow("SELECT ...", id)       // raw SQL in service layer
```

- SQL definitions: `internal/[module]/infrastructure/persistence/queries/*.sql`
- Generated Go code: `internal/[module]/infrastructure/persistence/GORMgen/`
- Repository interface: `internal/[module]/domain/repository/i_[name]_repository.go`
- Repository impl: `internal/[module]/infrastructure/persistence/repository/[name]_repository.go`

### 3.3 Cross-Module Communication (Interface + Adapter)

```go
// STEP 1 — Define port in CONSUMER module's domain layer
// internal/notification/domain/repository/i_user_reader.go
type UserInfo struct { ID, Name, Email string }
type IUserReader interface {
    FindByID(ctx context.Context, id string) (*UserInfo, error)
}

// STEP 2 — Implement LocalAdapter in CONSUMER module's infra layer
// internal/notification/infrastructure/adapter/local_user_adapter.go
type LocalUserAdapter struct { svc userservice.IUserService }
func (a *LocalUserAdapter) FindByID(ctx context.Context, id string) (*repository.UserInfo, error) { ... }

// STEP 3 — Wire in initializer (swap to HttpAdapter → zero service changes)
```

**❌ NEVER** import `internal/user/infrastructure/...` from another module.

### 3.4 Error Handling Convention

```go
// Repository → add location context
return nil, fmt.Errorf("userRepo.FindByEmail email=%s: %w", email, err)

// Application → add operation context
return nil, fmt.Errorf("userService.Login: %w", err)

// Interface → map to HTTP via middleware (NEVER expose raw errors)
response.HandleError(c, err) // reads error chain, maps to HTTP status
```

Rules:
- **Always `%w`**, never `%v` (preserves error chain for `errors.Is`/`errors.As`)
- Sentinel errors defined per module in `internal/[module]/domain/errors.go`
- Service layer has **zero knowledge** of HTTP status codes
- **Never:** `c.JSON(500, gin.H{"error": err.Error()})` — leaks internals

### 3.5 Context (ctx) Rules

```go
// ✅ ctx is ALWAYS first parameter of every I/O function
func (r *Repo) FindByID(ctx context.Context, id string) (*User, error) {}

// ✅ Background goroutine detached from request — use WithoutCancel (Go 1.21+)
detachedCtx := context.WithoutCancel(ctx)

// ✅ Worker goroutine MUST have exit condition + recover
go func() {
    defer func() {
        if r := recover(); r != nil { slog.Error("panic", "err", r) }
    }()
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done(): return
        case <-ticker.C: doWork(detachedCtx)
        }
    }
}()

// ❌ NEVER store ctx in struct fields
// ❌ NEVER use context.Background() inside a function that already has ctx
```

---

## 4. Naming Conventions

| Element | Rule | Example |
|---|---|---|
| Package | lowercase single word | `auth`, `notification` |
| File | snake_case + role suffix | `auth_service.go`, `user_repository.go` |
| Receiver | 1–3 char abbreviation (consistent) | `(s *AuthService)`, `(h *UserHandler)` |
| Repository interface | `I` prefix | `IAuthRepository`, `IUserReader` |
| Service interface | `I` prefix | `IAuthService`, `INotificationService` |
| Sentinel errors | `Err` prefix | `ErrUserNotFound`, `ErrTokenExpired` |
| Constants | PascalCase | `MaxRetries`, `DefaultTimeout` |
| Status enum values | SCREAMING_SNAKE | `StatusActive = "ACTIVE"` |
| JSON tags | snake_case | `json:"user_id"`, `json:"created_at"` |

---

## 5. Code Quality Gates (Check Before Finalizing)

```
Architecture
  [ ] No layer dependency violations
  [ ] No cross-module repo/entity imports (use Adapter)
  [ ] No GORM anywhere (grep -r "gorm.io" internal/)

Go Quality
  [ ] ctx first parameter in all I/O functions
  [ ] Errors wrapped with %w (not %v)
  [ ] No magic numbers/strings (use constants)
  [ ] Goroutines have exit conditions
  [ ] Background goroutines use context.WithoutCancel()
  [ ] Goroutines have recover() for panics
  [ ] No unnecessary exported identifiers

Database
  [ ] Only GORM methods in repositories
  [ ] Multi-table writes use UnitOfWork interface
  [ ] No SQL string concatenation

Security
  [ ] No sensitive data in logs (password, token, PII)
  [ ] External HTTP calls have timeout
  [ ] No internal errors exposed in HTTP response body

Testing
  [ ] Business logic covered by unit tests
  [ ] Tests mock interfaces (not concrete types)
  [ ] Table-driven tests for multiple scenarios
```

---

## 6. Workflow Rules

### Before writing any code → always read
1. `.ai/context/ARCHITECTURE.md` — system topology
2. `.ai/context/MODULES.md` — which module handles what
3. `.ai/context/ADR.md` — decisions constraining your task
4. `.ai/context/PATTERNS.md` — exact patterns to follow

### Task type → command mapping
| Task | Use command |
|---|---|
| New module | `/create` — never freestyle |
| New endpoint in existing module | `/create endpoint` |
| Bug investigation | `/debug` — RCA first, fix second |
| Refactor | `/refactor` with pattern reference |
| Review code | `/review` with scope specified |
| Plan complex feature | `/plan` — plan before implement |

### Mandatory: Clarify before coding
For any task > 20 lines of new code:
1. Ask 2+ clarifying questions
2. Present implementation plan
3. Wait for explicit approval
4. Then implement

---

## 7. Module Catalog (Update on every new module)

| Module | Path | Responsibility | Owner Interface |
|---|---|---|---|
| auth | `internal/auth/` | JWT, login, refresh token | `IAuthService` |
| user | `internal/user/` | User profiles, preferences | `IUserService` |
| notification | `internal/notification/` | Delivery orchestration, routing | `INotificationService` |
| mails | `internal/mails/` | Email delivery via SMTP | `IMailService` |
| sms | `internal/sms/` | SMS delivery via Twilio/etc | `ISMSService` |
| push | `internal/push/` | Mobile push via FCM/APNs | `IPushService` |
| webhook | `internal/webhook/` | Outbound HTTP webhooks | `IWebhookService` |
| template | `internal/template/` | Notification templates mgmt | `ITemplateService` |

---

## 8. References

| Document | Location | When to read |
|---|---|---|
| System Architecture | `.ai/context/ARCHITECTURE.md` | Start of every session |
| All ADRs | `.ai/context/ADR.md` | Before architectural decisions |
| Code Patterns | `.ai/context/PATTERNS.md` | Before writing any pattern code |
| Module Map | `.ai/context/MODULES.md` | When unsure where code goes |
| Golang Best Practices | `golang_best_practices.md` | Any Go quality question |
| Workflow Commands | `.agent/workflows/` | Task execution |
| Prompt Templates | `.ai/prompts/` | Starting a new task |

---

*Maintainer: Team Lead | Last updated: 2026-02-18*
*⚠️ Every merged ADR MUST update this file in the same PR.*
