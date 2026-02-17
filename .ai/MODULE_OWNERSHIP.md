# Module Ownership Matrix

> Mỗi module có 1 **PRIMARY** owner. Khi cần sửa cross-module, phải flag trong plan.  
> File này là **readonly** — chỉ Lead update.

---

## Module → Owner Mapping

| Module | Path | Owner | Backup |
|--------|------|-------|--------|
| **IAM & Tenancy** | `internal/iam/` | Minh | — |
| **Apps & Environments** | `internal/apps/` | Minh | — |
| **Messaging & Helpdesk** | `internal/messaging/` | Minh | — |
| **Notification Pipeline** | `internal/notification/` | Minh | — |
| **Workflow Engine** | `internal/workflow/` | Minh | — |
| **Shared Infrastructure** | `internal/infrastructure/`, `pkg/` | Lead | Any |
| **Middleware** | `internal/middleware/` | Lead | Any |
| **Health** | `internal/health/` | Lead | Any |

> **Note**: Khi team scale lên 3–5 members, Lead phân lại ownership. Ví dụ:
> - Member B → Messaging & Helpdesk
> - Member C → Apps & Environments
> - Member D → Notification Pipeline

---

## Shared Files (High Conflict Risk)

Những files này bị nhiều modules sử dụng — cần coordinate:

| File / Folder | Rule |
|---------------|------|
| `internal/initialize/*.go` | Coordinate với module owner trước khi sửa |
| `sql/schema/*.sql` | **Claim migration number trước khi viết** (check số tiếp theo trong folder) |
| `docs/swagger.*` | Auto-generated — **KHÔNG BAO GIỜ edit thủ công** |
| `global/` | Chỉ Lead edit |
| `go.mod` / `go.sum` | Notify Lead khi thêm dependency mới |

---

## Cross-Module Change Rules

```
Nếu task của bạn cần sửa file trong module của người khác:

1. Ghi rõ trong plan: "Cross-Module: requires {OWNER} review"
2. Liệt kê exact files + exact changes
3. Owner phải APPROVED trước khi bạn sửa
4. Nếu owner không available → escalate lên Lead
```

---

## How to Claim a Module

Khi member mới join team:
1. Lead assign module trong file này
2. Member đọc toàn bộ source code của module được assign
3. Member tạo file `.ai/members/{NAME}.md` từ template
4. Member bắt đầu nhận tasks từ `BACKLOG.md`
