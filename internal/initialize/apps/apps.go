package apps

import (
	"CONVERDA/global"
	"CONVERDA/internal/apps/application/service/impl"
	"CONVERDA/internal/apps/controller"
	"CONVERDA/internal/apps/infrastructure/persistence/repository"
	iamRepository "CONVERDA/internal/iam/infrastructure/persistence/repository"
	"CONVERDA/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func InitAppsModule(router *gin.RouterGroup, db *gorm.DB, authMiddleware gin.HandlerFunc) {

	// Repositories (Apps)
	appRepo := repository.NewAppRepository(db)
	envRepo := repository.NewEnvironmentRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)
	providerRepo := repository.NewProviderRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)
	metricsRepo := repository.NewUsageMetricRepository(db)
	systemEnvRepo := repository.NewSystemEnvironmentRepository(db)
	appsUoW := repository.NewAppsUnitOfWork(db)

	// Repositories (IAM - for membership checking)
	memberRepo := iamRepository.NewTenantMemberRepository(db)
	tenantRepo := iamRepository.NewTenantRepository(db)

	// Services
	systemEnvService := impl.NewSystemEnvironmentService(systemEnvRepo)
	metricsService := impl.NewMetricsService(metricsRepo)
	envService := impl.NewEnvironmentService(envRepo, appsUoW)
	apiKeyService := impl.NewAPIKeyService(apiKeyRepo, envRepo, appsUoW)

	// AppService depends on SystemEnv and APIKey services
	appService := impl.NewAppService(appRepo, envRepo, systemEnvService, apiKeyService, appsUoW)

	webhookService := impl.NewWebhookService(webhookRepo)
	providerService := impl.NewProviderService(providerRepo)

	// Middleware Adapters
	membershipChecker := middleware.NewMembershipCheckerAdapter(memberRepo, tenantRepo)
	tenantMembershipMiddleware := middleware.TenantMembershipMiddleware(membershipChecker)

	// Controllers
	appController := controller.NewAppController(appService, envService, apiKeyService, metricsService, systemEnvService)
	webhookController := controller.NewWebhookController(webhookService)
	providerController := controller.NewProviderController(providerService)

	// Register Routes
	controller.RegisterAppsRoutes(router, appController, webhookController, providerController, authMiddleware, tenantMembershipMiddleware)

	global.Logger.Info("Apps module initialized")
}
