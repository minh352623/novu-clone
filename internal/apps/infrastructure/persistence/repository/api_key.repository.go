package repository

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/domain/repository"
	"CONVERDA/internal/apps/infrastructure/persistence/mapper"
	"CONVERDA/internal/apps/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type apiKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) repository.APIKeyRepository {
	return &apiKeyRepository{db: db}
}

func (r *apiKeyRepository) Create(ctx context.Context, apiKey *entity.APIKey) (*entity.APIKey, error) {
	m := mapper.ToAPIKeyModel(apiKey)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, fmt.Errorf("failed to create api key: %w", err)
	}
	return mapper.ToAPIKeyEntity(m), nil
}

func (r *apiKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.APIKey, error) {
	var m model.APIKeyModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get api key by id: %w", err)
	}
	return mapper.ToAPIKeyEntity(&m), nil
}

func (r *apiKeyRepository) GetByHash(ctx context.Context, hash string) (*entity.APIKey, error) {
	var m model.APIKeyModel
	if err := r.db.WithContext(ctx).Where("key_hash = ?", hash).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get api key by hash: %w", err)
	}
	return mapper.ToAPIKeyEntity(&m), nil
}

func (r *apiKeyRepository) GetByEnvironment(ctx context.Context, envID uuid.UUID) ([]*entity.APIKey, error) {
	var models []model.APIKeyModel
	if err := r.db.WithContext(ctx).Where("environment_id = ?", envID).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get api keys by environment: %w", err)
	}

	keys := make([]*entity.APIKey, len(models))
	for i, m := range models {
		keys[i] = mapper.ToAPIKeyEntity(&m)
	}
	return keys, nil
}

func (r *apiKeyRepository) Update(ctx context.Context, apiKey *entity.APIKey) error {
	m := mapper.ToAPIKeyModel(apiKey)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return fmt.Errorf("failed to update api key: %w", err)
	}
	return nil
}

func (r *apiKeyRepository) RevokeAllByEnvironment(ctx context.Context, envID uuid.UUID) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&model.APIKeyModel{}).
		Where("environment_id = ? AND revoked_at IS NULL", envID).
		Update("revoked_at", now).Error; err != nil {
		return fmt.Errorf("failed to revoke all api keys by environment: %w", err)
	}
	return nil
}
