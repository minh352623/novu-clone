package impl

import (
	"context"
	"fmt"
	"time"

	"CONVERDA/global"
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
	userRepo       repository.UserRepository
	emailService   service.EmailService
	uow            repository.IAMUnitOfWork
}

// NewMemberService creates a new MemberService
func NewMemberService(
	memberRepo repository.TenantMemberRepository,
	roleRepo repository.RoleRepository,
	tenantRepo repository.TenantRepository,
	invitationRepo repository.InvitationRepository,
	userRepo repository.UserRepository,
	emailService service.EmailService,
	uow repository.IAMUnitOfWork,
) service.MemberService {
	return &memberServiceImpl{
		memberRepo:     memberRepo,
		roleRepo:       roleRepo,
		tenantRepo:     tenantRepo,
		invitationRepo: invitationRepo,
		userRepo:       userRepo,
		emailService:   emailService,
		uow:            uow,
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
	if err := s.memberRepo.Delete(ctx, memberID); err != nil {
		return fmt.Errorf("failed to delete member %s: %w", memberID, err)
	}
	return nil
}

func (s *memberServiceImpl) GetMembers(ctx context.Context, tenantID uuid.UUID) ([]*entity.TenantMember, error) {
	members, err := s.memberRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get members for tenant %s: %w", tenantID, err)
	}
	return members, nil
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

	if err := s.memberRepo.AssignRole(ctx, memberID, roleID); err != nil {
		return fmt.Errorf("failed to assign role %s to member %s: %w", roleID, memberID, err)
	}
	return nil
}

func (s *memberServiceImpl) RevokeRole(ctx context.Context, memberID uuid.UUID) error {
	if err := s.memberRepo.RemoveRole(ctx, memberID); err != nil {
		return fmt.Errorf("failed to revoke role for member %s: %w", memberID, err)
	}
	return nil
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

	// Resolve inviter display name
	inviterName := "A team member"
	inviterUser, err := s.userRepo.GetByID(ctx, invitedBy)
	if err == nil && inviterUser != nil {
		if inviterUser.FullName != nil && *inviterUser.FullName != "" {
			inviterName = *inviterUser.FullName
		} else {
			inviterName = inviterUser.Email
		}
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				global.Logger.Error("member_service: panic recovered in SendInvitation goroutine", "panic", r)
			}
		}()
		// Use context.WithoutCancel to ensure email delivery continues even if the original request context is cancelled
		// but still carries over tracing/correlation context if present.
		asyncCtx := context.WithoutCancel(ctx)
		if err := s.emailService.SendInvitation(asyncCtx, email, createdInvite.Token, roleName, tenantName, inviterName); err != nil {
			global.Logger.Warn("member_service: failed to send invitation email", "email", email, "error", err)
		}
	}()

	return createdInvite, nil
}

func (s *memberServiceImpl) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (*entity.TenantMember, error) {
	var createdMember *entity.TenantMember

	err := s.uow.Execute(ctx, func(tx repository.IAMTxRepository) error {
		// 1. Get Invitation
		invitation, err := tx.Invitations().GetByToken(ctx, token)
		if err != nil {
			return fmt.Errorf("failed to get invitation by token: %w", err)
		}
		if invitation == nil {
			return entity.ErrInvalidToken
		}

		// 2. Validate
		if err := invitation.Accept(); err != nil {
			return err
		}

		// 3. Check if user is already member
		exists, err := tx.Members().ExistsByTenantAndUser(ctx, invitation.TenantID, userID)
		if err != nil {
			return fmt.Errorf("failed to check membership existence: %w", err)
		}
		if exists {
			// Already member, just close invitation
			_ = tx.Invitations().Update(ctx, invitation)
			return entity.ErrMemberAlreadyExists
		}

		// 4. Create Member
		member, err := entity.NewTenantMember(invitation.TenantID, userID, invitation.RoleID, invitation.AppID)
		if err != nil {
			return err
		}

		createdMember, err = tx.Members().Create(ctx, member)
		if err != nil {
			return fmt.Errorf("failed to create member: %w", err)
		}

		// 5. Update Invitation Status
		if err := tx.Invitations().Update(ctx, invitation); err != nil {
			return fmt.Errorf("failed to update invitation status: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdMember, nil
}

func (s *memberServiceImpl) GetInvitation(ctx context.Context, token string) (*entity.TenantInvitation, error) {
	invite, err := s.invitationRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}
	return invite, nil
}
