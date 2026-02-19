package repository

import (
	"context"
	"fmt"

	"CONVERDA/internal/notification/domain/repository"

	"gorm.io/gorm"
)

type notificationTxRepository struct {
	db *gorm.DB
}

func (r *notificationTxRepository) Notifications() repository.NotificationRepository {
	return NewNotificationRepository(r.db)
}

func (r *notificationTxRepository) Jobs() repository.NotificationJobRepository {
	return NewNotificationJobRepository(r.db)
}

func (r *notificationTxRepository) Groups() repository.NotificationGroupRepository {
	return NewNotificationGroupRepository(r.db)
}

func (r *notificationTxRepository) Layouts() repository.NotificationLayoutRepository {
	return NewNotificationLayoutRepository(r.db)
}

func (r *notificationTxRepository) Templates() repository.TemplateRepository {
	return NewTemplateRepository(r.db)
}

func (r *notificationTxRepository) ProviderConfigs() repository.ProviderConfigRepository {
	return NewProviderConfigRepository(r.db)
}

func (r *notificationTxRepository) Webhooks() repository.WebhookRepository {
	return NewWebhookRepository(r.db)
}

func (r *notificationTxRepository) WebhookLogs() repository.WebhookLogRepository {
	return NewWebhookLogRepository(r.db)
}

type notificationUnitOfWork struct {
	db *gorm.DB
}

func NewNotificationUnitOfWork(db *gorm.DB) repository.NotificationUnitOfWork {
	return &notificationUnitOfWork{db: db}
}

func (u *notificationUnitOfWork) Execute(ctx context.Context, fn func(tx repository.NotificationTxRepository) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &notificationTxRepository{db: tx}
		if err := fn(txRepo); err != nil {
			return fmt.Errorf("transaction failed: %w", err)
		}
		return nil
	})
}
