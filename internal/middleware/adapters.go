package middleware

import (
	"context"
	"fmt"

	"CONVERDA/internal/iam/domain/repository"

	"github.com/google/uuid"
)

// MembershipCheckerAdapter adapts TenantMemberRepository to MembershipChecker interface
type MembershipCheckerAdapter struct {
	memberRepo repository.TenantMemberRepository
}

// NewMembershipCheckerAdapter creates a new adapter
func NewMembershipCheckerAdapter(memberRepo repository.TenantMemberRepository) *MembershipCheckerAdapter {
	return &MembershipCheckerAdapter{
		memberRepo: memberRepo,
	}
}

// GetMemberByTenantAndUser implements MembershipChecker interface
func (a *MembershipCheckerAdapter) GetMemberByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (memberID uuid.UUID, roleID *uuid.UUID, err error) {
	member, err := a.memberRepo.GetByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil {
		return uuid.Nil, nil, fmt.Errorf("user is not a member of this tenant")
	}
	return member.ID, member.RoleID, nil
}

// PermissionCheckerAdapter adapts RoleRepository to PermissionChecker interface
type PermissionCheckerAdapter struct {
	roleRepo repository.RoleRepository
}

// NewPermissionCheckerAdapter creates a new adapter
func NewPermissionCheckerAdapter(roleRepo repository.RoleRepository) *PermissionCheckerAdapter {
	return &PermissionCheckerAdapter{
		roleRepo: roleRepo,
	}
}

// GetRolePermissions implements PermissionChecker interface
func (a *PermissionCheckerAdapter) GetRolePermissions(ctx context.Context, roleID uuid.UUID) (map[string]interface{}, error) {
	role, err := a.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	return role.Permissions, nil
}
