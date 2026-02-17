# Plan: Frontend Integration Documentation (Novu-like UI)

> **Goal**: Create comprehensive documentation (`docs/FRONTEND_INTEGRATION.md`) to guide the Frontend team in building a UI for the Converda system, mimicking the functionality and UX of Novu (Notification Management) and general CRM/Dashboard features. This plan focuses on mapping the existing backend Modules (`Apps`, `Connectors`, `Workflows`, `IAM`) to UI requirements.

## 1. Context & Scope
- **Source**: `functions.md` (Functional Specs), `schema.sql` (DB Structure), Existing Codebase (`internal/apps/...`).
- **Target Audience**: Frontend Engineering Team.
- **Visual Reference**: Novu, Segment, Courier (modern SaaS Dashboards).

## 2. Documentation Structure (`docs/FRONTEND_INTEGRATION.md`)

I will create a single "Master Integration Guide" that breaks down into 4 key domains:

### 2.1. Domain: Application & Environment Management
*Maps to: `apps`, `environments`, `api_keys` tables*
- **UI Components**:
    - **Header/Sidebar**: Environment Switcher (Development / Staging / Production).
    - **API Keys Page**: List keys, Rotate/Regenerate key UI (with Copy to Clipboard).
    - **App Settings**: General App details.
- **Key Concepts**:
    - "Environment Context": The frontend must always send the `X-Environment-ID` or use the API Key associated with the selected environment.
    - Tenant isolation.

### 2.2. Domain: Connectors (Providers) Store
*Maps to: `providers`, `provider_configs` tables*
- **UI Components**:
    - **Integration Store**: Grid of cards for providers (SendGrid, Twilio, FCM, etc.).
    - **Config Drawer/Modal**: Form to input API/Secret keys specific to the active Environment.
    - **Connection Status**: Toggle switch (Active/Inactive).
- **Inspiration**: Novu Integrations Store / Segment Destinations.

### 2.3. Domain: Webhooks & Events
*Maps to: `webhooks`, `webhook_logs` tables*
- **UI Components**:
    - **Webhook Destinations**: List of URLs + Event triggers customization.
    - **Activity Log/Debugger**: Table showing recent webhook dispatches, request/response payloads (like Stripe Developer Dashboard).
- **Key Actions**:
    - "Send Test Event": Important for DX.

### 2.4. Domain: Notification Workflows (CMS)
*Maps to: `workflows`, `notification_templates`, `notification_groups` tables*
- **UI Components**:
    - **Workflow Editor**: Canvas or list-based editor for triggers and steps.
    - **Template Editor**: Rich text/HTML editor with variable injection `{{var}}`.
    - **Variable Manager**: Schema definition for dynamic content.
- **Inspiration**: Novu Workflow Editor.

## 3. Execution Steps

### Step 1: Draft `docs/FRONTEND_INTEGRATION.md`
- [ ] Define **Authentication & Global State** (Tenant + App + Environment headers).
- [ ] Document **URL Structure** (Routing guide).
- [ ] Detail **API Integration Patterns** (Status codes, Pagination, Error handling).

### Step 2: UI/UX Component Specifications
- [ ] **Data Grids**: Sorting, Filtering (Server-side).
- [ ] **Forms**: Validation rules (mapped from Backend DTOs).
- [ ] **Visual States**: Skeleton loaders, Empty states, Error toasts.

### Step 3: API Reference Linking
- [ ] Link to Swagger/OpenAPI (Using `/document api` output).
- [ ] Provide "Curl" examples for critical flows (e.g., triggering a workflow, testing a provider).

## 4. Verification
- **Review**: Ensure all fields in `schema.sql` (like `provider_configs`, `webhook_events`) have a corresponds UI representation or assumption.
- **Sign-off**: User approval on the plan before implementing the guide.

## 5. Agent Assignments
- **Tech Lead (Agent)**: Drafts the Architecture & Data Flow using user rules.
- **Product Owner (Agent)**: Ensures functional requirements from `functions.md` are met.
