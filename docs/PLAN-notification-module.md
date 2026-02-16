# Plan: Notification Module (Novu-like Engine)

> **Goal**: Implement the robust Notification Engine defined in `schema.sql` (Section 9). This enables the "CMS" capabilities (Templates, Layouts) and the "Delivery" capabilities (Jobs, Batches), allowing the Frontend to build the Workflow/Template editors.

## 1. Context & Scope
- **References**: `schema.sql` Section 9 ("Notification Module").
- **Dependency**: Uses `Apps/Providers` to actually send messages.
- **Objective**: Provide APIs for creating templates and triggering notification jobs.

## 2. Technical Architecture

### 2.1. Domain Entities (`internal/notification/domain/entity`)
- **NotificationGroup**: Categorization (Marketing, Transactional).
- **NotificationLayout**: HTML wrappers (`{{content}}`).
- **NotificationTemplate**: The core definition (Subject, Body per channel, Variables Schema).
- **NotificationJob**: A request to send notifications (Status: Pending -> Processing -> Completed).

### 2.2. Persistence Layer
- **Repositories**:
    - `TemplateRepository`: CRUD for templates + versioning.
    - `JobRepository`: Create job, update status, logging.
    - `LayoutRepository` & `GroupRepository`.

### 2.3. Service Layer (`internal/notification/application/service`)
- **TemplateService**: Manage CMS content.
- **NotificationService**:
    - `TriggerEvent(eventCode, payload, recipients)`: The main entry point.
    - *Logic*:
        1. Lookup Template by `eventCode`.
        2. Resolve Layout.
        3. Render Content (using payload).
        4. Select Provider (from `Apps` module).
        5. Dispatch (Mock or calls Provider Service).
        6. Update Job Status.

### 2.4. API Layer
- `POST /notifications/trigger`: Trigger a notification.
- `GET /notifications/templates`: For UI List.
- `POST /notifications/templates`: UI Editor Save.
- `GET /notifications/jobs`: Activity Feed.

## 3. Implementation Steps

### Phase 1: CMS Core (Templates & Layouts)
- [ ] Define Entities & GORM Models.
- [ ] Implement CRUD APIs for `Groups`, `Layouts`, `Templates`.
- **Deliverable**: Frontend can build the "Template Editor" UI.

### Phase 2: Engine (Trigger & Dispatch)
- [ ] Implement `TriggerEvent` logic.
- [ ] Implement `Job` tracking.
- [ ] Connect to `ProviderService` (from Apps module) to effectively send.

## 4. Verification
- **Test**: Create a template -> Trigger it via API -> Verify Job created and marked 'Generated'.
- **Mock**: Initially mock the actual email sending if Provider integration is complex, focus on the *Engine* logic first.

## 5. Agent Assignments
- **Backend Specialist**: Full Go implementation.
