package repository

import (
	"context"

	"CONVERDA/internal/iam/domain/model/entity"

	"github.com/google/uuid"
)

// PricingPlanFilters defines filter options for querying pricing plans
type PricingPlanFilters struct {
	IsActive  *bool
	IsDefault *bool
	Search    *string
	Limit     int
	Offset    int
}

// PricingPlanRepository defines the interface for pricing plan persistence
type PricingPlanRepository interface {
	// Create creates a new pricing plan
	Create(ctx context.Context, plan *entity.PricingPlan) (*entity.PricingPlan, error)

	// GetByID retrieves a plan by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PricingPlan, error)

	// GetBySlug retrieves a plan by slug
	GetBySlug(ctx context.Context, slug string) (*entity.PricingPlan, error)

	// GetDefault retrieves the default plan
	GetDefault(ctx context.Context) (*entity.PricingPlan, error)

	// GetAll retrieves all plans with filters
	GetAll(ctx context.Context, filters PricingPlanFilters) ([]*entity.PricingPlan, error)

	// CountAll counts all plans with filters
	CountAll(ctx context.Context, filters PricingPlanFilters) (int64, error)

	// Update updates a plan
	Update(ctx context.Context, plan *entity.PricingPlan) error

	// Delete deletes a plan
	Delete(ctx context.Context, id uuid.UUID) error

	// SetAsDefault sets a plan as default (and unsets others)
	SetAsDefault(ctx context.Context, id uuid.UUID) error
}
