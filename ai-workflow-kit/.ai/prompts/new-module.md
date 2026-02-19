# Prompt Template: Tạo module DDD mới

## Cách dùng
1. Copy toàn bộ phần giữa `---START---` và `---END---`
2. Điền thông tin vào các mục `[...]`
3. Paste vào chat với AI
4. Chờ AI hỏi clarifying questions (nếu AI không hỏi, nhắc: "Hỏi clarifying questions trước theo /create workflow")
5. Review implementation_plan.md do AI tạo ra → approve trước khi AI code

---START---
/create module [tên module, ví dụ: notification]

**Mô tả:** [1-2 câu về chức năng chính của module]

**Entities chính:**
- [Entity 1]: [các trường quan trọng, ví dụ: id, user_id, type, status, payload]
- [Entity 2 nếu có]: [...]

**Use cases (operations):**
1. [Use case 1, ví dụ: Tạo và queue notification mới]
2. [Use case 2]
3. [Use case 3]

**External services:**
- [SMTP / FCM / Twilio / Redis / None — ghi rõ]

**Cần dữ liệu từ module khác:**
- Từ module `[tên]`: cần [loại dữ liệu, ví dụ: tên và email của user]
- [Hoặc: Không cần dữ liệu từ module khác]

**API endpoints cần tạo:**
- `POST /api/v1/[path]` — [mô tả]
- `GET  /api/v1/[path]` — [mô tả]
- [Thêm endpoints nếu có]

**Yêu cầu đặc biệt:**
- [Performance: high throughput / low latency]
- [Consistency: idempotent / at-least-once / exactly-once]
- [Security: PII / sensitive data handling]
- [Hoặc: Không có yêu cầu đặc biệt]

---END---

## Ví dụ đã điền

---START---
/create module notification

**Mô tả:** Gửi thông báo đa kênh (email, in-app) khi có sự kiện trong hệ thống.

**Entities chính:**
- Notification: id, user_id, type, channel, status (pending/sent/failed), payload (jsonb), sent_at
- DeliveryLog: id, notification_id, attempt_count, result, error_message

**Use cases:**
1. Tạo notification và đẩy vào queue gửi async
2. Liệt kê notifications của user hiện tại (có pagination)
3. Đánh dấu notification đã đọc

**External services:**
- SMTP (Mailgun) cho email channel

**Cần dữ liệu từ module khác:**
- Từ module `user`: cần id, name, email của người nhận notification

**API endpoints cần tạo:**
- `POST /api/v1/notifications` — tạo và queue notification
- `GET  /api/v1/notifications` — list notifications của current user (với pagination)
- `PATCH /api/v1/notifications/:id/read` — đánh dấu đã đọc

**Yêu cầu đặc biệt:**
- Delivery phải async (không block request chính)
- Retry tối đa 3 lần với exponential backoff nếu gửi thất bại
- Không được gửi duplicate notification cho cùng event
---END---
