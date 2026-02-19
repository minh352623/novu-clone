# System Architecture

> **Status:** Living document — update when major architectural decisions change.
> **AI instruction:** Read this file before any task involving module creation, cross-module communication, or infrastructure choices.

---

## System Overview

**Project type:** Go monolith with microservice-ready architecture
**Design pattern:** Domain-Driven Design (DDD) — 4 layers per module
**Database:** PostgreSQL (via pgx/v5 + sqlc)
**Cache:** Redis
**Web framework:** GIN
**Auth:** JWT (access token 15min + refresh token 7 days)

---

## High-level Component Map

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client (HTTP)                           │
└──────────────────────────┬──────────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────────┐
│                       GIN Router                                │
│                   Middleware Stack:                             │
│           RequestLogger → Auth → RateLimit → CORS              │
└──────┬──────────────┬──────────────────┬────────────────────────┘
       │              │                  │
┌──────▼──────┐ ┌─────▼──────┐  ┌───────▼──────┐
│  Auth       │ │   Order    │  │ Notification │    ... modules
│  Module     │ │   Module   │  │   Module     │
└──────┬──────┘ └─────┬──────┘  └───────┬──────┘
       │              │                  │
┌──────▼──────────────▼──────────────────▼──────────────────────┐
│                     Infrastructure                              │
│    PostgreSQL (sqlc)    Redis (cache)    Event Bus (in-proc)   │
└────────────────────────────────────────────────────────────────┘
```

---

## Module Structure (per module)

Every module follows this exact structure:

```
internal/[module]/
├── domain/                          # Layer 1: Business logic kernel
│   ├── model/
│   │   ├── entity/                  # Core business entities
│   │   └── vo/                      # Value objects (optional)
│   ├── repository/                  # Repository INTERFACES
│   └── errors.go                    # Domain error variables
│
├── application/                     # Layer 2: Use case orchestration
│   ├── service/
│   │   ├── i_[module]_service.go    # Service interface
│   │   └── [module]_service.go      # Service implementation
│   └── schedule/                    # Background jobs (optional)
│
├── infrastructure/                  # Layer 3: Technical implementation
│   ├── persistence/
│   │   ├── queries/                 # SQL files (input to sqlc)
│   │   ├── sqlcgen/                 # sqlc generated (DO NOT EDIT)
│   │   └── repository/              # Repository implementations
│   ├── cache/                       # Redis cache implementations
│   └── adapter/                     # Cross-module adapters
│
├── controller/                      # Layer 4: HTTP interface
│   ├── dto/                         # Request / Response structs
│   └── http/                        # GIN handlers
│
└── [module].module.go               # DI wiring + route registration
```

---

## Existing Modules

### Auth Module (`internal/auth/`)
**Responsibility:** Authentication and token management
**Key entities:** Account
**API surface:**
- `POST /api/v1/auth/login` — login, returns access + refresh tokens
- `POST /api/v1/auth/register` — create account
- `POST /api/v1/auth/refresh` — exchange refresh token for new access token
- `POST /api/v1/auth/logout` — invalidate refresh token

**Notes:**
- JWT access token: 15min TTL, signed with RS256
- Refresh token: 7 days TTL, stored in Redis with rotation
- Passwords hashed with bcrypt cost=12

---

### User Module (`internal/user/`)
**Responsibility:** User profile management
**Key entities:** User
**API surface:**
- `GET /api/v1/users/me` — get current user profile
- `PUT /api/v1/users/me` — update profile
- `GET /api/v1/users/:id` — get public profile (limited fields)

**Exports to other modules:**
- `IUserReader` interface (see `domain/repository/i_user_reader.go`)
- Methods: `FindByID(ctx, id) → UserBasicInfo`

---

### Notification Module (`internal/notification/`)
**Responsibility:** Multi-channel notification delivery
**Key entities:** Notification, NotificationTemplate, DeliveryLog
**Channels:** Email (SMTP), in-app (WebSocket/polling)
**API surface:**
- `POST /api/v1/notifications` — create + queue notification
- `GET /api/v1/notifications` — list for current user
- `PATCH /api/v1/notifications/:id/read` — mark as read

**Notes:**
- Delivery is async via in-process event bus
- Failed deliveries are retried up to 3 times with exponential backoff
- Templates stored in DB with variable substitution ({{user_name}}, {{amount}})

---

## Cross-Module Communication

**Pattern:** Interface (Port) + Adapter (see PATTERNS.md)
**Rule:** Consumer module defines the interface, provider module is ignorant of it.

Current cross-module dependencies:
```
Notification → reads from → User (via IUserReader)
Order        → reads from → User (via IUserReader)
```

When switching to microservice: swap `LocalUserAdapter` → `HttpUserAdapter` in module wiring.

---

## Infrastructure Decisions

| Concern | Technology | Why |
|---------|-----------|-----|
| Web framework | GIN | Performance, mature, simple |
| Database driver | pgx/v5 | Best performance for PostgreSQL |
| Query layer | sqlc | Type-safe SQL, compile-time validation |
| Cache | Redis (go-redis/v9) | Standard, supports pub/sub for future |
| Auth | JWT (golang-jwt/v5) | Stateless, standard |
| Config | envconfig | Type-safe env parsing |
| Logging | slog (stdlib) | Go 1.21+, structured, zero dependency |
| UUID | google/uuid | Standard |
| Validation | go-playground/validator | Works with GIN binding |

---

## Deployment Architecture

```
[Client] → [Load Balancer] → [Go binary (multiple replicas)]
                                      ↓               ↓
                               [PostgreSQL]      [Redis]
```

- Binary: single binary, all modules compiled in
- Horizontal scaling: multiple replicas behind load balancer
- Session: stateless (JWT) — any replica can serve any request
- Migrations: run via `make migrate-up` before deployment (golang-migrate)
