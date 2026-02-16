# PLAN: Database Schema Refactor (NaaS)

## 0. Socratic Gate (Clarification)
- **Retention Policy**: Tôi dự định triển khai bằng một PostgreSQL Function và Trigger hoặc gợi ý cấu hình `pg_cron`. Bạn có ưu tiên phương pháp nào không? (Mặc định tôi sẽ viết Function).
- **Data Migration**: Kế hoạch này có yêu cầu migrate dữ liệu `configuration` cũ từ `providers` sang `provider_configs` không, hay chỉ cần cập nhật schema cho hệ thống mới?

## 1. Task Breakdown

### Phase 1: IAM & Core Updates
- Cập nhật bảng `pricing_plans` với cột `features`.

### Phase 2: Providers Refactoring
- Tạo bảng `provider_configs`.
- Cập nhật bảng `providers` (xóa cột cũ).
- Cập nhật bảng `workflow_steps` tham chiếu sang bảng config mới.

### Phase 3: Webhook Logging & Maintenance
- Tạo bảng `webhook_logs`.
- Viết function/trigger cho Retention Policy (90 ngày).

### Phase 4: Security & RLS
- Kích hoạt RLS cho các bảng mới.
- Thêm chính sách isolation dựa trên `tenant_id`.

## 2. Verification Checklist
- [ ] Chạy thành công file `schema.sql`.
- [ ] Ràng buộc FK hoạt động đúng.
- [ ] RLS ngăn chặn truy cập chéo tenant.
- [ ] Retention function xóa đúng dữ liệu cũ.
