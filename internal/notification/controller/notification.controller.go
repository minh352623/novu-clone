package controller

import (
	"net/http"

	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/controller/dto"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationController struct {
	svc service.NotificationService
}

func NewNotificationController(svc service.NotificationService) *NotificationController {
	return &NotificationController{svc: svc}
}

// SendNotification godoc
// @Summary Send Notification
// @Description Send a notification via configured channel
// @Tags Notification
// @Accept json
// @Produce json
// @Param request body dto.SendNotificationRequest true "Notification Request"
// @Success 200 {object} dto.NotificationResponse
// @Security BearerAuth
// @Router /notifications/send [post]
func (c *NotificationController) SendNotification(ctx *gin.Context) (interface{}, error) {
	var req dto.SendNotificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	// Extract Context (Tenant/Env)
	// Assuming Authentication Middleware sets these
	tenantID := uuid.Nil
	envID := uuid.Nil

	if tID, ok := ctx.Get("tenant_id"); ok {
		tenantID = tID.(uuid.UUID)
	}
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	// Call Service
	// Map DTO to Service Request
	svcReq := service.SendRequest{
		TenantID:      tenantID.String(),
		EnvironmentID: envID.String(),
		TemplateCode:  req.TemplateCode,
		Recipient:     req.Recipient,
		Channel:       req.Channel,
		Data:          req.Data,
		Language:      req.Language,
	}

	resp, err := c.svc.Send(ctx.Request.Context(), svcReq)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.NotificationResponse{
		NotificationID: resp.NotificationID,
		Status:         resp.Status,
	}, nil
}
