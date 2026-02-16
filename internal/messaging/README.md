# Messaging Module Documentation

Module này quản lý toàn bộ hệ thống hội thoại giữa khách hàng (Subscribers) và Agent. Ưu tiên tính đa kênh, bảo mật và khả năng mở rộng.

## 1. Kiến trúc Tổng quan (Architecture)

Module được xây dựng theo kiến trúc Clean Architecture:
- **Controller**: Xử lý HTTP request, parse DTO và thực hiện authentication.
- **Service**: Chứa Business Logic (xử lý luồng tin nhắn, gán hội thoại, logic kết thúc).
- **Repository**: Tương tác với Database bằng Gorm.

### Phân cấp tin nhắn (Closure Table)
Hệ thống sử dụng mô hình **Closure Table** để quản lý quan hệ cha-con (reply) giữa các tin nhắn. Bảng `message_closure` lưu trữ tất cả các đường dẫn trong cây hội thoại, cho phép:
- Truy vấn toàn bộ thread hoặc một nhánh con cực nhanh.
- Không giới hạn độ sâu của các câu trả lời.

## 2. Mô hình Dữ liệu (Key Entities)

- **Subscriber**: Khách hàng bên ngoài (từ SDK/Widget). Định danh bởi `SubscriberKey`.
- **ConversationPool**: Nhóm các tin nhắn lại thành một "phiên" làm việc. Có trạng thái: `unassigned`, `assigned`, `resolved`.
- **Message**: Nội dung tin nhắn. Có các loại `agent`, `contact`, `system`.
- **AssignmentLog**: Ghi lại lịch sử gán agent và thời gian xử lý phục vụ Analytics.

---

## 3. Danh sách API Endpoints

### 🛡️ Nhóm API cho Frontend / SDK (Public-facing)
Sử dụng Authentication dựa trên `X-API-Key`.

#### `POST /conversations/inbound`
- **Nhiệm vụ**: Nhận tin nhắn từ khách hàng (Subscriber).
- **Logic**: 
    - Nếu Subscriber chưa tồn tại -> Tự động tạo mới.
    - Nếu đã có hội thoại đang mở (`status != resolved`) -> Thêm tin nhắn vào đó.
    - Nếu chưa có hội thoại mở -> Tạo `ConversationPool` mới.
- **Header**: `X-API-Key: <environment_api_key>`

### 👤 Nhóm API cho Dashboard / Agent (Internal)
Sử dụng Authentication bằng Token của Agent.

#### `GET /conversations`
- **Nhiệm vụ**: Lấy danh sách các hội thoại (Inbox).
- **Query Params**:
    - `status`: Lọc theo trạng thái (`unassigned`, `assigned`, `resolved`).
    - `assigned_to`: `me` để lấy danh sách của chính mình, `all` để lấy toàn bộ.

#### `GET /conversations/:id`
- **Nhiệm vụ**: Lấy toàn bộ lịch sử tin nhắn của một hội thoại cụ thể.

#### `PATCH /conversations/:id/assign`
- **Nhiệm vụ**: Gán hội thoại cho một Agent.
- **Payload**: `{ "member_id": "UUID" }` (Gửi lên ID của agent muốn gán).
- **Hệ quả**: Chuyển trạng thái hội thoại sang `assigned` và ghi record vào `assignment_logs`.

#### `POST /conversations/:id/messages`
- **Nhiệm vụ**: Agent gửi tin nhắn trả lời khách hàng.

#### `POST /conversations/:id/resolve`
- **Nhiệm vụ**: Đánh dấu hội thoại đã được xử lý xong.
- **Hệ quả**: Cập nhật thống kê `response_time_seconds` cho Agent trong logs.

---

## 4. Bảo mật (Security)

Hệ thống áp dụng 2 tầng bảo mật:
1. **Environment Auth (Public)**: Dùng API Key để xác định `TenantID` và `EnvironmentID`. Ngăn chặn việc gửi tin nhắn giả mạo giữa các môi trường (Dev/Prod) hoặc giữa các khách hàng khác nhau.
2. **RBAC (Internal)**: Agent phải có quyền truy cập vào Tenant tương ứng mới có thể xem hoặc gán hội thoại.
