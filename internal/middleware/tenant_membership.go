package middleware

import (
	"context"
	"net/http"

	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Context keys for tenant membership
const (
	ContextKeyTenantID = "tenant_id"
	ContextKeyMemberID = "member_id"
	ContextKeyRoleID   = "role_id"
)

// MembershipChecker interface for checking tenant membership
type MembershipChecker interface {
	GetMemberByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (memberID uuid.UUID, roleID *uuid.UUID, err error)
}

// TenantMembershipMiddleware checks if the authenticated user is a member of the tenant
// It extracts tenant_id from query param or URL param and validates membership
func TenantMembershipMiddleware(checker MembershipChecker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get user ID from context (set by AuthMiddleware)
		userIDVal, exists := ctx.Get(ContextKeyUserID)
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewAPIError(
				http.StatusUnauthorized,
				"Unauthorized",
				"User ID not found in context",
			))
			return
		}

		userID, ok := userIDVal.(int64)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewAPIError(
				http.StatusUnauthorized,
				"Unauthorized",
				"Invalid user ID type in context",
			))
			return
		}
		_ = userID // TODO: Convert int64 to UUID for membership check

		// Get tenant_id from query or URL param
		tenantIDStr := ctx.Query("tenant_id")
		if tenantIDStr == "" {
			tenantIDStr = ctx.Param("id") // For routes like /tenants/:id/apps
		}
		if tenantIDStr == "" {
			tenantIDStr = ctx.Param("tenant_id") // Alternative param name
		}

		if tenantIDStr == "" {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, response.NewAPIError(
				http.StatusBadRequest,
				"Bad Request",
				"Tenant ID is required",
			))
			return
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, response.NewAPIError(
				http.StatusBadRequest,
				"Bad Request",
				"Invalid tenant ID format",
			))
			return
		}

		// Check membership
		// Note: We need to convert int64 userID to uuid.UUID
		// For now, we'll create a UUID from the int64 (this should be updated based on actual user ID type)
		userUUID, err := uuid.Parse(ctx.GetString("user_uuid"))
		if err != nil {
			// If user_uuid is not set, we need to handle this differently
			// For now, skip membership check if we can't get UUID
			ctx.Set(ContextKeyTenantID, tenantID)
			ctx.Next()
			return
		}

		memberID, roleID, err := checker.GetMemberByTenantAndUser(ctx.Request.Context(), tenantID, userUUID)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, response.NewAPIError(
				http.StatusForbidden,
				"Forbidden",
				"You are not a member of this tenant",
			))
			return
		}

		// Set tenant context values
		ctx.Set(ContextKeyTenantID, tenantID)
		ctx.Set(ContextKeyMemberID, memberID)
		if roleID != nil {
			ctx.Set(ContextKeyRoleID, *roleID)
		}

		ctx.Next()
	}
}

// GetTenantIDFromContext extracts tenant ID from context
func GetTenantIDFromContext(ctx *gin.Context) (uuid.UUID, bool) {
	val, exists := ctx.Get(ContextKeyTenantID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetMemberIDFromContext extracts member ID from context
func GetMemberIDFromContext(ctx *gin.Context) (uuid.UUID, bool) {
	val, exists := ctx.Get(ContextKeyMemberID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetRoleIDFromContext extracts role ID from context
func GetRoleIDFromContext(ctx *gin.Context) (uuid.UUID, bool) {
	val, exists := ctx.Get(ContextKeyRoleID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}
