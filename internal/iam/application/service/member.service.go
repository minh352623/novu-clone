package service

import (
	"context"

	"CONVERDA/internal/iam/domain/model/entity"

	"github.com/google/uuid"
)

// MemberService handles tenant membership operations
type MemberService interface {
	// AddMember adds a user to a tenant
	AddMember(ctx context.Context, tenantID, userID uuid.UUID, roleID, appID *uuid.UUID) (*entity.TenantMember, error)

	// RemoveMember removes a user from a tenant
	RemoveMember(ctx context.Context, memberID uuid.UUID) error

	// GetMembers gets all members of a tenant
	GetMembers(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantMember, error)

	// AssignRole assigns a role to a member
	AssignRole(ctx context.Context, memberID, roleID uuid.UUID) error

	// RevokeRole revokes a role from a member
	RevokeRole(ctx context.Context, memberID uuid.UUID) error

	// CheckMembership checks if a user is a member of a tenant
	CheckMembership(ctx context.Context, tenantID, userID uuid.UUID) (*entity.TenantMember, error)

	// InviteMember sends an invitation to an email
	InviteMember(ctx context.Context, tenantID uuid.UUID, email string, roleID, appID *uuid.UUID, invitedBy uuid.UUID) (*entity.TenantInvitation, error)

	// AcceptInvitation accepts an invitation
	AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (*entity.TenantMember, error)

	// GetInvitation retrieves an invitation by token
	GetInvitation(ctx context.Context, token string) (*entity.TenantInvitation, error)
}
