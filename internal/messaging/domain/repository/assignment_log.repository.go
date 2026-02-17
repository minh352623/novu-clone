package repository

import (
	"context"
	"time"

	"CONVERDA/internal/messaging/domain/model/entity"

	"github.com/google/uuid"
)

type AssignmentLogRepository interface {
	Create(ctx context.Context, log *entity.AssignmentLog) error
	GetLastByThread(ctx context.Context, threadID uuid.UUID) (*entity.AssignmentLog, error)
	GetByThread(ctx context.Context, threadID uuid.UUID) ([]*entity.AssignmentLog, error)
	Update(ctx context.Context, log *entity.AssignmentLog) error
	GetTeamStats(ctx context.Context, envID uuid.UUID, from, to time.Time, slaSeconds int) (*entity.TeamStats, error)
	GetAgentStats(ctx context.Context, envID uuid.UUID, memberID uuid.UUID, from, to time.Time, slaSeconds int) (*entity.AgentStats, error)
	GetActivityTimeline(ctx context.Context, envID uuid.UUID, memberID uuid.UUID, from, to time.Time, interval string) ([]*entity.ActivityPoint, error)
}
