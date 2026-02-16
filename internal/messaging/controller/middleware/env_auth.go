package middleware

import (
	"net/http"

	domainRepo "CONVERDA/internal/messaging/domain/repository"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
)

func EnvKeyAuth(repo domainRepo.EnvironmentAuthRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.APIResponse{
				Code:    http.StatusUnauthorized,
				Message: "Missing X-API-Key header",
			})
			return
		}

		envID, tenantID, err := repo.ValidateKey(c.Request.Context(), apiKey)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.APIResponse{
				Code:    http.StatusUnauthorized,
				Message: "Invalid API Key",
			})
			return
		}

		// Inject into context
		c.Set("environment_id", envID)
		c.Set("tenant_id", tenantID)
		c.Next()
	}
}
