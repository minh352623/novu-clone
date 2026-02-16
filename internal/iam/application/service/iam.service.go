package service

import (
	"context"
	"errors"

	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"

	"github.com/google/uuid"
)

// Service errors
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrTenantNotFound     = errors.New("tenant not found")
	ErrTenantSlugExists   = errors.New("tenant with this slug already exists")
	ErrRoleNotFound       = errors.New("role not found")
	ErrMemberNotFound     = errors.New("tenant member not found")
	ErrPlanNotFound       = errors.New("pricing plan not found")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized access")
)

// AuthService handles authentication operations
type AuthService interface {
	// Register creates a new user account
	Register(ctx context.Context, email, password string, fullName *string) (*entity.User, error)

	// Login authenticates a user and returns the user entity
	Login(ctx context.Context, email, password string) (*entity.User, error)

	// ChangePassword changes a user's password
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, userID uuid.UUID) (*entity.User, error)
}

// TenantService handles tenant operations
type TenantService interface {
	// CreateTenant creates a new tenant and adds the user as admin
	CreateTenant(ctx context.Context, name, slug string, ownerUserID uuid.UUID) (*entity.Tenant, error)

	// GetTenant retrieves a tenant by ID
	GetTenant(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)

	// GetTenantBySlug retrieves a tenant by slug
	GetTenantBySlug(ctx context.Context, slug string) (*entity.Tenant, error)

	// ListTenants lists all tenants with pagination
	ListTenants(ctx context.Context, filters repository.TenantFilters) ([]*entity.Tenant, int64, error)

	// UpdateTenant updates a tenant
	UpdateTenant(ctx context.Context, tenant *entity.Tenant) error

	// DeleteTenant deletes a tenant
	DeleteTenant(ctx context.Context, id uuid.UUID) error

	// GetUserTenants gets all tenants a user belongs to
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*entity.Tenant, error)
}

// RoleService handles role operations
type RoleService interface {
	CreateRole(ctx context.Context, name, slug string, permissions map[string]interface{}) (*entity.Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (*entity.Role, error)
	ListRoles(ctx context.Context, filters repository.RoleFilters) ([]*entity.Role, int64, error)
	UpdateRole(ctx context.Context, role *entity.Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
	UpdatePermissions(ctx context.Context, id uuid.UUID, permissions map[string]interface{}) error
}
