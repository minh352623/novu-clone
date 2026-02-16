package impl

import (
	"context"
	"fmt"

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
}

// NewTenantService creates a new TenantService
func NewTenantService(
	tenantRepo repository.TenantRepository,
	memberRepo repository.TenantMemberRepository,
	roleRepo repository.RoleRepository,
	planRepo repository.PricingPlanRepository,
) service.TenantService {
	return &tenantServiceImpl{
		tenantRepo: tenantRepo,
		memberRepo: memberRepo,
		roleRepo:   roleRepo,
		planRepo:   planRepo,
	}
}

func (s *tenantServiceImpl) CreateTenant(ctx context.Context, name, slug string, ownerUserID uuid.UUID) (*entity.Tenant, error) {
	// Check if slug exists
	existing, _ := s.tenantRepo.GetBySlug(ctx, slug)
	if existing != nil {
		return nil, service.ErrTenantSlugExists
	}

	// Create tenant
	tenant, err := entity.NewTenant(name, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Assign default plan
	defaultPlan, err := s.planRepo.GetDefault(ctx)
	if err == nil && defaultPlan != nil {
		tenant.AssignPlan(defaultPlan.ID)
	}

	// Persist tenant
	createdTenant, err := s.tenantRepo.Create(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to save tenant: %w", err)
	}

	// Get tenant_admin role
	adminRole, err := s.roleRepo.GetBySlug(ctx, entity.RoleTenantAdmin)
	var roleID *uuid.UUID
	if err == nil && adminRole != nil {
		roleID = &adminRole.ID
	} else {
		// Log error - this should not happen if migration 00016 ran
		fmt.Printf("Warning: Default admin role %s not found in database. Tenant owner will have no role.\n", entity.RoleTenantAdmin)
	}

	// Add owner as admin member
	member, err := entity.NewTenantMember(createdTenant.ID, ownerUserID, roleID, nil) // nil appID
	if err != nil {
		return nil, fmt.Errorf("failed to create member: %w", err)
	}
	if _, err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add owner as member: %w", err)
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
		return err
	}
	return s.tenantRepo.Update(ctx, tenant)
}

func (s *tenantServiceImpl) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	return s.tenantRepo.Delete(ctx, id)
}

func (s *tenantServiceImpl) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*entity.Tenant, error) {
	return s.tenantRepo.GetByUserID(ctx, userID)
}
