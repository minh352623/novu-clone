# PLAN-messaging-features

## Goal Description
Enhance the Messaging Module to support **Internal Chat** (1-1 Direct Chat and Group Chat) between Tenant Members, and propose broad improvements for module completion. This builds upon the existing `conversation_pools` architecture by introducing distinct thread types.

## User Review Required
> [!IMPORTANT]
> **Architecture Decision**: We are reusing the `conversation_pools` table for all thread types (`support`, `direct`, `group`) to maintain a unified messaging system. This requires ensuring `status` and `participant` logic handles both Support (Assignee-based) and Internal (Participant-based) flows correctly.

## Proposed Changes

### Domain & Entity
#### [MODIFY] [messaging.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/domain/entity/messaging.go)
- Ensure `Thread` struct fully supports `direct` and `group` types.
- Add necessary constants for Thread Types.

### Controller Layer
#### [MODIFY] [router.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/router.go)
- Register new endpoints:
  - `POST /conversations/direct` (Create/Get 1-1)
  - `POST /conversations/group` (Create Group)
  - `PATCH /conversations/group/:id` (Update Group)
  - `POST /conversations/group/:id/participants` (Add Members)
  - `DELETE /conversations/group/:id/participants/:memberId` (Remove Members)
  - `POST /conversations/:id/read` (Mark as Read)

#### [MODIFY] [conversation.controller.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/conversation.controller.go)
- Implement handlers for the above endpoints.
- Update `ListConversations` to support `type` filtering.

#### [MODIFY] [conversation.dto.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/dto/conversation.dto.go)
- Add DTOs: `CreateDirectChatRequest`, `CreateGroupChatRequest`, `AddParticipantsRequest`, etc.

### Service Layer
#### [MODIFY] [conversation.service.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/conversation.service.go)
- Add interface methods:
  - `GetOrCreateDirectThread(ctx, tenantID, memberA, memberB)`
  - `CreateGroupThread(ctx, tenantID, name, members)`
  - `AddGroupParticipants(...)`
  - `RemoveGroupParticipant(...)`
  - `MarkThreadRead(...)`

#### [MODIFY] [conversation.service.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/impl/conversation.service.impl.go)
- Implement valid logic for handling Direct/Group threads.
- Ensure `ListThreads` queries correctly for participants (not just assignees) when type is `direct` or `group`.

### Persistence Layer
#### [MODIFY] [thread.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/infrastructure/persistence/repository/thread.repository.go)
- Update `List` method to filter by `type` and join `thread_participants` correctly for Internal Chat visibility.
- Add `GetDirectThreadBetween(memberA, memberB)` method.

## Verification Plan

### Automated Tests
- Run existing messaging tests (if any) to ensure no regression.
- `go test ./internal/messaging/...`

### Manual Verification
1. **Direct Chat**:
   - Use API to create a chat between Member A and Member B.
   - Verify both can see the thread in `GET /conversations?type=direct`.
   - Send messages and verify delivery.
2. **Group Chat**:
   - Create a group with Members A, B, C.
   - Verify all see the thread.
   - Add Member D -> Verify Member D sees history (or new messages depending on policy).
3. **Legacy Support**:
   - Verify existing Support Chat flow (`inbound` -> `assign` -> `reply`) still works.
