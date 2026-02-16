# PLAN: Complete IAM Module Features

## 🎯 Goal
Hoàn thiện các tính năng "còn thiếu" (gaps) trong module IAM để chuyển trạng thái từ ⚠️ sang ✅ trong Audit Report.

## 📋 Scope

### 1. User Info (`/users/me`)
- **Vấn đề**: Hiện tại chỉ trả về `{id: userID}` giả lập.
- **Giải pháp**:
    - Thêm phương thức `GetUser` vào `AuthService` (hoặc `UserService`).
    - Query DB lấy full thông tin user + role/permissions.
    - Trả về DTO `UserResponse` đầy đủ.

### 2. Role Management (`/roles/{id}/assign` & `revoke`)
- **Vấn đề**: Controller đang trả về `NotImplemented`.
- **Giải pháp**:
    - Implement `AssignRole` trong `RoleService`: Thêm bản ghi vào bảng `tenant_members` (cập nhật `role_id`) hoặc bảng phụ trợ nếu design cho phép đa role (hiện tại schema `tenant_members` có 1 `role_id`).
    - *Lưu ý*: Kiểm tra xem user có thuộc tenant đó không trước khi gán.

### 3. Tenant Management (`Update`, `Delete`)
- **Vấn đề**: Controller đang trả về `NotImplemented`.
- **Giải pháp**:
    - Implement `UpdateTenant` trong `TenantService` (đổi tên, slug).
    - Implement `DeleteTenant`: Soft delete (set `deleted_at`) hoặc Hard delete tùy policy.

## 🚀 Execution Steps

1.  **Phase 1: User Info**
    - [ ] Update `UserService` interface & implementation.
    - [ ] Update `AuthController.Me` handler.

2.  **Phase 2: Role Assignment**
    - [ ] Update `RoleService` interface & implementation.
    - [ ] Update `RoleController.AssignRole` & `RevokeRole`.

3.  **Phase 3: Tenant Operations**
    - [ ] Update `TenantService` interface & implementation.
    - [ ] Update `TenantController.UpdateTenant` & `DeleteTenant`.

4.  **Phase 4: Verification**
    - [ ] Test gọi API `/me` ra full info.
    - [ ] Test update tenant info.

## ❓ Questions (Socratic Gate)
- **Q1**: Logic `AssignRole` sẽ thay đổi vai trò chính của member trong tenant, hay chúng ta hỗ trợ assign nhiều role phụ? (Hiện tại schema `tenant_members` chỉ có 1 `role_id` -> Thay đổi role chính).
- **Q2**: Khi `DeleteTenant`, có cần xóa hết Apps và Members liên quan không (Cascade)? -> Thông thường là **Soft Delete** hoặc Cascade Delete cứng.

*Mặc định: Assume Single Role per Member và Soft Delete Tenant.*
