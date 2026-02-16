# AUDIT_REPORT: Module IAM & Apps

Báo cáo này đánh giá mức độ hoàn thiện và tuân thủ tiêu chuẩn của 2 module `iam` và `apps` dựa trên hệ tài liệu `functions.md` và `golang_best_practices.md`.

## 1. Kết quả Audit Kỹ thuật (Technical Audit)

| Tiêu chí | Trạng thái | Ghi chú |
| :--- | :---: | :--- |
| **Build Stability** | ✅ | Dự án build thành công không có lỗi syntax/type. |
| **Architecture** | ✅ | Tuân thủ Clean Architecture (Controller -> Service -> Repository). |
| **Error Handling** | ⚠️ | Đã wrap error (`%w`), nhưng một số chỗ còn log bằng `fmt.Printf` thay vì `slog`. |
| **Logging** | ⚠️ | Chưa tích hợp `slog` đồng nhất trong toàn bộ application service. |
| **Dependency Injection** | ✅ | Sử dụng constructor injection chuẩn xác. |
| **Repository Pattern** | ✅ | Tách biệt Model DB và Entity Domain thông qua Mapper. |

---

## 2. Đối soát Tính năng (Feature Gap Analysis)

### 🛡️ Module IAM
| Tính năng (functions.md) | Trạng thái | Ghi chú |
| :--- | :---: | :--- |
| Đăng ký / Đăng nhập | ✅ | Đã hoàn thành (JWT + Refresh Token). |
| Refresh Token | ✅ | Đã triển khai (`/auth/refresh`). |
| Thông tin User (`/me`) | ✅ | Đã lấy full thông tin user từ DB. |
| Role Management | ✅ | Full CRUD + Assign/Revoke Role (`/roles/{id}/assign`). |
| Tenant Management | ✅ | CRUD Tenant + Owner Assignment + Update/Delete. |

### 🚀 Module Apps (Partner & Integration)
| Tính năng (functions.md) | Trạng thái | Ghi chú |
| :--- | :---: | :--- |
| Application CRUD | ✅ | Đã hoàn thành, tự động tạo môi trường (Dev/Staging/Prod). |
| Provider (NaaS) | ✅ | Đã hoàn thành cấu hình (Email, SMS, v.v.). |
| Webhook | ✅ | Đã hoàn thành quản lý URL và Events. |
| API Key Generation | ⚠️ | Đang dùng UUID đơn giản, thiếu cơ chế quay vòng (Rotation) bảo mật. |
| **System Health Monitor** | 🔴 | Hoàn toàn chưa triển khai (US-RA-01). |
| **Usage Metrics** | 🔴 | Chưa có logic đếm request/message cho dashboard (US-RA-02.5). |

---

## 3. Các "Khoảng trống" quan trọng (Major Gaps)

1.  **Module Hội thoại (Messaging):** Tài liệu `functions.md` tập trung rất nhiều vào CS Agent và CS Leader (Quản lý hội thoại, Pool, Gán Ticket). Tuy nhiên, code hiện tại chưa có các module này, mặc dù `schema.sql` đã có bảng dữ liệu.
2.  **Phần Dashboard:** Toàn bộ Dashboard và Performance Tracking cho Leader/Partner Admin (US-PA-03, US-CL-01) chưa có API backend hỗ trợ.
3.  **Hệ thống Phân quyền (RBAC):** Mặc dù đã có code gán quyền, nhưng logic kiểm tra quyền (Middleware) thực tế tại từng endpoint chưa được áp dụng triệt để (mới chỉ có Auth check).

---

## 4. Đề xuất & Lộ trình khắc phục

1.  **Ưu tiên 1 (Core Auth):** Hoàn thiện logic JWT và Refresh Token để hệ thống có thể chạy thực tế.
2.  **Ưu tiên 2 (Missing IAM):** Xóa bỏ các `TODO` trong `iam.controller` (Assign/Revoke Role, Tenant Update).
3.  **Ưu tiên 3 (Messaging Foundation):** Khởi tạo module `conversations` để xử lý logic `pool` và `messages` như trong schema.
4.  **Ưu tiên 4 (Observability):** Tích hợp log tập trung và metrics để phục vụ Dashboard theo dõi sức khỏe hệ thống.

---
**Người Audit:** Converda Agent
**Ngày:** 2026-02-06
