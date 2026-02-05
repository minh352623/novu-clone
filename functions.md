
---
1. Root Admin: System health, Partner management
2. Partner Admin: Tenant overview, team management
3. CS Leader: Team performance, assignment management
4. CS Agent: Personal workspace, productivity tracking

---
1. Root Admin – Quản trị nền tảng CRM

---
US-RA-01: Theo dõi sức khoẻ hệ thống
Với vai trò là Root Admin
Tôi muốn theo dõi sức khoẻ của toàn bộ hệ thống CRM
Để:
- Kịp thời phát hiện và xử lý các sự cố kỹ thuật
- Giám sát sức khỏe toàn hệ thống đa tenant, phát hiện sớm "thắt cổ chai" (bottleneck) và sự cố kỹ thuật để đảm bảo uptime 99.9%
Tiêu chí chấp nhận (Acceptance Criteria)
- Giả sử Root Admin đã đăng nhập hệ thống
- Khi truy cập màn hình System Health
- Thì hệ thống hiển thị các chỉ số:
  - Trạng thái hoạt động hệ thống (up/down)
  - Tốc độ phản hồi API
  - Tỷ lệ lỗi
  - Tình trạng xử lý tin nhắn (message queue)
- Và dữ liệu hiển thị là toàn bộ hệ thống, có filter theo từng Partner
- AC.US-RA-01.1: System Status Overview
GIVEN Root Admin đăng nhập vào hệ thống
WHEN truy cập "/admin/system-health"
THEN hiển thị dashboard với 4 khu vực chính:

┌─────────────────────────────────────────────┐
│ SYSTEM STATUS OVERVIEW                              │
│ ○ Status: UP (99.95% uptime last 30 days)           │
│ ○ Active Partners: 12/15                            │
│ ○ Total Conversations: 1,247 (today)                │
│ ○ Active CS Agents: 45/60                           │
└─────────────────────────────────────────────┘
- AC.US-RA-01.2: API Performance Metrics
┌─────────────────────────────────────────────┐
│ API PERFORMANCE (Real-time)                         │
├─────────────────────────────────────────────┤
│ Average Response Time: 145ms                        │
│ P95 Response Time: 320ms                            │
│ P99 Response Time: 680ms                            │
│ Requests/min: 1,250                                 │
│                                                     │
│ [Line Chart: Last 24 hours]                         │
│  Response Time Trend                                │
└─────────────────────────────────────────────┘
- AC.US-RA-01.3: Error Rate Monitoring
┌─────────────────────────────────────────────┐
│ ERROR TRACKING                                      │
├─────────────────────────────────────────────┤
│ Overall Error Rate: 0.12%                           │
│                                                     │
│ ⚠️ Top Errors (Last 1 hour):                        │
│ 1. 500 Internal Server - 8 occurrences              │
│ 2. 429 Rate Limit - 5 occurrences                   │
│ 3. 503 Service Unavailable - 2 occ.                 │
│                                                     │
│ [Filter: All Partners ▼] [Last 24h ▼]              │
└─────────────────────────────────────────────┘
- AC.US-RA-01.4: Message Queue Health
┌─────────────────────────────────────────────┐
│ MESSAGE QUEUE STATUS                                │
├─────────────────────────────────────────────┤
│ Pending Messages: 23                                │
│ Processing Rate: 1,200/min                          │
│ Avg Processing Time: 1.2s                           │
│ Failed Messages: 2 (retry in progress)              │
│                                                     │
│ Queue Depth Trend: [Mini Chart]                     │
└─────────────────────────────────────────────┘
- AC.US-RA-01.5: Filter theo Partner
  - WHEN Root Admin chọn filter "Partner: ABC Corp" 
  - THEN tất cả metrics chỉ hiển thị dữ liệu của Partner ABC Corp 
  - AND URL thay đổi thành "/admin/system-health?partner=abc-corp"

---
US-RA-02: Quản lý tích hợp Partner
Với vai trò là Root Admin
Tôi muốn tạo và cấu hình tích hợp với các hệ thống Partner
Để các ứng dụng bên thứ 3 có thể gửi tin nhắn vào CRM
Tiêu chí chấp nhận
- Giả sử Root Admin đang ở màn hình quản lý Partner
- Khi tạo mới một Partner
- Thì hệ thống cho phép cấu hình:
  - Trạng thái tích hợp (bật / tắt)
  - Thông tin kết nối (API key, webhook)
