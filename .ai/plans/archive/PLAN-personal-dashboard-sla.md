# PLAN: Personal Dashboard, SLA Tracking & Resolve Enforcement

This plan covers the implementation of the Personal Dashboard (US-CA-06), background SLA overdue tracking, and the enforcement of the "Resolved" thread status on new inbound messages.

## 1. Problem Description
- **Personal Dashboard**: Agents need a view of their own activity, performance trends (hourly/daily), and comparison against team averages/top performers.
- **SLA Tracking**: Threads exceeding the configured SLA threshold need to be flagged as "Overdue" automatically by a background process.
- **Resolve Enforcement**: Inbound messages to "Resolved" threads should automatically reopen them (status to "Unassigned" or "Open").

## 2. Proposed Changes

### 2.1 Messaging Module (`internal/messaging`)

#### [MODIFY] [conversation.dto.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/dto/conversation.dto.go)
- Add `AgentActivityPoint` struct (Time, Count).
- Add `AgentPerformance` struct (AverageResponseTime, ResolvedCount, etc.).
- Add `PersonalDashboardResponse` DTO (ActivityTimeline, Performance, TeamAverageComparison).

#### [MODIFY] [thread.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/domain/model/entity/thread.go)
- Add `IsOverdue` bool field to `Thread` entity (GORM model).

#### [MODIFY] [conversation.service.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/impl/conversation.service.impl.go)
- Implement `GetPersonalDashboard(ctx, memberID, req)`:
    - Hourly stats for today.
    - Daily stats for the week.
    - Team average comparison logic.
- Update inbound message handling:
    - Check if `Thread.Status == ThreadStatusResolved`.
    - If resolved, change status to `ThreadStatusUnassigned` (or `ThreadStatusOpen`).

#### [NEW] [sla_worker.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/worker/sla_worker.go)
- Implement `SLAWorker` with a Go ticker (e.g., every 1 minute).
- Logic:
    1. Fetch apps/environments with SLA configs.
    2. Query active threads (unassigned or assigned but not resolved).
    3. Calculate if `time.Now() - last_marked_at > threshold`.
    4. Update `IsOverdue = true` for matching threads.
    5. Trigger notification/event (optional but recommended for real-time).

#### [MODIFY] [messaging.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/initialize/messaging/messaging.go)
- Initialize and start the `SLAWorker` goroutine.

### 2.2 Controller Layer

#### [MODIFY] [conversation.controller.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/conversation.controller.go)
- Add `GetPersonalDashboard` handler.
- Register route: `GET /dashboards/personal`.

## 3. Review Levels (Collaboration Protocol)
- **Level 1 (Local Integration)**: Ensure `SLAWorker` doesn't overlap with existing `AssignmentLog` repository logic for performance calculation.
- **Level 2 (Proposal)**: This plan serves as the proposal.
- **Level 3 (Lead Approval)**: Awaiting `APPROVED` from Lead.

## 4. Verification Plan

### Automated Tests
- `go test ./internal/messaging/application/worker/...` (Unit test for worker logic).
- `go test ./internal/messaging/application/service/impl/...` (Verify dashboard metrics).

### Manual Verification
- Resolve a thread, send an inbound message (via Postman/Webhook), verify status changes to "Unassigned".
- Set a short SLA (e.g., 1 min), wait, and verify `is_overdue` updates in DB.

## 5. State Management
- Update `AI_STATE_MINH.md` after completion of each phase.
