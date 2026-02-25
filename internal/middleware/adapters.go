package middleware

import (
	"context"
	"fmt"

	"CONVERDA/internal/iam/domain/repository"

	"github.com/google/uuid"
)

// MembershipCheckerAdapter adapts TenantMemberRepository to MembershipChecker interface
type MembershipCheckerAdapter struct {
	memberRepo   repository.TenantMemberRepository
	tenantReader repository.TenantReader
}

// NewMembershipCheckerAdapter creates a new adapter
func NewMembershipCheckerAdapter(memberRepo repository.TenantMemberRepository, tenantReader repository.TenantReader) *MembershipCheckerAdapter {
	return &MembershipCheckerAdapter{
		memberRepo:   memberRepo,
		tenantReader: tenantReader,
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

// GetMemberByTenantUserAndApp implements MembershipChecker interface
func (a *MembershipCheckerAdapter) GetMemberByTenantUserAndApp(ctx context.Context, tenantID, userID uuid.UUID, appID *uuid.UUID) (memberID uuid.UUID, roleID *uuid.UUID, err error) {
	// 1. Try App-Specific Membership
	member, err := a.memberRepo.GetByTenantUserAndApp(ctx, tenantID, userID, appID)
	if err != nil {
		return uuid.Nil, nil, fmt.Errorf("failed to check app membership: %w", err)
	}
	if member != nil {
		return member.ID, member.RoleID, nil
	}

	// 2. Fallback to Tenant-Level (AppID = NIL)
	// Only if AppID was requested (not nil). If AppID was already nil, first query covered it.
	if appID != nil {
		tenantMember, err := a.memberRepo.GetByTenantUserAndApp(ctx, tenantID, userID, nil)
		if err != nil {
			return uuid.Nil, nil, fmt.Errorf("failed to check tenant fallback membership: %w", err)
		}
		if tenantMember != nil {
			return tenantMember.ID, tenantMember.RoleID, nil
		}
	}

	return uuid.Nil, nil, fmt.Errorf("user is not a member")
}

// GetTenantStatus implements MembershipChecker interface
func (a *MembershipCheckerAdapter) GetTenantStatus(ctx context.Context, tenantID uuid.UUID) (string, error) {
	tenant, err := a.tenantReader.GetByID(ctx, tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant: %w", err)
	}
	if tenant == nil {
		return "", fmt.Errorf("tenant not found")
	}
	return string(tenant.Status), nil
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
