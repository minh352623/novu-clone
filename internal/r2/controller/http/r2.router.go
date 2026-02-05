package http

import (
	"net/http"

	"CONVERDA/internal/middleware"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
)

// RegisterR2Routes registers all R2 routes
func RegisterR2Routes(r *gin.RouterGroup, handler *R2Handler) {
	r2 := r.Group("/r2")
	r2.Use(middleware.AuthMiddleware())
	{
		r2.POST("/upload", response.Wrap(handler.UploadFileBase64, http.StatusOK))
		r2.DELETE("/delete", response.Wrap(handler.DeleteFile, http.StatusOK))
	}
}
