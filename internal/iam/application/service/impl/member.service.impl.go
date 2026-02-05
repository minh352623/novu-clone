package impl

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/internal/iam/application/service"
	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"

	"github.com/google/uuid"
)

// memberServiceImpl implements MemberService
type memberServiceImpl struct {
	memberRepo     repository.TenantMemberRepository
	roleRepo       repository.RoleRepository
	invitationRepo repository.InvitationRepository
}

// NewMemberService creates a new MemberService
func NewMemberService(
	memberRepo repository.TenantMemberRepository,
	roleRepo repository.RoleRepository,
	invitationRepo repository.InvitationRepository,
) service.MemberService {
	return &memberServiceImpl{
		memberRepo:     memberRepo,
		roleRepo:       roleRepo,
		invitationRepo: invitationRepo,
	}
}

func (s *memberServiceImpl) AddMember(ctx context.Context, tenantID, userID uuid.UUID, roleID *uuid.UUID) (*entity.TenantMember, error) {
	// Check if already a member
	exists, err := s.memberRepo.ExistsByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if exists {
		return nil, entity.ErrMemberAlreadyExists
	}

	// Create member
	member, err := entity.NewTenantMember(tenantID, userID, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to create member: %w", err)
	}

	// Persist
	createdMember, err := s.memberRepo.Create(ctx, member)
	if err != nil {
		return nil, fmt.Errorf("failed to save member: %w", err)
	}

	return createdMember, nil
}

func (s *memberServiceImpl) RemoveMember(ctx context.Context, memberID uuid.UUID) error {
	return s.memberRepo.Delete(ctx, memberID)
}

func (s *memberServiceImpl) GetMembers(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantMember, error) {
	return s.memberRepo.GetByTenantID(ctx, tenantID)
}

func (s *memberServiceImpl) AssignRole(ctx context.Context, memberID, roleID uuid.UUID) error {
	// Verify role exists
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return service.ErrRoleNotFound
	}

	return s.memberRepo.AssignRole(ctx, memberID, roleID)
}

func (s *memberServiceImpl) CheckMembership(ctx context.Context, tenantID, userID uuid.UUID) (*entity.TenantMember, error) {
	member, err := s.memberRepo.GetByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if member == nil {
		return nil, service.ErrMemberNotFound
	}
	return member, nil
}

func (s *memberServiceImpl) InviteMember(ctx context.Context, tenantID uuid.UUID, email string, roleID *uuid.UUID, invitedBy uuid.UUID) (*entity.TenantInvitation, error) {
	// 1. Verify inviter is a member and has permission (Admin)
	inviter, err := s.memberRepo.GetByTenantAndUser(ctx, tenantID, invitedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to verification inviter: %w", err)
	}
	if inviter == nil || inviter.RoleID == nil {
		return nil, service.ErrUnauthorized
	}

	// Fetch Role to check slug
	inviterRole, err := s.roleRepo.GetByID(ctx, *inviter.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	// Assumption: Only TenantAdmin or Owner can invite. Or check permission "tenant.invite".
	// For simplicity, checking if role is 'tenant_admin' or 'owner'.
	if inviterRole.Slug != "tenant_admin" && inviterRole.Slug != "owner" {
		// Or check permissions map
		return nil, service.ErrUnauthorized
	}

	// 2. Check if user already exists in tenant? (Optional, but good UX)
	// We can't check by email easily against members without user table join or extra query.
	// We'll skip for now or rely on "Create member" uniqueness constraint later.

	// 3. Check if pending invitation exists
	existingInvite, _ := s.invitationRepo.GetByEmailAndTenant(ctx, email, tenantID)
	if existingInvite != nil {
		return nil, fmt.Errorf("invitation already pending for this email")
	}

	// 4. Create Invitation
	// 7 days expiration
	invitation, err := entity.NewTenantInvitation(tenantID, email, roleID, &invitedBy, 24*7*time.Hour)
	if err != nil {
		return nil, err
	}

	createdInvite, err := s.invitationRepo.Create(ctx, invitation)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// 5. Send Email (Mock)
	fmt.Printf("MOCK MAIL SEND: Invitation to %s with token %s\n", email, createdInvite.Token)

	return createdInvite, nil
}

func (s *memberServiceImpl) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (*entity.TenantMember, error) {
	// 1. Get Invitation
	invitation, err := s.invitationRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, entity.ErrInvalidToken
	}

	// 2. Validate
	if err := invitation.Accept(); err != nil {
		return nil, err
	}

	// 3. Check if user is already member
	exists, err := s.memberRepo.ExistsByTenantAndUser(ctx, invitation.TenantID, userID)
	if exists {
		// Already member, just close invitation
		_ = s.invitationRepo.Update(ctx, invitation)
		return nil, entity.ErrMemberAlreadyExists
	}

	// 4. Create Member
	member, err := entity.NewTenantMember(invitation.TenantID, userID, invitation.RoleID)
	if err != nil {
		return nil, err
	}

	createdMember, err := s.memberRepo.Create(ctx, member)
	if err != nil {
		return nil, fmt.Errorf("failed to create member: %w", err)
	}

	// 5. Update Invitation Status
	if err := s.invitationRepo.Update(ctx, invitation); err != nil {
		// Log error? Transaction would be better.
	}

	return createdMember, nil
}

func (s *memberServiceImpl) GetInvitation(ctx context.Context, token string) (*entity.TenantInvitation, error) {
	return s.invitationRepo.GetByToken(ctx, token)
}
