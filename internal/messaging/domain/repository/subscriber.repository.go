package repository

import (
	"context"

	"CONVERDA/internal/messaging/domain/model/entity"

	"github.com/google/uuid"
)

type SubscriberRepository interface {
	Create(ctx context.Context, sub *entity.Subscriber) (*entity.Subscriber, error)
	GetByKey(ctx context.Context, envID uuid.UUID, key string) (*entity.Subscriber, error)
	Update(ctx context.Context, sub *entity.Subscriber) error
}
