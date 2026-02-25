package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"CONVERDA/pkg/auditlog"
	"CONVERDA/pkg/tenantctx"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditMiddleware records audit log entries for mutating HTTP requests.
// Only records POST, PUT, PATCH, DELETE methods.
func AuditMiddleware(logger *auditlog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method
		// Only audit mutating operations
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			ctx.Next()
			return
		}

		// Extract tenant context
		tenantID := uuid.Nil
		if tID, ok := ctx.Get(ContextKeyTenantIDFromJWT); ok {
			if id, ok := tID.(uuid.UUID); ok {
				tenantID = id
			}
		}
		if tenantID == uuid.Nil {
			ctx.Next()
			return
		}

		// Extract actor
		actorID := uuid.Nil
		if uID, ok := ctx.Get(ContextKeyUserID); ok {
			if id, ok := uID.(uuid.UUID); ok {
				actorID = id
			}
		}

		// Determine action from HTTP method
		action := methodToAction(method)

		// Extract resource type from path
		resourceType := extractResourceType(ctx.FullPath())

		// Extract resource ID from params
		resourceID := ctx.Param("id")

		// Proceed with request
		ctx.Next()

		// Only log successful mutations (2xx status)
		status := ctx.Writer.Status()
		if status < 200 || status >= 300 {
			return
		}

		// Build context with tenant for audit logger
		auditCtx := tenantctx.WithTenantID(ctx.Request.Context(), tenantID)

		// Record audit entry
		opts := []auditlog.Option{
			auditlog.WithMetadata(map[string]interface{}{
				"ip":         ctx.ClientIP(),
				"user_agent": ctx.Request.UserAgent(),
				"method":     method,
				"path":       ctx.Request.URL.Path,
				"status":     status,
			}),
		}
		if actorID != uuid.Nil {
			opts = append(opts, auditlog.WithActor(actorID, "user"))
		}

		logger.Log(auditCtx, action, resourceType, resourceID, opts...)
	}
}

// methodToAction maps HTTP methods to audit action names
func methodToAction(method string) string {
	switch method {
	case http.MethodPost:
		return auditlog.ActionCreate
	case http.MethodPut, http.MethodPatch:
		return auditlog.ActionUpdate
	case http.MethodDelete:
		return auditlog.ActionDelete
	default:
		return method
	}
}

// extractResourceType extracts resource name from route path
// e.g., "/api/v1/tenants/:id/apps" -> "apps"
func extractResourceType(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return "unknown"
	}
	// Find last non-param segment
	for i := len(parts) - 1; i >= 0; i-- {
		if !strings.HasPrefix(parts[i], ":") {
			return parts[i]
		}
	}
	return "unknown"
}

// readBody reads the request body and restores it for downstream handlers
func readBody(ctx *gin.Context) []byte {
	if ctx.Request.Body == nil {
		return nil
	}
	body, _ := io.ReadAll(ctx.Request.Body)
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	return body
}
