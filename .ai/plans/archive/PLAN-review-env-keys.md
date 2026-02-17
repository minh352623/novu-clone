# PLAN: Review & Cleanup Environment API Keys

## 🎯 Goal
Phân tích sự khác biệt giữa cột `api_key` trong bảng `environments` và bảng quan hệ `app_api_keys`. Đề xuất kế hoạch chuẩn hóa để tránh nhầm lẫn.

## 📋 Analysis

### 1. `environments.api_key` (Legacy)
- **Kiểu dữ liệu**: `TEXT` (Nullable sau bản vá 00016).
- **Mục đích (Cũ)**: Lưu trữ 1 API Key duy nhất cho mỗi môi trường. Cách tiếp cận đơn giản ban đầu.
- **Hạn chế**:
    - Không hỗ trợ API Key Rotation (xooay key không downtime).
    - Không thể đặt tên (Name) cho key để phân biệt mục đích sử dụng.
    - Không thể set Expiration (thời hạn).
    - Hard schema change nếu muốn thêm metadata.

### 2. `app_api_keys` Table (Standard)
- **Kiểu dữ liệu**: Bảng riêng biệt.
- **Mục đích (Mới)**: Hỗ trợ nhiều API Key cho một môi trường.
- **Ưu điểm**:
    - **Multiple Keys**: Hỗ trợ Rolling Update (Key cũ còn hạn, Key mới đã active).
    - **Security**: Chỉ lưu `KeyHash`, `KeyPrefix`, `KeySuffix` (Không lưu Plain Key). An toàn hơn nhiều so với lưu plaintext ở cột cũ.
    - **Metadata**: Có `Name`, `ExpiresAt`, `RevokedAt`.

## 🏗️ Proposed Changes

### Phase 1: Deprecation (Current)
- [x] Đánh dấu `environments.api_key` là `Deprecated` trong Code (`entity`, `model`, `dto`).
- [x] Chuyển `api_key` sang `*string` (Nullable) để tránh lỗi duplicate constraint khi chưa có key.
- [ ] Logic tạo App mới (đã implement): Auto-generate key vào bảng `app_api_keys` thay vì cột `api_key`.

### Phase 2: Migration (Future)
- **Migration Script**: Copy dữ liệu từ `environments.api_key` sang bảng `app_api_keys` (nếu có dữ liệu legacy).
- **Drop Column**: Xóa hoàn toàn cột `api_key` khỏi bảng `environments` sau khi migrate xong.

## 🚀 Recommendation
User nên **chỉ sử dụng** danh sách `keys` (từ bảng `app_api_keys`) cho mọi logic mới. Cột `api_key` hiện tại chỉ nên giữ lại để tương thích ngược (backward compatibility) nếu có client cũ đang dùng, và sẽ bị xóa trong version major tiếp theo.

### Action Plan
1. **Frontend**: Chỉ hiển thị và quản lý key từ danh sách `Keys`. Ẩn field `api_key` cũ hoặc show warning "Legacy".
2. **Backend**: Giữ nguyên logic Auto-gen key vào bảng `app_api_keys`.
