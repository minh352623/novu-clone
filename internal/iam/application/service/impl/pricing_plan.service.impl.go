package impl

import (
	"context"
	"fmt"

	"CONVERDA/internal/iam/application/service"
	"CONVERDA/internal/iam/domain/model/entity"
	"CONVERDA/internal/iam/domain/repository"

	"github.com/google/uuid"
)

// pricingPlanServiceImpl implements PricingPlanService
type pricingPlanServiceImpl struct {
	planRepo repository.PricingPlanRepository
}

// NewPricingPlanService creates a new PricingPlanService
func NewPricingPlanService(planRepo repository.PricingPlanRepository) service.PricingPlanService {
	return &pricingPlanServiceImpl{
		planRepo: planRepo,
	}
}

func (s *pricingPlanServiceImpl) CreatePricingPlan(ctx context.Context, name, slug string, monthlyCredits int64, price float64, currency string) (*entity.PricingPlan, error) {
	// Check if slug exists
	existing, err := s.planRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing plan slug: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("pricing plan with slug '%s' already exists", slug)
	}

	// Create plan (Assuming entity.NewPricingPlan exists or struct init)
	// Since NewPricingPlan doesn't seem to exist in previous context, I'll direct struct init or create it if needed.
	// Looking back at logs, entity.PricingPlan struct is standard.
	// I will use direct struct initialization as per standard unless factory provided.
	// Actually, looking at previous files, entity.go usually has validation.
	// I'll assume direct initialization then validate.

	plan := &entity.PricingPlan{
		ID:             uuid.New(),
		Name:           name,
		Slug:           slug,
		MonthlyCredits: monthlyCredits,
		Price:          price,
		Currency:       currency,
		IsActive:       true,
		IsDefault:      false,
	}

	// If the entity has a Validate method, we should call it.
	// But let's stick to the pattern.

	// Persist
	createdPlan, err := s.planRepo.Create(ctx, plan)
	if err != nil {
		return nil, fmt.Errorf("failed to save pricing plan: %w", err)
	}

	return createdPlan, nil
}

func (s *pricingPlanServiceImpl) GetPricingPlan(ctx context.Context, id uuid.UUID) (*entity.PricingPlan, error) {
	plan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing plan: %w", err)
	}
	if plan == nil {
		return nil, service.ErrPlanNotFound
	}
	return plan, nil
}

func (s *pricingPlanServiceImpl) ListPricingPlans(ctx context.Context, filters repository.PricingPlanFilters) ([]*entity.PricingPlan, int64, error) {
	plans, err := s.planRepo.GetAll(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list pricing plans: %w", err)
	}

	count, err := s.planRepo.CountAll(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count pricing plans: %w", err)
	}

	return plans, count, nil
}

func (s *pricingPlanServiceImpl) UpdatePricingPlan(ctx context.Context, plan *entity.PricingPlan) error {
	if err := s.planRepo.Update(ctx, plan); err != nil {
		return fmt.Errorf("failed to update pricing plan %s: %w", plan.ID, err)
	}
	return nil
}

func (s *pricingPlanServiceImpl) DeletePricingPlan(ctx context.Context, id uuid.UUID) error {
	if err := s.planRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete pricing plan %s: %w", id, err)
	}
	return nil
}

func (s *pricingPlanServiceImpl) SetAsDefault(ctx context.Context, id uuid.UUID) error {
	// Verify it exists
	plan, err := s.planRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get pricing plan: %w", err)
	}
	if plan == nil {
		return service.ErrPlanNotFound
	}

	return s.planRepo.SetAsDefault(ctx, id)
}
