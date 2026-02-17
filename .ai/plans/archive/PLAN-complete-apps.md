# PLAN: Complete Apps Module

## 🎯 Goal
Hoàn thiện Module Apps đạt tiêu chuẩn "Production-Ready" bằng cách bổ sung bảo mật (API Key) và khả năng giám sát (Monitor/Metrics).

## 📋 Scope

### 1. API Key Security (Rotation)
- **Hiện tại**: Tạo khóa đơn giản, chưa có cơ chế xoay vòng.
- **Yêu cầu**: 
    - Implement `RotateKey(oldKeyID)`: Hủy key cũ (set `expires_at` = now) -> Tạo key mới.
    - Đảm bảo client có khoảng thời gian "grace period" (nếu cần) hoặc đổi ngay lập tức.

### 2. System Health Monitor (US-RA-01)
- **Yêu cầu**: Endpoint `/system/health` checking:
    - Database Connection (Ping).
    - Redis Connection (Ping).
    - Disk Space (Optional).
- **Output**: JSON Status `{ "status": "healthy", "components": { "db": "up", "redis": "up" } }`.

### 3. Usage Metrics (US-RA-02.5)
- **Yêu cầu**: Đếm số lượng request/API call của mỗi App.
- **Giải pháp**:
    - **Middleware**: `MetricsMiddleware` chặn các request có API Key.
    - **Storage**: Tăng counter trong Redis (Real-time) hoặc Log vào bảng `usage_logs` (Persistent).
    - **Dashboard**: API `GetMetrics(appID, timeRange)`.

## 🚀 Execution Steps

1.  **Phase 1: API Key Rotation**
    - [ ] Update `APIKeyService`: Thêm method `RotateKey`.
    - [ ] Update `AppController`: Add POST `/apps/{id}/api-keys/rotate`.

2.  **Phase 2: Health Monitor**
    - [ ] Create `HealthController` in `internal/apps/controller` (or shared `internal/system`).
    - [ ] Register route `/health`.

3.  **Phase 3: Metrics System**
    - [ ] Design DB Schema `usage_metrics` (Time-series data: `app_id`, `endpoint`, `count`, `timestamp`).
    - [ ] Create `MetricsMiddleware`.
    - [ ] Implement `MetricsService`.

## ❓ Questions (Socratic Gate)
- **Q1**: Metrics nên lưu vào DB chính (Postgres) hay cần Redis -> InfluxDB? 
    *   *Giả định*: Lưu vào Postgres (table partitioned by day) cho đơn giản giai đoạn này.
- **Q2**: Health Check có cần public không? -> Thường chỉ cho Internal hoặc Monitoring tool.

*Mặc định: Postgres Storage cho Metrics & Public Health Endpoint (đơn giản).*
