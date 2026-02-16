package service

import (
	"context"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type LayoutManager interface {
	CreateLayout(ctx context.Context, envID uuid.UUID, name, description, contentHTML string, variables map[string]interface{}, isDefault bool) (*entity.NotificationLayout, error)
	UpdateLayout(ctx context.Context, envID uuid.UUID, id uuid.UUID, name, description, contentHTML string, variables map[string]interface{}, isDefault bool) error
	DeleteLayout(ctx context.Context, envID uuid.UUID, id uuid.UUID) error
	GetLayout(ctx context.Context, envID uuid.UUID, id uuid.UUID) (*entity.NotificationLayout, error)
	ListLayouts(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationLayout, int64, error)
	GetDefaultLayout(ctx context.Context, envID uuid.UUID) (*entity.NotificationLayout, error)
}
