# Module Flows Documentation

This document describes the key flows within the **IAM** and **Apps** modules, including sequence diagrams for visual understanding.

## IAM Module

### 1. Authentication Flow
**Goal**: User registers, logs in, and obtains a JWT token.

```mermaid
sequenceDiagram
    participant Clients
    participant AuthController
    participant AuthService
    participant UserRepo
    participant DB

    Client->>AuthController: POST /auth/register
    AuthController->>AuthService: Register(email, password)
    AuthService->>UserRepo: CreateUser(user)
    UserRepo->>DB: INSERT INTO users...
    DB-->>UserRepo: success
    UserRepo-->>AuthService: user
    AuthService-->>AuthController: user, token
    AuthController-->>Client: 201 Created (Token)

    Client->>AuthController: POST /auth/login
    AuthController->>AuthService: Login(email, password)
    AuthService->>UserRepo: FindByEmail(email)
    UserRepo-->>AuthService: user
    AuthService->>AuthService: CheckPassword(hash, password)
    AuthService-->>AuthController: token
    AuthController-->>Client: 200 OK (Token)
```

### 2. Tenant & Member Invitation Flow
**Goal**: Admin invites a user to a tenant. User accepts.

```mermaid
sequenceDiagram
    participant Admin
    participant User
    participant TenantController
    participant MemberService
    participant InvitationRepo
    participant MemberRepo

    Admin->>TenantController: POST /tenants/:id/invitations (email)
    TenantController->>MemberService: InviteMember(tenantID, email)
    MemberService->>InvitationRepo: CreateInvitation(invite)
    InvitationRepo-->>MemberService: invite
    MemberService-->>TenantController: invite
    TenantController-->>Admin: 201 Created (Invite details)
    Note right of MemberService: Mocks sending email to User

    User->>TenantController: POST /invitations/accept (token)
    TenantController->>MemberService: AcceptInvitation(token)
    MemberService->>InvitationRepo: FindByToken(token)
    InvitationRepo-->>MemberService: invite
    MemberService->>MemberRepo: CreateMember(tenantID, userID)
    MemberRepo-->>MemberService: member
    MemberService->>InvitationRepo: UpdateStatus(accepted)
    MemberService-->>TenantController: member
    TenantController-->>User: 200 OK (Membership confirmed)
```

## Apps Module

### 1. App Creation Flow
**Goal**: Create an App and auto-provision environments.

```mermaid
sequenceDiagram
    participant User
    participant AppController
    participant AppService
    participant AppRepo
    participant EnvRepo

    User->>AppController: POST /tenants/:id/apps (name)
    AppController->>AppService: CreateApp(tenantID, name)
    AppService->>AppRepo: Create(app)
    AppRepo-->>AppService: app
    
    loop Default Environments
        AppService->>EnvRepo: Create(dev/stage/prod)
    end
    
    AppService-->>AppController: app + envs
    AppController-->>User: 201 Created (App with Envs)
```

### 2. Webhook Management
**Goal**: Create a webhook for an environment.

```mermaid
sequenceDiagram
    participant User
    participant WebhookController
    participant WebhookService
    participant WebhookRepo

    User->>WebhookController: POST /apps/:id/webhooks
    WebhookController->>WebhookService: CreateWebhook(appID, envID, url...)
    WebhookService->>WebhookService: GenerateSecret()
    WebhookService->>WebhookRepo: Create(webhook)
    WebhookRepo-->>WebhookService: webhook
    WebhookService-->>WebhookController: webhook
    WebhookController-->>User: 201 Created (Webhook + Secret)
```
