package repository

import (
	"context"
	"fmt"

	"CONVERDA/internal/apps/domain/model/entity"
	"CONVERDA/internal/apps/domain/repository"
	"CONVERDA/internal/apps/infrastructure/persistence/mapper"
	"CONVERDA/internal/apps/infrastructure/persistence/model"

	"gorm.io/gorm"
)

type systemEnvironmentRepository struct {
	db *gorm.DB
}

func NewSystemEnvironmentRepository(db *gorm.DB) repository.SystemEnvironmentRepository {
	return &systemEnvironmentRepository{db: db}
}

func (r *systemEnvironmentRepository) ListAll(ctx context.Context) ([]*entity.SystemEnvironment, error) {
	var models []model.SystemEnvironmentModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to list system environments: %w", err)
	}
	return mapper.ToSystemEnvironmentEntities(models), nil
}

func (r *systemEnvironmentRepository) GetByCode(ctx context.Context, code string) (*entity.SystemEnvironment, error) {
	var m model.SystemEnvironmentModel
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get system environment: %w", err)
	}
	return mapper.ToSystemEnvironmentEntity(&m), nil
}
