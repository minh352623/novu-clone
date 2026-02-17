package controller

import (
	"net/http"

	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
)

// RegisterAppsRoutes registers all Apps module routes
func RegisterAppsRoutes(
	router *gin.RouterGroup,
	appController *AppController,
	webhookController *WebhookController,
	providerController *ProviderController,
	authMiddleware gin.HandlerFunc,
	tenantMembershipMiddleware gin.HandlerFunc,
) {
	// Tenant scoped routes (require auth + membership check)
	tenants := router.Group("/tenants")
	if authMiddleware != nil {
		tenants.Use(authMiddleware)
	}
	if tenantMembershipMiddleware != nil {
		tenants.Use(tenantMembershipMiddleware)
	}
	{
		tenants.POST("/:id/apps", response.Wrap(appController.CreateApp, http.StatusCreated))
		tenants.GET("/:id/apps", response.Wrap(appController.ListApps, http.StatusOK))
	}

	// App scoped routes
	apps := router.Group("/apps")
	if authMiddleware != nil {
		apps.Use(authMiddleware)
	}
	{
		apps.GET("/:app_id", response.Wrap(appController.GetApp, http.StatusOK))
		apps.PUT("/:app_id", response.Wrap(appController.UpdateApp, http.StatusOK))
		apps.DELETE("/:app_id", response.Wrap(appController.DeleteApp, http.StatusOK))
		apps.GET("/:app_id/metrics", response.Wrap(appController.GetAppMetrics, http.StatusOK))
		apps.GET("/:app_id/metrics/detailed", response.Wrap(appController.GetDetailedMetrics, http.StatusOK))
		apps.GET("/:app_id/metrics/timeseries", response.Wrap(appController.GetDailyTimeSeries, http.StatusOK))

		// Environments
		apps.POST("/:app_id/environments", response.Wrap(appController.CreateEnvironment, http.StatusCreated))
		apps.GET("/:app_id/environments", response.Wrap(appController.ListEnvironments, http.StatusOK))
		apps.PUT("/:app_id/environments/:env_id", response.Wrap(appController.UpdateEnvironmentConfig, http.StatusOK))

		// Webhooks
		apps.POST("/:app_id/webhooks", response.Wrap(webhookController.CreateWebhook, http.StatusCreated))
		apps.GET("/:app_id/webhooks", response.Wrap(webhookController.ListWebhooks, http.StatusOK))

		// Providers
		apps.POST("/:app_id/providers", response.Wrap(providerController.CreateProvider, http.StatusCreated))
		apps.GET("/:app_id/providers", response.Wrap(providerController.ListProviders, http.StatusOK))

		// API Keys (Environment scoped)
		envGroup := router.Group("/environments")
		if authMiddleware != nil {
			envGroup.Use(authMiddleware)
		}
		{
			envGroup.GET("/:id/api-keys", response.Wrap(appController.ListAPIKeys, http.StatusOK))
			envGroup.POST("/:id/api-keys/rotate", response.Wrap(appController.RotateAPIKey, http.StatusOK))
		}

		apiKeyGroup := router.Group("/api-keys")
		if authMiddleware != nil {
			apiKeyGroup.Use(authMiddleware)
		}
		{
			apiKeyGroup.POST("/:key_id/revoke", response.Wrap(appController.RevokeAPIKey, http.StatusOK))
		}
	}

	// Webhook routes
	webhooks := router.Group("/webhooks")
	if authMiddleware != nil {
		webhooks.Use(authMiddleware)
	}
	{
		webhooks.GET("/:webhook_id", response.Wrap(webhookController.GetWebhook, http.StatusOK))
		webhooks.PUT("/:webhook_id", response.Wrap(webhookController.UpdateWebhook, http.StatusOK))
		webhooks.DELETE("/:webhook_id", response.Wrap(webhookController.DeleteWebhook, http.StatusOK))
	}

	// Provider routes
	providers := router.Group("/providers")
	if authMiddleware != nil {
		providers.Use(authMiddleware)
	}
	{
		providers.GET("/:provider_id", response.Wrap(providerController.GetProvider, http.StatusOK))
		providers.PUT("/:provider_id", response.Wrap(providerController.UpdateProvider, http.StatusOK))
		providers.DELETE("/:provider_id", response.Wrap(providerController.DeleteProvider, http.StatusOK))
	}

	// System routes
	system := router.Group("/system")
	if authMiddleware != nil {
		system.Use(authMiddleware)
	}
	{
		system.GET("/environments", response.Wrap(appController.ListSystemEnvironments, http.StatusOK))
	}
}
