package middleware

import (
	"context"
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

		// Tenant-level RPM from plan (set by auth middleware from JWT claims)
		if planVal, exists := ctx.Get(ContextKeyPlan); exists {
			if planSlug, ok := planVal.(string); ok && planSlug != "" {
				// Use tenant-scoped key for plan-level rate limiting
				if tenantVal, exists := ctx.Get(ContextKeyTenantIDFromJWT); exists {
					if tenantID, ok := tenantVal.(uuid.UUID); ok {
						tenantRPMKey := fmt.Sprintf("tenant_rpm:%s", tenantID.String())
						// Plan-level RPM is enforced if set (checked by caller or looked up)
						// We use a conservative default; actual enforcement comes from env config below
						ctx.Set("tenant_rate_key", tenantRPMKey)
					}
				}
			}
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

// TenantRateLimitMiddleware enforces tenant-level rate limits based on the pricing plan.
// Must be placed AFTER AuthMiddleware + TenantMembershipMiddleware.
func TenantRateLimitMiddleware(limiter ratelimit.RateLimiter, planRPMProvider TenantPlanRPMProvider) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tenantIDVal, exists := ctx.Get(ContextKeyTenantIDFromJWT)
		if !exists {
			ctx.Next()
			return
		}

		tenantID, ok := tenantIDVal.(uuid.UUID)
		if !ok {
			ctx.Next()
			return
		}

		// Get plan RPM for this tenant
		rpm := planRPMProvider.GetTenantPlanRPM(ctx.Request.Context(), tenantID)
		if rpm <= 0 {
			ctx.Next()
			return
		}

		tenantKey := fmt.Sprintf("tenant_rpm:%s", tenantID.String())
		if !limiter.Allow(tenantKey, rpm, 1*time.Minute) {
			remaining := limiter.Remaining(tenantKey, rpm, 1*time.Minute)
			ctx.Header("X-Tenant-RateLimit-Limit", strconv.Itoa(rpm))
			ctx.Header("X-Tenant-RateLimit-Remaining", strconv.Itoa(remaining))
			ctx.Header("Retry-After", "60")
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, response.NewAPIError(
				http.StatusTooManyRequests,
				"Tenant rate limit exceeded",
				fmt.Sprintf("Tenant RPM limit of %d exceeded", rpm),
			))
			return
		}

		remaining := limiter.Remaining(tenantKey, rpm, 1*time.Minute)
		ctx.Header("X-Tenant-RateLimit-Limit", strconv.Itoa(rpm))
		ctx.Header("X-Tenant-RateLimit-Remaining", strconv.Itoa(remaining))

		ctx.Next()
	}
}

// TenantPlanRPMProvider looks up the rate limit RPM for a tenant's plan.
type TenantPlanRPMProvider interface {
	GetTenantPlanRPM(ctx context.Context, tenantID uuid.UUID) int
}
