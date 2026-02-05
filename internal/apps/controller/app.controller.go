package controller

import (
	"net/http"

	"CONVERDA/internal/apps/application/service"
	"CONVERDA/internal/apps/controller/dto"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AppController struct {
	appService service.AppService
	envService service.EnvironmentService
}

func NewAppController(appService service.AppService, envService service.EnvironmentService) *AppController {
	return &AppController{
		appService: appService,
		envService: envService,
	}
}

// CreateApp godoc
// @Summary Create a new app
// @Tags Apps
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param request body dto.CreateAppRequest true "App data"
// @Success 201 {object} dto.AppResponse
// @Router /tenants/{id}/apps [post]
func (c *AppController) CreateApp(ctx *gin.Context) (interface{}, error) {
	tenantID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid tenant ID", err)
	}

	var req dto.CreateAppRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	app, err := c.appService.CreateApp(ctx.Request.Context(), tenantID, req.Name, req.Description)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToAppResponse(app), nil
}

// ListApps godoc
// @Summary List apps
// @Description List all apps for a tenant
// @Tags Apps
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {array} dto.AppResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /tenants/{id}/apps [get]
func (c *AppController) ListApps(ctx *gin.Context) (interface{}, error) {
	tenantID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid tenant ID", err)
	}

	apps, err := c.appService.ListApps(ctx.Request.Context(), tenantID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToAppResponseList(apps), nil
}

// GetApp godoc
// @Summary Get app
// @Description Get details of a specific app
// @Tags Apps
// @Produce json
// @Param app_id path string true "App ID"
// @Success 200 {object} dto.AppResponse
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /apps/{app_id} [get]
func (c *AppController) GetApp(ctx *gin.Context) (interface{}, error) {
	appID, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	app, err := c.appService.GetApp(ctx.Request.Context(), appID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusNotFound, "App not found", err)
	}

	return dto.ToAppResponse(app), nil
}

// UpdateApp godoc
// @Summary Update app
// @Description Update an existing app
// @Tags Apps
// @Accept json
// @Produce json
// @Param app_id path string true "App ID"
// @Param request body dto.UpdateAppRequest true "Update data"
// @Success 200 {object} dto.AppResponse
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /apps/{app_id} [put]
func (c *AppController) UpdateApp(ctx *gin.Context) (interface{}, error) {
	appID, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	var req dto.UpdateAppRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	app, err := c.appService.GetApp(ctx.Request.Context(), appID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusNotFound, "App not found", err)
	}

	if req.Name != nil {
		app.Name = *req.Name
	}
	if req.Description != nil {
		app.Description = req.Description
	}

	if err := c.appService.UpdateApp(ctx.Request.Context(), app); err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToAppResponse(app), nil
}

// DeleteApp godoc
// @Summary Delete app
// @Description Delete an app
// @Tags Apps
// @Produce json
// @Param app_id path string true "App ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /apps/{app_id} [delete]
func (c *AppController) DeleteApp(ctx *gin.Context) (interface{}, error) {
	appID, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	if err := c.appService.DeleteApp(ctx.Request.Context(), appID); err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return gin.H{"message": "App deleted"}, nil
}

// CreateEnvironment godoc
// @Summary Create environment
// @Description Create a new environment for an app
// @Tags Apps
// @Accept json
// @Produce json
// @Param app_id path string true "App ID"
// @Param request body dto.CreateEnvironmentRequest true "Environment data"
// @Success 201 {object} dto.EnvironmentResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /apps/{app_id}/environments [post]
func (c *AppController) CreateEnvironment(ctx *gin.Context) (interface{}, error) {
	appID, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	var req dto.CreateEnvironmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	env, err := c.envService.CreateEnvironment(ctx.Request.Context(), appID, req.Code)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	return dto.ToEnvironmentResponse(env), nil
}

// ListEnvironments godoc
// @Summary List environments
// @Description List all environments for an app
// @Tags Apps
// @Produce json
// @Param app_id path string true "App ID"
// @Success 200 {array} dto.EnvironmentResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /apps/{app_id}/environments [get]
func (c *AppController) ListEnvironments(ctx *gin.Context) (interface{}, error) {
	appID, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	envs, err := c.envService.ListEnvironments(ctx.Request.Context(), appID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToEnvironmentResponseList(envs), nil
}
