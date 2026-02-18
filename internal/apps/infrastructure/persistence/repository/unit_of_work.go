package repository

import (
	"context"

	"CONVERDA/internal/apps/domain/repository"

	"gorm.io/gorm"
)

type gormAppsUnitOfWork struct {
	db *gorm.DB
}

// NewAppsUnitOfWork creates a new GORM-based implementation of AppsUnitOfWork
func NewAppsUnitOfWork(db *gorm.DB) repository.AppsUnitOfWork {
	return &gormAppsUnitOfWork{db: db}
}

func (u *gormAppsUnitOfWork) Execute(ctx context.Context, fn func(repository.AppsTxRepository) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &gormAppsTxRepository{
			apps:         NewAppRepository(tx),
			environments: NewEnvironmentRepository(tx),
			apiKeys:      NewAPIKeyRepository(tx),
			webhooks:     NewWebhookRepository(tx),
			providers:    NewProviderRepository(tx),
			metrics:      NewUsageMetricRepository(tx),
		}
		return fn(txRepo)
	})
}

type gormAppsTxRepository struct {
	apps         repository.AppRepository
	environments repository.EnvironmentRepository
	apiKeys      repository.APIKeyRepository
	webhooks     repository.WebhookRepository
	providers    repository.ProviderRepository
	metrics      repository.UsageMetricRepository
}

func (r *gormAppsTxRepository) Apps() repository.AppRepository {
	return r.apps
}

func (r *gormAppsTxRepository) Environments() repository.EnvironmentRepository {
	return r.environments
}

func (r *gormAppsTxRepository) APIKeys() repository.APIKeyRepository {
	return r.apiKeys
}

func (r *gormAppsTxRepository) Webhooks() repository.WebhookRepository {
	return r.webhooks
}

func (r *gormAppsTxRepository) Providers() repository.ProviderRepository {
	return r.providers
}

func (r *gormAppsTxRepository) Metrics() repository.UsageMetricRepository {
	return r.metrics
}
