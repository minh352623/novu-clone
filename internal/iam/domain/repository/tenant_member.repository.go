package repository

import (
	"context"

	"CONVERDA/internal/iam/domain/model/entity"

	"github.com/google/uuid"
)

// TenantMemberFilters defines filter options for querying tenant members
type TenantMemberFilters struct {
	TenantID *uuid.UUID
	UserID   *uuid.UUID
	RoleID   *uuid.UUID
	Limit    int
	Offset   int
}

// TenantMemberRepository defines the interface for tenant member persistence
type TenantMemberRepository interface {
	// Create creates a new tenant member
	Create(ctx context.Context, member *entity.TenantMember) (*entity.TenantMember, error)

	// GetByID retrieves a member by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.TenantMember, error)

	// GetByTenantAndUser retrieves a member by tenant and user ID
	GetByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (*entity.TenantMember, error)

	// GetAll retrieves all members with filters
	GetAll(ctx context.Context, filters TenantMemberFilters) ([]*entity.TenantMember, error)

	// CountAll counts all members with filters
	CountAll(ctx context.Context, filters TenantMemberFilters) (int64, error)

	// GetByTenantID retrieves all members of a tenant
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantMember, error)

	// GetByUserID retrieves all memberships for a user
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.TenantMember, error)

	// Update updates a member
	Update(ctx context.Context, member *entity.TenantMember) error

	// Delete deletes a member
	Delete(ctx context.Context, id uuid.UUID) error

	// AssignRole assigns a role to a member
	AssignRole(ctx context.Context, memberID, roleID uuid.UUID) error

	// RemoveRole removes the role from a member
	RemoveRole(ctx context.Context, memberID uuid.UUID) error

	// ExistsByTenantAndUser checks if a membership exists
	ExistsByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (bool, error)

	// GetByTenantUserAndApp checks for specific app-level membership (or tenant-level if appID is nil)
	GetByTenantUserAndApp(ctx context.Context, tenantID, userID uuid.UUID, appID *uuid.UUID) (*entity.TenantMember, error)
}
