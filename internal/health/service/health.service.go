package service

import (
	"context"
	"time"

	"CONVERDA/internal/health/dto"

	"github.com/google/uuid"
)

// HealthService aggregates metrics from messaging + notification modules.
type HealthService interface {
	GetSystemHealth(ctx context.Context, envID uuid.UUID, from, to time.Time) (*dto.SystemHealthResponse, error)
}
