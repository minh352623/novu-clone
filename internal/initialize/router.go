package initialize

import (
	"database/sql"
	"fmt"
	"time"

	"CONVERDA/global"
	iamRepo "CONVERDA/internal/iam/infrastructure/persistence/repository"
	initializeApps "CONVERDA/internal/initialize/apps"
	initializeHealth "CONVERDA/internal/initialize/health"
	initializeIAM "CONVERDA/internal/initialize/iam"
	initializeMessaging "CONVERDA/internal/initialize/messaging"
	initializeNotification "CONVERDA/internal/initialize/notification"
	initializeR2 "CONVERDA/internal/initialize/r2"
	initializeWorkflow "CONVERDA/internal/initialize/workflow"
	"CONVERDA/internal/middleware"
	r2Http "CONVERDA/internal/r2/controller/http"
	"CONVERDA/pkg/auditlog"
	"CONVERDA/pkg/ratelimit"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitRouter initializes the router with all routes
func InitRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()
	fmt.Printf("Server mode: %s\n", global.Config.Server.Mode)

	if global.Config.Server.Mode != "production" {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor()
		r = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
	}

	// -- Multi-tenant infrastructure --
	auditLogRepo := iamRepo.NewAuditLogRepository(global.GormDB)
	auditLogger := auditlog.NewLogger(auditLogRepo)
	tenantLimiter := ratelimit.NewMemoryRateLimiter(5 * time.Minute)
	tenantRPMProvider := middleware.NewTenantPlanRPMAdapter(global.GormDB)

	// API v1/2025 group
	v1 := r.Group("/v1/api")
	{
		// Global middleware: audit logging for all mutating requests
		v1.Use(middleware.AuditMiddleware(auditLogger))

		// Tenant-level rate limiting (after auth sets tenant context)
		v1.Use(middleware.TenantRateLimitMiddleware(tenantLimiter, tenantRPMProvider))

		// R2 module routes
		r2Handler := initializeR2.InitR2()
		if r2Handler != nil {
			r2Http.RegisterR2Routes(v1, r2Handler)
		}

		// IAM module routes
		initializeIAM.InitIAMModule(v1, global.GormDB, middleware.AuthMiddleware())

		// Apps module routes
		initializeApps.InitAppsModule(v1, global.GormDB, middleware.AuthMiddleware())

		// Messaging module routes
		initializeMessaging.InitMessagingModule(v1, global.GormDB, middleware.AuthMiddleware())

		// Notification module routes
		initializeNotification.InitNotificationModule(global.GormDB, v1)

		// Health module routes (shared across modules)
		initializeHealth.InitHealthModule(global.GormDB, v1, middleware.AuthMiddleware())

		// Workflow Engine module routes
		initializeWorkflow.InitWorkflowModule(global.GormDB, v1, middleware.AuthMiddleware())
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
