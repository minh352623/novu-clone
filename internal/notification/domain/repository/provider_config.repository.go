package repository

import (
	"context"

	"CONVERDA/internal/notification/domain/entity"

	"github.com/google/uuid"
)

type ProviderConfigRepository interface {
	GetActive(ctx context.Context, envID uuid.UUID, providerType string) (*entity.ProviderConfig, error)
}
