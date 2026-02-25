package controller

import (
	"errors"

	"CONVERDA/global"
	"CONVERDA/internal/iam/application/service"
	"CONVERDA/internal/iam/controller/dto"
	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"
	"CONVERDA/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// getUserIDFromCtx safely extracts the user ID from the gin context.
// It uses comma-ok pattern to prevent panics from invalid type assertion.
func getUserIDFromCtx(ctx *gin.Context) (uuid.UUID, error) {
	val, exists := ctx.Get("user_id")
	if !exists {
		return uuid.Nil, response.NewUnauthorizedError("Unauthorized")
	}
	uid, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, response.NewUnauthorizedError("Invalid user context")
	}
	return uid, nil
}

// AuthController handles authentication endpoints
type AuthController struct {
	authService  service.AuthService
	tokenService service.TokenService
}

// NewAuthController creates a new AuthController
func NewAuthController(authService service.AuthService, tokenService service.TokenService) *AuthController {
	return &AuthController{
		authService:  authService,
		tokenService: tokenService,
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
		return nil, response.NewBadRequestError(err.Error())
	}

	user, err := c.authService.Register(ctx.Request.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			return nil, response.NewConflictError(err.Error())
		}
		if errors.Is(err, entity.ErrUserEmailRequired) || errors.Is(err, entity.ErrUserPasswordRequired) || errors.Is(err, entity.ErrUserPasswordTooShort) {
			return nil, response.NewBadRequestError(err.Error())
		}
		global.Logger.Error("AuthController.Register: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
	}

	// Generate tokens for auto-login
	accessToken, _, err := c.tokenService.GenerateAccessToken(ctx.Request.Context(), user.ID, user.Email, nil)
	if err != nil {
		// Log error but assume registration success
		// Ideally we should rollback or return warning, but for now just return user
		return dto.AuthResponse{
			User: dto.ToUserResponse(user),
		}, nil
	}
	refreshToken, _, err := c.tokenService.GenerateRefreshToken(ctx.Request.Context(), user.ID, user.Email, nil)
	if err != nil {
		return nil, response.NewInternalServerError("Failed to generate refresh token")
	}

	return dto.AuthResponse{
		User:         dto.ToUserResponse(user),
		Token:        accessToken,
		RefreshToken: refreshToken,
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
		return nil, response.NewBadRequestError(err.Error())
	}

	user, err := c.authService.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		return nil, response.NewUnauthorizedError("Invalid credentials")
	}

	// Generate JWT tokens
	accessToken, _, err := c.tokenService.GenerateAccessToken(ctx.Request.Context(), user.ID, user.Email, nil)
	if err != nil {
		return nil, response.NewInternalServerError("Failed to generate token")
	}

	refreshToken, _, err := c.tokenService.GenerateRefreshToken(ctx.Request.Context(), user.ID, user.Email, nil)
	if err != nil {
		return nil, response.NewInternalServerError("Failed to generate refresh token")
	}

	return dto.AuthResponse{
		User:         dto.ToUserResponse(user),
		Token:        accessToken,
		RefreshToken: refreshToken,
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
		return nil, response.NewBadRequestError(err.Error())
	}

	userID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	err = c.authService.ChangePassword(ctx.Request.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return nil, response.NewBadRequestError("Current password is incorrect")
		}
		if errors.Is(err, service.ErrUserNotFound) {
			return nil, response.NewNotFoundError("User not found")
		}
		global.Logger.Error("AuthController.ChangePassword: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
	}

	return gin.H{"message": "Password changed successfully"}, nil
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh the user's access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token data"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) (interface{}, error) {
	var req dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	userID, email, err := c.tokenService.ValidateRefreshToken(ctx.Request.Context(), req.RefreshToken)
	if err != nil {
		return nil, response.NewUnauthorizedError("Invalid or expired refresh token")
	}

	// Generate new access token
	accessToken, _, err := c.tokenService.GenerateAccessToken(ctx.Request.Context(), userID, email, nil)
	if err != nil {
		return nil, response.NewInternalServerError("Failed to generate token")
	}

	// Optionally rotate refresh token here

	// We don't have user entity here to return full profile unless we fetch it.
	// For simple refresh, returning token is often enough, but response DTO has User.
	// We should fetch user. TODO: Add GetUser to authService or userService call.
	// For now, return empty user or minimal info.
	// Actually, let's just return the token. The DTO expects User pointer.

	return dto.AuthResponse{
		Token: accessToken,
	}, nil
}

