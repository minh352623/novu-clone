# Prompt Template: Thêm API endpoint mới vào module đã có

## Cách dùng
Copy phần giữa `---START---` và `---END---`, điền thông tin, paste vào chat.

---START---
/create endpoint

**Module:** `internal/[tên module, ví dụ: order]/`
**Method + Path:** `[GET/POST/PUT/PATCH/DELETE] /api/v1/[path]`
**Mô tả:** [Endpoint này làm gì?]

**Request:**
- Body (JSON): [mô tả các fields, hoặc "Không có body"]
  - `[field_name]` ([kiểu dữ liệu], [required/optional]): [mô tả]
- Query params: [mô tả, hoặc "Không có"]
- Path params: `/:id` ([kiểu dữ liệu]) [hoặc không có]

**Response thành công:**
- Status: [200/201/204]
- Body: [mô tả dữ liệu trả về, hoặc "Không có body (204)"]

**Business rules:**
1. [Rule 1, ví dụ: Chỉ owner của resource mới được phép xóa]
2. [Rule 2]

**Có cần transaction không?**
- [Có — ghi rõ các bảng nào được write trong cùng transaction]
- [Không]

**Auth:**
- [Yêu cầu đăng nhập (JWT required)]
- [Public — không cần auth]

---END---

## Ví dụ đã điền

---START---
/create endpoint

**Module:** `internal/order/`
**Method + Path:** `PATCH /api/v1/orders/:id/cancel`
**Mô tả:** User hủy đơn hàng của mình.

**Request:**
- Body: Không có body
- Path params: `:id` (UUID string) — ID của order cần hủy

**Response thành công:**
- Status: 200
- Body: `{ "id": "...", "status": "cancelled", "updated_at": "..." }`

**Business rules:**
1. Chỉ owner của order mới được hủy (user_id phải match)
2. Chỉ được hủy order đang ở trạng thái `pending` hoặc `confirmed`
3. Order đã `completed` hoặc đã `cancelled` → trả lỗi conflict

**Transaction:** Không cần — chỉ update 1 bảng.

**Auth:** JWT required — lấy user_id từ token.
---END---
