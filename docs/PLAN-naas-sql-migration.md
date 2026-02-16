# PLAN: NaaS SQL Migration

## 0. Socratic Gate (Clarification)
- **Goose Migration**: Dự án đang sử dụng Goose cho việc quản lý migration. Tôi sẽ tạo file `.sql` chuẩn Goose (có kèm `-- +goose Up` và `-- +goose Down`).
- **Data Safety**: Vì bảng `providers` hiện tại có thể đang chứa dữ liệu, tôi sẽ thực hiện migrate dữ liệu từ `providers.configuration` sang `provider_configs` trước khi xóa cột `configuration`. Bạn có đồng ý không?

## 1. Task Breakdown

### Phase 1: Create Migration File
- Tạo file `sql/schema/00012_naas_schema_refactor.sql`.
- Định nghĩa phần `-- +goose Up`:
    - Thêm cột `features` vào `pricing_plans`.
    - Tạo bảng `webhook_logs`.
    - Tạo bảng `provider_configs`.
    - Migrate dữ liệu (nếu cần).
    - Cập nhật `workflow_steps`.
    - Xóa cột `configuration` trong `providers`.
    - Tạo chính sách RLS.
- Định nghĩa phần `-- +goose Down`:
    - Rollback toàn bộ các thay đổi trên.

### Phase 2: Verification
- Kiểm tra cú pháp SQL.
- Kiểm tra thứ tự thực thi.

## 2. Verification Checklist
- [ ] File migration được tạo đúng định dạng.
- [ ] Logic Up/Down đầy đủ và an toàn.
