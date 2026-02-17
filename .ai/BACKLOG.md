# Project Backlog
> Last updated: **2026-02-17** by Lead  
> File này chỉ Lead edit. Members cập nhật trạng thái trong `.ai/members/{NAME}.md`.

---

## Active Sprint

| # | Task | Module | Owner | Status | Branch |
|---|------|--------|-------|--------|--------|
| 2 | Email template for invitations | IAM | Minh | ✅ Done | `feature/email-invitation-template` |
| 3 | Pool status refinement | Messaging | Minh | ✅ Done | `feature/pool-status-refinement` |
| 4 | WebSocket real-time updates | Messaging | Minh | ✅ Done | `feature/websocket-realtime` |

---

## Backlog (Prioritized)

| # | Task | Module | Priority | Notes |
|---|------|--------|----------|-------|
| 5 | Audit trail pagination | Messaging | P2 | ✅ Done |

---

## Cancelled

| Task | Reason |
|------|--------|
| SSO Integration (Google OAuth) | Cancelled by Lead — no longer needed |

---

## Completed

| Task | Owner | Date | Key PR/Files |
|------|-------|------|--------------|
| Pool status refinement | Minh | 2026-02-17 | `messaging.go`, `conversation.service.impl.go` |
| Email template for invitations | Minh | 2026-02-17 | `invitation_template.go`, `ses_email.service.go` |
| Workflow Digest Steps | Minh | 2026-02-17 | `digest_handler.go`, `workflow_executor.go` |
| Notification i18n | Minh | 2026-02-17 | `template_content.controller.go` |
| Rate Limiting per Partner | Minh | 2026-02-16 | `ratelimit.go`, `rate_limit.go` |
| Per-Partner Usage Metrics | Minh | 2026-02-16 | `metrics.service.impl.go` |
| Workflow Engine Phase 2 | Minh | 2026-02-15 | `trigger_service.go`, `workflow_executor.go` |
| Webhook Retry Policy Config | Minh | 2026-02-14 | `webhook.repository.go` |
| System Health Dashboard | Minh | 2026-02-14 | `health_service.go` |
| Dashboard APIs | Minh | 2026-02-13 | Multiple controller files |
| SLA Tracking & Enforcement | Minh | 2026-02-13 | `sla_worker.go`, `resolve.go` |
| Codebase Audit (9 issues) | Minh | 2026-02-12 | Multiple fixes |
| FCM Provider | Minh | 2026-02-11 | `fcm_provider.go` |
| SMTP Provider | Minh | 2026-02-10 | `smtp_provider.go` |
| Notification Pipeline | Minh | 2026-02-10 | Full module |
| Workflow Engine Phase 1 | Minh | 2026-02-09 | Full module CRUD |

---

## Module Progress

| Module | Progress | Notes |
|--------|----------|-------|
| **IAM & Tenancy** | ✅ ~95% | Email template done, SSO cancelled |
| **Apps & Environments** | ✅ ~100% | Rate limiting, metrics, webhooks all done |
| **Messaging & Helpdesk** | ✅ ~95% | WebSocket, pool status, audit trail all done |
| **Notification Pipeline** | ✅ 100% | i18n, providers, templates all done |
| **Workflow Engine** | ✅ 100% | CRUD + Trigger + Digest all done |

---

## Technical Notes
- DDD Architecture: `controller → application/service → domain → infrastructure`
- Database: PostgreSQL, schema covers all modules
- Auth: JWT (access + refresh), RBAC with permissions map
- API Framework: Gin
- Middleware: Auth, TenantMembership, PermissionChecker, EnvAuth, RateLimiter