- Và mỗi Partner được tạo một không gian dữ liệu (tenant) riêng biệt
- AC.US-RA-02.1: Danh sách Partners
┌────────────────────────────────────────────────┐
│ PARTNER INTEGRATIONS                [+ Add Partner]     │
├────────────────────────────────────────────────┤
│ Partner Name    │ Status  │ Messages/Day │ Last Active │
├────────────────────────────────────────────────┤
│ ABC Corp        │ ● ON    │ 1,247        │ 2 min ago   │
│ XYZ Ltd         │ ● ON    │ 892          │ 5 min ago   │
│ Test Partner    │ ○ OFF   │ 0            │ 3 days ago  │
└────────────────────────────────────────────────┘
- AC.US-RA-02.2: Form tạo/chỉnh sửa Partner
┌──────────────────────────────────────┐
│ CREATE NEW PARTNER                          │
├──────────────────────────────────────┤
│ Partner Name: [____________]                │
│ Partner ID: [auto-generated]                │
│                                             │
│ Integration Status:                         │
│ ○ Enabled   ○ Disabled                      │
│                                             │
│ Connection Settings:                        │
│ • API Key: [******************] [Regenerate]│
│ • API Secret: [******************]          │
│ • Webhook URL: [_________________________]  │
│                                             │
│ Rate Limiting:                              │
│ • Max Requests/min: [1000]                  │
│ • Max Messages/day: [10000]                 │
│                                             │
│ Tenant Isolation:                           │
│ • Database Schema: partner_abc_corp         │
│ • Storage Bucket: s3://crm-abc-corp/        │
│                                             │
│ [Cancel]  [Save & Create Tenant]            │
└──────────────────────────────────────┘
- AC.US-RA-02.3: API Key Management
  - WHEN Root Admin click "Regenerate" API Key
  - THEN hệ thống hiển thị trình thiết lập API Key gồm các bước:
    - Đặt tên KEY và chọn Environment
    - Chọn Permission Scopes
    - Chọn Expiration Date (Never/Custom)
    - Hiển thị new key ONE TIME only (chỉ hiển thị và sử dụng một lần)
  - AND Partner Admin có thể sử dụng NEW KEY được cung cấp để integrate
- AC.US-RA-02.4: Webhook Configuration
  - Webhook Events: 
    □ message.received 
    □ message.sent 
    □ conversation.assigned 
    □ conversation.closed  
  - Webhook Retry Policy: 
    - Max Retries: [3] 
    - Backoff Strategy: [Exponential ▼] 
    - Timeout: [5000ms]
- AC.US-RA-02.5: Usage Metrics per Partner
┌──────────────────────────────────────┐
│ PARTNER: ABC Corp                           │
├──────────────────────────────────────┤
│ API Usage (Last 30 days):                   │
│ • Total Requests: 125,847                   │
│ • Avg Requests/day: 4,195                   │
│ • Peak Hour: 14:00-15:00 (847 req/h)        │
│                                             │
│ [Line Chart: Daily API Requests]            │
│                                             │
│ Messages Processed:                         │
│ • Inbound: 45,234                           │
│ • Outbound: 38,902                          │
│ • Failed: 127 (0.14%)                       │
└──────────────────────────────────────┘
Quy tắc nghiệp vụ
- Chỉ Root Admin mới có quyền tạo, sửa, xoá tích hợp
- Partner không thể xem hoặc chỉnh cấu hình tích hợp của Partner khác

---
2. Partner Admin – Quản trị tenant & đội CS

---
US-PA-01: Mời tài khoản người dùng qua email SSO
Với vai trò là Partner Admin
Tôi muốn mời thêm người dùng bằng email SSO và gán sẵn vai trò
Để xây dựng và quản lý đội chăm sóc khách hàng
Tiêu chí chấp nhận
- Giả sử Partner Admin đang ở màn hình quản lý người dùng
- Khi gửi lời mời tới một địa chỉ email
- Thì hệ thống gửi email kích hoạt SSO
- Và người dùng được gán vai trò ngay khi chấp nhận lời mời
WHEN click "+ Invite User":
┌──────────────────────────────────────┐
│ INVITE NEW USER                             │
├──────────────────────────────────────┤
│ Email: [________________________]           │
│                                             │
│ Role: ○ CS Leader                           │
│       ○ CS Agent                            │
│       ○ Viewer (read-only)                  │
│                                             │
│ ☑ Send invitation email                     │
│                                             │
│ [Cancel] [Send Invitation]                  │
└──────────────────────────────────────┘
- Email Template:
Subject: You've been invited to join CRM - ABC Corp  

