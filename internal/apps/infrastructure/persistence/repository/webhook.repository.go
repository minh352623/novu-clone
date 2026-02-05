package repository

import (
	"context"

	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/domain/repository"
	"CONVERDA/internal/apps/infrastructure/persistence/mapper"
	"CONVERDA/internal/apps/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type webhookRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) repository.WebhookRepository {
	return &webhookRepository{db: db}
}

func (r *webhookRepository) Create(ctx context.Context, webhook *entity.Webhook) (*entity.Webhook, error) {
	m := mapper.ToWebhookModel(webhook)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return mapper.ToWebhookEntity(m), nil
}

func (r *webhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Webhook, error) {
	var m model.WebhookModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return mapper.ToWebhookEntity(&m), nil
}

func (r *webhookRepository) ListByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*entity.Webhook, error) {
	var models []model.WebhookModel
	if err := r.db.WithContext(ctx).Where("environment_id = ?", environmentID).Find(&models).Error; err != nil {
		return nil, err
	}
	return mapper.ToWebhookEntities(models), nil
}

func (r *webhookRepository) ListByApp(ctx context.Context, appID uuid.UUID) ([]*entity.Webhook, error) {
	var models []model.WebhookModel
	if err := r.db.WithContext(ctx).Where("app_id = ?", appID).Find(&models).Error; err != nil {
		return nil, err
	}
	return mapper.ToWebhookEntities(models), nil
}

func (r *webhookRepository) Update(ctx context.Context, webhook *entity.Webhook) error {
	m := mapper.ToWebhookModel(webhook)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *webhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.WebhookModel{}, "id = ?", id).Error
}
