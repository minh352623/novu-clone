package repository

import (
	"context"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type NotificationGroupRepository interface {
	Create(ctx context.Context, group *entity.NotificationGroup) error
	Update(ctx context.Context, group *entity.NotificationGroup) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.NotificationGroup, error)
	GetByKey(ctx context.Context, envID uuid.UUID, key string) (*entity.NotificationGroup, error)
	List(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationGroup, int64, error)
}
