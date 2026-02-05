package repository

import (
	"context"

	"CONVERDA/internal/apps/domain/model/entity"

	"github.com/google/uuid"
)

type WebhookRepository interface {
	Create(ctx context.Context, webhook *entity.Webhook) (*entity.Webhook, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Webhook, error)
	ListByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*entity.Webhook, error)
	ListByApp(ctx context.Context, appID uuid.UUID) ([]*entity.Webhook, error)
	Update(ctx context.Context, webhook *entity.Webhook) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ProviderRepository interface {
	Create(ctx context.Context, provider *entity.Provider) (*entity.Provider, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Provider, error)
	ListByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*entity.Provider, error)
	ListByApp(ctx context.Context, appID uuid.UUID) ([]*entity.Provider, error)
	GetByTypeAndEnv(ctx context.Context, environmentID uuid.UUID, providerType string) (*entity.Provider, error)
	Update(ctx context.Context, provider *entity.Provider) error
	Delete(ctx context.Context, id uuid.UUID) error
}
