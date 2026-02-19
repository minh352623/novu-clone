package controller

import (
	"strconv"

	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/controller/dto"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LayoutController struct {
	manager service.LayoutManager
}

func NewLayoutController(manager service.LayoutManager) *LayoutController {
	return &LayoutController{manager: manager}
}

// CreateLayout godoc
// @Summary Create Notification Layout
// @Description Create a new notification layout
// @Tags Notification
// @Accept json
// @Produce json
// @Param request body dto.CreateLayoutRequest true "Layout Data"
// @Success 201 {object} dto.LayoutResponse
// @Security BearerAuth
// @Router /notifications/layouts [post]
func (c *LayoutController) CreateLayout(ctx *gin.Context) (interface{}, error) {
	var req dto.CreateLayoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	layout, err := c.manager.CreateLayout(ctx.Request.Context(), envID, req.Name, req.Description, req.ContentHTML, req.VariablesSchema, req.IsDefault)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return dto.ToLayoutResponse(layout), nil
}

func (c *LayoutController) ListLayouts(ctx *gin.Context) (interface{}, error) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	layouts, total, err := c.manager.ListLayouts(ctx.Request.Context(), envID, limit, offset)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return dto.NewPaginatedResponse(dto.ToLayoutResponseList(layouts), page, limit, total), nil
}

func (c *LayoutController) GetLayout(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid ID")
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	layout, err := c.manager.GetLayout(ctx.Request.Context(), envID, id)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}
	if layout == nil {
		return nil, response.NewNotFoundError("Layout not found")
	}

	return dto.ToLayoutResponse(layout), nil
}

// UpdateLayout godoc
// @Summary Update Notification Layout
// @Description Update a notification layout
// @Tags Notification
// @Accept json
// @Produce json
// @Param id path string true "Layout ID"
// @Param request body dto.UpdateLayoutRequest true "Layout Data"
// @Success 200
// @Security BearerAuth
// @Router /notifications/layouts/{id} [patch]
func (c *LayoutController) UpdateLayout(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid ID")
	}

	var req dto.UpdateLayoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	if err := c.manager.UpdateLayout(ctx.Request.Context(), envID, id, req.Name, req.Description, req.ContentHTML, req.VariablesSchema, req.IsDefault); err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "updated"}, nil
}

func (c *LayoutController) DeleteLayout(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid ID")
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	if err := c.manager.DeleteLayout(ctx.Request.Context(), envID, id); err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "deleted"}, nil
}
