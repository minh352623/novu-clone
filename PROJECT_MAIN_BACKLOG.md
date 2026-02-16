# PROJECT MAIN BACKLOG
> File dùng chung cho cả team. Cập nhật lần cuối: **2026-02-11**

---

## 1. TỔNG QUAN TIẾN ĐỘ MODULES

| Module | Tiến độ | Ghi chú |
|---|---|---|
| **IAM & Tenancy** | ✅ ~85% | Auth, Tenant, Role, Member, Invitation, Pricing Plan đã xong |
| **Apps & Environments** | ✅ ~80% | App, Env, API Key, Webhook, Provider, Metrics đã xong |
| **Messaging & Helpdesk** | ⚠️ ~55% | Core engine xong, thiếu layer CS Operations & Real-time |
| **Notification Pipeline** | � ~20% | Core service & interface xong, thiếu DB & Providers |
| **Workflow Engine** | 🔴 ~5% | Schema DB có, chưa có Go service |

---

## 2. CHI TIẾT TỪNG MODULE

### 2.1 IAM & Tenancy (`internal/iam`)

#### ✅ Đã hoàn thành
- [x] US-PA-01: Invite user qua email (InviteMember, AcceptInvitation)
- [x] US-PA-02: Phân quyền vai trò (AssignRole, RevokeRole, RBAC permissions)
- [x] Flow A: Tạo Tenant → Invite Member → Accept → Gán Role
- [x] Auth: Register, Login, RefreshToken, ChangePassword, Me
- [x] Tenant CRUD + ListTenants + GetMyTenants
- [x] Role CRUD + permissions map
- [x] Pricing Plan CRUD + SetAsDefault
- [x] Token Service: Access + Refresh token

#### ⚠️ Còn thiếu
- [ ] **SSO thực tế** (Google, Azure AD OAuth) — hiện chỉ có email/password
- [ ] **Gửi email invitation** — chưa tích hợp SMTP service
- [ ] **Email template** cho invitation (Subject, Body theo spec US-PA-01)

---

### 2.2 Apps & Environments (`internal/apps`)

#### ✅ Đã hoàn thành
- [x] App CRUD (per Tenant)
- [x] Environment: Create, List, Get, VerifyAPIKey
- [x] API Key: Generate, Rotate, Validate, Revoke, List
- [x] Webhook CRUD + Toggle active
- [x] Provider CRUD + Toggle active + GetActiveProvider
- [x] Metrics: RecordUsage, GetAppMetrics
- [x] System Environment: ListAll (dev/staging/production)

#### ⚠️ Còn thiếu
- [ ] **US-RA-02.4**: Webhook retry policy config (max retries, backoff, timeout)
- [ ] **Webhook dispatch engine** — bảng `webhook_logs` có, chưa có service gửi & retry
- [ ] **US-RA-02.5**: Per-partner usage metrics chi tiết (inbound/outbound/failed messages)
- [ ] **Rate limiting** config per Partner (max req/min, max msg/day)

---

### 2.3 Messaging & Helpdesk (`internal/messaging`)

#### ✅ Đã hoàn thành
- [x] Flow B: Receive inbound message → Tạo/update Thread
- [x] US-CA-02: Reply message
- [x] US-CA-03: Internal notes
- [x] US-CL-02: Assign / Reassign / Bulk assign thread
- [x] US-CA-04: Unassign (trả về pool)
- [x] US-CA-05: Resolve thread
- [x] Internal chat: Direct + Group (CRUD participants)
- [x] Cursor-based pagination (GetMessagesByCursor)
- [x] Analytics cơ bản: GetTeamStats, GetAgentStats
- [x] AssignmentLog entity
- [x] Mark thread as read

#### 🔴 Còn thiếu (ưu tiên cao)
- [ ] **WebSocket** cho real-time updates (agent nhận tin nhắn mới)
- [ ] **US-RA-01**: System Health Dashboard (status, API perf, error tracking, queue health)
- [x] **US-PA-03**: Partner Dashboard (conversation volume, response time distribution, team performance)
- [x] **US-CL-03**: Team dashboard chi tiết (workload formula, pool management, agent load status)
- [ ] **US-CA-06**: Personal dashboard chi tiết (activity timeline, performance trend, team comparison)
- [ ] **SLA tracking** — chưa có config SLA + compliance calculation
- [x] Channel concept — Added `channel` field to Thread/DTO/Service
- [ ] **Audit trail endpoint** — AssignmentLog có entity, chưa expose API
- [ ] **US-CA-05 enforcement** — chưa block message sau khi thread resolved
- [ ] **Pool status** rõ ràng — hiện chỉ có unassigned/assigned/resolved

---

### 2.4 Notification Pipeline (`internal/notification` — CHƯA CÓ)

> Tham chiếu: Flow C trong `project_flows.md`

- [x] Notification Service: trigger send (template_code, subscriber_key, variables)
- [x] Template management: load template + active version (Core logic)
- [x] Layout system: header/footer compilation
- [x] CMS Engine: compile Layout + Template + Variables → final content
- [x] Provider dispatch: gọi FCM/SMTP/Twilio (SMTP & FCM DONE)
- [x] Notification logging: ghi kết quả vào `notifications`
- [ ] Multi-language support (i18n)

---

### 2.5 Workflow Engine (`internal/workflow` — CHƯA CÓ)

> Tham chiếu: Flow D trong `project_flows.md`

- [ ] Workflow CRUD (active/inactive)
- [ ] Step execution engine
- [ ] Step types: Channel (gửi tin), Delay (chờ), Digest (gộp)
- [ ] Trigger event handling
- [ ] Sequential step execution với state tracking

---

## 3. BACKLOG ƯU TIÊN

### 🔴 P0 — Critical (cần làm trước)
1. **[DONE] WebSocket Real-time** — Messaging module (Hub & Client implemented)
2. **[DONE] Gửi email invitation** — IAM flow (Implemented via AWS SES)
3. **[DONE] Channel concept** — Added `channel` field to Thread

### 🟡 P1 — High (sprint tiếp theo)
4. **Notification Pipeline** — Flow C, cần cho push/email notifications
5. **[DONE] Dashboard APIs** (US-PA-03, US-CL-03)
6. **SLA configuration & tracking**
7. **Webhook dispatch engine** + retry logic

### 🟢 P2 — Medium (backlog)
8. **System Health Dashboard** (US-RA-01)
9. **Workflow Engine** (Flow D)
10. **SSO integration** (OAuth providers)
11. **Rate limiting per Partner**
12. **Per-partner detailed usage metrics**

---

## 4. GHI CHÚ KỸ THUẬT
- Dự án follow DDD: `controller → application/service → domain → infrastructure`
- Database: PostgreSQL, schema đã cover cả Workflow + Notification tables
- Auth: JWT (access + refresh token), RBAC với permissions map
- API framework: Gin
- Middleware: Auth, TenantMembership, PermissionChecker, EnvAuth (API Key)
