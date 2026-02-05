package controller

import (
	"net/http"

	"CONVERDA/internal/apps/application/service"
	"CONVERDA/internal/apps/controller/dto"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WebhookController struct {
	webhookService service.WebhookService
}

func NewWebhookController(webhookService service.WebhookService) *WebhookController {
	return &WebhookController{webhookService: webhookService}
}

// CreateWebhook godoc
// @Summary Create a webhook
// @Description Create a new webhook for an app
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param app_id path string true "App ID"
// @Param tenant_id query string true "Tenant ID"
// @Param request body dto.CreateWebhookRequest true "Webhook data"
// @Success 201 {object} dto.WebhookResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /apps/{app_id}/webhooks [post]
func (c *WebhookController) CreateWebhook(ctx *gin.Context) (interface{}, error) {
	appID, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	var req dto.CreateWebhookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	// Get tenant_id from context (set by middleware based on app ownership)
	tenantID := ctx.Query("tenant_id")
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid tenant ID", err)
	}

	webhook, err := c.webhookService.CreateWebhook(ctx.Request.Context(), tid, appID, req.EnvironmentID, req.URL, req.Events, req.Description)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToWebhookResponse(webhook), nil
}

// ListWebhooks godoc
// @Summary List webhooks
// @Description List all webhooks for an app
// @Tags Webhooks
// @Produce json
// @Param app_id path string true "App ID"
// @Success 200 {array} dto.WebhookResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /apps/{app_id}/webhooks [get]
func (c *WebhookController) ListWebhooks(ctx *gin.Context) (interface{}, error) {
	appID, err := uuid.Parse(ctx.Param("app_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid app ID", err)
	}

	webhooks, err := c.webhookService.ListWebhooks(ctx.Request.Context(), appID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToWebhookResponseList(webhooks), nil
}

// GetWebhook godoc
// @Summary Get webhook
// @Description Get details of a specific webhook
// @Tags Webhooks
// @Produce json
// @Param webhook_id path string true "Webhook ID"
// @Success 200 {object} dto.WebhookResponse
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /webhooks/{webhook_id} [get]
func (c *WebhookController) GetWebhook(ctx *gin.Context) (interface{}, error) {
	webhookID, err := uuid.Parse(ctx.Param("webhook_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid webhook ID", err)
	}

	webhook, err := c.webhookService.GetWebhook(ctx.Request.Context(), webhookID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusNotFound, "Webhook not found", err)
	}

	return dto.ToWebhookResponse(webhook), nil
}

// UpdateWebhook godoc
// @Summary Update webhook
// @Description Update an existing webhook
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param webhook_id path string true "Webhook ID"
// @Param request body dto.UpdateWebhookRequest true "Update data"
// @Success 200 {object} dto.WebhookResponse
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /webhooks/{webhook_id} [put]
func (c *WebhookController) UpdateWebhook(ctx *gin.Context) (interface{}, error) {
	webhookID, err := uuid.Parse(ctx.Param("webhook_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid webhook ID", err)
	}

	var req dto.UpdateWebhookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	webhook, err := c.webhookService.GetWebhook(ctx.Request.Context(), webhookID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusNotFound, "Webhook not found", err)
	}

	if req.URL != nil {
		webhook.URL = *req.URL
	}
	if req.Events != nil {
		webhook.Events = req.Events
	}
	if req.Description != nil {
		webhook.Description = req.Description
	}
	if req.IsActive != nil {
		webhook.IsActive = *req.IsActive
	}

	if err := c.webhookService.UpdateWebhook(ctx.Request.Context(), webhook); err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToWebhookResponse(webhook), nil
}

// DeleteWebhook godoc
// @Summary Delete webhook
// @Description Delete a webhook
// @Tags Webhooks
// @Produce json
// @Param webhook_id path string true "Webhook ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /webhooks/{webhook_id} [delete]
func (c *WebhookController) DeleteWebhook(ctx *gin.Context) (interface{}, error) {
	webhookID, err := uuid.Parse(ctx.Param("webhook_id"))
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid webhook ID", err)
	}

	if err := c.webhookService.DeleteWebhook(ctx.Request.Context(), webhookID); err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return gin.H{"message": "Webhook deleted"}, nil
}
