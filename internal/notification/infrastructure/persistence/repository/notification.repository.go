package repository

import (
	"context"
	"fmt"

	"CONVERDA/internal/notification/domain/entity"
	"CONVERDA/internal/notification/domain/repository"
	"CONVERDA/internal/notification/infrastructure/persistence/mapper"
	"CONVERDA/internal/notification/infrastructure/persistence/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db     *gorm.DB
	mapper *mapper.NotificationMapper
}

func NewNotificationRepository(db *gorm.DB) repository.NotificationRepository {
	return &notificationRepository{
		db:     db,
		mapper: mapper.NewNotificationMapper(),
	}
}

func (r *notificationRepository) Create(ctx context.Context, notif *entity.Notification) error {
	m := r.mapper.ToModel(notif)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

func (r *notificationRepository) Update(ctx context.Context, notif *entity.Notification) error {
	m := r.mapper.ToModel(notif)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return fmt.Errorf("failed to update notification: %w", err)
	}
	return nil
}

func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error) {
	var m model.NotificationModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return r.mapper.ToDomain(&m), nil
}
