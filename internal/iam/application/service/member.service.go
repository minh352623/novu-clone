package service

import (
	"context"

	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"

	"github.com/google/uuid"
)

// MemberService handles tenant membership operations
type MemberService interface {
	// AddMember adds a user to a tenant
	AddMember(ctx context.Context, tenantID, userID uuid.UUID, roleID *uuid.UUID) (*entity.TenantMember, error)

	// RemoveMember removes a user from a tenant
	RemoveMember(ctx context.Context, memberID uuid.UUID) error

	// GetMembers gets all members of a tenant
	GetMembers(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantMember, error)

	// AssignRole assigns a role to a member
	AssignRole(ctx context.Context, memberID, roleID uuid.UUID) error

	// CheckMembership checks if a user is a member of a tenant
	CheckMembership(ctx context.Context, tenantID, userID uuid.UUID) (*entity.TenantMember, error)

	// InviteMember sends an invitation to an email
	InviteMember(ctx context.Context, tenantID uuid.UUID, email string, roleID *uuid.UUID, invitedBy uuid.UUID) (*entity.TenantInvitation, error)

	// AcceptInvitation accepts an invitation
	AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (*entity.TenantMember, error)

	// GetInvitation retrieves an invitation by token
	GetInvitation(ctx context.Context, token string) (*entity.TenantInvitation, error)
}

// RoleService handles role operations
type RoleService interface {
	// CreateRole creates a new role
	CreateRole(ctx context.Context, name, slug string, permissions map[string]interface{}) (*entity.Role, error)

	// GetRole retrieves a role by ID
	GetRole(ctx context.Context, id uuid.UUID) (*entity.Role, error)

	// ListRoles lists all roles
	ListRoles(ctx context.Context, filters repository.RoleFilters) ([]*entity.Role, int64, error)

	// UpdateRole updates a role
	UpdateRole(ctx context.Context, role *entity.Role) error

	// DeleteRole deletes a role
	DeleteRole(ctx context.Context, id uuid.UUID) error

	// UpdatePermissions updates role permissions
	UpdatePermissions(ctx context.Context, roleID uuid.UUID, permissions map[string]interface{}) error
}
