package service

import (
	"context"

	"CONVERDA/internal/apps/domain/model/entity"

	"github.com/google/uuid"
)

type AppService interface {
	CreateApp(ctx context.Context, tenantID uuid.UUID, name string, description *string, slaThreshold *int) (*entity.App, error)
	GetApp(ctx context.Context, appID uuid.UUID) (*entity.App, error)
	ListApps(ctx context.Context, tenantID uuid.UUID) ([]*entity.App, error)
	UpdateApp(ctx context.Context, app *entity.App) error
	DeleteApp(ctx context.Context, appID uuid.UUID) error
}

type EnvironmentService interface {
	CreateEnvironment(ctx context.Context, appID uuid.UUID, code string, slaThreshold *int) (*entity.Environment, error)
	ListEnvironments(ctx context.Context, appID uuid.UUID) ([]*entity.Environment, error)
	GetEnvironment(ctx context.Context, appID uuid.UUID, code string) (*entity.Environment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Environment, error)
	VerifyAPIKey(ctx context.Context, apiKey string) (*entity.Environment, error)
	UpdateEnvironment(ctx context.Context, env *entity.Environment) error
}

type WebhookService interface {
	CreateWebhook(ctx context.Context, tenantID, appID, envID uuid.UUID, url string, events []string, desc *string) (*entity.Webhook, error)
	GetWebhook(ctx context.Context, id uuid.UUID) (*entity.Webhook, error)
	ListWebhooks(ctx context.Context, appID uuid.UUID) ([]*entity.Webhook, error)
	UpdateWebhook(ctx context.Context, webhook *entity.Webhook) error
	DeleteWebhook(ctx context.Context, id uuid.UUID) error
	ToggleWebhook(ctx context.Context, id uuid.UUID, active bool) error
}

type ProviderService interface {
	CreateProvider(ctx context.Context, tenantID, appID, envID uuid.UUID, providerType, providerName string, config map[string]interface{}) (*entity.Provider, error)
	GetProvider(ctx context.Context, id uuid.UUID) (*entity.Provider, error)
	ListProviders(ctx context.Context, appID uuid.UUID) ([]*entity.Provider, error)
	GetActiveProvider(ctx context.Context, envID uuid.UUID, providerType string) (*entity.Provider, error)
	UpdateProvider(ctx context.Context, provider *entity.Provider) error
	DeleteProvider(ctx context.Context, id uuid.UUID) error
	ToggleProvider(ctx context.Context, id uuid.UUID, active bool) error
}
