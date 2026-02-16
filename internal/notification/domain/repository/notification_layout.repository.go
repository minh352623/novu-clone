package repository

import (
	"context"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type NotificationLayoutRepository interface {
	Create(ctx context.Context, layout *entity.NotificationLayout) error
	Update(ctx context.Context, layout *entity.NotificationLayout) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.NotificationLayout, error)
	List(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationLayout, int64, error)
	GetDefault(ctx context.Context, envID uuid.UUID) (*entity.NotificationLayout, error)
}
