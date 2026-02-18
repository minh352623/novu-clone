package health

import (
	"CONVERDA/internal/health/controller"
	healthDomain "CONVERDA/internal/health/domain"
	healthRepo "CONVERDA/internal/health/infrastructure/repository"
	"CONVERDA/internal/health/service/impl"
	"CONVERDA/pkg/response"

	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InitHealthModule initializes the shared Health module.
func InitHealthModule(db *gorm.DB, router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// Repository Adapters
	threadAdapter := healthRepo.NewThreadHealthAdapter(db)
	slaAdapter := healthRepo.NewSLAHealthAdapter(db)
	webhookAdapter := healthRepo.NewWebhookHealthAdapter(db)

	// Service
	healthService := impl.NewHealthService(threadAdapter, slaAdapter, webhookAdapter, healthDomain.DefaultSLAThreshold)

	// Controller
	healthController := controller.NewHealthDashboardController(healthService)

	// Register route under existing /dashboards group
	dashboardGroup := router.Group("/dashboards")
	if authMiddleware != nil {
		dashboardGroup.Use(authMiddleware)
	}
	{
		dashboardGroup.GET("/health", response.Wrap(healthController.GetSystemHealth, http.StatusOK))
	}
}
