package service

import (
	"context"

	"github.com/google/uuid"
)

type WebhookDispatcher interface {
	Dispatch(ctx context.Context, envID uuid.UUID, eventType string, payload interface{})
}
