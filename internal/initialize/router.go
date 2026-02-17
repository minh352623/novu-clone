package initialize

import (
	"database/sql"
	"fmt"
	"net/http"

	"CONVERDA/global"
	"CONVERDA/internal/apps/controller"
	initializeApps "CONVERDA/internal/initialize/apps"
	initializeHealth "CONVERDA/internal/initialize/health"
	initializeIAM "CONVERDA/internal/initialize/iam"
	initializeMessaging "CONVERDA/internal/initialize/messaging"
	initializeNotification "CONVERDA/internal/initialize/notification"
	initializeR2 "CONVERDA/internal/initialize/r2"
	initializeWorkflow "CONVERDA/internal/initialize/workflow"
	"CONVERDA/internal/middleware"
	r2Http "CONVERDA/internal/r2/controller/http"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitRouter initializes the router with all routes
func InitRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()
	fmt.Printf("Server mode: %s\n", global.Config.Server.Mode)

	// Health check
	healthController := controller.NewHealthController()
	r.GET("/health", response.Wrap(healthController.Check, http.StatusOK))

	if global.Config.Server.Mode != "production" {
		gin.SetMode(gin.DebugMode)
		gin.ForceConsoleColor()
		r = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
	}
	// API v1/2025 group
	v1 := r.Group("/v1/api")
	{
		// R2 module routes
		r2Handler := initializeR2.InitR2()
		if r2Handler != nil {
			r2Http.RegisterR2Routes(v1, r2Handler)
		}

		// IAM module routes
		initializeIAM.InitIAMModule(v1, middleware.AuthMiddleware())

		// Apps module routes
		initializeApps.InitAppsModule(v1, middleware.AuthMiddleware())

		// Messaging module routes
		initializeMessaging.InitMessagingModule(v1, middleware.AuthMiddleware())

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
