package impl

import (
	"context"
	"fmt"

	"CONVERDA/global"
	"CONVERDA/internal/iam/application/service"
	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"

	"github.com/google/uuid"
)

// tenantServiceImpl implements TenantService
type tenantServiceImpl struct {
	tenantRepo repository.TenantRepository
	memberRepo repository.TenantMemberRepository
	roleRepo   repository.RoleRepository
	planRepo   repository.PricingPlanRepository
	uow        repository.IAMUnitOfWork
}

// NewTenantService creates a new TenantService
func NewTenantService(
	tenantRepo repository.TenantRepository,
	memberRepo repository.TenantMemberRepository,
	roleRepo repository.RoleRepository,
	planRepo repository.PricingPlanRepository,
	uow repository.IAMUnitOfWork,
) service.TenantService {
	return &tenantServiceImpl{
		tenantRepo: tenantRepo,
		memberRepo: memberRepo,
		roleRepo:   roleRepo,
		planRepo:   planRepo,
		uow:        uow,
	}
}

func (s *tenantServiceImpl) CreateTenant(ctx context.Context, name, slug string, ownerUserID uuid.UUID) (*entity.Tenant, error) {
	var createdTenant *entity.Tenant

	err := s.uow.Execute(ctx, func(tx repository.IAMTxRepository) error {
		// 1. Check if slug exists
		existing, err := tx.Tenants().GetBySlug(ctx, slug)
		if err != nil {
			return fmt.Errorf("failed to check existing tenant slug: %w", err)
		}
		if existing != nil {
			return service.ErrTenantSlugExists
		}

		// 2. Create tenant entity
		tenant, err := entity.NewTenant(name, slug)
		if err != nil {
			return fmt.Errorf("failed to create tenant: %w", err)
		}

		// 3. Assign default plan (Read-only, can be done before or within tx)
		dp, err := s.planRepo.GetDefault(ctx)
		if err == nil && dp != nil {
			tenant.AssignPlan(dp.ID)
		}

		// 4. Persist tenant
		createdTenant, err = tx.Tenants().Create(ctx, tenant)
		if err != nil {
			return fmt.Errorf("failed to save tenant: %w", err)
		}

		// 5. Get tenant_admin role
		adminRole, err := tx.Roles().GetBySlug(ctx, entity.RoleTenantAdmin)
		var roleID *uuid.UUID
		if err == nil && adminRole != nil {
			roleID = &adminRole.ID
		} else {
			global.Logger.Warn("CreateTenant: default admin role not found in database", "roleSlug", string(entity.RoleTenantAdmin))
		}

		// 6. Add owner as admin member
		member, err := entity.NewTenantMember(createdTenant.ID, ownerUserID, roleID, nil)
		if err != nil {
			return fmt.Errorf("failed to create member: %w", err)
		}
		if _, err := tx.Members().Create(ctx, member); err != nil {
			return fmt.Errorf("failed to add owner as member: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdTenant, nil
}

func (s *tenantServiceImpl) GetTenant(ctx context.Context, id uuid.UUID) (*entity.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	if tenant == nil {
		return nil, service.ErrTenantNotFound
	}
	return tenant, nil
}

func (s *tenantServiceImpl) GetTenantBySlug(ctx context.Context, slug string) (*entity.Tenant, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	if tenant == nil {
		return nil, service.ErrTenantNotFound
	}
	return tenant, nil
}

func (s *tenantServiceImpl) ListTenants(ctx context.Context, filters repository.TenantFilters) ([]*entity.Tenant, int64, error) {
	tenants, err := s.tenantRepo.GetAll(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}

	count, err := s.tenantRepo.CountAll(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	return tenants, count, nil
}

func (s *tenantServiceImpl) UpdateTenant(ctx context.Context, tenant *entity.Tenant) error {
	if err := tenant.Validate(); err != nil {
		return fmt.Errorf("invalid tenant: %w", err)
	}
	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return fmt.Errorf("failed to update tenant %s: %w", tenant.ID, err)
	}
	return nil
}

func (s *tenantServiceImpl) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	if err := s.tenantRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tenant %s: %w", id, err)
	}
	return nil
}

func (s *tenantServiceImpl) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*entity.Tenant, error) {
	tenants, err := s.tenantRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants for user %s: %w", userID, err)
	}
	return tenants, nil
}
