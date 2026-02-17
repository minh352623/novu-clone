# Plan: SLA Configuration & Audit Trail

> **Goal**: Persist SLA settings in Environment/App configuration and expose Thread Audit Trail enriched with agent names.

## 1. Requirement Refinement (Socratic Gate - DONE)

- **SLA Config Scope**: Defined at **Environment** level with a fallback at the **App** level.
- **Audit Trail Data**: Enrich logs with **Member/Agent display names**.
- **Calculation strategy**: Use the **current** configuration for calculations (simpler Dashboard logic).
- **Default Value**: **900 seconds** (15 minutes).

## 2. Technical Architecture

### 2.1. Apps Module (`internal/apps`)
- **Domain**: 
  - Add `SLAThresholdSeconds` (int) to `App` entity (as default).
  - Add `SLAThresholdSeconds` (int) to `Environment` entity (as override).
- **DTO**: 
  - Update `CreateEnvironmentRequest`, `UpdateEnvironmentRequest`, `EnvironmentResponse`.
  - Update `CreateAppRequest`, `UpdateAppRequest`, `AppResponse`.
- **Service**: Implement fallback logic: `env.SLA != 0 ? env.SLA : app.SLA != 0 ? app.SLA : 900`.

### 2.2. Messaging Module (`internal/messaging`)
- **Domain**: Add `GetByThread(ctx, threadID)` to `AssignmentLogRepository`.
- **Service**: 
  - Add `GetThreadAuditTrail(ctx, threadID)` to `ConversationService`.
  - Fetch display names from `IAM` module or internal cache if available.
- **Controller**:
  - `GET /v1/api/conversations/:id/audit-trail` -> Returns `[]AssignmentLogResponse` (enriched).

## 3. Review Levels
- **Level 1 (Local)**: Verify no breaking changes to existing Dashboard APIs.
- **Level 2 (Proposal)**: Verify REST endpoint structure and DTO mapping.
- **Level 3 (Lead Approval)**: APPROVED by Lead.

## 4. State Management
- Update `AI_STATE_MINH.md` after each Phase.
- Update `PROJECT_MAIN_BACKLOG.md` to mark SLA Configuration as complete.

## 5. Implementation Steps

### Phase 1: Apps Module (SLA Configuration)
- [ ] Update `entity.App` and `entity.Environment` with SLA fields.
- [ ] Update GORM models and run migrations.
- [ ] Update DTOs and Controllers in `internal/apps`.

### Phase 2: Messaging Module (Audit Trail Repository)
- [ ] Implement `GetByThread` in `AssignmentLogRepository`.
- [ ] Create `ToAssignmentLogResponse` DTO with display name fields.

### Phase 3: Messaging Module (Service & Controller)
- [ ] Implement `GetThreadAuditTrail` in `ConversationService` with enrichment logic.
- [ ] Add `GET /v1/api/conversations/:id/audit-trail` handler.

### Phase 4: Integration & Verification
- [ ] Update `Dashboard` methods to use dynamic SLA from Apps module.
- [ ] Verify enrichment works correctly with real agent data.
