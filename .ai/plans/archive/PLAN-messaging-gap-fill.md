# IMPL-messaging-gap-fill: Complementary Messaging Functions

> **Goal**: Fill the functional gaps identified in the Messaging Module Audit to fully satisfy user requirements (Internal Notes, Advanced Assignment, Analytics).

## 1. Domain Model Updates (`internal/messaging/domain`)

### 1.1 Support Internal Notes (US-CA-03)
- **Modify `Message` Entity**:
  - Add `Type` field (enum: `standard`, `internal_note`).
  - Default to `standard`.
- **Logic**:
  - `Internal Notes` must NOT be sent to the Subscriber (customer).
  - Only visible to Agents/Admins in `GetThread`.

### 1.2 "Return to Pool" Logic (US-CA-04)
- **Clarify `Unassign`**:
  - Action: Set `assigned_to` = NULL.
  - Status: Transition from `assigned` -> `unassigned` (or `open` if we want to distinguish).
  - Audit: Log the unassignment event.

## 2. Service Layer Extensions (`internal/messaging/application/service`)

### 2.1 New Methods in `ConversationService`
```go
// Internal Notes
AddInternalNote(ctx, threadID, authorID, content) (*Message, error)

// Assignment Management
UnassignThread(ctx, threadID) error
BulkAssignThreads(ctx, threadIDs []uuid.UUID, memberID uuid.UUID) error

// Analytics (Aggregations)
GetTeamStats(ctx, envID, timeRange) (*TeamStats, error)
GetAgentStats(ctx, agentID, timeRange) (*AgentStats, error)
```

## 3. Controller & API Specifications (`internal/messaging/controller`)

### 3.1 New Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `POST` | `/conversations/{id}/notes` | Create an internal note (invisible to customer) |
| `POST` | `/conversations/{id}/unassign` | Return thread to pool |
| `PATCH` | `/conversations/assign` | Bulk assign threads (Body: `{thread_ids: [], ...}`) |
| `GET` | `/analytics/conversations/team` | Team performance metrics |
| `GET` | `/analytics/conversations/me` | My personal performance metrics |

## 4. Implementation Steps

- [ ] **Step 1: Domain & Persistence**
  - Add `type` column to `messages` table (migration needed).
  - Update `MessageRepository` to support filtering by type (optional, for "public view").
- [ ] **Step 2: Service Logic**
  - Implement `AddInternalNote`: same as Reply but skip `Outbound` logic (no webhook to component).
  - Implement `UnassignThread`: clear assignee, update status, log.
  - Implement `BulkAssign`: loop `AssignThread` or batch update.
- [ ] **Step 3: Analytics Queries**
  - Implement SQL queries for `GetTeamStats` in `AssignmentLogRepository` (avg response time, counts).
- [ ] **Step 4: API Exposure**
  - Map new endpoints in `ConversationController`.
  - Add Swagger docs.

## 5. Verification
- **Automated**: Unit tests for `Unassign` and `BulkAssign`.
- **Manual**:
  - Verify Internal Note does NOT trigger webhook to customer.
  - Verify "Return to Pool" makes thread appear in "Unassigned" list.
