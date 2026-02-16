# Documentation Update Plan: Subscriber Chat Support

## Goal
Update `docs/MESSAGING_MODULE.vi.md` to accurately reflect the recently implemented polymorphic participant support for Direct and Group chats.

## Proposed Changes

### Documentation
#### [MODIFY] [MESSAGING_MODULE.vi.md](file:///Users/tekix/Documents/company/converda/converda-service/docs/MESSAGING_MODULE.vi.md)
- Update **Section 2.2 User / Agent APIs**:
    - Update `POST /conversations/direct` payload description to use `target: {id, type}`.
    - Update `POST /conversations/group` payload description to use `participants: [{id, type}]`.
    - Update `POST /conversations/group/:id/participants` payload description to use `participants: [{id, type}]`.
- Update **Section 3. Feature Flows**:
    - Update Mermaid diagrams for Direct Chat and Group Chat to show the new payload structure.
    - Mention that $Subscriber$ can now be part of these chats.

## Verification Plan
### Manual Verification
- Review the markdown file for clarity and accuracy against the implementation in `conversation.dto.go` and `conversation.controller.go`.
