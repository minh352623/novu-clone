package impl

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/internal/apps/application/service"
	"CONVERDA/internal/apps/controller/dto"
	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/domain/repository"

	"github.com/google/uuid"
)

type metricsServiceImpl struct {
	metricsRepo repository.UsageMetricRepository
}

func NewMetricsService(metricsRepo repository.UsageMetricRepository) service.MetricsService {
	return &metricsServiceImpl{metricsRepo: metricsRepo}
}

func (s *metricsServiceImpl) RecordUsage(ctx context.Context, appID, envID uuid.UUID, providerType, direction string, statusCode int) error {
	metric := entity.NewUsageMetric(appID, envID, providerType, direction, statusCode)
	if err := s.metricsRepo.Create(ctx, metric); err != nil {
		return fmt.Errorf("failed to record usage metric: %w", err)
	}
	return nil
}

func (s *metricsServiceImpl) GetAppMetrics(ctx context.Context, appID uuid.UUID, days int) (map[string]int64, error) {
	if days <= 0 {
		days = 30
	}
	to := time.Now()
	from := to.AddDate(0, 0, -days)

	metrics, err := s.metricsRepo.GetSummaryByApp(ctx, appID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get app metrics summary: %w", err)
	}
	return metrics, nil
}

func (s *metricsServiceImpl) GetDetailedMetrics(ctx context.Context, appID uuid.UUID, days int) (*dto.DetailedMetricsResponse, error) {
	if days <= 0 {
		days = 30
	}
	to := time.Now()
	from := to.AddDate(0, 0, -days)

	summary, err := s.metricsRepo.GetDetailedSummary(ctx, appID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get detailed metrics: %w", err)
	}
	return dto.ToDetailedMetricsResponse(summary), nil
}

func (s *metricsServiceImpl) GetDailyTimeSeries(ctx context.Context, appID uuid.UUID, envID *uuid.UUID, days int) (*dto.TimeSeriesResponse, error) {
	if days <= 0 {
		days = 30
	}
	to := time.Now()
	from := to.AddDate(0, 0, -days)

	dailyMetrics, err := s.metricsRepo.GetDailyTimeSeries(ctx, appID, envID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily time series: %w", err)
	}
	return dto.ToTimeSeriesResponse(dailyMetrics), nil
}

func (s *metricsServiceImpl) GetEnvironmentBreakdown(ctx context.Context, appID uuid.UUID, days int) (*dto.EnvironmentBreakdownResponse, error) {
	if days <= 0 {
		days = 30
	}
	to := time.Now()
	from := to.AddDate(0, 0, -days)

	envMetrics, err := s.metricsRepo.GetEnvironmentBreakdown(ctx, appID, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment breakdown: %w", err)
	}
	return dto.ToEnvironmentBreakdownResponse(envMetrics), nil
}