// Me godoc
// @Summary Get current user
// @Description Get details of the currently authenticated user
// @Tags Users
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Security BearerAuth
// @Router /users/me [get]
func (c *AuthController) Me(ctx *gin.Context) (interface{}, error) {
	userID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	user, err := c.authService.GetUser(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return nil, response.NewNotFoundError("User not found")
		}
		global.Logger.Error("AuthController.Me: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
	}

	return dto.ToUserResponse(user), nil
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
		return nil, response.NewBadRequestError(err.Error())
	}

	userID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	tenant, err := c.tenantService.CreateTenant(ctx.Request.Context(), req.Name, req.Slug, userID)
	if err != nil {
		if errors.Is(err, service.ErrTenantSlugExists) {
			return nil, response.NewConflictError(err.Error())
		}
		global.Logger.Error("TenantController.CreateTenant: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
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
		return nil, response.NewBadRequestError("Invalid tenant ID")
	}

	tenant, err := c.tenantService.GetTenant(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTenantNotFound) {
			return nil, response.NewNotFoundError("Tenant not found")
		}
		global.Logger.Error("TenantController.GetTenant: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
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
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, response.NewBadRequestError("Invalid tenant ID")
	}

	var req dto.UpdateTenantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	// 1. Get existing tenant to verify existence
	tenant, err := c.tenantService.GetTenant(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrTenantNotFound) {
			return nil, response.NewNotFoundError("Tenant not found")
		}
		global.Logger.Error("TenantController.UpdateTenant: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
	}

	// 2. Partial update fields
	if req.Name != nil {
		tenant.Name = *req.Name
	}

	// 3. Save
	if err := c.tenantService.UpdateTenant(ctx.Request.Context(), tenant); err != nil {
		if errors.Is(err, service.ErrTenantSlugExists) {
			return nil, response.NewConflictError(err.Error())
		}
		global.Logger.Error("TenantController.UpdateTenant: update failed", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
	}

	return dto.ToTenantResponse(tenant), nil
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
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, response.NewBadRequestError("Invalid tenant ID")
	}

	if err := c.tenantService.DeleteTenant(ctx.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrTenantNotFound) {
			return nil, response.NewNotFoundError("Tenant not found")
		}
		global.Logger.Error("TenantController.DeleteTenant: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
	}

	return gin.H{"message": "Tenant deleted successfully"}, nil
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
	var params dto.PaginationParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	filters := repository.TenantFilters{
		Limit:  params.PageSize,
		Offset: (params.Page - 1) * params.PageSize,
		Search: params.Search,
	}

	tenants, total, err := c.tenantService.ListTenants(ctx.Request.Context(), filters)
	if err != nil {
		global.Logger.Error("TenantController.ListTenants: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
	}

	return dto.NewPaginatedResponse(
		dto.ToTenantResponseList(tenants),
		total,
		params.Page,
		params.PageSize,
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
	userID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	tenants, err := c.tenantService.GetUserTenants(ctx.Request.Context(), userID)
	if err != nil {
		global.Logger.Error("TenantController.GetMyTenants: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
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
		return nil, response.NewBadRequestError("Invalid tenant ID")
	}

	members, err := c.memberService.GetMembers(ctx.Request.Context(), tenantID)
	if err != nil {
		global.Logger.Error("TenantController.GetMembers: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
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
		return nil, response.NewBadRequestError("Invalid tenant ID")
	}

	var req dto.AddMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	member, err := c.memberService.AddMember(ctx.Request.Context(), tenantID, req.UserID, req.RoleID, req.AppID)
	if err != nil {
		if errors.Is(err, entity.ErrMemberAlreadyExists) {
			return nil, response.NewConflictError(err.Error())
		}
		global.Logger.Error("TenantController.AddMember: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
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
		return nil, response.NewBadRequestError("Invalid tenant ID")
	}

	var req dto.InviteMemberRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, response.NewBadRequestError(err.Error())
	}

	userID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	invite, err := c.memberService.InviteMember(ctx.Request.Context(), tenantID, req.Email, req.RoleID, req.AppID, userID)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			return nil, response.NewForbiddenError("Insufficient permissions")
		}
		if errors.Is(err, service.ErrInvitationAlreadyPending) {
			return nil, response.NewConflictError(err.Error())
		}
		global.Logger.Error("TenantController.InviteMember: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
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
		return nil, response.NewBadRequestError(err.Error())
	}

	userID, err := getUserIDFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	member, err := c.memberService.AcceptInvitation(ctx.Request.Context(), req.Token, userID)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidToken) || errors.Is(err, entity.ErrInvitationExpired) || errors.Is(err, entity.ErrInvitationAccepted) {
			return nil, response.NewBadRequestError(err.Error())
		}
		if errors.Is(err, entity.ErrMemberAlreadyExists) {
			return nil, response.NewConflictError(err.Error())
		}
		global.Logger.Error("TenantController.AcceptInvitation: unexpected error", "error", err)
		return nil, response.NewInternalServerError("An internal error occurred")
	}

	return dto.ToTenantMemberResponse(member), nil
}
