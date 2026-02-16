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
	// GetMemberByTenantUserAndApp checks membership for a specific app (or tenant-wide if appID is nil).
	GetMemberByTenantUserAndApp(ctx context.Context, tenantID, userID uuid.UUID, appID *uuid.UUID) (memberID uuid.UUID, roleID *uuid.UUID, err error)
}

// TenantMembershipMiddleware checks if the authenticated user is a member of the tenant
// It extracts tenant_id (and optional app_id) from query param or URL param and validates membership
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

		// Try to get UUID directly if available (best effort)
		var userUUID uuid.UUID
		var err error

		userUUIDStr := ctx.GetString("user_uuid")
		if userUUIDStr != "" {
			userUUID, err = uuid.Parse(userUUIDStr)
		} else {
			// Fallback: Check if userIDVal is already UUID
			if id, ok := userIDVal.(uuid.UUID); ok {
				userUUID = id
			} else {
				// If it's int64 or something else and we can't get UUID, fail
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewAPIError(
					http.StatusUnauthorized,
					"Unauthorized",
					"User UUID not found",
				))
				return
			}
		}

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.NewAPIError(
				http.StatusUnauthorized,
				"Unauthorized",
				"Invalid User UUID",
			))
			return
		}

		// Get tenant_id from query or URL param
		tenantIDStr := ctx.Query("tenant_id")
		if tenantIDStr == "" {
			tenantIDStr = ctx.Param("id")
		}
		if tenantIDStr == "" {
			tenantIDStr = ctx.Param("tenant_id")
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

		// Get app_id from query or URL param (Optional)
		appIDStr := ctx.Query("app_id")
		if appIDStr == "" {
			appIDStr = ctx.Param("app_id")
		}

		var appID *uuid.UUID
		if appIDStr != "" {
			id, err := uuid.Parse(appIDStr)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, response.NewAPIError(
					http.StatusBadRequest,
					"Bad Request",
					"Invalid app ID format",
				))
				return
			}
			appID = &id
		}

		// Check membership with optional appID
		memberID, roleID, err := checker.GetMemberByTenantUserAndApp(ctx.Request.Context(), tenantID, userUUID, appID)
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
