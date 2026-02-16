# Frontend Integration Guide (Converda Console)

This document serves as the implementation guide for the **Converda Console UI**. The goal is to build a modern, developer-first dashboard similar to **Novu**, **Segment**, or **Stripe**, allowing users to manage their communication infrastructure.

---

## 1. Global Context & Authentication

The Console operates within a strict hierarchy: **Tenant > App > Environment**.
The frontend must maintain this context globally and pass it in API requests.

### 1.1. Context Headers
Every API request to `/api/v1/apps/*` or `/api/v1/system/*` MUST include the following headers:

| Header | Description | Source |
|--------|-------------|--------|
| `Authorization` | `Bearer <JWT>` | Logged-in User Token |
| `X-Tenant-ID` | UUID | Selected Tenant (from URL or State) |
| `X-App-ID` | UUID | Selected App (from URL or State) |
| `X-Environment-ID` | UUID | **Crucial**. Defines if we are editing `Dev`, `Staging`, or `Prod` |

> **Hard Requirement**: The UI must have a globally accessible **Environment Switcher** (Dropdown) in the top navigation bar. Changing this switcher reloads all data on the current page for the new environment.

---

## 2. Feature: Application & Keys Management

**Goal**: Manage the API Keys used by the SDKs to connect to Converda.

### 2.1. UI Layout
- **Path**: `/apps/:appId/settings/keys`
- **Components**:
    - **Environment Tabs/Switcher**: (If not in global header).
    - **Key List**: Display the masked key (e.g., `sk_live_...4a2b`).
    - **Actions**:
        - `Copy to Clipboard`
        - `Regenerate Key` (Destructive action, requires confirmation modal).
        - `Revoke Key`.

### 2.2. API Integration
- **List Keys**: `GET /api/v1/apps/:app_id/environments/:env_id/api-keys`
    - Response: `{ id, key_prefix, created_at, last_used_at }`
- **Rotate Key**: `POST /api/v1/apps/:app_id/environments/:env_id/api-keys/rotate`
    - **UX**: Show the full key **ONCE** in a modal. Never display it again.

---

## 3. Feature: Connectors Store (Integrations)

**Goal**: A "Marketplace" feel where users enable/disable providers (Email, SMS, Chat, Push).

### 3.1. UI Layout
- **Path**: `/apps/:appId/integrations`
- **Components**:
    - **Provider Grid**: Cards with Logos (SendGrid, Twilio, Firebase, Slack, etc.).
    - **Status Indicators**: `Active` (Green Dot) / `Inactive` (Grey).
    - **Configuration Drawer**: Clicking a card opens a drawer/modal to input credentials.

### 3.2. Configuration Drawer
Each provider type requires specific fields. The UI should render dynamic forms based on the provider type, OR hardcode forms for supported providers initially.

**Example: SendGrid Form**
- `API Key` (Password field)
- `From Email` (Text)
- `From Name` (Text)
- Toggle: `Active`

### 3.3. API Integration
- **List Providers**: `GET /api/v1/apps/:app_id/providers`
    - Filter response in UI to separate `installed` vs `available`.
- **Connect Provider**: `POST /api/v1/apps/:app_id/providers`
    - Payload:
      ```json
      {
        "environment_id": "...",
        "provider_type": "email",
        "provider_name": "sendgrid",
        "configuration": { "api_key": "..." }
      }
      ```
- **Update Config**: `PUT /api/v1/providers/:provider_id`

---

## 4. Feature: Webhooks

**Goal**: Allow users to subscribe to Converda events (e.g., `message.received`, `workflow.completed`).

### 4.1. UI Layout
- **Path**: `/apps/:appId/webhooks`
- **Components**:
    - **Endpoint List**: Table showing `URL`, `Active Events`, `Status`, `Secret (Hidden)`.
    - **Add Endpoint Button**: Opens creation modal.
    - **Secret Reveal**: Click to reveal signing secret (`whsec_...`).

### 4.2. API Integration
- **Create Webhook**: `POST /api/v1/apps/:app_id/webhooks`
    - Input: `URL`, `Description`, `Events` (Multi-select dropdown).
    - Supported Events:
        - `message.sent`
        - `message.delivered`
        - `conversation.started`
- **List Webhooks**: `GET /api/v1/apps/:app_id/webhooks`

---

## 5. Feature: Workflows (CMS Preview)

**Goal**: Visualizer for notification flows. *Note: Backend implementation is pending full CMS logic, but UI can be prepped.*

### 5.1. UI Layout (Vision)
- **Visual Builder**: A React Flow / diagram canvas.
    - **Nodes**: `Trigger` -> `Delay` -> `Email` -> `In-App` -> `Push`.
    - **Edges**: Connection lines.
- **Sidebar**: Drag-and-drop steps.

---

## 6. Standard Responses & Errors

All API responses follow the envelope pattern:

**Success (200/201)**
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

**Error (4xx/5xx)**
```json
{
  "code": 400,
  "message": "Invalid request parameters",
  "error": "Detail error message or validation errors"
}
```

**UX Requirements**:
- On `401 Unauthorized`: Redirect to Login.
- On `403 Forbidden`: Show "Access Denied" page (Permission check).
- On `400 Bad Request`: Show form validation errors or Toast notification.
