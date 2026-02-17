package repository

import (
	"context"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type WebhookRepository interface {
	GetByEvent(ctx context.Context, envID uuid.UUID, eventType string) ([]*entity.Webhook, error)
}

type WebhookLogRepository interface {
	Create(ctx context.Context, log *entity.WebhookLog) error
	Update(ctx context.Context, log *entity.WebhookLog) error
	GetPendingRetries(ctx context.Context, limit int) ([]*entity.WebhookLog, error)
}
