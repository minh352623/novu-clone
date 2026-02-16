package service

import (
	"context"

	"github.com/google/uuid"
)

type MetricsService interface {
	RecordUsage(ctx context.Context, appID, envID uuid.UUID, providerType string, statusCode int) error
	GetAppMetrics(ctx context.Context, appID uuid.UUID, days int) (map[string]int64, error)
}
