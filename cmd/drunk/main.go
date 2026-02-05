package main

import (
	"strconv"

	"CONVERDA/internal/initialize"

	_ "CONVERDA/docs"
)

// @title Go Drunk Backend API by DDD
// @version 1.0
// @description This is a server for a Go Drunk Backend API, demonstrating DDD principles.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. "Bearer abcde12345"

// Base path
// @BasePath /v1/api

// @externalDocs.description OpenAPI
// @externalDocs.url https://swagger.io/resources/open-api/
func main() {
	r, port := initialize.Run()

	// prometheus.MustRegister(pingCounter)

	// r.GET("/ping/200", ping)
	// r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	// Start gRPC server

	r.Run(":" + strconv.Itoa(port)) // listen and serve on configured port
}
