package dto

import (
	"time"

	"CONVERDA/internal/iam/domain/model/entity"

	"github.com/google/uuid"
)

// ================ Request DTOs ================

// PaginationParams represents query parameters for pagination
type PaginationParams struct {
	Page     int     `form:"page" binding:"omitempty,min=1"`
	PageSize int     `form:"page_size" binding:"omitempty,min=1,max=100"`
	Search   *string `form:"search"`
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email    string  `json:"email" binding:"required,email"`
	Password string  `json:"password" binding:"required,min=8"`
	FullName *string `json:"full_name,omitempty"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// CreateTenantRequest represents tenant creation request
type CreateTenantRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required,alphanum"`
}

// UpdateTenantRequest represents tenant update request
type UpdateTenantRequest struct {
	Name *string `json:"name,omitempty"`
}

// AddMemberRequest represents adding a member to a tenant
type AddMemberRequest struct {
	UserID uuid.UUID  `json:"user_id" binding:"required"`
	RoleID *uuid.UUID `json:"role_id,omitempty"`
	AppID  *uuid.UUID `json:"app_id,omitempty"`
}

// InviteMemberRequest represents inviting a member via email
type InviteMemberRequest struct {
	Email  string     `json:"email" binding:"required,email"`
	RoleID *uuid.UUID `json:"role_id,omitempty"`
	AppID  *uuid.UUID `json:"app_id,omitempty"`
}

// AcceptInvitationRequest represents accepting an invitation
type AcceptInvitationRequest struct {
	Token string `json:"token" binding:"required"`
}

// AssignRoleRequest represents role assignment request
type AssignRoleRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

// RevokeRoleRequest represents role revocation request
type RevokeRoleRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

// CreateRoleRequest represents role creation request
type CreateRoleRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Slug        string                 `json:"slug" binding:"required,alphanum"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
}

// UpdateRoleRequest represents role update request
type UpdateRoleRequest struct {
	Name        *string                `json:"name,omitempty"`
	Permissions map[string]interface{} `json:"permissions,omitempty"`
}

// CreatePricingPlanRequest represents pricing plan creation request
type CreatePricingPlanRequest struct {
	Name           string  `json:"name" binding:"required"`
	Slug           string  `json:"slug" binding:"required,alphanum"`
	MonthlyCredits int64   `json:"monthly_credits" binding:"min=0"`
	Price          float64 `json:"price" binding:"min=0"`
	Currency       string  `json:"currency" binding:"required,len=3"`
}

// UpdatePricingPlanRequest represents pricing plan update request
type UpdatePricingPlanRequest struct {
	Name           *string  `json:"name,omitempty"`
	MonthlyCredits *int64   `json:"monthly_credits,omitempty" binding:"omitempty,min=0"`
	Price          *float64 `json:"price,omitempty" binding:"omitempty,min=0"`
	Currency       *string  `json:"currency,omitempty" binding:"omitempty,len=3"`
	IsActive       *bool    `json:"is_active,omitempty"`
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ================ Response DTOs ================

// UserResponse represents user response
type UserResponse struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	FullName    *string    `json:"full_name,omitempty"`
	IsRootAdmin bool       `json:"is_root_admin"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ToUserResponse converts entity to response
func ToUserResponse(user *entity.User) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		IsRootAdmin: user.IsRootAdmin,
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
	}
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         *UserResponse `json:"user"`
	Token        string        `json:"token,omitempty"`         // Access Token
	RefreshToken string        `json:"refresh_token,omitempty"` // Refresh Token
}

// TenantResponse represents tenant response
type TenantResponse struct {
	ID            uuid.UUID            `json:"id"`
	Name          string               `json:"name"`
	Slug          string               `json:"slug"`
	Status        string               `json:"status"`
	PricingPlanID *uuid.UUID           `json:"pricing_plan_id,omitempty"`
	PricingPlan   *PricingPlanResponse `json:"pricing_plan,omitempty"`
	PlanStartDate time.Time            `json:"plan_start_date"`
	CreatedAt     time.Time            `json:"created_at"`
}

// ToTenantResponse converts entity to response
func ToTenantResponse(tenant *entity.Tenant) *TenantResponse {
	if tenant == nil {
		return nil
	}
	resp := &TenantResponse{
		ID:            tenant.ID,
		Name:          tenant.Name,
		Slug:          tenant.Slug,
		Status:        string(tenant.Status),
		PricingPlanID: tenant.PricingPlanID,
		PlanStartDate: tenant.PlanStartDate,
		CreatedAt:     tenant.CreatedAt,
	}
	if tenant.PricingPlan != nil {
		resp.PricingPlan = ToPricingPlanResponse(tenant.PricingPlan)
	}
	return resp
}

// ToTenantResponseList converts list of entities to responses
func ToTenantResponseList(tenants []*entity.Tenant) []*TenantResponse {
	result := make([]*TenantResponse, 0, len(tenants))
	for _, t := range tenants {
		result = append(result, ToTenantResponse(t))
	}
	return result
}

