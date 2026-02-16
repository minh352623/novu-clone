package repository

import (
	"context"
	"errors"

	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"
	"CONVERDA/internal/notification/infrastructure/persistence/mapper"
	"CONVERDA/internal/notification/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type notificationLayoutRepository struct {
	db *gorm.DB
}

func NewNotificationLayoutRepository(db *gorm.DB) repository.NotificationLayoutRepository {
	return &notificationLayoutRepository{db: db}
}

func (r *notificationLayoutRepository) Create(ctx context.Context, layout *entity.NotificationLayout) error {
	m := mapper.ToLayoutModel(layout)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *notificationLayoutRepository) Update(ctx context.Context, layout *entity.NotificationLayout) error {
	m := mapper.ToLayoutModel(layout)
	return r.db.WithContext(ctx).Model(&model.NotificationLayoutModel{}).Where("id = ?", layout.ID).Updates(m).Error
}

func (r *notificationLayoutRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.NotificationLayoutModel{}, id).Error
}

func (r *notificationLayoutRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.NotificationLayout, error) {
	var m model.NotificationLayoutModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToLayoutDomain(&m), nil
}

func (r *notificationLayoutRepository) List(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationLayout, int64, error) {
	var models []*model.NotificationLayoutModel
	var total int64

	db := r.db.WithContext(ctx).Model(&model.NotificationLayoutModel{}).Where("environment_id = ?", envID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ToLayoutDomainList(models), total, nil
}

func (r *notificationLayoutRepository) GetDefault(ctx context.Context, envID uuid.UUID) (*entity.NotificationLayout, error) {
	var m model.NotificationLayoutModel
	if err := r.db.WithContext(ctx).Where("environment_id = ? AND is_default = ?", envID, true).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToLayoutDomain(&m), nil
}
