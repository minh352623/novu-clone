package service

import (
	"context"

	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"

	"github.com/google/uuid"
)

// PricingPlanService handles pricing plan operations
type PricingPlanService interface {
	// CreatePricingPlan creates a new pricing plan
	CreatePricingPlan(ctx context.Context, name, slug string, monthlyCredits int64, price float64, currency string) (*entity.PricingPlan, error)

	// GetPricingPlan retrieves a pricing plan by ID
	GetPricingPlan(ctx context.Context, id uuid.UUID) (*entity.PricingPlan, error)

	// ListPricingPlans lists all pricing plans
	ListPricingPlans(ctx context.Context, filters repository.PricingPlanFilters) ([]*entity.PricingPlan, int64, error)

	// UpdatePricingPlan updates a pricing plan
	UpdatePricingPlan(ctx context.Context, plan *entity.PricingPlan) error

	// DeletePricingPlan deletes a pricing plan
	DeletePricingPlan(ctx context.Context, id uuid.UUID) error

	// SetAsDefault sets a pricing plan as the default plan
	SetAsDefault(ctx context.Context, id uuid.UUID) error
}
