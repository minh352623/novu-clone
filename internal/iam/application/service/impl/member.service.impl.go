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
	tenantRepo     repository.TenantRepository
	invitationRepo repository.InvitationRepository
	emailService   service.EmailService
}

// NewMemberService creates a new MemberService
func NewMemberService(
	memberRepo repository.TenantMemberRepository,
	roleRepo repository.RoleRepository,
	tenantRepo repository.TenantRepository,
	invitationRepo repository.InvitationRepository,
	emailService service.EmailService,
) service.MemberService {
	return &memberServiceImpl{
		memberRepo:     memberRepo,
		roleRepo:       roleRepo,
		tenantRepo:     tenantRepo,
		invitationRepo: invitationRepo,
		emailService:   emailService,
	}
}

func (s *memberServiceImpl) AddMember(ctx context.Context, tenantID, userID uuid.UUID, roleID, appID *uuid.UUID) (*entity.TenantMember, error) {
	// Check if already a member
	exists, err := s.memberRepo.ExistsByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if exists {
		return nil, entity.ErrMemberAlreadyExists
	}

	// Create member
	member, err := entity.NewTenantMember(tenantID, userID, roleID, appID)
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

func (s *memberServiceImpl) RevokeRole(ctx context.Context, memberID uuid.UUID) error {
	return s.memberRepo.RemoveRole(ctx, memberID)
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

func (s *memberServiceImpl) InviteMember(ctx context.Context, tenantID uuid.UUID, email string, roleID, appID *uuid.UUID, invitedBy uuid.UUID) (*entity.TenantInvitation, error) {
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
	if inviterRole.Slug != "tenant_admin" && inviterRole.Slug != "owner" {
		return nil, service.ErrUnauthorized
	}

	// 2. Check if pending invitation exists
	existingInvite, _ := s.invitationRepo.GetByEmailAndTenant(ctx, email, tenantID)
	if existingInvite != nil {
		return nil, fmt.Errorf("invitation already pending for this email")
	}

	// 3. Create Invitation
	invitation, err := entity.NewTenantInvitation(tenantID, email, roleID, appID, &invitedBy, 24*7*time.Hour)
	if err != nil {
		return nil, err
	}

	createdInvite, err := s.invitationRepo.Create(ctx, invitation)
	if err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	// 4. Send Email
	// Fetch Tenant Name
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	tenantName := "Converda Team"
	if err == nil && tenant != nil {
		tenantName = tenant.Name
	}

	// Fetch Role Name for email
	roleName := inviterRole.Name
	// Wait, we want the *invited* role name, not inviter's role.
	// roleID is passed arg.
	if roleID != nil {
		invitedRole, err := s.roleRepo.GetByID(ctx, *roleID)
		if err == nil && invitedRole != nil {
			roleName = invitedRole.Name
		} else {
			roleName = "Member"
		}
	} else {
		roleName = "Member"
	}

	go func() {
		// Use background context for sending email to avoid cancellation if request ends?
		// But we should probably use a separate context with timeout.
		// For now using todo/background context.
		// Ignoring error for async simplicity, but logging would be better.
		_ = s.emailService.SendInvitation(context.Background(), email, createdInvite.Token, roleName, tenantName)
	}()

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
	member, err := entity.NewTenantMember(invitation.TenantID, userID, invitation.RoleID, invitation.AppID)
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