// PricingPlanResponse represents pricing plan response
type PricingPlanResponse struct {
	ID                  uuid.UUID `json:"id"`
	Name                string    `json:"name"`
	Slug                string    `json:"slug"`
	MonthlyCredits      int64     `json:"monthly_credits"`
	Price               float64   `json:"price"`
	Currency            string    `json:"currency"`
	IsActive            bool      `json:"is_active"`
	IsDefault           bool      `json:"is_default"`
	MaxApps             int       `json:"max_apps"`
	MaxMembers          int       `json:"max_members"`
	MaxWorkflows        int       `json:"max_workflows"`
	MaxMessagesPerMonth int64     `json:"max_messages_per_month"`
	RateLimitRPM        int       `json:"rate_limit_rpm"`
}

// ToPricingPlanResponse converts entity to response
func ToPricingPlanResponse(plan *entity.PricingPlan) *PricingPlanResponse {
	if plan == nil {
		return nil
	}

	return &PricingPlanResponse{
		ID:                  plan.ID,
		Name:                plan.Name,
		Slug:                plan.Slug,
		MonthlyCredits:      plan.MonthlyCredits,
		Price:               plan.Price,
		Currency:            plan.Currency,
		IsActive:            plan.IsActive,
		IsDefault:           plan.IsDefault,
		MaxApps:             plan.MaxApps,
		MaxMembers:          plan.MaxMembers,
		MaxWorkflows:        plan.MaxWorkflows,
		MaxMessagesPerMonth: plan.MaxMessagesPerMonth,
		RateLimitRPM:        plan.RateLimitRPM,
	}
}

// ToPricingPlanResponseList converts list of entities to responses
func ToPricingPlanResponseList(plans []*entity.PricingPlan) []*PricingPlanResponse {
	result := make([]*PricingPlanResponse, 0, len(plans))
	for _, p := range plans {
		result = append(result, ToPricingPlanResponse(p))
	}
	return result
}

// RoleResponse represents role response
type RoleResponse struct {
	ID          uuid.UUID              `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Permissions map[string]interface{} `json:"permissions"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ToRoleResponse converts entity to response
func ToRoleResponse(role *entity.Role) *RoleResponse {
	if role == nil {
		return nil
	}
	return &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Slug:        role.Slug,
		Permissions: role.Permissions,
		CreatedAt:   role.CreatedAt,
	}
}

// ToRoleResponseList converts list of entities to responses
func ToRoleResponseList(roles []*entity.Role) []*RoleResponse {
	result := make([]*RoleResponse, 0, len(roles))
	for _, r := range roles {
		result = append(result, ToRoleResponse(r))
	}
	return result
}

// TenantMemberResponse represents tenant member response
type TenantMemberResponse struct {
	ID        uuid.UUID     `json:"id"`
	TenantID  uuid.UUID     `json:"tenant_id"`
	UserID    uuid.UUID     `json:"user_id"`
	RoleID    *uuid.UUID    `json:"role_id,omitempty"`
	AppID     *uuid.UUID    `json:"app_id,omitempty"`
	User      *UserResponse `json:"user,omitempty"`
	Role      *RoleResponse `json:"role,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
}

// TenantInvitationResponse represents tenant invitation response
type TenantInvitationResponse struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	Email     string     `json:"email"`
	RoleID    *uuid.UUID `json:"role_id,omitempty"`
	AppID     *uuid.UUID `json:"app_id,omitempty"`
	Status    string     `json:"status"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// ToTenantInvitationResponse converts entity to response
func ToTenantInvitationResponse(inv *entity.TenantInvitation) *TenantInvitationResponse {
	if inv == nil {
		return nil
	}
	return &TenantInvitationResponse{
		ID:        inv.ID,
		TenantID:  inv.TenantID,
		Email:     inv.Email,
		RoleID:    inv.RoleID,
		AppID:     inv.AppID,
		Status:    string(inv.Status),
		ExpiresAt: inv.ExpiresAt,
		CreatedAt: inv.CreatedAt,
	}
}

// ToTenantMemberResponse converts entity to response
func ToTenantMemberResponse(member *entity.TenantMember) *TenantMemberResponse {
	if member == nil {
		return nil
	}
	resp := &TenantMemberResponse{
		ID:        member.ID,
		TenantID:  member.TenantID,
		UserID:    member.UserID,
		RoleID:    member.RoleID,
		AppID:     member.AppID,
		CreatedAt: member.CreatedAt,
	}
	if member.User != nil {
		resp.User = ToUserResponse(member.User)
	}
	if member.Role != nil {
		resp.Role = ToRoleResponse(member.Role)
	}
	return resp
}

// ToTenantMemberResponseList converts list of entities to responses
func ToTenantMemberResponseList(members []*entity.TenantMember) []*TenantMemberResponse {
	result := make([]*TenantMemberResponse, 0, len(members))
	for _, m := range members {
		result = append(result, ToTenantMemberResponse(m))
	}
	return result
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// NewPaginatedResponse creates a new paginated response
func NewPaginatedResponse(data interface{}, total int64, page, pageSize int) *PaginatedResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return &PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
