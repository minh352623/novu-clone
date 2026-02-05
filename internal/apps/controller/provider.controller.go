package controller

import (
	"net/http"

	"CONVERDA/internal/apps/application/service"
	"CONVERDA/internal/apps/controller/dto"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProviderController struct {
	providerService service.ProviderService
}

func NewProviderController(providerService service.ProviderService) *ProviderController {
	return &ProviderController{providerService: providerService}
}

// CreateProvider godoc
// @Summary Create a provider
// @Description Create a new provider for an app
// @Tags Providers
// @Accept json
// @Produce json
// @Param app_id path string true "App ID"
// @Param tenant_id query string true "Tenant ID"
// @Param request body dto.CreateProviderRequest true "Provider data"
// @Success 201 {object} dto.ProviderResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /apps/{app_id}/providers [post]
func (c *ProviderController) CreateProvider(ctx *gin.Context) (interface{}, error) {
	appID, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	var req dto.CreateProviderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	tenantID := ctx.Query("tenant_id")
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid tenant ID", err)
	}

	provider, err := c.providerService.CreateProvider(ctx.Request.Context(), tid, appID, req.EnvironmentID, req.ProviderType, req.ProviderName, req.Configuration)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToProviderResponse(provider), nil
}

// ListProviders godoc
// @Summary List providers
// @Description List all providers for an app
// @Tags Providers
// @Produce json
// @Param app_id path string true "App ID"
// @Success 200 {array} dto.ProviderResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /apps/{app_id}/providers [get]
func (c *ProviderController) ListProviders(ctx *gin.Context) (interface{}, error) {
	appID, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	providers, err := c.providerService.ListProviders(ctx.Request.Context(), appID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToProviderResponseList(providers), nil
}

// GetProvider godoc
// @Summary Get provider
// @Description Get details of a specific provider
// @Tags Providers
// @Produce json
// @Param provider_id path string true "Provider ID"
// @Success 200 {object} dto.ProviderResponse
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /providers/{provider_id} [get]
func (c *ProviderController) GetProvider(ctx *gin.Context) (interface{}, error) {
	providerID, err := uuid.Parse(ctx.Param("provider_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid provider ID", err)
	}

	provider, err := c.providerService.GetProvider(ctx.Request.Context(), providerID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusNotFound, "Provider not found", err)
	}

	return dto.ToProviderResponse(provider), nil
}

// UpdateProvider godoc
// @Summary Update provider
// @Description Update an existing provider
// @Tags Providers
// @Accept json
// @Produce json
// @Param provider_id path string true "Provider ID"
// @Param request body dto.UpdateProviderRequest true "Update data"
// @Success 200 {object} dto.ProviderResponse
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /providers/{provider_id} [put]
func (c *ProviderController) UpdateProvider(ctx *gin.Context) (interface{}, error) {
	providerID, err := uuid.Parse(ctx.Param("provider_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid provider ID", err)
	}

	var req dto.UpdateProviderRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	provider, err := c.providerService.GetProvider(ctx.Request.Context(), providerID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusNotFound, "Provider not found", err)
	}

	if req.ProviderName != nil {
		provider.ProviderName = *req.ProviderName
	}
	if req.Configuration != nil {
		provider.Configuration = req.Configuration
	}
	if req.IsActive != nil {
		provider.IsActive = *req.IsActive
	}

	if err := c.providerService.UpdateProvider(ctx.Request.Context(), provider); err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToProviderResponse(provider), nil
}

// DeleteProvider godoc
// @Summary Delete provider
// @Description Delete a provider
// @Tags Providers
// @Produce json
// @Param provider_id path string true "Provider ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /providers/{provider_id} [delete]
func (c *ProviderController) DeleteProvider(ctx *gin.Context) (interface{}, error) {
	providerID, err := uuid.Parse(ctx.Param("provider_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid provider ID", err)
	}

	if err := c.providerService.DeleteProvider(ctx.Request.Context(), providerID); err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return gin.H{"message": "Provider deleted"}, nil
}
