package apps

import (
	"CONVERDA/global"
	"CONVERDA/internal/apps/application/service/impl"
	"CONVERDA/internal/apps/controller"
	"CONVERDA/internal/apps/infrastructure/persistence/repository"
	iamRepository "CONVERDA/internal/iam/infrastructure/persistence/repository"
	"CONVERDA/internal/middleware"

	"github.com/gin-gonic/gin"
)

func InitAppsModule(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	db := global.GormDB

	// Repositories (Apps)
	appRepo := repository.NewAppRepository(db)
	envRepo := repository.NewEnvironmentRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)
	providerRepo := repository.NewProviderRepository(db)

	// Repositories (IAM - for membership checking)
	memberRepo := iamRepository.NewTenantMemberRepository(db)

	// Services
	appService := impl.NewAppService(appRepo, envRepo)
	envService := impl.NewEnvironmentService(envRepo)
	webhookService := impl.NewWebhookService(webhookRepo)
	providerService := impl.NewProviderService(providerRepo)

	// Middleware Adapters
	membershipChecker := middleware.NewMembershipCheckerAdapter(memberRepo)
	tenantMembershipMiddleware := middleware.TenantMembershipMiddleware(membershipChecker)

	// Controllers
	appController := controller.NewAppController(appService, envService)
	webhookController := controller.NewWebhookController(webhookService)
	providerController := controller.NewProviderController(providerService)

	// Register Routes
	controller.RegisterAppsRoutes(router, appController, webhookController, providerController, authMiddleware, tenantMembershipMiddleware)

	global.Logger.Info("Apps module initialized")
}
