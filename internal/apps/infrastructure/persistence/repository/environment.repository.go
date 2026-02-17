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

type environmentRepository struct {
	db *gorm.DB
}

func NewEnvironmentRepository(db *gorm.DB) repository.EnvironmentRepository {
	return &environmentRepository{db: db}
}

func (r *environmentRepository) Create(ctx context.Context, env *entity.Environment) (*entity.Environment, error) {
	m := mapper.ToEnvironmentModel(env)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return mapper.ToEnvironmentEntity(m), nil
}

func (r *environmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Environment, error) {
	var m model.EnvironmentModel
	if err := r.db.WithContext(ctx).Preload("Keys").Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToEnvironmentEntity(&m), nil
}

func (r *environmentRepository) GetByAppAndCode(ctx context.Context, appID uuid.UUID, code string) (*entity.Environment, error) {
	var m model.EnvironmentModel
	if err := r.db.WithContext(ctx).Preload("Keys").Where("app_id = ? AND environment_code = ?", appID, code).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToEnvironmentEntity(&m), nil
}

func (r *environmentRepository) ListByApp(ctx context.Context, appID uuid.UUID) ([]*entity.Environment, error) {
	var models []model.EnvironmentModel
	if err := r.db.WithContext(ctx).Preload("Keys").Where("app_id = ?", appID).Find(&models).Error; err != nil {
		return nil, err
	}
	return mapper.ToEnvironmentEntities(models), nil
}

func (r *environmentRepository) GetByAPIKey(ctx context.Context, apiKey string) (*entity.Environment, error) {
	var m model.EnvironmentModel
	if err := r.db.WithContext(ctx).Where("api_key = ?", apiKey).First(&m).Error; err != nil {
		return nil, err
	}
	return mapper.ToEnvironmentEntity(&m), nil
}

func (r *environmentRepository) Update(ctx context.Context, env *entity.Environment) error {
	m := mapper.ToEnvironmentModel(env)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *environmentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.EnvironmentModel{}, "id = ?", id).Error
}