Hi,  

You've been invited to join ABC Corp's customer service team.  

Role: CS Agent 
Invited by: admin@abccorp.com  

Click here to activate your account: 
[Activate Account Button] (expires in 7 days)  

This link uses SSO authentication.

---
US-PA-02: Phân quyền và vai trò người dùng
Với vai trò là Partner Admin
Tôi muốn gán vai trò cho từng người dùng
Để phân tách rõ trách nhiệm trong quá trình vận hành
Tiêu chí chấp nhận
- Giả sử Partner Admin đang quản lý danh sách người dùng
- Khi gán vai trò (CS Leader, CS Agent, chỉ xem)
- Thì quyền truy cập của người dùng được cập nhật ngay lập tức
Quy tắc nghiệp vụ
- Partner Admin chỉ quản lý người dùng trong tenant của mình
- Không có quyền chỉnh sửa Root Admin
┌────────────────────────────────────────────────┐
│ USER MANAGEMENT                      [+ Invite User]    │
├────────────────────────────────────────────────┤
│ Search: [_____________]  Role: [All ▼]  Status: [All ▼]│
├─────────────────────────────────────────────────┤
│ Name          │ Email             │ Role      │ Status  │
├──────── ────┼────────────────┼─────────┼────────┤
│ Nguyen Van A  │ a@company.com     │ CS Leader │ Active  │
│ Tran Thi B    │ b@company.com     │ CS Agent  │ Active  │
│ Le Van C      │ c@company.com     │ Viewer    │ Pending │
└─────────────────────────────────────────────────┘


---
US-PA-03: Xem dashboard và báo cáo tổng quát
Với vai trò là Partner Admin
Tôi muốn xem dashboard và báo cáo tổng thể
Để nhìn tổng quan về hoạt động chăm sóc khách hàng trong tenant, đánh giá hiệu quả team và xu hướng
Tiêu chí chấp nhận
- Khi truy cập dashboard
- Thì chỉ hiển thị dữ liệu thuộc tenant của Partner đó
- Và cho phép lọc theo thời gian, trạng thái hội thoại, đội CS
- AC.US-PA-03.1: Key Metrics Summary
┌──────────────────────────────────────────────────────┐
│ DASHBOARD OVERVIEW - ABC Corp                                  │
│ [Last 7 days ▼] [Export PDF]                                  │
├──────────────────────────────────────────────────────┤
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐    │
│ │ Total Conv.     │ │ Avg Response   │ │ SLA Met         │    │
│ │    1,247        │ │    12 min      │ │   94.5%         │    │
│ │ ↑ 12% vs LW     │ │ ↓ 8% vs LW     │ │ ↓ 2% vs LW      │    │
│ └──────────────┘ └──────────────┘ └──────────────┘    │
│                                                                │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐    │
│ │ Closed Conv.    │ │ In Pool        │ │ Active Agents.  │    │
│ │    1,089        │ │      23        │ │      15         │    │
│ │ 87.3% rate      │ │ ↑ 5 vs LW      │ │ 15/18 total.    │    │
│ └──────────────┘ └──────────────┘ └──────────────┘    │
└──────────────────────────────────────────────────────┘
- AC.US-PA-03.2: Conversation Volume Trend
┌──────────────────────────────────────┐
│ CONVERSATION VOLUME                         │
├──────────────────────────────────────┤
│ [Bar + Line Chart: Last 30 days]            │
│                                             │
│ ▆ New Conversations                         │
│ ━ Resolved Conversations                    │
│                                             │
│ [Zoom: 7D | 30D | 90D]                      │
└──────────────────────────────────────┘
- AC.US-PA-03.3: Response Time Distribution
┌──────────────────────────────────────┐
│ RESPONSE TIME BREAKDOWN                     │
├──────────────────────────────────────┤
│ < 5 min:   ████████████ 45% (562 conv.)   │
│ 5-15 min:  ██████████ 38% (474 conv.)     │
│ 15-30 min: ████ 12% (150 conv.)            │
│ > 30 min:  ██ 5% (61 conv.)                │
│                                             │
│ Target SLA: First response < 15 min         │
└──────────────────────────────────────┘
- AC.US-PA-03.4: Team Performance Overview
┌──────────────────────────────────────┐
│ CS TEAM PERFORMANCE                         │
├──────────────────────────────────────┤
│ Agent Name    │ Assigned │ Closed │ Avg RT │
├─────────────┼────────┼───────┤───────┤
│ Nguyen Van A  │    87    │   82   │ 11m    │
│ Tran Thi B    │    92    │   89   │ 9m     │
│ Le Van C      │    65    │   58   │ 14m    │
│ ...                                         │
│ [View Full Report →]                       │
└──────────────────────────────────────┘
- AC.US-PA-03.5: Filters và Segmentation
  - Time Range: [Last 7 days ▼]
    - Today
    - Last 7 days
    - Last 30 days
    - Custom range: [From /] [To /]
  - Conversation Status: [All ▼]
    - Open
    - Assigned
    - In Pool
    - Closed
  - Priority: [All ▼]
    - High
    - Medium
    - Low

