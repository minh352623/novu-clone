# Implementation Plan - Messaging Module Hardening

## Goal
Nâng cấp bảo mật và độ tin cậy của Module Messaging thông qua việc chống lỗi IDOR, xử lý Race Condition khi tạo chat Direct, và hoàn thiện các tính năng polymorphic.

## Proposed Changes

### 1. Bảo mật & Cô lập môi trường (IDOR Protection)
- **Service Layer**: Cập nhật tất cả các phương thức trong `conversationServiceImpl` để bắt buộc kiểm tra `environment_id` của Thread trước khi thực hiện thao tác (Reply, Resolve, Assign, v.v.).
- **Repository Layer**: Thêm phương thức `GetByIDAndEnv(ctx, id, envID)` để truy vấn an toàn.

### 2. Xử lý Concurrency (Direct Chat)
- **Database Schema**: Thêm cột `reference_hash` (unique) vào bảng `threads`.
- **Logic**: Khi tạo chat Direct, tạo hash từ cặp ID người tham gia (đã sắp xếp). Nếu gặp lỗi trùng lặp (Unique Violation), service sẽ tự động fetch lại bản ghi đã tồn tại.

### 3. Hoàn thiện Polymorphic & Pagination
- **Polymorphic Fix**: Cập nhật `RemoveGroupParticipant` để nhận cả `type` và `id` thay vì hardcode `user`.
- **Pagination**: Cập nhật API lấy tin nhắn (`GetThread`) để hỗ trợ phân trang với `limit` mặc định là 20.

---

## Danh sách file thay đổi (Dự kiến)

### [Domain]
#### [MODIFY] [messaging.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/domain/entity/messaging.go)
- Thêm trường `ReferenceHash` vào struct `Thread`.

### [Infrastructure]
#### [MODIFY] [messaging.model.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/infrastructure/persistence/model/messaging.model.go)
- Thêm trường `ReferenceHash` vào `ThreadModel` với tag `gorm:"uniqueIndex"`.
#### [MODIFY] [common.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/infrastructure/persistence/repository/common.repository.go)
- Thêm `GetByIDAndEnv` vào `ThreadRepository`.
- Cập nhật `RemoveParticipant` để dùng cả `entityType`.

### [Application]
#### [MODIFY] [conversation.service.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/impl/conversation.service.impl.go)
- Triển khai logic kiểm tra Environment ID cho tất cả các hàm.
- Triển khai logic tạo `ReferenceHash` cho chat Direct.
- Cập nhật `GetMessagesByThread` với phân trang.

### [Controller]
#### [MODIFY] [conversation.controller.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/conversation.controller.go)
- Cập nhật API `RemoveGroupParticipant` và các API lấy tin nhắn.

---

## Verification Plan

### Automated Tests
- Cập nhật `conversation_service_test.go` để giả lập Race Condition và lỗi IDOR.
- Chạy test check isolation:
  ```bash
  go test -v ./internal/messaging/application/service/impl/...
  ```

### Manual Verification
- Dùng Postman tạo cùng lúc 2 request chat Direct giữa cùng 2 User và kiểm tra DB chỉ có 1 thread.
- Thử lấy tin nhắn của 1 Thread ID thuộc Environment khác bằng Token của Environment hiện tại (phải trả về lỗi 403/404).
