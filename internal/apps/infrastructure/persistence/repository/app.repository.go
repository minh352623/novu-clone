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

type appRepository struct {
	db *gorm.DB
}

func NewAppRepository(db *gorm.DB) repository.AppRepository {
	return &appRepository{db: db}
}

func (r *appRepository) Create(ctx context.Context, app *entity.App) (*entity.App, error) {
	m := mapper.ToAppModel(app)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return mapper.ToAppEntity(m), nil
}

func (r *appRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.App, error) {
	var m model.AppModel
	if err := r.db.WithContext(ctx).Preload("Environments").First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return mapper.ToAppEntity(&m), nil
}

func (r *appRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*entity.App, error) {
	var models []model.AppModel
	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&models).Error; err != nil {
		return nil, err
	}
	return mapper.ToAppEntities(models), nil
}

func (r *appRepository) Update(ctx context.Context, app *entity.App) error {
	m := mapper.ToAppModel(app)
	return r.db.WithContext(ctx).Save(m).Error
}

func (r *appRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.AppModel{}, "id = ?", id).Error
}
