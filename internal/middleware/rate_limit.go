package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"CONVERDA/internal/apps/application/service"
	"CONVERDA/pkg/ratelimit"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RateLimitMiddleware enforces per-environment rate limits.
// Must be placed AFTER APIKeyMiddleware (needs env_id in context).
func RateLimitMiddleware(envService service.EnvironmentService, limiter ratelimit.RateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		envIDVal, exists := ctx.Get(ContextKeyEnvID)
		if !exists {
			// No env_id in context — skip rate limiting (e.g., non-API-key routes)
			ctx.Next()
			return
		}

		envID, ok := envIDVal.(uuid.UUID)
		if !ok {
			ctx.Next()
			return
		}

		// Look up environment config
		env, err := envService.GetByID(ctx.Request.Context(), envID)
		if err != nil || env == nil {
			// Environment not found — skip rate limiting, let downstream handle
			ctx.Next()
			return
		}

		// Check RPM (requests per minute)
		if env.RateLimitRPM > 0 {
			rpmKey := fmt.Sprintf("rpm:%s", envID.String())
			if !limiter.Allow(rpmKey, env.RateLimitRPM, 1*time.Minute) {
				remaining := limiter.Remaining(rpmKey, env.RateLimitRPM, 1*time.Minute)
				ctx.Header("X-RateLimit-Limit", strconv.Itoa(env.RateLimitRPM))
				ctx.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
				ctx.Header("Retry-After", "60")
				ctx.AbortWithStatusJSON(http.StatusTooManyRequests, response.NewAPIError(
					http.StatusTooManyRequests,
					"Rate limit exceeded",
					fmt.Sprintf("RPM limit of %d exceeded, retry after 60 seconds", env.RateLimitRPM),
				))
				return
			}

			// Set rate limit headers on success
			remaining := limiter.Remaining(rpmKey, env.RateLimitRPM, 1*time.Minute)
			ctx.Header("X-RateLimit-Limit", strconv.Itoa(env.RateLimitRPM))
			ctx.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		}

		// Check daily limit
		if env.RateLimitDaily > 0 {
			dailyKey := fmt.Sprintf("daily:%s", envID.String())
			if !limiter.Allow(dailyKey, env.RateLimitDaily, 24*time.Hour) {
				ctx.Header("Retry-After", "3600")
				ctx.AbortWithStatusJSON(http.StatusTooManyRequests, response.NewAPIError(
					http.StatusTooManyRequests,
					"Daily rate limit exceeded",
					fmt.Sprintf("Daily limit of %d exceeded", env.RateLimitDaily),
				))
				return
			}
		}

		ctx.Next()
	}
}
