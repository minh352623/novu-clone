# Plan: Messaging Module Implementation

> **Goal**: Implement the Core Messaging Engine to support CS Agent operations (Conversation Pools, Assignment, Nested Set Messages) as defined in `functions.md` (Section 4) and `schema.sql` (Section 4).

## 1. Context & Scope
- **Current State**: `internal/messaging` skeleton exists but is empty.
- **Dependencies**: 
  - `IAM` (for Member assignment).
  - `Apps/Environments` (for context).
  - `Schema`: `subscribers`, `conversation_pools`, `assignment_logs`, `messages`.
- **References**: 
  - `functions.md`: US-CL-01 (View), US-CL-02 (Assign), US-CA-02 (Reply), US-CA-05 (Close).

## 2. Technical Architecture

### 2.1. Domain Entities (`internal/messaging/domain/entity`)
- **Subscriber**: Represents the external user (customer).
    - Fields: `ID`, `SubscriberKey`, `Data`.
- **ConversationPool** (The "Ticket"): 
    - Fields: `ID`, `Status` (Unassigned/Assigned/Closed), `AssignedToMemberID`.
- **Message**: 
    - Fields: `ID`, `Content` (JSONB), `SenderType`, `NestedSet` (Lft, Rgt, Depth).
    - **Logic**: Use `Nested Set` for efficient tree retrieval (though implementation might start simple with Adjacency List + Recursive query or just Adjacency if depth is flat). *Note: Schema uses Nested Set.*

### 2.2. Persistence Layer (`internal/messaging/infrastructure/persistence`)
- **Repository Pattern**:
    - `SubscriberRepository`: FindOrCreate.
    - `PoolRepository`: List by Status, Assign Member.
    - `MessageRepository`: Insert (using Stored Proc `add_message_node`?), GetByPool.
- **Optimization**: Use the `add_message_node` SQL function defined in schema if complex, or Implement logic in Go. *Recommendation: Use Go logic for better control/testing, but respect Schema design.*

### 2.3. Service Layer (`internal/messaging/application/service`)
- **ConversationService**:
    - `StartConversation(subscriberKey, initialMsg)`
    - `ReplyToConversation(poolID, content, sender)`
    - `GetHistory(poolID)`
- **AssignmentService**:
    - `AssignToAgent(poolID, memberID)`
    - `ReturnToPool(poolID)`

### 2.4. API Layer (`internal/messaging/controller`)
- **REST Endpoints**:
    - `POST /api/v1/conversations/inbound` (Webhook/SDK trigger)
    - `GET /api/v1/conversations` (Filter: Status, Agent)
    - `POST /api/v1/conversations/:id/messages` (Reply)
    - `PATCH /api/v1/conversations/:id/assign` (Assignment)
- **Real-time**:
    - *Future*: WebSocket for live chat. (For now, focus on REST).

## 3. Execution Steps

### Phase 1: Foundation (Entities & Repos)
- [ ] Define `Subscriber`, `ConversationPool`, `Message` entities.
- [ ] Implement GORM models & Mappers.
- [ ] Implement `SubscriberRepo` and `PoolRepo`.

### Phase 2: Message Handling (The Heavy Lift)
- [ ] Implement `MessageRepo`.
- [ ] **Crucial**: Implement `AddMessage` logic (Nested Set calculation: `UPDATE rgt/lft` logic).
    - *Decision*: Port the SQL Stored Proc logic to Go Gorm Transaction to ensure testability and safety.

### Phase 3: Service Logic
- [ ] `ConversationService`: Orchestrate flow "New Msg -> Find/Create Subscriber -> Find/Create Open Pool -> Add Msg".
- [ ] `AssignmentService`: Handle state transitions & Logging(`assignment_logs`).

### Phase 4: API & Integration
- [ ] Create `ConversationController`.
- [ ] Register Routes in `router.go`.
- [ ] **Verification**: Create Unit Test `TestConversationFlow`.

## 4. Verification Plan
- **Manual**: Use Postman to simulate a customer sending a message, then an Agent replying.
- **Automated**: `TestNestedSetInsertion` to ensure `lft/rgt` values are correct after multiple replies.

## 5. Agent Assignments
- **Backend Specialist**: Full implementation.
- **Database Architect**: Review Nested Set logic performance.
