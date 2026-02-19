package controller

import (
	"strconv"

	"CONVERDA/internal/notification/application/service"
	"CONVERDA/internal/notification/controller/dto"
	"CONVERDA/internal/notification/domain"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GroupController struct {
	manager service.GroupManager
}

func NewGroupController(manager service.GroupManager) *GroupController {
	return &GroupController{manager: manager}
}

// CreateGroup godoc
// @Summary Create Notification Group
// @Description Create a new notification group (category)
// @Tags Notification
// @Accept json
// @Produce json
// @Param request body dto.CreateGroupRequest true "Group Data"
// @Success 201 {object} dto.GroupResponse
// @Security BearerAuth
// @Router /notifications/groups [post]
func (c *GroupController) CreateGroup(ctx *gin.Context) (interface{}, error) {
	var req dto.CreateGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	group, err := c.manager.CreateGroup(ctx.Request.Context(), envID, req.Name, req.Key, req.Description, req.IsDefault)
	if err != nil {
		if err == domain.ErrGroupDuplicateKey {
			return nil, response.NewBadRequestError(err.Error())
		}
		return nil, response.NewInternalServerError(err.Error())
	}

	return dto.ToGroupResponse(group), nil
}

// ListGroups godoc
// @Summary List Notification Groups
// @Description List notification groups
// @Tags Notification
// @Produce json
// @Param page query int false "Page"
// @Param limit query int false "Limit"
// @Success 200 {object} dto.PaginatedResponse
// @Security BearerAuth
// @Router /notifications/groups [get]
func (c *GroupController) ListGroups(ctx *gin.Context) (interface{}, error) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	groups, total, err := c.manager.ListGroups(ctx.Request.Context(), envID, limit, offset)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return dto.NewPaginatedResponse(dto.ToGroupResponseList(groups), page, limit, total), nil
}

// GetGroup godoc
// @Summary Get Notification Group
// @Description Get a notification group by ID
// @Tags Notification
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} dto.GroupResponse
// @Security BearerAuth
// @Router /notifications/groups/{id} [get]
func (c *GroupController) GetGroup(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid ID")
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	group, err := c.manager.GetGroup(ctx.Request.Context(), envID, id)
	if err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}
	if group == nil {
		return nil, response.NewNotFoundError("Group not found")
	}

	return dto.ToGroupResponse(group), nil
}

// UpdateGroup godoc
// @Summary Update Notification Group
// @Description Update a notification group
// @Tags Notification
// @Accept json
// @Produce json
// @Param id path string true "Group ID"
// @Param request body dto.UpdateGroupRequest true "Group Data"
// @Success 200
// @Security BearerAuth
// @Router /notifications/groups/{id} [patch]
func (c *GroupController) UpdateGroup(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid ID")
	}

	var req dto.UpdateGroupRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	if err := c.manager.UpdateGroup(ctx.Request.Context(), envID, id, req.Name, req.Description); err != nil {
		if err == domain.ErrGroupNotFound {
			return nil, response.NewNotFoundError(err.Error())
		}
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "updated"}, nil
}

// DeleteGroup godoc
// @Summary Delete Notification Group
// @Description Delete a notification group
// @Tags Notification
// @Produce json
// @Param id path string true "Group ID"
// @Success 200
// @Security BearerAuth
// @Router /notifications/groups/{id} [delete]
func (c *GroupController) DeleteGroup(ctx *gin.Context) (interface{}, error) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		return nil, response.NewBadRequestError("Invalid ID")
	}

	envID := uuid.Nil
	if eID, ok := ctx.Get("environment_id"); ok {
		envID = eID.(uuid.UUID)
	}

	if err := c.manager.DeleteGroup(ctx.Request.Context(), envID, id); err != nil {
		return nil, response.NewInternalServerError(err.Error())
	}

	return map[string]string{"status": "deleted"}, nil
}
