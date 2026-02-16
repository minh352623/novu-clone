# 🧠 Brainstorm: Messaging Module Enhancements

## Context
Hiện tại module Messaging được thiết kế tập trung vào tính năng **Support (CSKH)**:
- **Subscriber** (Khách hàng) tạo hội thoại -> **Agent** (Nhân viên) reply.
- Table `conversation_pools` (Threads) lưu trữ hội thoại.
- Table `messages` lưu nội dung (với mô hình Closure Table cho thread/reply).

**Mục tiêu**: Mở rộng để hỗ trợ **Chat nội bộ (Internal Chat)**:
1. **1-1 Chat**: Member chat với Member.
2. **Group Chat**: Nhóm nhiều Member.

---

## 🏗️ Architectural Options for Internal Chat

### Option A: Reuse `conversation_pools` (Unified Model)
Sử dụng chung bảng `conversation_pools` hiện có, phân biệt bằng cột `type`.

- **Type `support`**: Subscriber <-> Agent (Hiện tại).
- **Type `direct`**: Member <-> Member (1-1).
- **Type `group`**: Multiple Members.

✅ **Pros:**
- Tận dụng lại logic `messages`, `message_closure`, `repositories`.
- API thống nhất cho việc lấy danh sách tin nhắn, reply, assignments.
- Dễ dàng mở rộng (VD: thêm member vào chat 1-1 -> thành group).

❌ **Cons:**
- Logic phân quyền (RLS) phức tạp hơn chút (Support thì check assigned, Internal thì check participant).
- Cần điều chỉnh bảng `conversation_pools` nếu có field `subscriber_id` bắt buộc (Hiện tại nullable -> OK).

📊 **Effort:** Low

### Option B: Separate Tables (`direct_chats`, `group_chats`)
Tạo bảng riêng cho chat nội bộ.

✅ **Pros:**
- Tách biệt hoàn toàn logic Support và Internal.
- Schema tối ưu riêng cho từng loại.

❌ **Cons:**
- Duplicate code xử lý tin nhắn (gửi, nhận, file, read status).
- Khó quản lý "All Conversations" của một user.

📊 **Effort:** High

---

## 💡 Recommendation
**Chọn Option A (Unified Model)** vì tính nhất quán, giảm thiểu code thừa và tận dụng kiến trúc linh hoạt hiện có.

---

## 🔌 API Definitions (Proposed)

### 1. 1-1 Chat API

**Create/Get Direct Chat**
- `POST /api/v1/conversations/direct`
- Body: `{ "target_member_id": "uuid" }`
- Logic:
  - Kiểm tra xem đã tồn tại thread `type=direct` giữa 2 members chưa.
  - Nếu có -> Return existing ID.
  - Nếu chưa -> Create new thread & add participants.

### 2. Group Chat API

**Create Group**
- `POST /api/v1/conversations/group`
- Body: `{ "name": "Team A", "member_ids": ["uuid1", "uuid2"] }`

**Update Group Info**
- `PATCH /api/v1/conversations/group/:id`

**Add Members**
- `POST /api/v1/conversations/group/:id/participants`
- Body: `{ "member_ids": ["uuid"] }`

**Leave/Remove Member**
- `DELETE /api/v1/conversations/group/:id/participants/:memberId`

### 3. General Enhancements

**List Conversations**
- `GET /api/v1/conversations?type=direct,group,support`
- Cần update method `ListThreads` để lọc theo `type` và check `participant` thay vì chỉ `assigned_to` (support).

**Mark as Read**
- `POST /api/v1/conversations/:id/read`
- Cập nhật `last_read_at` trong bảng `thread_participants`.

---

## 🚀 Improvements for Module Completion

1.  **Real-time Updates (Socket/SSE)**
    - Cần cơ chế push notification khi có tin nhắn mới.
    - Tích hợp Module Notification hoặc dùng Socket.IO riêng.

2.  **Rich Media (Attachments)**
    - API upload file/image (Tích hợp R2 Service).
    - Message type `image`, `file`.

3.  **Typing Indicators**
    - Ephemeral events qua Socket.

4.  **Reaction/Emoji**
    - Thêm bảng `message_reactions`.

5.  **Search**
    - Full-text search cho tin nhắn (PostgreSQL TSVECTOR).