---
3. CS Leader – Điều phối & quản lý đội CS

---
US-CL-01: Xem toàn bộ hội thoại của Partner
Với vai trò là CS Leader
Tôi muốn xem tất cả hội thoại thuộc Partner của mình
Để giám sát hoạt động chăm sóc khách hàng
Tiêu chí chấp nhận
- Giả sử CS Leader đăng nhập hệ thống
- Khi mở danh sách hội thoại
- Thì hiển thị:
  - Hội thoại đã assign
  - Hội thoại trong pool
  - Hội thoại đã đóng

---
US-CL-02: Assign và reassign hội thoại cho CS Agent
Với vai trò là CS Leader
Tôi muốn gán hoặc điều chỉnh hội thoại cho CS Agent
Để phân bổ công việc hợp lý
Tiêu chí chấp nhận
- Giả sử hội thoại chưa ở trạng thái Closed
- Khi CS Leader gán CS Agent
- Thì hội thoại xuất hiện trong inbox của CS Agent đó
- Và thao tác được ghi log để audit
Quy tắc nghiệp vụ
- CS Leader có quyền override việc assign
- CS Agent không có quyền assign cho người khác
┌────────────────────────────────────────────────────┐
│ CONVERSATION MANAGEMENT                                     │
├────────────────────────────────────────────────────┤
│ Filters: [All Status ▼] [All Agents ▼] [All Channels ▼]   │
├────────────────────────────────────────────────────┤
│ □ Customer    │ Channel │ Status   │ Assigned To │ Last Msg │
├─────────────┼───────┼─────────┼───────────┼────────┤
│ □ Khách A     │ Zalo    │ 🟢 Open  │ Agent A     │ 2 min ago│
│ □ Khách B     │ Website │ 🟡 Pool  │ Unassigned  │ 5 min ago│
│ □ Khách C     │ FB      │ 🟢 Open  │ Agent B     │ 1 min ago│
│ □ Khách D     │ Zalo    │ ⚪ Closed│ Agent A     │ 1 hr ago │
├─────────────────────────────────────────────────────┤
│ ☑ Select all | [Bulk Assign ▼]                              │
└─────────────────────────────────────────────────────┘
- WHEN CS Leader selects conversations và click "Bulk Assign":
  - THEN Dropdown danh sách Agents
  - AND Confirm assignment
  - AND Log audit trail

---
US-CL-03: Xem dashboard hiệu suất đội CS
Với vai trò là CS Leader
Tôi muốn xem dashboard hiệu suất của đội CS
Để đánh giá năng suất và SLA
Tiêu chí chấp nhận
- Khi truy cập dashboard đội
- Thì hiển thị:
  - Số hội thoại theo từng CS Agent
  - Thời gian phản hồi trung bình
  - Tỷ lệ đạt SLA
  - Số hội thoại trả về pool
