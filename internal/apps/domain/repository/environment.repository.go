package repository

import (
	"context"

	"CONVERDA/internal/apps/domain/model/entity"

	"github.com/google/uuid"
)

type EnvironmentRepository interface {
	Create(ctx context.Context, env *entity.Environment) (*entity.Environment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Environment, error)
	GetByAppAndCode(ctx context.Context, appID uuid.UUID, code string) (*entity.Environment, error)
	ListByApp(ctx context.Context, appID uuid.UUID) ([]*entity.Environment, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*entity.Environment, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
