package service

import (
	"context"
	"time"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type JobScheduler interface {
	ScheduleJob(ctx context.Context, envID uuid.UUID, tenantID *uuid.UUID, channel string, templateID *uuid.UUID, scheduledAt *time.Time, recipients []map[string]interface{}, metadata map[string]interface{}) (*entity.NotificationJob, error)
	CancelJob(ctx context.Context, envID uuid.UUID, id uuid.UUID) error
	GetJob(ctx context.Context, envID uuid.UUID, id uuid.UUID) (*entity.NotificationJob, error)
	ListJobs(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationJob, int64, error)
	ProcessJob(ctx context.Context, jobID uuid.UUID) error // Triggers processing (can be called by worker or API)
}
