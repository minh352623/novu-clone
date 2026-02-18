package middleware

import (
	"context"
	"net/http"

	"CONVERDA/global"
	"CONVERDA/internal/apps/application/service"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	ContextKeyAppID = "app_id"
	ContextKeyEnvID = "env_id"
)

// APIKeyMiddleware validates API Key and records usage
func APIKeyMiddleware(apiKeyService service.APIKeyService, metricsService service.MetricsService, providerType string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		apiKey := ctx.GetHeader("X-API-Key")
		if apiKey == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewAPIError(
				http.StatusUnauthorized,
				"Invalid API Key",
				"X-API-Key header is required",
			))
			return
		}

		// Validate Key
		key, err := apiKeyService.ValidateKey(ctx.Request.Context(), apiKey)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewAPIError(
				http.StatusUnauthorized,
				"Invalid API Key",
				err.Error(),
			))
			return
		}

		// Set context
		ctx.Set(ContextKeyEnvID, key.EnvironmentID)
		ctx.Set(ContextKeyAppID, key.AppID)

		ctx.Next()

		// Record usage after request
		go func() {
			defer func() {
				if r := recover(); r != nil {
					global.Logger.Error("api_key_middleware: panic recovered in RecordUsage", zap.Any("panic", r))
				}
			}()

			// Use a background context as the request context may be cancelled
			// In a real production app, you might want to use a more robust queue/buffer
			_ = metricsService.RecordUsage(context.Background(), key.AppID, key.EnvironmentID, providerType, "outbound", ctx.Writer.Status())
		}()
	}
}
