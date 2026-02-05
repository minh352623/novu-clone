package middleware

import (
	"context"
	"net/http"

	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PermissionChecker interface for checking user permissions
type PermissionChecker interface {
	// GetRolePermissions returns the permissions map for a role
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) (map[string]interface{}, error)
}

// RequirePermission creates a middleware that checks if the user has the required permission
// It expects TenantMembershipMiddleware to have run first to set the role_id in context
func RequirePermission(checker PermissionChecker, permission string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Get role ID from context (set by TenantMembershipMiddleware)
		roleIDVal, exists := ctx.Get(ContextKeyRoleID)
		if !exists {
			// No role assigned, deny access
			ctx.AbortWithStatusJSON(http.StatusForbidden, response.NewAPIError(
				http.StatusForbidden,
				"Forbidden",
				"No role assigned",
			))
			return
		}

		roleID, ok := roleIDVal.(uuid.UUID)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, response.NewAPIError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Invalid role ID type in context",
			))
			return
		}

		// Get permissions for the role
		permissions, err := checker.GetRolePermissions(ctx.Request.Context(), roleID)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, response.NewAPIError(
				http.StatusInternalServerError,
				"Internal Server Error",
				"Failed to get role permissions",
			))
			return
		}

		// Check if permission exists and is true
		if !hasPermission(permissions, permission) {
			ctx.AbortWithStatusJSON(http.StatusForbidden, response.NewAPIError(
				http.StatusForbidden,
				"Forbidden",
				"You do not have permission to perform this action",
			))
			return
		}

		ctx.Next()
	}
}

// hasPermission checks if the permissions map contains the required permission
// Supports dot notation for nested permissions (e.g., "apps.create")
func hasPermission(permissions map[string]interface{}, permission string) bool {
	if permissions == nil {
		return false
	}

	// Check for wildcard permission (admin)
	if val, exists := permissions["*"]; exists {
		if b, ok := val.(bool); ok && b {
			return true
		}
	}

	// Check for exact permission
	if val, exists := permissions[permission]; exists {
		if b, ok := val.(bool); ok {
			return b
		}
		// If it exists but is not a boolean, consider it as granted
		return true
	}

	return false
}

// RequireTenantAdmin creates a middleware that checks if the user is a tenant admin
// This is a convenience wrapper for common tenant admin operations
func RequireTenantAdmin(checker PermissionChecker) gin.HandlerFunc {
	return RequirePermission(checker, "tenant.admin")
}
