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

type notificationJobRepository struct {
	db *gorm.DB
}

func NewNotificationJobRepository(db *gorm.DB) repository.NotificationJobRepository {
	return &notificationJobRepository{db: db}
}

func (r *notificationJobRepository) Create(ctx context.Context, job *entity.NotificationJob) error {
	m := mapper.ToJobModel(job)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *notificationJobRepository) Update(ctx context.Context, job *entity.NotificationJob) error {
	m := mapper.ToJobModel(job)
	return r.db.WithContext(ctx).Model(&model.NotificationJobModel{}).Where("id = ?", job.ID).Updates(m).Error
}

func (r *notificationJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.NotificationJobModel{}, id).Error
}

func (r *notificationJobRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.NotificationJob, error) {
	var m model.NotificationJobModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return mapper.ToJobDomain(&m), nil
}

func (r *notificationJobRepository) List(ctx context.Context, envID uuid.UUID, limit, offset int) ([]*entity.NotificationJob, int64, error) {
	var models []*model.NotificationJobModel
	var total int64

	db := r.db.WithContext(ctx).Model(&model.NotificationJobModel{}).Where("environment_id = ?", envID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Limit(limit).Offset(offset).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ToJobDomainList(models), total, nil
}

func (r *notificationJobRepository) GetPendingJobs(ctx context.Context) ([]*entity.NotificationJob, error) {
	var models []*model.NotificationJobModel
	if err := r.db.WithContext(ctx).Where("status IN ?", []string{"pending", "scheduled"}).Order("scheduled_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	return mapper.ToJobDomainList(models), nil
}