- AC.US-CL-03.1: Team Overview Metrics
┌────────────────────────────────────────────────┐
│ TEAM PERFORMANCE                                        │
│ [Today ▼] [This Week] [This Month]                     │
├────────────────────────────────────────────────┤
│ Team Metrics:                                           │
│ • Total Agents: 8 (7 active, 1 away)                    │
│ • Conversations Handled: 487 (today)                    │
│ • Avg First Response: 11.5 min                          │
│ • SLA Compliance: 92.8%                                 │
│ • Conversations in Pool: 12 (↑ 3 vs yesterday)          │
└────────────────────────────────────────────────┘
- AC.US-CL-03.2: Agent Performance Table
┌─────────────────────────────────────────────────────────┐
│ AGENT PERFORMANCE DETAILS                                         │
├─────────────────────────────────────────────────────────┤
│ Agent      │Status│Assigned│Closed│Avg RT│SLA%│Pool Returns│Load│
├──────────┼─────┼───────┼─────┼─────┼────┼──────────┼────┤
│ Agent A    │ 🟢   │   25   │  23  │ 9m   │96% │     1      │ ███ │
│ Agent B    │ 🟢   │   28   │  26  │ 12m  │89% │     2      │████ │
│ Agent C    │ 🟡   │   15   │  14  │ 8m   │98% │     0      │ ██  │
│ Agent D    │ 🔴   │   32   │  28  │ 18m  │78% │     4      │█████│
│ Agent E    │ 🟢   │   22   │  20  │ 10m  │95% │     1      │███  │
├──────────┴─────┴───────┴─────┴─────┴────┴──────────┴─────┤
│ Legend: 🟢 Available | 🟡 Busy | 🔴 Overloaded | ⚪ Away           │
└──────────────────────────────────────────────────────────┘
  - WHEN click vào Agent: 
  - THEN Hiển thị detail panel bên phải:   
    - List conversations assigned   
    - Activity timeline   
    - Performance trend (7 days)
- AC.US-CL-03.3: Workload Distribution Chart
┌───────────────────────────────────────┐
│ WORKLOAD DISTRIBUTION                        │
├───────────────────────────────────────┤
│ [Stacked Bar Chart]                          │
│                                              │
│ Agent A  ██████████ 25 (10 open, 15 closed)│
│ Agent B  ████████████ 28 (12 open, 16...)  │
│ Agent C  ██████ 15 (5 open, 10 closed)      │
│ ...                                          │
│                                              │
│ Recommended Action:                          │
│ ⚠️ Agent D is overloaded (32 conv.)          │
│    [Reassign 5 conversations]                │
└───────────────────────────────────────┘
- AC.US-CL-03.5: Pool Management View
┌───────────────────────────────────────┐
│ UNASSIGNED CONVERSATIONS (Pool)              │
├───────────────────────────────────────┤
│ Customer      │ Channel │ Waiting │ Priority│
├─────────────┼───────┼────────┼────────┤
│ Khách 001     │ Zalo    │  5 min  │  High   │
│ Khách 002     │ Website │  2 min  │  Medium │
│ Khách 003     │ FB Msg  │ 12 min  │  High   │
│                                              │
│ [Quick Assign to Available Agent]            │
└───────────────────────────────────────┘
- WHEN CS Leader click "Quick Assign"
  - THEN Hệ thống suggest Agent có workload thấp nhất 
  - AND CS Leader confirm 
  - OR chọn Agent khác
Workload Calculation
  - Agent Load = (Open Conversations * 2) + (Assigned but not replied * 1.5) 
  - Load Status:   
    - Normal: Load < 20   
    - Busy: 20 ≤ Load < 30   
    - Overloaded: Load ≥ 30

---
4. CS Agent – Thực hiện chăm sóc khách hàng

---
US-CA-01: Nhận hội thoại được assign
Với vai trò là CS Agent
Tôi muốn nhận các hội thoại được giao
Để xử lý yêu cầu của khách hàng
Tiêu chí chấp nhận
- Khi hội thoại được assign cho CS Agent
- Thì hội thoại xuất hiện trong danh sách của CS Agent đó

---
US-CA-02: Trả lời tin nhắn khách hàng
Với vai trò là CS Agent
Tôi muốn trả lời tin nhắn của khách hàng
Để giải quyết vấn đề của họ
Tiêu chí chấp nhận
- Giả sử hội thoại đang ở trạng thái Open
- Khi CS Agent gửi tin nhắn
- Thì tin nhắn được gửi về ứng dụng client
- Và trạng thái hội thoại được cập nhật

---
US-CA-03: Ghi chú nội bộ (Internal Note)
Với vai trò là CS Agent
Tôi muốn ghi chú nội bộ cho hội thoại
Để lưu lại thông tin xử lý mà khách hàng không nhìn thấy
Tiêu chí chấp nhận
- Giả sử hội thoại tồn tại
- Khi CS Agent thêm internal note
- Thì ghi chú chỉ hiển thị cho nội bộ
Quy tắc nghiệp vụ
- Không bắt buộc phải trả lời khách hàng
- Không bắt buộc phải đóng hội thoại
- Không gửi thông báo cho khách hàng

