package repository

import (
	"context"
	"time"

	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"
	"CONVERDA/internal/notification/infrastructure/persistence/mapper"
	"CONVERDA/internal/notification/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type webhookRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) repository.WebhookRepository {
	return &webhookRepository{db: db}
}

func (r *webhookRepository) GetByEvent(ctx context.Context, envID uuid.UUID, eventType string) ([]*entity.Webhook, error) {
	var models []*model.WebhookModel
	// Find active webhooks in the environment that contain the eventType in their events array
	// JSONB query: events @> '["eventType"]'
	err := r.db.WithContext(ctx).
		Where("environment_id = ? AND is_active = ?", envID, true).
		Where("events @> ?", "\""+eventType+"\""). // Simple JSONB contains check for string array
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	var parameters []*entity.Webhook
	for _, m := range models {
		parameters = append(parameters, mapper.ToWebhookDomain(m))
	}
	return parameters, nil
}

type webhookLogRepository struct {
	db *gorm.DB
}

func NewWebhookLogRepository(db *gorm.DB) repository.WebhookLogRepository {
	return &webhookLogRepository{db: db}
}

func (r *webhookLogRepository) Create(ctx context.Context, log *entity.WebhookLog) error {
	m := mapper.ToWebhookLogModel(log)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *webhookLogRepository) Update(ctx context.Context, log *entity.WebhookLog) error {
	m := mapper.ToWebhookLogModel(log)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *webhookLogRepository) GetPendingRetries(ctx context.Context, limit int) ([]*entity.WebhookLog, error) {
	var models []*model.WebhookLogModel
	err := r.db.WithContext(ctx).
		Where("status = ? AND retry_count < max_retries AND next_retry_at <= ?", entity.WebhookLogStatusFailed, time.Now()).
		Order("next_retry_at ASC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	var logs []*entity.WebhookLog
	for _, m := range models {
		logs = append(logs, mapper.ToWebhookLogDomain(m))
	}
	return logs, nil
}
