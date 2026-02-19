package repository

import (
	"context"
)

// NotificationTxRepository defines the set of repositories accessible within a transaction.
type NotificationTxRepository interface {
	Notifications() NotificationRepository
	Jobs() NotificationJobRepository
	Groups() NotificationGroupRepository
	Layouts() NotificationLayoutRepository
	Templates() TemplateRepository
	ProviderConfigs() ProviderConfigRepository
	Webhooks() WebhookRepository
	WebhookLogs() WebhookLogRepository
}

// NotificationUnitOfWork defines the interface for managing transactions in the Notification module.
type NotificationUnitOfWork interface {
	Execute(ctx context.Context, fn func(tx NotificationTxRepository) error) error
}
