package impl

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/global"
	"CONVERDA/internal/health/dto"
	"CONVERDA/internal/health/service"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// QueueReader provides read access to the messaging queue state.
type QueueReader interface {
	CountByStatus(ctx context.Context, envID uuid.UUID, status string) (int64, error)
	CountOverdue(ctx context.Context, envID uuid.UUID) (int64, error)
	AvgWaitTime(ctx context.Context, envID uuid.UUID) (float64, error)
}

// SLAReader provides read access to SLA compliance data.
type SLAReader interface {
	GetSLACompliance(ctx context.Context, envID uuid.UUID, slaSeconds int, from, to time.Time) (*dto.SLACompliance, error)
}

// WebhookHealthReader provides read access to webhook dispatch stats.
type WebhookHealthReader interface {
	GetHealthStats(ctx context.Context, from, to time.Time) (*dto.WebhookHealth, error)
}

type healthServiceImpl struct {
	queueReader   QueueReader
	slaReader     SLAReader
	webhookReader WebhookHealthReader
	slaThreshold  int
}

func NewHealthService(
	queueReader QueueReader,
	slaReader SLAReader,
	webhookReader WebhookHealthReader,
	defaultSLAThreshold int,
) service.HealthService {
	return &healthServiceImpl{
		queueReader:   queueReader,
		slaReader:     slaReader,
		webhookReader: webhookReader,
		slaThreshold:  defaultSLAThreshold,
	}
}

func (s *healthServiceImpl) GetSystemHealth(ctx context.Context, envID uuid.UUID, from, to time.Time) (*dto.SystemHealthResponse, error) {
	// 1. Queue Health
	queueHealth, err := s.buildQueueHealth(ctx, envID)
	if err != nil {
		global.Logger.Error("health_service: failed to build queue health", zap.Error(err), zap.String("environmentID", envID.String()))
		return nil, fmt.Errorf("failed to build queue health: %w", err)
	}

	// 2. SLA Compliance
	sla, err := s.slaReader.GetSLACompliance(ctx, envID, s.slaThreshold, from, to)
	if err != nil {
		global.Logger.Error("health_service: failed to fetch SLA compliance", zap.Error(err), zap.String("environmentID", envID.String()))
		return nil, fmt.Errorf("failed to fetch SLA compliance: %w", err)
	}

	// 3. Webhook Health
	wh, err := s.webhookReader.GetHealthStats(ctx, from, to)
	if err != nil {
		global.Logger.Error("health_service: failed to fetch webhook health", zap.Error(err), zap.String("environmentID", envID.String()))
		return nil, fmt.Errorf("failed to fetch webhook health: %w", err)
	}

	return &dto.SystemHealthResponse{
		QueueHealth:   *queueHealth,
		SLACompliance: *sla,
		WebhookHealth: *wh,
		GeneratedAt:   time.Now(),
	}, nil
}

func (s *healthServiceImpl) buildQueueHealth(ctx context.Context, envID uuid.UUID) (*dto.QueueHealth, error) {
	unassigned, err := s.queueReader.CountByStatus(ctx, envID, "unassigned")
	if err != nil {
		return nil, fmt.Errorf("failed to count unassigned messages: %w", err)
	}

	assigned, err := s.queueReader.CountByStatus(ctx, envID, "assigned")
	if err != nil {
		return nil, fmt.Errorf("failed to count assigned messages: %w", err)
	}

	overdue, err := s.queueReader.CountOverdue(ctx, envID)
	if err != nil {
		return nil, fmt.Errorf("failed to count overdue messages: %w", err)
	}

	avgWait, err := s.queueReader.AvgWaitTime(ctx, envID)
	if err != nil {
		return nil, fmt.Errorf("failed to get average wait time: %w", err)
	}

	return &dto.QueueHealth{
		UnassignedCount: int(unassigned),
		AssignedCount:   int(assigned),
		OverdueCount:    int(overdue),
		AvgWaitTimeSec:  avgWait,
		TotalActive:     int(unassigned + assigned),
	}, nil
}
