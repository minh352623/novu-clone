# PLAN: System Environment Refactor & Hardcoded Data Cleanup

## 🎯 Goal
Loại bỏ việc gán cứng (hardcoding) dữ liệu trong mã nguồn, đặc biệt là danh sách môi trường mặc định, và cung cấp API để Frontend có thể lấy danh sách môi trường hệ thống.

## 📋 Vấn đề hiện tại
1. `AppService.CreateApp` gán cứng `[]string{"development", "staging", "production"}`.
2. Chưa có API để UI hiển thị danh sách `system_environments`.
3. Một số giá trị như `SHARED_SECRET_KEY` trong `guards.go` đang bị gán cứng.

## 🏗️ Giải pháp đề xuất

### 1. System Environments Module
- **Repository**: Tạo `SystemEnvironmentRepository` để query bảng `system_environments`.
- **Service**: Tạo `SystemEnvironmentService` cung cấp method `ListAll`.
- **Controller**: Thêm `GET /api/v1/system/environments` (Public hoặc Authenticated tùy domain).
- **Refactor AppService**: Inject `SystemEnvironmentService` vào `AppService` và gọi `ListAll()` để lấy danh sách cần init thay vì dùng array gán cứng.

### 2. Hardcoded Cleanup
- **Guards**: Chuyển `SHARED_SECRET_KEY` sang sử dụng `global.Config`.
- **Constants**: Gom các giá trị định danh (slugs) vào một file constant trung tâm nếu cần.

## 🚀 Các bước thực thi

1.  **Step 1: System Environment Core**
    - Tạo model/entity cho `SystemEnvironment` (Read-only).
    - Implement Repository & Service.
2.  **Step 2: API & Integration**
    - Đăng ký route: `GET /api/v1/system/environments` (Yêu cầu Authentication).
    - Refactor `CreateApp` logic: Inject `SystemEnvironmentService` để lấy list code.
3.  **Step 3: Verification**
    - Test API với token.
    - Test tạo App xem có đủ môi trường từ DB không.

## ❓ Câu hỏi đã làm rõ (User Feedback)
1. **Auth**: API dành cho user đã login.
2. **Permission**: Chỉ cần Read-only (Frontend show data).
3. **Security Refactor**: Tạm thời bỏ qua `SHARED_SECRET_KEY`.
