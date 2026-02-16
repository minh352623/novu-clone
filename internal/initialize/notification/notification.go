package notification

import (
	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/application/service/impl"
	"CONVERDA/internal/notification/controller"
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
	tmplManager := service.NewTemplateManager(tmplRepo, layoutRepo)
	dispatcher := provider.NewDispatcher(configRepo)
	groupManager := impl.NewGroupManager(groupRepo)
	layoutManager := impl.NewLayoutManager(layoutRepo)
	jobScheduler := impl.NewJobScheduler(jobRepo, NotificationService, tmplManager) // Need NotificationService here, but NotificationService is init below?
	// Circular dependency? JobScheduler needs NotificationService. NotificationService needs TmplManager.
	// NotificationService = impl.NewNotificationService(notifRepo, tmplManager, dispatcher)
	// JobScheduler depends on NotificationService to Send.

	// We must init NotificationService BEFORE JobScheduler.

	// 3. Application Service
	NotificationService = impl.NewNotificationService(notifRepo, tmplManager, dispatcher)

	// Now init JobScheduler
	jobScheduler = impl.NewJobScheduler(jobRepo, NotificationService, tmplManager)

	// Init Webhook Dispatcher
	webhookRepo := repository.NewWebhookRepository(db)
	webhookLogRepo := repository.NewWebhookLogRepository(db)
	webhookDispatcher := impl.NewWebhookDispatcher(webhookRepo, webhookLogRepo)

	// (Optional) We might want to pass webhookDispatcher to controller or export it globally like NotificationService
	// For now, just initializing it is enough for the module pattern, or if we want to expose it via controller later.
	// But `webhookDispatcher` is a service that other modules might use.
	// Since we don't have a specific controller for webhooks yet (management is separate, triggering is internal),
	// we will leave it as is, or assign to a package level var if needed.
	// Let's create a global variable for it too, similar to NotificationService, for easier access if needed.
	_ = webhookDispatcher // to avoid unused variable error until we use it

	// 4. Register Routes
	controller.RegisterRoutes(router, NotificationService, groupManager, layoutManager, jobScheduler)
}
