# Plan: Frontend Documentation v2 (Messaging & OpenAPI)

> **Goal**: Update the frontend integration guide to include the newly implemented **Messaging Module** (Inbox, Conversations, Closure Table logic) and generate formal OpenAPI credentials.

## 1. Context
- **Previous State**: `docs/FRONTEND_INTEGRATION.md` covers Infra (Apps, providers, webhooks).
- **New Feature**: Messaging Module (`/conversations/*`) is now ready.
- **Requirement**: Frontend team needs to know how to build the **Agent Inbox** (like Intercom/Zendesk) and **Chat Widget** (SDK).

## 2. Documentation Updates (`docs/FRONTEND_INTEGRATION.md`)

I will add a new major section **6. Feature: Messaging Inbox (CS Agent)** covering:

### 2.1. Agent Inbox UI
- **Path**: `/apps/:appId/inbox`
- **sidebar**: List of Conversation Pools (Filtered by `status=unassigned` or `assigned_to=me`).
- **Endpoints**: `GET /api/v1/conversations?status=...`

### 2.2. Conversation Detail & Reply
- **Main View**: Chat bubble list.
- **Nested Replies**: Visualizing the "Thread" structure if we support it in UI, or flat list if typical chat. *Note: Backend supports nested (closure), but UI might flatten it.*
- **Action**: "Reply" -> `POST /conversations/:id/messages`.

### 2.3. Assignment Logic
- **Action**: "Assign to Me" button.
- **Endpoint**: `PATCH /conversations/:id/assign`.

## 3. OpenAPI Generation (`/document api`)
- Run `swag init` to generate `docs/swagger.json` and `docs/swagger.yaml`.
- This provides the "Source of Truth" for API types.

## 4. Execution Steps

### Step 1: Update Integration Guide
- [ ] Append "Feature: Messaging Inbox" to `docs/FRONTEND_INTEGRATION.md`.
- [ ] Detail the Data Flow for "Real-time" (Simulated via polling for now, or specifying WebSocket future).

### Step 2: Generate Swagger
- [ ] Run `swag init` in `cmd/server/main.go` or root (depending on setup).
- [ ] Verify `docs/docs.go` is updated.

## 5. Agent Assignments
- **Tech Writer**: Update MD files.
- **Backend Spec**: Run swaggo.
