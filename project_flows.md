# Dự án Converda - Quy trình hệ thống (System Flows)

Tài liệu này mô tả các luồng hoạt động chính của hệ thống bằng sơ đồ Mermaid.

## 1. Tổng quan kiến trúc (System Architecture Overview)

Hệ thống được thiết kế theo mô hình Multi-tenant (Đa người thuê), phân tách dữ liệu theo Tenant và Môi trường.

```mermaid
graph TD
    subgraph "Clients & External"
        U[End User / Customer]
        PA[Partner Admin / Agent]
        P[External Providers: SMTP, FCM, Twilio]
    end

    subgraph "Converda Platform (NaaS)"
        API[API Gateway / Router]
        IAM[IAM & Tenancy Module]
        WE[Workflow Engine]
        NM[Notification Module]
        MM[Messaging Module]
        DB[(PostgreSQL)]
    end

    U <-->|Chat / Push| API
    PA <-->|Dashboard / API Key| API
    API <--> IAM
    API <--> NM
    API <--> MM
    NM --> WE
    NM --> P
    WE --> MM
    MM --> DB
    NM --> DB
    IAM --> DB
```

---

## 2. Luồng tính năng chi tiết (Feature Flows)

### A. Quản lý định danh và thành viên (IAM & Onboarding)
Quy trình từ lúc tạo Tenant đến khi mời nhân viên CS.

```mermaid
sequenceDiagram
    participant RA as Root Admin
    participant PA as Partner Admin
    participant Sys as System
    participant U as User (Email)

    RA->>Sys: Tạo Partner/Tenant (Slug, Plan)
    Sys->>Sys: Cấp Tenant ID & Default App/Env
    PA->>Sys: Đăng nhập Dashboard
    PA->>Sys: Invite Member (Email, Role: Agent)
    Sys->>U: Gửi Email SSO Activation
    U->>Sys: Accept Invite (SSO Login)
    Sys->>Sys: Gán Member vào Tenant & Role
```

---

### B. Luồng nhắn tin và điều phối (Messaging & Helpdesk)
Cách tin nhắn từ khách hàng được đưa vào Pool và giao cho Agent.

```mermaid
flowchart LR
    Start([Khách gửi tin nhắn]) --> API[API tiếp nhận]
    API --> Store[Lưu vào bảng Messages]
    Store --> CheckPool{Đã có Pool?}
    
    CheckPool -- No --> CreatePool[Tạo Pool mới: unassigned]
    CheckPool -- Yes --> UpdatePool[Cập nhật Pool: last_msg_at]
    
    CreatePool --> Routing[Luồng điều phối]
    UpdatePool --> Routing
    
    Routing --> Assign{Có tự động gán?}
    Assign -- Yes --> Assigned[Trạng thái: assigned]
    Assign -- No --> InPool[Trạng thái: in_pool]
    
    InPool --> Leader[CS Leader gán thủ công]
    Leader --> Assigned
    Assigned --> WS[Thông báo Agent qua WebSocket]
```

---

### C. Pipeline thông báo (Notification Pipeline - CMS Style)
Quy trình xử lý một thông báo từ Trigger đến lúc gửi đi, tích hợp layout và đa ngôn ngữ.

```mermaid
sequenceDiagram
    participant App as External App
    participant API as Notification API
    participant CMS as CMS Engine
    participant DB as Database
    participant Pr as Provider (FCM/SMTP)

    App->>API: Trigger Send (template_code, subscriber_key, variables)
    API->>DB: Load Template & Active Version
    DB-->>API: Template Metadata
    API->>DB: Load Content (Language, Version)
    API->>DB: Load Layout (Header/Footer)
    API->>CMS: Compile (Layout + Template Content + Variables)
    CMS-->>API: Final HTML/Text/JSON
    API->>DB: Get Provider Config for Env
    API->>Pr: Dispatch Dispatch Order
    Pr-->>API: Response (Success/Fail)
    API->>DB: Log Result (notification_logs)
```

---

### D. Quy trình thực thi Workflow (Workflow Engine)
Luồng tự động hóa dựa trên các bước (Steps).

```mermaid
graph TD
    T[Trigger Event] --> W[Fetch Active Workflow]
    W --> S1[Thực thi Step 1]
    S1 --> Type{Loại Step?}
    
    Type -- "Channel (Gửi tin)" --> NT[Call Notification Pipeline]
    Type -- "Delay (Chờ)" --> Wait[Hệ thống chờ X phút/giờ]
    Type -- "Digest (Gộp)" --> Collect[Thu thập tin nhắn]
    
    NT --> Next{Hết Steps?}
    Wait --> Next
    Collect --> Next
    
    Next -- No --> S2[Thực thi Step kế tiếp]
    Next -- Yes --> End([Hoàn thành])
```
