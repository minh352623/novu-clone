# PLAN: Auto-generate API Key on App & Env Creation

## 🎯 Goal
Tự động tạo một API Key mặc định (Primary Key) cho mỗi môi trường khi một App mới được khởi tạo. Điều này đảm bảo User có thể sử dụng App ngay lập tức mà không cần bước tạo Key thủ công.

## 📋 Proposed Changes

### 1. Apps Module (Service)
- [MODIFY] `internal/apps/application/service/impl/app.service.impl.go`:
    - Inject `APIKeyService` vào `appServiceImpl`.
    - Trong hàm `CreateApp`, sau khi tạo thành công từng `Environment`, gọi `apiKeyService.GenerateKey` với tên mặc định là "Default Key".
    - Log cảnh báo nếu việc tạo Key thất bại nhưng không làm gián đoạn luồng tạo App.

### 2. Initialization
- [MODIFY] `internal/initialize/apps/apps.go`:
    - Thay đổi thứ tự khởi tạo service: Khởi tạo `apiKeyService` trước `appService`.
    - Pass `apiKeyService` vào constructor của `appService`.

## 🚀 Verification Plan
1. Khởi động server.
2. Gọi API `POST /v1/api/tenants/{id}/apps` để tạo App mới.
3. Kiểm tra DB (bảng `app_api_keys`) xem có 3 keys tương ứng với 3 môi trường mặc định (Dev, Stage, Prod) vừa được tạo hay không.
4. Gọi API `GET /v1/api/environments/{env_id}/api-keys` để kiểm tra key hiển thị trên list.
