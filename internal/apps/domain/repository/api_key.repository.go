package repository

import (
	"context"

	"CONVERDA/internal/apps/domain/model/entity"

	"github.com/google/uuid"
)

type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *entity.APIKey) (*entity.APIKey, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.APIKey, error)
	GetByHash(ctx context.Context, hash string) (*entity.APIKey, error)
	GetByEnvironment(ctx context.Context, envID uuid.UUID) ([]*entity.APIKey, error)
	Update(ctx context.Context, apiKey *entity.APIKey) error
	RevokeAllByEnvironment(ctx context.Context, envID uuid.UUID) error
}
