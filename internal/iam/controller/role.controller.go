package controller

import (
	"net/http"

	"CONVERDA/internal/iam/application/service"
	"CONVERDA/internal/iam/controller/dto"
	"CONVERDA/internal/iam/domain/repository"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RoleController handles role endpoints
type RoleController struct {
	roleService service.RoleService
}

// NewRoleController creates a new RoleController
func NewRoleController(roleService service.RoleService) *RoleController {
	return &RoleController{
		roleService: roleService,
	}
}

// CreateRole godoc
// @Summary Create a new role
// @Description Create a new role with permissions
// @Tags Roles
// @Accept json
// @Produce json
// @Param request body dto.CreateRoleRequest true "Role data"
// @Success 201 {object} dto.RoleResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /roles [post]
func (c *RoleController) CreateRole(ctx *gin.Context) (interface{}, error) {
	var req dto.CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	role, err := c.roleService.CreateRole(ctx.Request.Context(), req.Name, req.Slug, req.Permissions)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	return dto.ToRoleResponse(role), nil
}

// GetRole godoc
// @Summary Get role by ID
// @Description Get a specific role by its ID
// @Tags Roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 200 {object} dto.RoleResponse
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /roles/{id} [get]
func (c *RoleController) GetRole(ctx *gin.Context) (interface{}, error) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid role ID", err)
	}

	role, err := c.roleService.GetRole(ctx.Request.Context(), id)
	if err != nil {
		if err == service.ErrRoleNotFound {
			return nil, response.NewAPIError(http.StatusNotFound, err.Error(), err)
		}
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToRoleResponse(role), nil
}

// ListRoles godoc
// @Summary List all roles
// @Description Get a paginated list of all roles
// @Tags Roles
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param search query string false "Search term"
// @Success 200 {object} dto.PaginatedResponse
// @Security BearerAuth
// @Router /roles [get]
func (c *RoleController) ListRoles(ctx *gin.Context) (interface{}, error) {
	page := 1
	pageSize := 20
	// TODO: parse page/pageSize

	filters := repository.RoleFilters{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	roles, total, err := c.roleService.ListRoles(ctx.Request.Context(), filters)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.NewPaginatedResponse(
		dto.ToRoleResponseList(roles),
		total,
		page,
		pageSize,
	), nil
}

// UpdateRole godoc
// @Summary Update a role
// @Description Update an existing role
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body dto.UpdateRoleRequest true "Role update data"
// @Success 200 {object} dto.RoleResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /roles/{id} [put]
func (c *RoleController) UpdateRole(ctx *gin.Context) (interface{}, error) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid role ID", err)
	}

	var req dto.UpdateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	// Get existing role (verification)
	role, err := c.roleService.GetRole(ctx.Request.Context(), id)
	if err != nil {
		if err == service.ErrRoleNotFound {
			return nil, response.NewAPIError(http.StatusNotFound, err.Error(), err)
		}
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	// Update fields
	if req.Name != nil {
		role.Name = *req.Name
	}
	if req.Permissions != nil {
		role.Permissions = req.Permissions
	}

	if err := c.roleService.UpdateRole(ctx.Request.Context(), role); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	return dto.ToRoleResponse(role), nil
}

// DeleteRole godoc
// @Summary Delete a role
// @Description Delete an existing role
// @Tags Roles
// @Produce json
// @Param id path string true "Role ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /roles/{id} [delete]
func (c *RoleController) DeleteRole(ctx *gin.Context) (interface{}, error) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid role ID", err)
	}

	if err := c.roleService.DeleteRole(ctx.Request.Context(), id); err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return nil, nil // 204 No Content handled by wrapper if returning nil data? Wrapper uses 200 usually.
	// For 204, maybe wrapper needs custom code or we return nil and wrapper sets 204 if code=204?
	// Handler func signature wraps with 'code int'.
	// So we will pass http.StatusNoContent to Wrap.
}

// UpdatePermissions godoc
// @Summary Update role permissions
// @Description Update the permissions of a role
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body map[string]interface{} true "Permissions map"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /roles/{id}/permissions [put]
func (c *RoleController) UpdatePermissions(ctx *gin.Context) (interface{}, error) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid role ID", err)
	}

	var permissions map[string]interface{}
	if err := ctx.ShouldBindJSON(&permissions); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	if err := c.roleService.UpdatePermissions(ctx.Request.Context(), id, permissions); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	return gin.H{"message": "Permissions updated successfully"}, nil
}

// AssignRole godoc
// @Summary Assign role to user
// @Description Assign a role to a specific user
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body dto.AssignRoleRequest true "User data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /roles/{id}/assign [post]
func (c *RoleController) AssignRole(ctx *gin.Context) (interface{}, error) {
	// TODO: Implement logic to assign role to user
	return nil, response.NewAPIError(http.StatusNotImplemented, "Not implemented", nil)
}

// RevokeRole godoc
// @Summary Revoke role from user
// @Description Revoke a role from a specific user
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "Role ID"
// @Param request body dto.RevokeRoleRequest true "User data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /roles/{id}/revoke [post]
func (c *RoleController) RevokeRole(ctx *gin.Context) (interface{}, error) {
	// TODO: Implement logic to revoke role from user
	return nil, response.NewAPIError(http.StatusNotImplemented, "Not implemented", nil)
}
