package controller

import (
	"net/http"
	"time"

	"CONVERDA/internal/health/dto"
	"CONVERDA/internal/health/service"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HealthDashboardController struct {
	svc service.HealthService
}

func NewHealthDashboardController(svc service.HealthService) *HealthDashboardController {
	return &HealthDashboardController{svc: svc}
}

// GetSystemHealth godoc
// @Summary Get System Health Dashboard
// @Description Get operational health overview: queue state, SLA compliance, webhook reliability
// @Tags System Health
// @Produce json
// @Param environment_id query string true "Environment ID"
// @Param from query string false "From Date (YYYY-MM-DD)"
// @Param to query string false "To Date (YYYY-MM-DD)"
// @Success 200 {object} dto.SystemHealthResponse
// @Security BearerAuth
// @Router /dashboards/health [get]
func (c *HealthDashboardController) GetSystemHealth(ctx *gin.Context) (interface{}, error) {
	var req dto.HealthRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid request parameters", err)
	}

	envID, err := uuid.Parse(req.EnvironmentID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid Environment ID", err)
	}

	now := time.Now()
	fromStr := req.From
	toStr := req.To
	if fromStr == "" {
		fromStr = now.AddDate(0, 0, -1).Format("2006-01-02")
	}
	if toStr == "" {
		toStr = now.Format("2006-01-02")
	}

	from, _ := time.Parse("2006-01-02", fromStr)
	to, _ := time.Parse("2006-01-02", toStr)
	to = to.Add(24*time.Hour - time.Second)

	result, err := c.svc.GetSystemHealth(ctx.Request.Context(), envID, from, to)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return result, nil
}
