package repository

import (
	"context"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type NotificationRepository interface {
	Create(ctx context.Context, notif *entity.Notification) error
	Update(ctx context.Context, notif *entity.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error)
}
