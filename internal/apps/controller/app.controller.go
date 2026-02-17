package controller

import (
	"net/http"
	"strconv"

	"CONVERDA/internal/apps/application/service"
	"CONVERDA/internal/apps/controller/dto"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AppController struct {
	appService       service.AppService
	envService       service.EnvironmentService
	apiKeyService    service.APIKeyService
	metricsService   service.MetricsService
	systemEnvService service.SystemEnvironmentService
}

func NewAppController(
	appService service.AppService,
	envService service.EnvironmentService,
	apiKeyService service.APIKeyService,
	metricsService service.MetricsService,
	systemEnvService service.SystemEnvironmentService,
) *AppController {
	return &AppController{
		appService:       appService,
		envService:       envService,
		apiKeyService:    apiKeyService,
		metricsService:   metricsService,
		systemEnvService: systemEnvService,
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
// @Security BearerAuth
func (c *AppController) CreateApp(ctx *gin.Context) (interface{}, error) {
	tenantID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid tenant ID", err)
	}

	var req dto.CreateAppRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	app, err := c.appService.CreateApp(ctx.Request.Context(), tenantID, req.Name, req.Description, req.SLAThresholdSeconds)
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
	if req.SLAThresholdSeconds != nil {
		app.SLAThresholdSeconds = *req.SLAThresholdSeconds
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

	env, err := c.envService.CreateEnvironment(ctx.Request.Context(), appID, req.Code, req.SLAThresholdSeconds)
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

// UpdateEnvironmentConfig godoc
// @Summary Update environment configuration
// @Description Update rate limit and SLA settings for an environment
// @Tags Apps
// @Accept json
// @Produce json
// @Param app_id path string true "App ID"
// @Param env_id path string true "Environment ID"
// @Param body body dto.UpdateEnvironmentConfigRequest true "Config update"
// @Success 200 {object} dto.EnvironmentResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /apps/{app_id}/environments/{env_id} [put]
func (c *AppController) UpdateEnvironmentConfig(ctx *gin.Context) (interface{}, error) {
	_, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	envID, err := uuid.Parse(ctx.Param("env_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid environment ID", err)
	}

	var req dto.UpdateEnvironmentConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	env, err := c.envService.GetByID(ctx.Request.Context(), envID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusNotFound, "Environment not found", err)
	}

	// Apply partial updates
	if req.SLAThresholdSeconds != nil {
		env.SLAThresholdSeconds = *req.SLAThresholdSeconds
	}
	if req.RateLimitRPM != nil {
		env.RateLimitRPM = *req.RateLimitRPM
	}
	if req.RateLimitDaily != nil {
		env.RateLimitDaily = *req.RateLimitDaily
	}

	if err := c.envService.UpdateEnvironment(ctx.Request.Context(), env); err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToEnvironmentResponse(env), nil
}

// ListAPIKeys godoc
// @Summary List API Keys
// @Tags Apps
// @Produce json
// @Param id path string true "Env ID"
// @Success 200 {array} dto.APIKeyResponse
// @Router /environments/{id}/api-keys [get]
// @Security BearerAuth
func (c *AppController) ListAPIKeys(ctx *gin.Context) (interface{}, error) {
	envID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid environment ID", err)
	}

	keys, err := c.apiKeyService.ListKeys(ctx.Request.Context(), envID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToAPIKeyResponseList(keys), nil // Still need ToAPIKeyResponseList in DTO
}

// RotateAPIKey godoc
// @Summary Rotate API Key
// @Description Revokes all old keys and generates a new one
// @Tags Apps
// @Accept json
// @Produce json
// @Param id path string true "Env ID"
// @Param request body dto.RotateKeyRequest true "Key data"
// @Success 200 {object} dto.APIKeyFullResponse
// @Router /environments/{id}/api-keys/rotate [post]
// @Security BearerAuth
func (c *AppController) RotateAPIKey(ctx *gin.Context) (interface{}, error) {
	envID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid environment ID", err)
	}

	var req dto.RotateKeyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	plainKey, key, err := c.apiKeyService.RotateKey(ctx.Request.Context(), envID, req.Name)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	resp := dto.APIKeyFullResponse{
		APIKeyResponse: *dto.ToAPIKeyResponse(key),
		PlainKey:       plainKey,
	}

	return resp, nil
}

// RevokeAPIKey godoc
// @Summary Revoke API Key
// @Tags Apps
// @Param key_id path string true "Key ID"
// @Success 200 {object} map[string]string
// @Router /api-keys/{key_id}/revoke [post]
// @Security BearerAuth
func (c *AppController) RevokeAPIKey(ctx *gin.Context) (interface{}, error) {
	keyID, err := uuid.Parse(ctx.Param("key_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid key ID", err)
	}

	if err := c.apiKeyService.RevokeKey(ctx.Request.Context(), keyID); err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return gin.H{"message": "API Key revoked"}, nil
}

// GetAppMetrics godoc
// @Summary Get App usage metrics
// @Tags Apps
// @Produce json
// @Param app_id path string true "App ID"
// @Param days query int false "Days to look back"
// @Success 200 {object} map[string]int64
// @Router /apps/{app_id}/metrics [get]
// @Security BearerAuth
func (c *AppController) GetAppMetrics(ctx *gin.Context) (interface{}, error) {
	appID, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	days := 30
	// Optional: parse days from query

	metrics, err := c.metricsService.GetAppMetrics(ctx.Request.Context(), appID, days)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return metrics, nil
}

// GetDetailedMetrics godoc
// @Summary Get detailed usage metrics for an app
// @Tags Apps
// @Produce json
// @Param app_id path string true "App ID"
// @Param days query int false "Days to look back (default: 30)"
// @Success 200 {object} dto.DetailedMetricsResponse
// @Router /apps/{app_id}/metrics/detailed [get]
// @Security BearerAuth
func (c *AppController) GetDetailedMetrics(ctx *gin.Context) (interface{}, error) {
	appID, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	days := 30
	if d := ctx.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}

	result, err := c.metricsService.GetDetailedMetrics(ctx.Request.Context(), appID, days)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}
	return result, nil
}

// GetDailyTimeSeries godoc
// @Summary Get daily time-series metrics for an app
// @Tags Apps
// @Produce json
// @Param app_id path string true "App ID"
// @Param days query int false "Days to look back (default: 30)"
// @Param environment_id query string false "Optional environment filter"
// @Success 200 {object} dto.TimeSeriesResponse
// @Router /apps/{app_id}/metrics/timeseries [get]
// @Security BearerAuth
func (c *AppController) GetDailyTimeSeries(ctx *gin.Context) (interface{}, error) {
	appID, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	days := 30
	if d := ctx.Query("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 {
			days = v
		}
	}

	var envID *uuid.UUID
	if envStr := ctx.Query("environment_id"); envStr != "" {
		parsed, err := uuid.Parse(envStr)
		if err != nil {
			return nil, response.NewAPIError(http.StatusBadRequest, "Invalid environment ID", err)
		}
		envID = &parsed
	}

	result, err := c.metricsService.GetDailyTimeSeries(ctx.Request.Context(), appID, envID, days)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}
	return result, nil
}

// ListSystemEnvironments godoc
// @Summary List system environments
// @Tags System
// @Produce json
// @Success 200 {array} dto.SystemEnvironmentResponse
// @Security BearerAuth
// @Router /system/environments [get]
// @Security BearerAuth
func (c *AppController) ListSystemEnvironments(ctx *gin.Context) (interface{}, error) {
	envs, err := c.systemEnvService.ListAll(ctx.Request.Context())
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToSystemEnvironmentResponseList(envs), nil
}
