package repository

import (
	"context"
	"time"

	"CONVERDA/internal/apps/domain/model/entity"

	"github.com/google/uuid"
)

type UsageMetricRepository interface {
	Create(ctx context.Context, metric *entity.UsageMetric) error
	GetSummaryByApp(ctx context.Context, appID uuid.UUID, from, to time.Time) (map[string]int64, error)
	GetEnvironmentUsage(ctx context.Context, envID uuid.UUID, from, to time.Time) (int64, error)
}
