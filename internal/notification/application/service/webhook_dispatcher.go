package service

import (
	"context"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type WebhookDispatcher interface {
	Dispatch(ctx context.Context, envID uuid.UUID, eventType string, payload interface{})
	RetryWebhook(ctx context.Context, log *entity.WebhookLog)
}
