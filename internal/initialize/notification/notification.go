package notification

import (
	"context"

	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/application/service/impl"
	"CONVERDA/internal/notification/application/worker"
	"CONVERDA/internal/notification/controller"
	"CONVERDA/internal/notification/infrastructure/adapter"
	"CONVERDA/internal/notification/infrastructure/persistence/repository"
	"CONVERDA/internal/notification/infrastructure/provider"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

var (
	NotificationService service.NotificationService
)

// InitNotificationModule initializes the Notification module
func InitNotificationModule(db *gorm.DB, router *gin.RouterGroup) {
	// 1. Repositories (Postgres)
	notifRepo := repository.NewNotificationRepository(db)
	tmplRepo := repository.NewTemplateRepository(db)
	configRepo := repository.NewProviderConfigRepository(db)
	groupRepo := repository.NewNotificationGroupRepository(db)
	layoutRepo := repository.NewNotificationLayoutRepository(db)
	jobRepo := repository.NewNotificationJobRepository(db)

	// 2. Domain Services
	notifUoW := repository.NewNotificationUnitOfWork(db)
	tmplManager := service.NewTemplateManager(tmplRepo, layoutRepo)
	dispatcher := provider.NewDispatcher(configRepo)
	dispatcherAdapter := adapter.NewDispatcherAdapter(dispatcher)
	groupManager := impl.NewGroupManager(groupRepo, notifUoW)
	layoutManager := impl.NewLayoutManager(layoutRepo, notifUoW)

	// 3. Application Service
	NotificationService = impl.NewNotificationService(notifRepo, notifUoW, tmplManager, dispatcherAdapter)

	// Now init JobScheduler
	jobScheduler := impl.NewJobScheduler(jobRepo, notifUoW, NotificationService, tmplManager)

	// 4. Webhook Dispatcher + Retry Worker
	webhookRepo := repository.NewWebhookRepository(db)
	webhookLogRepo := repository.NewWebhookLogRepository(db)
	webhookDispatcher := impl.NewWebhookDispatcher(webhookRepo, webhookLogRepo)

	retryWorker := worker.NewWebhookRetryWorker(webhookDispatcher, webhookLogRepo)
	go retryWorker.Run(context.Background())

	// 5. Register Routes
	controller.RegisterRoutes(router, NotificationService, groupManager, layoutManager, jobScheduler, tmplManager)
}
