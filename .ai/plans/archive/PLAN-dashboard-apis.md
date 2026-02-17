# Implementation Plan - Dashboard APIs (US-PA-03, US-CL-03)

This plan outlines the implementation of REST API endpoints for the Partner and Team Dashboards in the `messaging` module. These endpoints will provide key performance metrics, response time distributions, and agent workload status.

## User Review Required

> [!IMPORTANT]
> **SLA Configuration**: For now, the SLA threshold will be passed as a query parameter (defaulting to 15m). In the future, this should be moved to an Environment-level configuration.
> **Workload Formula**: Currently just a count of open assigned threads.

## Proposed Changes

### [Messaging Module]

#### [MODIFY] [messaging.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/domain/model/entity/messaging.go)
- Enhance `TeamStats` and `AgentStats` to include:
    - `ResponseTimeDistribution` (struct with buckets: `<5m`, `5-15m`, `>15m`).
    - `AverageResponseTime` (existing).
    - `Workload` (for agents).

#### [MODIFY] [assignment_log.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/domain/repository/assignment_log.repository.go)
- Update `GetTeamStats` and `GetAgentStats` signatures to include `SLASeconds int`.

#### [MODIFY] [assignment_log.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/infrastructure/persistence/repository/assignment_log.repository.go)
- Implement distribution bucket logic using `CASE WHEN` in SQL for efficiency.
- Update SLA compliance calculation to use the provided `SLASeconds`.

#### [NEW] [dashboard.dto.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/dto/dashboard.dto.go)
- Define `DashboardStatsRequest` (EnvironmentID, From, To, SLAThreshold).
- Define `TeamDashboardResponse` and `PartnerDashboardResponse`.

#### [MODIFY] [conversation.service.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/conversation.service.go)
- Add `GetTeamDashboard` and `GetPartnerDashboard` methods.

#### [MODIFY] [conversation.service.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/application/service/impl/conversation.service.impl.go)
- Implement dashboard logic, handling default 24h time range if `From`/`To` are zero.

#### [MODIFY] [conversation.controller.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/conversation.controller.go)
- Add `GetTeamDashboard` and `GetPartnerDashboard` handlers.

#### [MODIFY] [router.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/messaging/controller/router.go)
- Register new GET endpoints: `/dashboards/team` and `/dashboards/partner`.

## Verification Plan

### Automated Tests
- `go test ./internal/messaging/infrastructure/persistence/repository/...` to verify SQL logic for distribution buckets.
- `go test ./internal/messaging/application/service/impl/...` to verify service logic and default values.

### Manual Verification
- Use `curl` or Postman to hit the new endpoints.
- Verify that response time distribution buckets sum up correctly to the resolved count.
- Verify that agent workload matches the number of 'assigned' threads for that agent.
