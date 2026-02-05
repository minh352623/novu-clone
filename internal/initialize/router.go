package initialize

import (
	"database/sql"
	"fmt"

	"CONVERDA/global"
	initializeApps "CONVERDA/internal/initialize/apps"
	initializeIAM "CONVERDA/internal/initialize/iam"
	initializeR2 "CONVERDA/internal/initialize/r2"
	"CONVERDA/internal/middleware"
	r2Http "CONVERDA/internal/r2/controller/http"

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
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
