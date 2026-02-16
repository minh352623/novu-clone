# Subscriber Chat Implementation Plan

## Goal
Enable 1-1 and Group chats to support both **Users** (internal agents/members) and **Subscribers** (external contacts) by refactoring the API to be polymorphic.

## User Review Required
> [!IMPORTANT]
> **API Breaking Change**: `POST /conversations/group` and `POST /conversations/direct` payloads will change structure to support `type`.
>
> **Old Payload (Group):**
> ```json
> { "name": "Team", "member_ids": ["uuid1", "uuid2"] }
> ```
>
> **New Payload (Group):**
> ```json
> {
>   "name": "Team",
>   "participants": [
>     { "id": "uuid1", "type": "user" },
>     { "id": "uuid2", "type": "subscriber" }
>   ]
> }
> ```

## Proposed Changes

### Domain & DTOs
#### [MODIFY] [conversation.dto.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/dto/conversation.dto.go)
- Define `ParticipantDTO` struct:
  ```go
  type ParticipantDTO struct {
      ID   uuid.UUID `json:"id" binding:"required"`
      Type string    `json:"type" binding:"required,oneof=user subscriber"`
  }
  ```
- Update `CreateGroupChatRequest` to use `[]ParticipantDTO`.
- Update `CreateDirectChatRequest` to use `ParticipantDTO` for the target.
- Update `AddParticipantsRequest` to use `[]ParticipantDTO`.

### Service Layer
#### [MODIFY] [conversation.service.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/conversation.service.go)
- Update `CreateGroupThread` signature to accept `[]entity.ThreadParticipant` (or a struct that holds ID+Type).
- Update `GetOrCreateDirectThread` signature to accept target ID and Type.
- Update `AddGroupParticipants` signature.

#### [MODIFY] [conversation.service.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/impl/conversation.service.impl.go)
- Refactor `GetOrCreateDirectThread`:
  - Handle `GetDirectThreadBetween` to support mixed types (User-Subscriber). *Note: Repository update might be needed if `GetDirectThreadBetween` assumed "user" types implicitly or if we want to maintain the current repo signature, we might need `GetDirectThreadBetweenEntities`.*
- Refactor `CreateGroupThread`: Use the provided type when creating participants.
- Refactor `AddGroupParticipants`.

### Persistence Layer
#### [MODIFY] [common.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/infrastructure/persistence/repository/common.repository.go)
- Update `GetDirectThreadBetween` to accept `(entityTypeA, entityIDA, entityTypeB, entityIDB)` instead of just 2 IDs, or ensure it checks the correct types.
- Current implementation joins `thread_participants` but doesn't filter by `entity_type` in the join condition explicitly? It does: `p1.entity_id = ?`. It assumes uniqueness of ID across tables? No, UUIDs are unique but good practice is to check type.
- **Action**: Update `GetDirectThreadBetween` to be type-safe.

### Controller Layer
#### [MODIFY] [conversation.controller.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/conversation.controller.go)
- Update handlers to map `ParticipantDTO` to service arguments.

## Verification Plan

### Automated Tests
- None planned (Manual verification step).

### Manual Verification
1. **Direct Chat (User-Subscriber)**:
   - Call `POST /conversations/direct` with `{ "target": { "id": "sub_uuid", "type": "subscriber" } }`.
   - Verify thread created with `type=direct`.
   - Verify participants in DB are `user` (creator) and `subscriber` (target).
2. **Group Chat (Mixed)**:
   - Call `POST /conversations/group` with mixed participants.
   - Verify db records.
3. **Legacy Support**:
   - Ensure existing logic for "User-User" chat still works with new payload format.
