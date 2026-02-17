package service

import (
	"context"

	"CONVERDA/internal/apps/controller/dto"

	"github.com/google/uuid"
)

type MetricsService interface {
	RecordUsage(ctx context.Context, appID, envID uuid.UUID, providerType, direction string, statusCode int) error
	GetAppMetrics(ctx context.Context, appID uuid.UUID, days int) (map[string]int64, error)

	// US-RA-02.5: Detailed metrics
	GetDetailedMetrics(ctx context.Context, appID uuid.UUID, days int) (*dto.DetailedMetricsResponse, error)
	GetDailyTimeSeries(ctx context.Context, appID uuid.UUID, envID *uuid.UUID, days int) (*dto.TimeSeriesResponse, error)
	GetEnvironmentBreakdown(ctx context.Context, appID uuid.UUID, days int) (*dto.EnvironmentBreakdownResponse, error)
}
