# Implementation Plan - Messaging Module Unit Tests

## Goal
Tạo các unit test cho `ConversationService` để đảm bảo tính năng chat polymorphic (hỗ trợ cả User và Subscriber) hoạt động chính xác và không có regression.

## Proposed Changes

### Tests
#### [NEW] [conversation_service_test.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/impl/conversation_service_test.go)
- Triển khai các Mock Repository (`Thread`, `Message`, `Subscriber`, `AssignmentLog`) sử dụng `testify/mock`.
- **Test cases cho `GetOrCreateDirectThread`**:
    - Tạo mới thread giữa User và Subscriber.
    - Trả về thread có sẵn giữa User và Subscriber.
    - Trả về thread có sẵn giữa User và User.
- **Test cases cho `CreateGroupThread`**:
    - Tạo nhóm với danh sách participants hỗn hợp (User & Subscriber).
    - Kiểm tra logic tự động thêm người tạo (creator) vào danh sách participants nếu chưa có.
- **Test cases cho `AddGroupParticipants`**:
    - Thêm một danh sách participants hỗn hợp vào group có sẵn.

## Verification Plan

### Automated Tests
- Chạy riêng file test mới:
  ```bash
  go test -v internal/messaging/application/service/impl/conversation_service_test.go
  ```
- Chạy toàn bộ test của module messaging:
  ```bash
  go test -v ./internal/messaging/...
  ```
