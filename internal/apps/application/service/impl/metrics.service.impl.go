package impl

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/internal/apps/application/service"
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

func (s *metricsServiceImpl) RecordUsage(ctx context.Context, appID, envID uuid.UUID, providerType string, statusCode int) error {
	metric := entity.NewUsageMetric(appID, envID, providerType, statusCode)
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
