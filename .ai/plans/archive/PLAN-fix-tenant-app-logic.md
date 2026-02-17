# Kế hoạch sửa lỗi logic khởi tạo Tenant/App và Init Role data

## 🎯 Mục tiêu
Đảm bảo khi khởi tạo Tenant và App, dữ liệu được gán đúng quyền (Owner), môi trường (Environment) được tạo tự động thành công, và hệ thống Role được thiết lập sẵn sàng.

## 📋 Vấn đề hiện tại
1.  **Environment missing**: Logic `CreateApp` gọi `NewEnvironment` nhưng không gán `api_key`, trong khi DB yêu cầu `NOT NULL`.
2.  **Owner Not Assigned**: `CreateTenant` gán role `tenant_admin` nhưng table `roles` đang trống -> User không có quyền gì.
3.  **Role Data**: Chưa có dữ liệu khởi tạo cho bảng `roles`.

## 🏗️ Giải pháp đề xuất

### 1. Database (Migration)
- **Action**: Tạo migration mới để:
  - Seed dữ liệu vào bảng `roles` (tenant_admin, tenant_member).
  - Cập nhật cột `environments.api_key` thành `NULLABLE` (vì Go entity đã đánh dấu deprecated và chuyển sang dùng table `app_api_keys`).
  - Gán Permission cơ bản cho các Role.

### 2. IAM Module Fixes
- **Target**: `internal/iam/application/service/impl/tenant.service.impl.go`
- **Action**: Đảm bảo logic `CreateTenant` tìm đúng Role `tenant_admin` vừa seed. Thêm check log nếu không tìm thấy role.

### 3. Apps Module Fixes
- **Target**: `internal/apps/application/service/impl/app.service.impl.go`
- **Action**: Kiểm tra lại flow tạo Environment. Nếu DB đã allow NULL `api_key`, logic cũ sẽ chạy tốt.
- **Target**: `internal/apps/domain/model/entity/apps.go`
- **Action**: Cập nhật `NewEnvironment` để có thể nhận hoặc tự tạo 1 API Key tạm thời nếu vẫn cần thiết cho tương thích ngược.

### 4. Role Initialization (Dựa trên tài liệu)
- **Roles**:
  - `tenant_admin`: Quyền quản trị toàn bộ Tenant (Iam, Apps, Billing).
  - `tenant_member`: Quyền truy cập các tính năng cơ bản của Apps.
- **Permissions**: Danh sách permissions dạng `iam.*`, `apps.*`, `system.*`.

## 🚀 Các bước thực thi

1.  **Step 1**: Tạo migration `00016_seed_roles_and_refactor_env.sql`.
2.  **Step 2**: Sửa logic `NewEnvironment` và `CreateApp` để đồng bộ.
3.  **Step 3**: Kiểm tra lại `CreateTenant` để đảm bảo gán role Admin thành công.
4.  **Step 4**: Chạy integration test kiểm tra `POST /tenants` và `POST /tenants/:id/apps`.

## ❓ Câu hỏi cần làm rõ (Socratic Gate)
1. Bạn muốn Permission trong bảng `roles` là một JSON object phức tạp (ví dụ: `{"apps.create": true}`) hay list string? (Hiện tại struct Role là `map[string]interface{}`).
2. Ngoài `tenant_admin` và `tenant_member`, bạn có muốn thêm Role `owner` riêng biệt không? (Thường `owner` có toàn quyền hơn `admin` ở mốc xoá Tenant).
