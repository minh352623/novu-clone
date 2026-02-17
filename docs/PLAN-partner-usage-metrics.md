# PLAN: US-RA-02.5 — Per-Partner Detailed Usage Metrics

## Problem
Current metrics are **too basic**: `GetAppMetrics` returns `map[string]int64` grouped only by provider_type. Partners (Apps) need **detailed breakdowns**:
- **Message direction**: inbound vs outbound
- **Success/failure** rates
- **Per-environment** granularity
- **Daily time-series** for trend analysis

### Existing Code
| Component | Current State |
|-----------|--------------|
| Entity: `UsageMetric` | ✅ Has: appID, envID, providerType, statusCode, timestamp |
| Repo: `GetSummaryByApp` | Only groups by provider_type, returns `map[string]int64` |
| Repo: `GetEnvironmentUsage` | Returns raw count per environment |
| Service: `MetricsService` | 2 methods: `RecordUsage`, `GetAppMetrics` |
| Controller: `GetAppMetrics` | Single endpoint: `GET /:app_id/metrics?days=30` |

---

## Review Level: L2 (Proposal)
- No new tables required — extends existing `usage_metrics`
- Adds `direction` column to entity + table
- New repository query methods + service methods + API endpoints
- No breaking changes to existing API

---

## Proposed Changes

### Component 1: Database Migration

#### [NEW] [00025_usage_metrics_direction.sql](file:///Users/tekix/Documents/company/converda/converda-service/sql/schema/00025_usage_metrics_direction.sql)

```sql
-- Add direction column to usage_metrics table
ALTER TABLE usage_metrics ADD COLUMN direction TEXT NOT NULL DEFAULT 'outbound';
-- Values: 'inbound', 'outbound'

-- Add index for common query patterns
CREATE INDEX idx_usage_metrics_app_direction ON usage_metrics(app_id, direction, timestamp);
CREATE INDEX idx_usage_metrics_env_direction ON usage_metrics(environment_id, direction, timestamp);
```

---

### Component 2: Domain Layer Updates

#### [MODIFY] [metrics.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/apps/domain/model/entity/metrics.go)

- Add `Direction` field (`"inbound"` / `"outbound"`) to `UsageMetric` entity
- Add constants: `DirectionInbound`, `DirectionOutbound`
- Update `NewUsageMetric` to accept direction parameter

#### [MODIFY] [metrics.model.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/apps/infrastructure/persistence/model/metrics.model.go)

- Add `Direction` column to GORM model

---

### Component 3: Repository — New Query Methods

#### [MODIFY] [metrics.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/apps/domain/repository/metrics.repository.go)

Add to `UsageMetricRepository` interface:
```go
GetDetailedSummary(ctx context.Context, appID uuid.UUID, from, to time.Time) (*DetailedMetricsSummary, error)
GetDailyTimeSeries(ctx context.Context, appID uuid.UUID, envID *uuid.UUID, from, to time.Time) ([]DailyMetric, error)
GetEnvironmentBreakdown(ctx context.Context, appID uuid.UUID, from, to time.Time) ([]EnvironmentMetric, error)
```

New result types (in entity):
```go
type DetailedMetricsSummary struct {
    TotalMessages     int64
    InboundMessages   int64
    OutboundMessages  int64
    SuccessCount      int64   // statusCode 2xx
    FailureCount      int64   // statusCode 4xx/5xx
    ByProvider        map[string]ProviderMetric
}

type ProviderMetric struct {
    Total   int64
    Success int64
    Failed  int64
}

type DailyMetric struct {
    Date     string  // "2026-02-17"
    Inbound  int64
    Outbound int64
    Success  int64
    Failed   int64
}

type EnvironmentMetric struct {
    EnvironmentID uuid.UUID
    Total         int64
    Inbound       int64
    Outbound      int64
    Success       int64
    Failed        int64
}
```

#### [MODIFY] [metrics.repository.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/apps/infrastructure/persistence/repository/metrics.repository.go)

Implement the 3 new queries using GORM + SQL aggregation.

---

### Component 4: Service Layer

#### [MODIFY] [metrics.service.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/apps/application/service/metrics.service.go)

Add to `MetricsService` interface:
```go
GetDetailedMetrics(ctx context.Context, appID uuid.UUID, days int) (*dto.DetailedMetricsResponse, error)
GetDailyTimeSeries(ctx context.Context, appID uuid.UUID, envID *uuid.UUID, days int) (*dto.TimeSeriesResponse, error)
GetEnvironmentBreakdown(ctx context.Context, appID uuid.UUID, days int) (*dto.EnvironmentBreakdownResponse, error)
```

#### [MODIFY] [metrics.service.impl.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/apps/application/service/impl/metrics.service.impl.go)

- Update `RecordUsage` to accept `direction` parameter
- Implement the 3 new service methods

---

### Component 5: DTOs + Controller

#### [NEW] [metrics.dto.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/apps/controller/dto/metrics.dto.go)

```go
type DetailedMetricsResponse struct { ... }
type TimeSeriesResponse struct { ... }
type EnvironmentBreakdownResponse struct { ... }
```

#### [MODIFY] [app.controller.go](file:///Users/tekix/Documents/company/converda/converda-service/internal/apps/controller/app.controller.go)

Add 2 new endpoints:
| Method | Route | Description |
|--------|-------|-------------|
| `GET` | `/:app_id/metrics/detailed` | Full breakdown (direction + provider + success/fail) |
| `GET` | `/:app_id/metrics/timeseries` | Daily aggregated time-series |

---

### Component 6: Update Callers

#### [MODIFY] RecordUsage callers

Add `direction` parameter to all call sites that invoke `RecordUsage`. These primarily exist in the notification dispatcher when sending outbound notifications.

---

## Verification Plan

### Automated Tests
```bash
go build ./...
go test -race ./internal/apps/...
```
- Extend existing `MetricsService` tests with new methods
- Unit test `GetDetailedSummary`, `GetDailyTimeSeries`, `GetEnvironmentBreakdown` queries via mocks

### Manual Verification
1. Run migration
2. Record sample usage data via API
3. Query `GET /:app_id/metrics/detailed` — verify inbound/outbound/success/fail breakdown
4. Query `GET /:app_id/metrics/timeseries?days=7` — verify daily data points

---

## Effort Estimate

| Component | Effort |
|-----------|--------|
| Migration + Entity + Model | ~30 min |
| Repository queries (3 new methods) | ~1.5h |
| Service + DTOs | ~1h |
| Controller endpoints | ~30 min |
| Update RecordUsage callers | ~30 min |
| Tests | ~1h |
| **Total** | **~5 hours** |

---

## State Management
After completion, update:
- `AI_STATE_MINH.md` — mark US-RA-02.5 complete
- `PROJECT_MAIN_BACKLOG.md` — check off US-RA-02.5, update Apps progress to ~90%
