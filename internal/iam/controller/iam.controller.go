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

// AuthController handles authentication endpoints
type AuthController struct {
	authService service.AuthService
}

// NewAuthController creates a new AuthController
func NewAuthController(authService service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /auth/register [post]

func (c *AuthController) Register(ctx *gin.Context) (interface{}, error) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	user, err := c.authService.Register(ctx.Request.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			return nil, response.NewAPIError(http.StatusConflict, err.Error(), err)
		}
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	return dto.AuthResponse{
		User: dto.ToUserResponse(user),
	}, nil
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) (interface{}, error) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	user, err := c.authService.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Invalid credentials", err)
	}

	// TODO: Generate JWT token here
	return dto.AuthResponse{
		User:  dto.ToUserResponse(user),
		Token: "", // JWT token to be implemented
	}, nil
}

// ChangePassword godoc
// @Summary Change password
// @Description Change the current user's password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "Password change data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /auth/change-password [post]
func (c *AuthController) ChangePassword(ctx *gin.Context) (interface{}, error) {
	var req dto.ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := ctx.Get("user_id")
	if !exists {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", nil)
	}

	err := c.authService.ChangePassword(ctx.Request.Context(), userID.(uuid.UUID), req.OldPassword, req.NewPassword)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return nil, response.NewAPIError(http.StatusBadRequest, "Current password is incorrect", err)
		}
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	return gin.H{"message": "Password changed successfully"}, nil
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh the user's access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Refresh token data"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) (interface{}, error) {
	// TODO: Implement RefreshToken logic using token service
	return nil, response.NewAPIError(http.StatusNotImplemented, "Not implemented", nil)
}

// Me godoc
// @Summary Get current user
// @Description Get details of the currently authenticated user
// @Tags Users
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} map[string]string
// @Security BearerAuth
// @Router /users/me [get]
func (c *AuthController) Me(ctx *gin.Context) (interface{}, error) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", nil)
	}
	// TODO: Fetch user details from service
	return gin.H{"id": userID}, nil
}

// TenantController handles tenant endpoints
type TenantController struct {
	tenantService service.TenantService
	memberService service.MemberService
}

// NewTenantController creates a new TenantController
func NewTenantController(tenantService service.TenantService, memberService service.MemberService) *TenantController {
	return &TenantController{
		tenantService: tenantService,
		memberService: memberService,
	}
}

// CreateTenant godoc
// @Summary Create a new tenant
// @Description Create a new tenant/organization
// @Tags Tenants
// @Accept json
// @Produce json
// @Param request body dto.CreateTenantRequest true "Tenant data"
// @Success 201 {object} dto.TenantResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Security BearerAuth
// @Router /tenants [post]
func (c *TenantController) CreateTenant(ctx *gin.Context) (interface{}, error) {
	var req dto.CreateTenantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	// Get user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", nil)
	}

	tenant, err := c.tenantService.CreateTenant(ctx.Request.Context(), req.Name, req.Slug, userID.(uuid.UUID))
	if err != nil {
		if err == service.ErrTenantSlugExists {
			return nil, response.NewAPIError(http.StatusConflict, err.Error(), err)
		}
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	return dto.ToTenantResponse(tenant), nil
}

// GetTenant godoc
// @Summary Get tenant by ID
// @Description Get a specific tenant by its ID
// @Tags Tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} dto.TenantResponse
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /tenants/{id} [get]
func (c *TenantController) GetTenant(ctx *gin.Context) (interface{}, error) {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid tenant ID", err)
	}

	tenant, err := c.tenantService.GetTenant(ctx.Request.Context(), id)
	if err != nil {
		if err == service.ErrTenantNotFound {
			return nil, response.NewAPIError(http.StatusNotFound, err.Error(), err)
		}
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToTenantResponse(tenant), nil
}

// UpdateTenant godoc
// @Summary Update tenant
// @Description Update an existing tenant
// @Tags Tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param request body dto.CreateTenantRequest true "Update data"
// @Success 200 {object} dto.TenantResponse
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /tenants/{id} [put]
func (c *TenantController) UpdateTenant(ctx *gin.Context) (interface{}, error) {
	idStr := ctx.Param("id")
	_, err := uuid.Parse(idStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid tenant ID", err)
	}
	// TODO: Implement UpdateTenant logic in service and controller
	return nil, response.NewAPIError(http.StatusNotImplemented, "Not implemented", nil)
}

// DeleteTenant godoc
// @Summary Delete tenant
// @Description Delete a tenant
// @Tags Tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /tenants/{id} [delete]
func (c *TenantController) DeleteTenant(ctx *gin.Context) (interface{}, error) {
	idStr := ctx.Param("id")
	_, err := uuid.Parse(idStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid tenant ID", err)
	}
	// TODO: Implement DeleteTenant logic
	return nil, response.NewAPIError(http.StatusNotImplemented, "Not implemented", nil)
}

