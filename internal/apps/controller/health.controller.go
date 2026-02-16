package controller

import (
	"context"
	"net/http"
	"time"

	"CONVERDA/global"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// Check godoc
// @Summary System health check
// @Description Check the health status of the system and its dependencies
// @Tags System
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (c *HealthController) Check(ctx *gin.Context) (interface{}, error) {
	status := "UP"
	checks := make(map[string]string)

	// 1. Check Database
	dbStatus := "UP"
	if global.Pdbc == nil {
		dbStatus = "DOWN"
		status = "PARTIAL_DOWN"
	} else {
		// Create a context with timeout for the ping
		pingCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Second)
		defer cancel()

		if err := global.Pdbc.PingContext(pingCtx); err != nil {
			dbStatus = "DOWN (Ping failed)"
			status = "PARTIAL_DOWN"
		}
	}
	checks["database"] = dbStatus

	resp := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Checks:    checks,
	}

	if status != "UP" {
		return resp, response.NewAPIError(http.StatusServiceUnavailable, "System is not fully healthy", nil)
	}

	return resp, nil
}
