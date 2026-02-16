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

type notificationGroupRepository struct {
	db *gorm.DB
}

func NewNotificationGroupRepository(db *gorm.DB) repository.NotificationGroupRepository {
	return &notificationGroupRepository{db: db}
}

func (r *notificationGroupRepository) Create(ctx context.Context, group *entity.NotificationGroup) error {
	m := mapper.ToGroupModel(group)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *notificationGroupRepository) Update(ctx context.Context, group *entity.NotificationGroup) error {
	m := mapper.ToGroupModel(group)
	return r.db.WithContext(ctx).Model(&model.NotificationGroupModel{}).Where("id = ?", group.ID).Updates(m).Error
}

func (r *notificationGroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.NotificationGroupModel{}, id).Error
}

func (r *notificationGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.NotificationGroup, error) {
	var m model.NotificationGroupModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Or custom Not Found error
		}
		return nil, err
	}
	return mapper.ToGroupDomain(&m), nil
}

func (r *notificationGroupRepository) GetByKey(ctx context.Context, envID uuid.UUID, key string) (*entity.NotificationGroup, error) {
	var m model.NotificationGroupModel
	if err := r.db.WithContext(ctx).Where("environment_id = ? AND key = ?", envID, key).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToGroupDomain(&m), nil
}

func (r *notificationGroupRepository) List(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationGroup, int64, error) {
	var models []*model.NotificationGroupModel
	var total int64

	db := r.db.WithContext(ctx).Model(&model.NotificationGroupModel{}).Where("environment_id = ?", envID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ToGroupDomainList(models), total, nil
}