// ListTenants godoc
// @Summary List all tenants
// @Description Get a paginated list of all tenants
// @Tags Tenants
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param search query string false "Search term"
// @Success 200 {object} dto.PaginatedResponse
// @Security BearerAuth
// @Router /tenants [get]
func (c *TenantController) ListTenants(ctx *gin.Context) (interface{}, error) {
	page := 1
	pageSize := 20

	filters := repository.TenantFilters{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	tenants, total, err := c.tenantService.ListTenants(ctx.Request.Context(), filters)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.NewPaginatedResponse(
		dto.ToTenantResponseList(tenants),
		total,
		page,
		pageSize,
	), nil
}

// GetMyTenants godoc
// @Summary Get current user's tenants
// @Description Get all tenants the current user is a member of
// @Tags Tenants
// @Produce json
// @Success 200 {array} dto.TenantResponse
// @Security BearerAuth
// @Router /tenants/me [get]
func (c *TenantController) GetMyTenants(ctx *gin.Context) (interface{}, error) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", nil)
	}

	tenants, err := c.tenantService.GetUserTenants(ctx.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToTenantResponseList(tenants), nil
}

// GetMembers godoc
// @Summary Get tenant members
// @Description Get all members of a tenant
// @Tags Tenants
// @Produce json
// @Param id path string true "Tenant ID"
// @Success 200 {array} dto.TenantMemberResponse
// @Security BearerAuth
// @Router /tenants/{id}/members [get]
func (c *TenantController) GetMembers(ctx *gin.Context) (interface{}, error) {
	idStr := ctx.Param("id")
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid tenant ID", err)
	}

	members, err := c.memberService.GetMembers(ctx.Request.Context(), tenantID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToTenantMemberResponseList(members), nil
}

// AddMember godoc
// @Summary Add member to tenant
// @Description Add a user as a member of a tenant
// @Tags Tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param request body dto.AddMemberRequest true "Member data"
// @Success 201 {object} dto.TenantMemberResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Security BearerAuth
// @Router /tenants/{id}/members [post]
func (c *TenantController) AddMember(ctx *gin.Context) (interface{}, error) {
	idStr := ctx.Param("id")
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid tenant ID", err)
	}

	var req dto.AddMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	member, err := c.memberService.AddMember(ctx.Request.Context(), tenantID, req.UserID, req.RoleID)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	return dto.ToTenantMemberResponse(member), nil
}

// InviteMember godoc
// @Summary Invite member to tenant
// @Description Invite a user via email to join a tenant
// @Tags Tenants
// @Accept json
// @Produce json
// @Param id path string true "Tenant ID"
// @Param request body dto.InviteMemberRequest true "Invitation data"
// @Success 201 {object} dto.TenantInvitationResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Security BearerAuth
// @Router /tenants/{id}/invitations [post]
func (c *TenantController) InviteMember(ctx *gin.Context) (interface{}, error) {
	idStr := ctx.Param("id")
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, "Invalid tenant ID", err)
	}

	var req dto.InviteMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", nil)
	}

	invite, err := c.memberService.InviteMember(ctx.Request.Context(), tenantID, req.Email, req.RoleID, userID.(uuid.UUID))
	if err != nil {
		if err == service.ErrUnauthorized {
			return nil, response.NewAPIError(http.StatusForbidden, "Insufficient permissions", err)
		}
		return nil, response.NewAPIError(http.StatusInternalServerError, err.Error(), err)
	}

	return dto.ToTenantInvitationResponse(invite), nil
}

// AcceptInvitation godoc
// @Summary Accept invitation
// @Description Accept an invitation using a token
// @Tags Invitations
// @Accept json
// @Produce json
// @Param request body dto.AcceptInvitationRequest true "Token data"
// @Success 200 {object} dto.TenantMemberResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /invitations/accept [post]
func (c *TenantController) AcceptInvitation(ctx *gin.Context) (interface{}, error) {
	var req dto.AcceptInvitationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	userID, exists := ctx.Get("user_id")
	if !exists {
		return nil, response.NewAPIError(http.StatusUnauthorized, "Unauthorized", nil)
	}

	member, err := c.memberService.AcceptInvitation(ctx.Request.Context(), req.Token, userID.(uuid.UUID))
	if err != nil {
		// Handle errors like Expired, InvalidToken etc.
		return nil, response.NewAPIError(http.StatusBadRequest, err.Error(), err)
	}

	return dto.ToTenantMemberResponse(member), nil
}
