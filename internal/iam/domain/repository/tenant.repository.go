package repository

import (
	"context"

	"CONVERDA/internal/iam/domain/model/entity"

	"github.com/google/uuid"
)

// TenantFilters defines filter options for querying tenants
type TenantFilters struct {
	PricingPlanID *uuid.UUID
	Search        *string
	Limit         int
	Offset        int
}

// TenantReader defines the interface for reading tenant data
type TenantReader interface {
	// GetByID retrieves a tenant by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tenant, error)

	// GetBySlug retrieves a tenant by slug
	GetBySlug(ctx context.Context, slug string) (*entity.Tenant, error)

	// GetAll retrieves all tenants with filters
	GetAll(ctx context.Context, filters TenantFilters) ([]*entity.Tenant, error)

	// CountAll counts all tenants with filters
	CountAll(ctx context.Context, filters TenantFilters) (int64, error)

	// GetByUserID retrieves all tenants that a user is a member of
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Tenant, error)
}

// TenantWriter defines the interface for writing tenant data
type TenantWriter interface {
	// Create creates a new tenant
	Create(ctx context.Context, tenant *entity.Tenant) (*entity.Tenant, error)

	// Update updates a tenant
	Update(ctx context.Context, tenant *entity.Tenant) error

	// Delete deletes a tenant
	Delete(ctx context.Context, id uuid.UUID) error

	// AssignPricingPlan assigns a pricing plan to a tenant
	AssignPricingPlan(ctx context.Context, tenantID, planID uuid.UUID) error
}

// TenantRepository defines the complete interface for tenant persistence
type TenantRepository interface {
	TenantReader
	TenantWriter
}
