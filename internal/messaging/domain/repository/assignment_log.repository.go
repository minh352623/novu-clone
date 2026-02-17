package repository

import (
	"context"
	"time"

	"CONVERDA/internal/messaging/domain/model/entity"

	"github.com/google/uuid"
)

type AuditTrailFilter struct {
	ThreadID uuid.UUID
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

type AssignmentLogRepository interface {
	Create(ctx context.Context, log *entity.AssignmentLog) error
	GetLastByThread(ctx context.Context, threadID uuid.UUID) (*entity.AssignmentLog, error)
	GetByThread(ctx context.Context, threadID uuid.UUID) ([]*entity.AssignmentLog, error)
	ListByThread(ctx context.Context, filter AuditTrailFilter) ([]*entity.AssignmentLog, int64, error)
	Update(ctx context.Context, log *entity.AssignmentLog) error
	GetTeamStats(ctx context.Context, envID uuid.UUID, from, to time.Time, slaSeconds int) (*entity.TeamStats, error)
	GetAgentStats(ctx context.Context, envID uuid.UUID, memberID uuid.UUID, from, to time.Time, slaSeconds int) (*entity.AgentStats, error)
	GetActivityTimeline(ctx context.Context, envID uuid.UUID, memberID uuid.UUID, from, to time.Time, interval string) ([]*entity.ActivityPoint, error)
}
