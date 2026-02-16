package service

import (
	"context"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type GroupManager interface {
	CreateGroup(ctx context.Context, envID uuid.UUID, name, key, description string, isDefault bool) (*entity.NotificationGroup, error)
	UpdateGroup(ctx context.Context, envID uuid.UUID, id uuid.UUID, name, description string) error
	DeleteGroup(ctx context.Context, envID uuid.UUID, id uuid.UUID) error
	GetGroup(ctx context.Context, envID uuid.UUID, id uuid.UUID) (*entity.NotificationGroup, error)
	ListGroups(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationGroup, int64, error)
	GetByKey(ctx context.Context, envID uuid.UUID, key string) (*entity.NotificationGroup, error)
}
