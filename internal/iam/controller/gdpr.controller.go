package controller

import (
	"CONVERDA/pkg/gdpr"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GDPRController handles GDPR compliance endpoints
type GDPRController struct {
	gdprService *gdpr.Service
}

// NewGDPRController creates a new GDPRController
func NewGDPRController(gdprService *gdpr.Service) *GDPRController {
	return &GDPRController{gdprService: gdprService}
}

// ExportRequest is the request body for GDPR data export
type ExportRequest struct {
	UserID     uuid.UUID           `json:"user_id" binding:"required"`
	Categories []gdpr.DataCategory `json:"categories"` // empty = all
}

// ExportResponse is the response for GDPR data export
type ExportResponse struct {
	UserID uuid.UUID           `json:"user_id"`
	Data   []gdpr.ExportedData `json:"data"`
}

// EraseRequest is the request body for GDPR data erasure
type EraseRequest struct {
	UserID     uuid.UUID           `json:"user_id" binding:"required"`
	Categories []gdpr.DataCategory `json:"categories"` // empty = all
}

// EraseResponse is the response for GDPR data erasure
type EraseResponse struct {
	UserID uuid.UUID                   `json:"user_id"`
	Erased map[gdpr.DataCategory]int64 `json:"erased"`
}

// ExportData godoc
// @Summary Export user data (GDPR)
// @Description Export all personal data for a user within a tenant
// @Tags GDPR
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param request body ExportRequest true "Export request"
// @Success 200 {object} ExportResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /tenants/{id}/gdpr/export [post]
func (c *GDPRController) ExportData(ctx *gin.Context) (interface{}, error) {
	tenantIDStr := ctx.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, response.NewBadRequestError("invalid tenant ID")
	}

	var req ExportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	data, err := c.gdprService.Export(ctx.Request.Context(), tenantID, req.UserID, req.Categories)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return ExportResponse{
		UserID: req.UserID,
		Data:   data,
	}, nil
}

// EraseData godoc
// @Summary Erase user data (GDPR Right to be Forgotten)
// @Description Erase/anonymize personal data for a user within a tenant
// @Tags GDPR
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param request body EraseRequest true "Erase request"
// @Success 200 {object} EraseResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /tenants/{id}/gdpr/erase [post]
func (c *GDPRController) EraseData(ctx *gin.Context) (interface{}, error) {
	tenantIDStr := ctx.Param("id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, response.NewBadRequestError("invalid tenant ID")
	}

	var req EraseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	erased, err := c.gdprService.Erase(ctx.Request.Context(), tenantID, req.UserID, req.Categories)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return EraseResponse{
		UserID: req.UserID,
		Erased: erased,
	}, nil
}