---
US-CA-04: Trả hội thoại về pool
Với vai trò là CS Agent
Tôi muốn chuyển hội thoại về pool
Để CS Leader gán cho người phù hợp hơn
Tiêu chí chấp nhận
- Giả sử CS Agent không xử lý được hội thoại
- Khi chuyển hội thoại về pool
- Thì hội thoại không còn trong danh sách của CS Agent
- Và thay đổi trạng thái được ghi log

---
US-CA-05: Đóng hội thoại
Với vai trò là CS Agent
Tôi muốn đóng hội thoại đã xử lý xong
Để hoàn tất nghiệp vụ chăm sóc khách hàng
Tiêu chí chấp nhận
- Giả sử hội thoại đã có ít nhất một hành động từ CS Agent
- Khi CS Agent đóng hội thoại
- Thì trạng thái chuyển sang Closed
- Và khách hàng không thể gửi thêm tin nhắn

---
US-CA-06: Xem dashboard hiệu suất cá nhân
Với vai trò là CS Agent
Tôi muốn xem dashboard hiệu suất của bản thân
Để dõi năng suất, quản lý workload cá nhân, cải thiện performance
Tiêu chí chấp nhận
- Khi truy cập dashboard cá nhân
- Thì chỉ hiển thị dữ liệu của CS Agent đó
- AC.US-CA-06.1: Personal Metrics Today
──────────────────────────────────────────────────────┐
│ MY PERFORMANCE - CS Agent: Nguyen Van A                       │
│ Today | This Week | This Month                                │
├──────────────────────────────────────────────────────┤
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐    │
│ │ Assigned       │ │    Closed       │ │    Avg Response │    │
│ │     25         │ │        23       │ │       9 min     │    │
│ │                │ │    92% rate     │ │    ✅ Under SLA │    │
│ └──────────────┘ └──────────────┘ └──────────────┘    │
│                                                                │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐    │
│ │ In Progress     │ │   Returned     │ │    SLA Met      │    │
│ │      2          │ │        1       │ │       96%       │    │
│ └──────────────┘ └──────────────┘ └──────────────┘    │
└──────────────────────────────────────────────────────┘
- AC.US-CA-06.2: Activity Timeline
┌─────────────────────────────────────┐
│ MY ACTIVITY TODAY                          │
├─────────────────────────────────────┤
│ 14:35 ✅ Closed conversation #1247         │
│ 14:20 💬 Replied to Khách A                │
│ 14:05 📥 Assigned conversation #1245       │
│ 13:50 ✅ Closed conversation #1243         │
│ 13:30 📝 Added internal note to #1240      │
│ ...                                        │
│ [Load More]                                │
└─────────────────────────────────────┘
- AC.US-CA-06.3: Performance Comparison
┌──────────────────────────────────────┐
│ COMPARE WITH TEAM AVERAGE                   │
├──────────────────────────────────────┤
│ Metric              │ You  │ Team Avg       │
├─────────────────────┼──────┼─────────┤
│ Conversations/day   │  25  │  22 (+3 😊)    │
│ Avg Response Time   │ 9min │ 12min (Better!)│
│ Resolution Rate     │ 92%  │  88% (+4%)     │
│ SLA Compliance      │ 96%  │  91% (+5%)     │
└───────────────────────────────────────┘
- AC.US-CA-06.4: Weekly Performance Trend
┌──────────────────────────────────────┐
│ MY PERFORMANCE TREND (Last 7 Days)          │
├──────────────────────────────────────┤
│ [Multi-line Chart]                          │
│                                             │
│ ━ Conversations Handled                     │
│ ━ Avg Response Time                         │
│ ━ SLA Compliance %                          │
│                                             │
│ Mon  Tue  Wed  Thu  Fri  Sat  Sun           │
│  22   24   26   25   28   20   18           │
└──────────────────────────────────────┘
- Metrics update ngay khi Agent thực hiện hành động 
- Activity timeline: Real-time WebSocket updates 
- Charts: Cache 5 phút
- CS Agent CHỈ xem metrics của chính mình
- Không xem được metrics của Agent khác
- CS Leader và Partner Admin có thể xem tất cả
