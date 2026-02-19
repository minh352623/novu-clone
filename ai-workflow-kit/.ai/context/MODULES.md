# Module Catalog

> **AI instruction:** Check this file to find which module owns what, and who to call when you need data from another domain.
> Update this file whenever a new module is added or cross-module dependencies change.

---

## Module Registry

| Module | Package path | Status | Owner |
|--------|-------------|--------|-------|
| auth | `internal/auth/` | Production | [Team] |
| user | `internal/user/` | Production | [Team] |
| notification | `internal/notification/` | Production | [Team] |
| _[Add new modules here]_ | | | |

---

## Module Details

### `auth` — Authentication

**Purpose:** Account creation, login, JWT issuance, token refresh, logout.

**Key entities:**
- `Account` (id, email, password_hash, role, status)

**Public service interface:** `IAuthService`
```go
// Methods available to other modules (via DI, not direct import)
Register(ctx, req) (*dto.AccountResp, error)
Login(ctx, req) (*dto.TokenResp, error)
RefreshToken(ctx, refreshToken) (*dto.TokenResp, error)
Logout(ctx, refreshToken) error
```

**Exports to other modules:** Nothing (auth is a leaf module — others don't need auth data)

**HTTP endpoints:**
```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
POST /api/v1/auth/logout
```

---

### `user` — User Profile

**Purpose:** User profile management, avatar, preferences.

**Key entities:**
- `User` (id, account_id, name, email, avatar_url, bio)

**Public service interface:** `IUserService`
```go
FindByID(ctx, id) (*entity.User, error)
FindByIDs(ctx, ids) ([]*entity.User, error)
GetProfile(ctx, id) (*dto.UserProfileResp, error)
UpdateProfile(ctx, id, req) (*dto.UserProfileResp, error)
```

**Exports to other modules:**
```go
// Port defined in EACH CONSUMER module (not here)
// Any module needing user data should define:
type IUserReader interface {
    FindByID(ctx context.Context, id string) (*UserBasicInfo, error)
}
// Then use LocalUserAdapter to call IUserService.FindByID()
```

**HTTP endpoints:**
```
GET  /api/v1/users/me
PUT  /api/v1/users/me
GET  /api/v1/users/:id
```

---

### `notification` — Notification

**Purpose:** Create, queue, deliver, and track notifications across channels.

**Key entities:**
- `Notification` (id, user_id, type, channel, status, payload, sent_at)
- `NotificationTemplate` (id, type, channel, subject_template, body_template)
- `DeliveryLog` (id, notification_id, attempt, result, error)

**Public service interface:** `INotificationService`
```go
Send(ctx, req) (*dto.NotificationResp, error)           // create + queue
ListByUser(ctx, userID, filter) (*dto.PaginatedResp, error)
MarkAsRead(ctx, id, userID) error
```

**Cross-module dependencies:**
- Needs user data → defines `IUserReader` in `internal/notification/domain/repository/`
- Wired to `LocalUserAdapter` → `UserService.FindByID()`

**HTTP endpoints:**
```
POST  /api/v1/notifications
GET   /api/v1/notifications
PATCH /api/v1/notifications/:id/read
```

---

## Adding a New Module

When creating a new module, add its entry to this file:

```markdown
### `[name]` — [Short description]

**Purpose:** [What problem does this module solve?]

**Key entities:**
- `[Entity]` (list key fields)

**Public service interface:** `I[Name]Service`
(list methods that other modules could call)

**Exports to other modules:**
(list what data other modules can request via Port+Adapter)

**Cross-module dependencies:**
(list what this module needs from others)

**HTTP endpoints:**
(list all API endpoints)
```

---

## Dependency Graph

```
auth ─────────────────────────────────── (leaf, nothing depends on auth for data)
                                            ↑ JWT middleware validates token from auth
user ────────────────────────────────────── (provider)
  ↑ IUserReader
notification ──── (consumer of user)
  ↑ IUserReader
order ─────────── (consumer of user)

[future modules add to this graph]
```
