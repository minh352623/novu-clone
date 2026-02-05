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

type providerRepository struct {
	db *gorm.DB
}

func NewProviderRepository(db *gorm.DB) repository.ProviderRepository {
	return &providerRepository{db: db}
}

func (r *providerRepository) Create(ctx context.Context, provider *entity.Provider) (*entity.Provider, error) {
	m := mapper.ToProviderModel(provider)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return mapper.ToProviderEntity(m), nil
}

func (r *providerRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Provider, error) {
	var m model.ProviderModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return mapper.ToProviderEntity(&m), nil
}

func (r *providerRepository) ListByEnvironment(ctx context.Context, environmentID uuid.UUID) ([]*entity.Provider, error) {
	var models []model.ProviderModel
	if err := r.db.WithContext(ctx).Where("environment_id = ?", environmentID).Find(&models).Error; err != nil {
		return nil, err
	}
	return mapper.ToProviderEntities(models), nil
}

func (r *providerRepository) ListByApp(ctx context.Context, appID uuid.UUID) ([]*entity.Provider, error) {
	var models []model.ProviderModel
	if err := r.db.WithContext(ctx).Where("app_id = ?", appID).Find(&models).Error; err != nil {
		return nil, err
	}
	return mapper.ToProviderEntities(models), nil
}

func (r *providerRepository) GetByTypeAndEnv(ctx context.Context, environmentID uuid.UUID, providerType string) (*entity.Provider, error) {
	var m model.ProviderModel
	if err := r.db.WithContext(ctx).Where("environment_id = ? AND provider_type = ? AND is_active = true", environmentID, providerType).First(&m).Error; err != nil {
		return nil, err
	}
	return mapper.ToProviderEntity(&m), nil
}

func (r *providerRepository) Update(ctx context.Context, provider *entity.Provider) error {
	m := mapper.ToProviderModel(provider)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *providerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ProviderModel{}, "id = ?", id).Error
}
