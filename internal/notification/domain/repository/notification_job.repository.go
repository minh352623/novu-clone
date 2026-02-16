package repository

import (
	"context"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type NotificationJobRepository interface {
	Create(ctx context.Context, job *entity.NotificationJob) error
	Update(ctx context.Context, job *entity.NotificationJob) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.NotificationJob, error)
	List(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationJob, int64, error)
	GetPendingJobs(ctx context.Context) ([]*entity.NotificationJob, error)
}
