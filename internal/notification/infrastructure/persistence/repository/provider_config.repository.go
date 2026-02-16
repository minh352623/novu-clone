package repository

import (
	"context"
	"errors"
	"fmt"

	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"
	"CONVERDA/internal/notification/infrastructure/persistence/mapper"
	"CONVERDA/internal/notification/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type providerConfigRepository struct {
	db *gorm.DB
}

func NewProviderConfigRepository(db *gorm.DB) repository.ProviderConfigRepository {
	return &providerConfigRepository{db: db}
}

func (r *providerConfigRepository) GetActive(ctx context.Context, envID uuid.UUID, providerType string) (*entity.ProviderConfig, error) {
	var pc model.ProviderConfigModel

	// Join with Providers table to filter by provider_type
	if err := r.db.WithContext(ctx).
		Joins("JOIN providers ON providers.id = provider_configs.provider_id").
		Where("provider_configs.environment_id = ? AND provider_configs.is_active = true AND providers.provider_type = ?", envID, providerType).
		Preload("Provider").
		First(&pc).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("no active provider config found for type %s in env %s", providerType, envID)
		}
		return nil, err
	}

	return mapper.ToProviderConfigDomain(&pc), nil
}
